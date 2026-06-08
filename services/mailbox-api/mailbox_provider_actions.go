package main

import (
	"context"
	"fmt"

	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/pb"
)

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
