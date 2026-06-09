package mailboxmodel

func (m *Record) GetCreatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.CreatedAt
}

func (m *Record) GetUpdatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.UpdatedAt
}
