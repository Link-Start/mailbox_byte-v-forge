package mailboxmodel

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

func (m *Record) GetLastError() string {
	if m == nil {
		return ""
	}
	return m.LastError
}

func (m *Record) GetAuthStatus() string {
	if m == nil {
		return ""
	}
	return m.AuthStatus
}

func (m *Record) GetLatestSignal() *mailboxv1.EmailSignal {
	if m == nil {
		return nil
	}
	return m.LatestSignal
}
