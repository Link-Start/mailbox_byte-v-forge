package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/natseventbus"
)

type mailboxWorkerConsumerConfig struct {
	label      string
	definition eventcatalog.Definition
	batch      int
	ackWait    time.Duration
}

func startMailboxEventWorkers(ctx context.Context, group *errgroup.Group, cfg config, bus *natseventbus.Bus, repo mailboxRepository, service *EmailService, operations operationStore, activities *mailboxActivities) error {
	if bus == nil {
		return nil
	}
	pollConsumer, err := mailboxWorkerConsumer(bus, cfg, mailboxWorkerConsumerConfig{label: "mailbox email poll", definition: eventcatalog.MailboxEmailPollRequested, batch: 10, ackWait: 60 * time.Second})
	if err != nil {
		return err
	}
	fetchConsumer, err := mailboxWorkerConsumer(bus, cfg, mailboxWorkerConsumerConfig{label: "mailbox inbox fetch", definition: mailboxInboxFetchRequested, batch: 5, ackWait: 5 * time.Minute})
	if err != nil {
		return err
	}
	registrationConsumer, err := mailboxWorkerConsumer(bus, cfg, mailboxWorkerConsumerConfig{label: "mailbox registration", definition: mailboxRegistrationRequested, batch: 2, ackWait: 5 * time.Minute})
	if err != nil {
		return err
	}
	oauthConsumer, err := mailboxWorkerConsumer(bus, cfg, mailboxWorkerConsumerConfig{label: "mailbox OAuth", definition: mailboxOAuthRequested, batch: 2, ackWait: 5 * time.Minute})
	if err != nil {
		return err
	}
	events := newMailboxEvents(bus)
	group.Go(func() error { return repo.RunOutboxWorker(ctx, mailboxEventOutboxTable, events, logWarning) })
	group.Go(func() error { return runMailboxEmailPollWorker(ctx, pollConsumer, service) })
	group.Go(func() error { return runMailboxInboxFetchWorker(ctx, fetchConsumer, service, operations) })
	group.Go(func() error { return runMailboxRegistrationWorker(ctx, registrationConsumer, operations, activities) })
	group.Go(func() error { return runMailboxOAuthWorker(ctx, oauthConsumer, operations, activities) })
	return nil
}

func mailboxWorkerConsumer(bus *natseventbus.Bus, cfg config, worker mailboxWorkerConsumerConfig) (eventbus.Consumer, error) {
	consumer, err := bus.PullWorkerForDefinition(cfg.eventStreamName, worker.definition, worker.batch, worker.ackWait)
	if err != nil {
		return nil, fmt.Errorf("initialize %s worker: %w", worker.label, err)
	}
	return consumer, nil
}
