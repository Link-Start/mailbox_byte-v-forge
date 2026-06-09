package mailboxpg

import (
	"fmt"
	"strings"
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
		if err := mergeStorageFields(&fields, tokenFields, alias); err != nil {
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
