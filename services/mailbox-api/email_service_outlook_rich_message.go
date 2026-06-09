package main

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (s *EmailService) hydrateOutlookRichInboxMessage(ctx context.Context, email string, parserProfile string, resp *mailboxv1.GetMailboxInboxMessageResponse) *mailboxv1.GetMailboxInboxMessageResponse {
	if !shouldHydrateOutlookRichInboxMessage(resp) {
		return resp
	}
	mailbox, err := s.mailboxRepo.PollMailboxForEmail(ctx, email)
	if err != nil {
		logWarning("hydrate Outlook rich email body mailbox=%s: %v", emailx.Redact(email), err)
		return resp
	}
	source, err := s.watcher.sources.SourceForMailbox(mailbox)
	if err != nil {
		logWarning("hydrate Outlook rich email body source mailbox=%s: %v", emailx.Redact(email), err)
		return resp
	}
	outlook, ok := source.(*outlookInboxSource)
	if !ok {
		return resp
	}
	input, ok, err := outlook.FetchInboxMessage(ctx, mailbox, resp.GetMessage().GetId())
	if err != nil {
		logWarning("hydrate Outlook rich email body fetch mailbox=%s: %v", emailx.Redact(email), err)
		return resp
	}
	if !ok || input.HTMLBody == "" {
		return resp
	}
	if _, err := s.inbox.RecordMessageInputs(ctx, outlook.ProviderKey(), []inboxapp.MessageInput{input}, true); err != nil {
		logWarning("hydrate Outlook rich email body persist mailbox=%s: %v", emailx.Redact(email), err)
		return resp
	}
	fresh, err := s.inbox.GetMessage(ctx, email, resp.GetMessage().GetId(), outlook.ProviderKey(), parserProfile)
	if err != nil {
		return resp
	}
	return fresh
}

func shouldHydrateOutlookRichInboxMessage(resp *mailboxv1.GetMailboxInboxMessageResponse) bool {
	return resp != nil &&
		resp.GetMessage() != nil &&
		resp.GetHtmlBody() == "" &&
		resp.GetMessage().GetProviderKey() == emailProviderOutlook &&
		resp.GetMessage().GetId() != ""
}
