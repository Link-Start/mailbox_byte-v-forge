package main

import (
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxmodel"
)

func publicFetchInboxesResponse(resp *mailboxv1.FetchMailboxInboxesResponse, operationID string) *mailboxv1.FetchMailboxInboxesResponse {
	if resp == nil {
		return &mailboxv1.FetchMailboxInboxesResponse{OperationId: operationID}
	}
	return &mailboxv1.FetchMailboxInboxesResponse{
		Results:      append([]*mailboxv1.FetchMailboxInboxResult{}, resp.GetResults()...),
		MailboxCount: resp.GetMailboxCount(),
		FetchedCount: resp.GetFetchedCount(),
		FailedCount:  resp.GetFailedCount(),
		MessageCount: resp.GetMessageCount(),
		OperationId:  operationID,
	}
}

func publicMailbox(mailbox *mailboxmodel.Record) *mailboxv1.EmailMailbox {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.EmailMailbox{
		EmailAddress:    mailbox.GetEmailAddress(),
		LastError:       mailbox.GetLastError(),
		CreatedAt:       mailbox.GetCreatedAt(),
		UpdatedAt:       mailbox.GetUpdatedAt(),
		AuthStatus:      publicMailboxAuthStatus(mailbox.GetAuthStatus()),
		ProviderKey:     mailbox.GetProviderKey(),
		LatestSignal:    mailbox.GetLatestSignal(),
		Domain:          mailbox.GetDomain(),
		CredentialState: publicMailboxCredentialState(mailbox),
	}
}

func publicMailboxList(mailboxes []*mailboxmodel.Record) []*mailboxv1.EmailMailbox {
	out := make([]*mailboxv1.EmailMailbox, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		if public := publicMailbox(mailbox); public != nil {
			out = append(out, public)
		}
	}
	return out
}

func publicMailboxCredentialState(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialState {
	if mailbox == nil {
		return nil
	}
	passwordPresent := strings.TrimSpace(mailbox.GetPassword()) != ""
	refreshTokenPresent := strings.TrimSpace(mailbox.GetRefreshToken()) != ""
	accessTokenPresent := strings.TrimSpace(mailbox.GetAccessToken()) != ""
	present := []mailboxv1.MailboxCredentialKind{}
	appendPresent := func(kind mailboxv1.MailboxCredentialKind, exists bool) {
		if exists {
			present = append(present, kind)
		}
	}
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD, passwordPresent)
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, refreshTokenPresent)
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, accessTokenPresent)
	return &mailboxv1.MailboxCredentialState{
		PasswordPresent:          passwordPresent,
		OauthRefreshTokenPresent: refreshTokenPresent,
		OauthAccessTokenPresent:  accessTokenPresent,
		PresentCredentials:       present,
	}
}

func mailboxRecordFromCredentialInput(input *mailboxv1.MailboxCredentialInput) *mailboxmodel.Record {
	if input == nil {
		return nil
	}
	return applyCredentialInput(&mailboxmodel.Record{
		EmailAddress: emailx.Normalize(input.GetEmailAddress()),
		ProviderKey:  strings.TrimSpace(input.GetProviderKey()),
		AuthStatus:   mailboxAuthStatusValue(input.GetAuthStatus()),
		LastError:    safeMailboxText(input.GetLastError()),
	}, input)
}

func applyCredentialInput(record *mailboxmodel.Record, input *mailboxv1.MailboxCredentialInput) *mailboxmodel.Record {
	if record == nil {
		return nil
	}
	for _, credential := range input.GetCredentials() {
		value := strings.TrimSpace(credential.GetValue())
		switch credential.GetKind() {
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD:
			record.Password = value
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN:
			record.RefreshToken = value
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN:
			record.AccessToken = value
		default:
		}
	}
	return record
}
