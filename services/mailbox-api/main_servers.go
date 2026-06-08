package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"mailboxapi/internal/grpcclient"
	"mailboxapi/internal/grpchealth"
	"mailboxapi/internal/hotstream"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/redisx"

	"mailboxapi/pb"
)

type mailboxServerRuntime struct {
	emailBackend    *EmailService
	operations      operationStore
	activities      *mailboxActivities
	providers       mailboxProviderRuntimeConfig
	hot             *mailboxHotStream
	work            *mailboxWorkDispatcher
	inbox           *inboxapp.Service
	watcher         *MailWatcher
	inboxLock       *redisx.BestEffortLocker
	dashboardEvents hotstream.Subscriber
}

func startMailboxServers(ctx context.Context, group *errgroup.Group, cfg config, runtime mailboxServerRuntime) error {
	errCh := make(chan error, 3)
	startWebhookServer(ctx, cfg.webhookHTTPAddr, cfg.webhook, runtime.providers, runtime.inbox, runtime.watcher, runtime.inboxLock, errCh)
	if err := startMailboxGRPCServer(ctx, group, cfg.listenAddr, runtime); err != nil {
		return err
	}
	if err := startMailboxDashboard(ctx, group, cfg, runtime.dashboardEvents, errCh); err != nil {
		return err
	}
	group.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			return err
		}
	})
	return nil
}

func startMailboxGRPCServer(ctx context.Context, group *errgroup.Group, listenAddr string, runtime mailboxServerRuntime) error {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen mailbox API on %s: %w", listenAddr, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMailboxServiceServer(grpcServer, &server{
		emailBackend: runtime.emailBackend,
		operations:   runtime.operations,
		providers:    runtime.providers,
		hot:          runtime.hot,
		work:         runtime.work,
		activities:   runtime.activities,
	})
	grpchealth.RegisterServing(grpcServer)
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()
	log.Printf("mailbox API listening on %s", listenAddr)
	group.Go(func() error {
		if err := grpcServer.Serve(listener); err != nil {
			return fmt.Errorf("mailbox API failed: %w", err)
		}
		return nil
	})
	return nil
}

func startMailboxDashboard(ctx context.Context, group *errgroup.Group, cfg config, events hotstream.Subscriber, errCh chan<- error) error {
	if strings.TrimSpace(cfg.dashboardHTTPAddr) == "" {
		return nil
	}
	dashboardConn, err := grpcclient.NewInsecure(grpcclient.SelfTarget(cfg.listenAddr))
	if err != nil {
		return fmt.Errorf("connect mailbox dashboard API: %w", err)
	}
	group.Go(func() error {
		<-ctx.Done()
		_ = dashboardConn.Close()
		return nil
	})
	startDashboardHTTP(ctx, cfg.dashboardHTTPAddr, cfg.dashboardStaticDir, cfg.dashboard, pb.NewMailboxServiceClient(dashboardConn), events, errCh)
	return nil
}
