package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
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

func navigateCommand(commandID, targetURL string, timeout time.Duration) *browserautomationv1.BrowserCommand {
	return &browserautomationv1.BrowserCommand{
		CommandId:  commandID,
		CommandKey: commandID,
		Timeout:    durationpb.New(timeout),
		Operation: &browserautomationv1.BrowserCommand_Navigate{
			Navigate: &browserautomationv1.NavigateCommand{
				Url:       targetURL,
				WaitUntil: browserautomationv1.BrowserNavigationWaitUntil_BROWSER_NAVIGATION_WAIT_UNTIL_DOM_CONTENT_LOADED,
				Timeout:   durationpb.New(timeout),
			},
		},
	}
}

func evaluateCommand(commandID, expression string, args map[string]any, timeout time.Duration) *browserautomationv1.BrowserCommand {
	structArgs, err := structpb.NewStruct(args)
	if err != nil {
		structArgs = &structpb.Struct{}
	}
	return &browserautomationv1.BrowserCommand{
		CommandId:  commandID,
		CommandKey: commandID,
		Timeout:    durationpb.New(timeout),
		Operation: &browserautomationv1.BrowserCommand_Evaluate{
			Evaluate: &browserautomationv1.EvaluateCommand{
				InlineExpression: expression,
				Args:             structArgs,
				Timeout:          durationpb.New(timeout),
			},
		},
	}
}

func commandResultMap(results []*browserautomationv1.BrowserCommandResult, commandID string) map[string]any {
	for _, result := range results {
		if result.GetCommandId() != commandID || result.GetJsonValue() == nil {
			continue
		}
		if value, ok := result.GetJsonValue().AsInterface().(map[string]any); ok {
			return value
		}
	}
	return nil
}

func hashLabel(value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%x", uuid.NewSHA1(uuid.NameSpaceOID, []byte(strings.ToLower(value))))[:12]
}
