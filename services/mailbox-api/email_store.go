package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mailboxRow struct {
	ID           string
	Email        string
	Provider     string
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
	CreatedAt    int64
	UpdatedAt    int64
}

type inboxMessageRow struct {
	ID             string
	MailboxEmail   string
	Subject        string
	FromAddress    string
	BodyPreview    string
	ReceivedAtUnix int64
	RecipientsJSON string
	Provider       string
	SourceEmail    string
	BodyText       string
	HTMLBody       string
	RawSize        int64
}

type inboxPersistMessage struct {
	key            string
	id             string
	mailboxEmail   string
	subject        string
	fromAddress    string
	bodyPreview    string
	receivedAtUnix int64
	recipients     []string
	provider       string
	sourceEmail    string
	bodyText       string
	htmlBody       string
	rawSize        int64
}

type inboxMessageKey struct {
	provider     string
	mailboxEmail string
	messageKey   string
}

func (row inboxMessageRow) toProto() (*mailboxv1.EmailInboxMessage, error) {
	return row.toProtoForProfile("")
}

func (row inboxMessageRow) toProtoForProfile(profile string) (*mailboxv1.EmailInboxMessage, error) {
	recipients := []string{}
	if strings.TrimSpace(row.RecipientsJSON) != "" {
		if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
			return nil, err
		}
	}
	return emailMessageWithSignals(&mailboxv1.EmailInboxMessage{
		Id:                 row.ID,
		MailboxEmail:       emailx.Normalize(row.MailboxEmail),
		Subject:            row.Subject,
		FromAddress:        row.FromAddress,
		BodyPreview:        row.BodyPreview,
		ReceivedAtUnix:     row.ReceivedAtUnix,
		Recipients:         uniqueStrings(recipients),
		ProviderKey:        normalizeEmailProvider(row.Provider),
		SourceMailboxEmail: emailx.Normalize(row.SourceEmail),
		BodyArtifactRef:    inboxArtifactRef(row.Provider, row.MailboxEmail, row.ID, "body_text", int64(len(row.BodyText))),
		HtmlArtifactRef:    inboxArtifactRef(row.Provider, row.MailboxEmail, row.ID, "html_body", int64(len(row.HTMLBody))),
		RawSize:            row.RawSize,
	}, profile), nil
}

func inboxArtifactRef(provider string, mailboxEmail string, messageID string, purpose string, sizeBytes int64) *commonv1.ArtifactRef {
	provider = normalizeEmailProvider(provider)
	mailboxEmail = emailx.Normalize(mailboxEmail)
	messageID = strings.TrimSpace(messageID)
	purpose = strings.TrimSpace(purpose)
	if provider == "" || mailboxEmail == "" || messageID == "" || purpose == "" || sizeBytes <= 0 {
		return nil
	}
	artifactID := strings.Join([]string{"mailbox", provider, mailboxEmail, messageID, purpose}, ":")
	return &commonv1.ArtifactRef{
		ArtifactId:  artifactID,
		Uri:         "mailbox://inbox/" + artifactID,
		ContentType: mailboxArtifactContentType(purpose),
		SizeBytes:   sizeBytes,
		Purpose:     purpose,
	}
}

func mailboxArtifactContentType(purpose string) string {
	if purpose == "html_body" {
		return "text/html"
	}
	return "text/plain"
}

type rowScanner interface {
	Scan(dest ...any) error
}

type MailboxStore struct {
	pool   *pgxpool.Pool
	recent *recentEmailCache
}

func NewMailboxStore(ctx context.Context, dsn string, recent *recentEmailCache) (*MailboxStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("PG_DSN is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	store := &MailboxStore{pool: pool, recent: recent}
	if err := store.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *MailboxStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
