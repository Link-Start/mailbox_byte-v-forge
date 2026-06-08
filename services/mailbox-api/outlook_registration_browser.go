package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
)

func (r *outlookRegistrationRunner) startSession(ctx context.Context, email string) (string, error) {
	if r.browserClient == nil {
		return "", errors.New("browser-automation client is not initialized")
	}
	reqCtx, cancel := context.WithTimeout(ctx, r.cfg.commandTimeout)
	defer cancel()
	resp, err := r.browserClient.StartBrowserSession(reqCtx, &browserautomationv1.StartBrowserSessionRequest{
		RequestId: "outlook-oauth-" + uuid.NewString(),
		Profile: &browserautomationv1.BrowserProfile{
			BrowserKind: browserautomationv1.BrowserKind_BROWSER_KIND_FIREFOX,
			Locale:      r.cfg.locale,
			Timezone:    r.cfg.timezone,
			UserAgent:   r.cfg.userAgent,
			Viewport: &browserautomationv1.BrowserViewport{
				Width:  int32(r.cfg.windowWidth),
				Height: int32(r.cfg.windowHeight),
			},
			ProxyRef: r.cfg.proxyRef,
			ExtraHttpHeaders: map[string]string{
				"Accept-Language": r.cfg.acceptLanguage,
			},
		},
		Labels: map[string]string{
			"domain":     "mailbox",
			"provider":   "outlook",
			"workflow":   "oauth",
			"email_hash": hashLabel(email),
		},
		Ttl: durationpb.New(r.cfg.sessionTTL),
	})
	if err != nil {
		return "", err
	}
	if resp.GetError() != nil {
		return "", errors.New(safeMailboxText(resp.GetError().GetMessage()))
	}
	sessionID := resp.GetSession().GetSessionId()
	if sessionID == "" {
		return "", errors.New("browser-automation returned empty session_id")
	}
	return sessionID, nil
}

func (r *outlookRegistrationRunner) stopSession(sessionID string) {
	if sessionID == "" || r.browserClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, _ = r.browserClient.StopBrowserSession(ctx, &browserautomationv1.StopBrowserSessionRequest{
		SessionId: sessionID,
		Reason:    "outlook oauth finished",
	})
}

func (r *outlookRegistrationRunner) execute(ctx context.Context, sessionID string, taskKey string, commands []*browserautomationv1.BrowserCommand) ([]*browserautomationv1.BrowserCommandResult, error) {
	timeout := r.cfg.commandTimeout + 15*time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	resp, err := r.browserClient.ExecuteBrowserCommands(reqCtx, &browserautomationv1.ExecuteBrowserCommandsRequest{
		RequestId: "outlook-oauth-" + uuid.NewString(),
		Input: &browserautomationv1.BrowserTaskInput{
			SessionId:   sessionID,
			TaskKey:     taskKey,
			ScenarioKey: "outlook.oauth",
			Timeout:     durationpb.New(timeout),
			Commands:    commands,
			Labels: map[string]string{
				"domain":   "mailbox",
				"provider": "outlook",
				"workflow": "oauth",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.GetError() != nil {
		return resp.GetResults(), errors.New(safeMailboxText(resp.GetError().GetMessage()))
	}
	for _, result := range resp.GetResults() {
		if result.GetStatus() == browserautomationv1.BrowserCommandStatus_BROWSER_COMMAND_STATUS_FAILED ||
			result.GetStatus() == browserautomationv1.BrowserCommandStatus_BROWSER_COMMAND_STATUS_TIMEOUT {
			if result.GetError() != nil {
				return resp.GetResults(), errors.New(safeMailboxText(result.GetError().GetMessage()))
			}
			return resp.GetResults(), fmt.Errorf("browser command %s failed", result.GetCommandKey())
		}
	}
	return resp.GetResults(), nil
}
