package mailboxprovider

type TokenFields struct {
	Table              string
	EmailColumn        string
	PasswordColumn     string
	RefreshTokenColumn string
	AccessTokenColumn  string
	AuthStatusColumn   string
	LastErrorColumn    string
	CreatedAtColumn    string
	UpdatedAtColumn    string
}

func (f TokenFields) HasTokenStorage() bool {
	return f.Table != "" && f.EmailColumn != "" && f.RefreshTokenColumn != "" && f.AccessTokenColumn != ""
}
