package main

import (
	"context"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventoutbox"
	"mailboxapi/internal/inboxapp"
)

func (d *mailboxWorkDispatcher) metadata(eventName string, subject string, eventID string, correlationID string) *commonv1.EventMetadata {
	return eventbus.NewEventMetadata(eventbus.EventMetadataConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  inboxapp.EventVersion,
		SourceService: d.source,
		Subject:       subject,
		CorrelationID: correlationID,
	})
}

func (d *mailboxWorkDispatcher) enqueue(ctx context.Context, record eventoutbox.Record) error {
	tx, err := d.beginner.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	if err := eventoutbox.InsertRecordPgx(ctx, tx, mailboxEventOutboxTable, record, time.Now().Unix()); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
