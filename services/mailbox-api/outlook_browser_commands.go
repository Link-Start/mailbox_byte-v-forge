package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
)

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
