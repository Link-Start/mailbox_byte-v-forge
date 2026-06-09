package secretref

import (
	"errors"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	commonv1 "mailboxapi/internal/contracts/commonv1"
)

func New(provider string, purpose string, secretID string, expiresAt time.Time) *commonv1.SecretRef {
	secretID = strings.TrimSpace(secretID)
	if secretID == "" {
		return nil
	}
	ref := &commonv1.SecretRef{
		SecretId: secretID,
		Provider: strings.TrimSpace(provider),
		Purpose:  strings.TrimSpace(purpose),
	}
	if !expiresAt.IsZero() {
		ref.ExpiresAt = timestamppb.New(expiresAt)
	}
	return ref
}

func Clone(ref *commonv1.SecretRef, defaultProvider string, defaultPurpose string) *commonv1.SecretRef {
	if !Configured(ref) {
		return nil
	}
	return &commonv1.SecretRef{
		SecretId:  strings.TrimSpace(ref.GetSecretId()),
		Provider:  firstNonEmpty(ref.GetProvider(), defaultProvider),
		Purpose:   firstNonEmpty(ref.GetPurpose(), defaultPurpose),
		ExpiresAt: ref.GetExpiresAt(),
	}
}

func Configured(ref *commonv1.SecretRef) bool {
	return strings.TrimSpace(ref.GetSecretId()) != ""
}

func Validate(ref *commonv1.SecretRef) error {
	if !Configured(ref) {
		return errors.New("secret_id is required")
	}
	if strings.TrimSpace(ref.GetProvider()) == "" {
		return errors.New("secret provider is required")
	}
	if strings.TrimSpace(ref.GetPurpose()) == "" {
		return errors.New("secret purpose is required")
	}
	return nil
}
