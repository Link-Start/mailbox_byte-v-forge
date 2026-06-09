package mailboxpg

import "mailboxapi/internal/mailboxprovider"

func mergeStorageFields(fields *mailboxSelectFields, tokenFields mailboxprovider.TokenFields, alias string) error {
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
