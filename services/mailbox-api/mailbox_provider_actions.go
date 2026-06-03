package main

import (
	"context"
	"fmt"

	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/pb"
)

type mailboxRegistrationRunner interface {
	RunMailboxRegistration(context.Context, *pb.RunMailboxRegistrationRequest) (*pb.RunMailboxRegistrationResponse, error)
}

type mailboxOAuthRunner interface {
	RunMailboxOAuth(context.Context, *pb.RunMailboxOAuthRequest) (*pb.RunMailboxOAuthResponse, error)
}

type mailboxProviderActionRegistry struct {
	defaultProvider string
	registration    map[string]mailboxRegistrationRunner
	oauth           map[string]mailboxOAuthRunner
}

func newMailboxProviderActionRegistry(defaultProvider string) *mailboxProviderActionRegistry {
	return &mailboxProviderActionRegistry{
		defaultProvider: mailboxprovider.NormalizeKey(defaultProvider),
		registration:    map[string]mailboxRegistrationRunner{},
		oauth:           map[string]mailboxOAuthRunner{},
	}
}

func (r *mailboxProviderActionRegistry) RegisterRegistration(provider string, runner mailboxRegistrationRunner) {
	if r == nil || runner == nil {
		return
	}
	if key := mailboxprovider.NormalizeKey(provider); key != "" {
		r.registration[key] = runner
	}
}

func (r *mailboxProviderActionRegistry) RegisterOAuth(provider string, runner mailboxOAuthRunner) {
	if r == nil || runner == nil {
		return
	}
	if key := mailboxprovider.NormalizeKey(provider); key != "" {
		r.oauth[key] = runner
	}
}

func (r *mailboxProviderActionRegistry) RunRegistration(ctx context.Context, provider string, req *pb.RunMailboxRegistrationRequest) (*pb.RunMailboxRegistrationResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("mailbox provider action registry is required")
	}
	key := r.providerKey(provider)
	runner := r.registration[key]
	if runner == nil {
		return nil, fmt.Errorf("mailbox provider cannot run registration: %s", key)
	}
	return runner.RunMailboxRegistration(ctx, req)
}

func (r *mailboxProviderActionRegistry) RunOAuth(ctx context.Context, provider string, req *pb.RunMailboxOAuthRequest) (*pb.RunMailboxOAuthResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("mailbox provider action registry is required")
	}
	key := r.providerKey(provider)
	runner := r.oauth[key]
	if runner == nil {
		return nil, fmt.Errorf("mailbox provider cannot run OAuth: %s", key)
	}
	return runner.RunMailboxOAuth(ctx, req)
}

func (r *mailboxProviderActionRegistry) providerKey(provider string) string {
	key := mailboxprovider.NormalizeKey(provider)
	if key == "" && r != nil {
		key = r.defaultProvider
	}
	return key
}
