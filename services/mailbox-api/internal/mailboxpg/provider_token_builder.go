package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func newProviderTokenUpdate(fields mailboxprovider.TokenFields, refreshToken string, accessToken string) (providerTokenUpdate, error) {
	table, err := mailboxprovider.SQLIdentifier(fields.Table)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	emailColumn, err := mailboxprovider.SQLIdentifier(fields.EmailColumn)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	refreshColumn, err := mailboxprovider.SQLIdentifier(fields.RefreshTokenColumn)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	accessColumn, err := mailboxprovider.SQLIdentifier(fields.AccessTokenColumn)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	return providerTokenUpdate{
		table:       table,
		emailColumn: emailColumn,
		args:        []any{strings.TrimSpace(refreshToken), strings.TrimSpace(accessToken)},
		assignments: []string{
			fmt.Sprintf("%s = $1", refreshColumn),
			fmt.Sprintf("%s = $2", accessColumn),
		},
	}, nil
}

func (u *providerTokenUpdate) addOptional(rawColumn string, value any) error {
	if rawColumn == "" {
		return nil
	}
	column, err := mailboxprovider.SQLIdentifier(rawColumn)
	if err != nil {
		return err
	}
	u.args = append(u.args, value)
	u.assignments = append(u.assignments, fmt.Sprintf("%s = $%d", column, len(u.args)))
	return nil
}
