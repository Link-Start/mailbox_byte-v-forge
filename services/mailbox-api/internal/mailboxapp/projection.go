package mailboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/redactx"

	"mailboxapi/internal/mailboxmodel"
)

const mailboxErrorSnippetLimit = 600

func PublicMailbox(mailbox *mailboxmodel.Record) *mailboxv1.EmailMailbox {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.EmailMailbox{
		EmailAddress:    mailbox.GetEmailAddress(),
		LastError:       mailbox.GetLastError(),
		CreatedAt:       mailbox.GetCreatedAt(),
		UpdatedAt:       mailbox.GetUpdatedAt(),
		AuthStatus:      mailboxmodel.PublicAuthStatus(mailbox.GetAuthStatus()),
		ProviderKey:     mailbox.GetProviderKey(),
		LatestSignal:    mailbox.GetLatestSignal(),
		Domain:          mailbox.GetDomain(),
		CredentialState: credentialState(mailbox),
	}
}

func PublicMailboxList(mailboxes []*mailboxmodel.Record) []*mailboxv1.EmailMailbox {
	out := make([]*mailboxv1.EmailMailbox, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		if public := PublicMailbox(mailbox); public != nil {
			out = append(out, public)
		}
	}
	return out
}

func RecordFromCredentialInput(input *mailboxv1.MailboxCredentialInput) *mailboxmodel.Record {
	if input == nil {
		return nil
	}
	return applyCredentialInput(&mailboxmodel.Record{
		EmailAddress: emailx.Normalize(input.GetEmailAddress()),
		ProviderKey:  strings.TrimSpace(input.GetProviderKey()),
		AuthStatus:   mailboxmodel.AuthStatusValue(input.GetAuthStatus()),
		LastError:    safeText(input.GetLastError()),
	}, input)
}

func CredentialInput(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialInput {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.MailboxCredentialInput{
		EmailAddress: mailbox.GetEmailAddress(),
		ProviderKey:  mailbox.GetProviderKey(),
		Credentials:  credentials(mailbox),
		AuthStatus:   mailboxmodel.PublicAuthStatus(mailbox.GetAuthStatus()),
		LastError:    mailbox.GetLastError(),
	}
}

func credentialState(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialState {
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

func credentials(mailbox *mailboxmodel.Record) []*mailboxv1.MailboxCredentialValue {
	if mailbox == nil {
		return nil
	}
	values := []*mailboxv1.MailboxCredentialValue{}
	appendValue := func(kind mailboxv1.MailboxCredentialKind, value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, &mailboxv1.MailboxCredentialValue{Kind: kind, Value: value})
		}
	}
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD, mailbox.GetPassword())
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, mailbox.GetRefreshToken())
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, mailbox.GetAccessToken())
	return values
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

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
