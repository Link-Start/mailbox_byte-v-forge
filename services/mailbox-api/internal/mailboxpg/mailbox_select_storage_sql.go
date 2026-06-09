package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func (q *mailboxSelectQuery) selectSQL() (string, error) {
	fields := mailboxSelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	joins := []string{}
	for _, provider := range q.providers.StorageExtensions() {
		tokenFields, ok := provider.TokenFields()
		if !ok {
			continue
		}
		alias := providerStorageAlias(provider.Key())
		join, err := storageJoinSQL(provider, tokenFields, alias)
		if err != nil {
			return "", err
		}
		joins = append(joins, join)
		if err := q.mergeStorageFields(&fields, tokenFields, alias); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf(`
	SELECT m.id, m.email, m.provider,
		%s AS password,
		%s AS refresh_token,
		%s AS access_token,
		%s AS auth_status,
		%s AS last_error,
		m.created_at, m.updated_at
	FROM mailboxes m%s
`, fields.Password, fields.RefreshToken, fields.AccessToken, fields.AuthStatus, fields.LastError, strings.Join(joins, "")), nil
}

func (q *mailboxSelectQuery) mergeStorageFields(fields *mailboxSelectFields, tokenFields mailboxprovider.TokenFields, alias string) error {
	var err error
	if fields.Password, err = coalesceStorageField(fields.Password, alias, tokenFields.PasswordColumn); err != nil {
		return err
	}
	if fields.RefreshToken, err = coalesceStorageField(fields.RefreshToken, alias, tokenFields.RefreshTokenColumn); err != nil {
		return err
	}
	if fields.AccessToken, err = coalesceStorageField(fields.AccessToken, alias, tokenFields.AccessTokenColumn); err != nil {
		return err
	}
	if fields.AuthStatus, err = coalesceStorageField(fields.AuthStatus, alias, tokenFields.AuthStatusColumn); err != nil {
		return err
	}
	if fields.LastError, err = coalesceStorageField(fields.LastError, alias, tokenFields.LastErrorColumn); err != nil {
		return err
	}
	return nil
}
