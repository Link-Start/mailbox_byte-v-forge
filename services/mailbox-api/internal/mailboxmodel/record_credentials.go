package mailboxmodel

func (m *Record) GetPassword() string {
	if m == nil {
		return ""
	}
	return m.Password
}

func (m *Record) GetRefreshToken() string {
	if m == nil {
		return ""
	}
	return m.RefreshToken
}

func (m *Record) GetAccessToken() string {
	if m == nil {
		return ""
	}
	return m.AccessToken
}
