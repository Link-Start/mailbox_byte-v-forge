package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
)

func readMailboxRecords(dir string, includePasswordOnly bool) ([]mailboxRecord, error) {
	recordsByEmail := map[string]mailboxRecord{}
	tokenRecords, err := parseTokenFile(filepath.Join(dir, "outlook_token.txt"))
	if err != nil {
		return nil, err
	}
	for _, record := range tokenRecords {
		recordsByEmail[record.email] = record
	}
	if includePasswordOnly {
		for _, name := range []string{"logged_email.txt", "unlogged_email.txt"} {
			passwordRecords, err := parsePasswordFile(filepath.Join(dir, name))
			if err != nil {
				return nil, err
			}
			for _, record := range passwordRecords {
				if _, exists := recordsByEmail[record.email]; !exists {
					recordsByEmail[record.email] = record
				}
			}
		}
	}
	records := make([]mailboxRecord, 0, len(recordsByEmail))
	for _, record := range recordsByEmail {
		records = append(records, record)
	}
	return records, nil
}

func parseTokenFile(path string) ([]mailboxRecord, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	records := []mailboxRecord{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "---", 5)
		if len(parts) < 3 {
			continue
		}
		email := emailx.Normalize(parts[0])
		if email == "" {
			continue
		}
		record := mailboxRecord{
			email:        email,
			password:     strings.TrimSpace(parts[1]),
			refreshToken: strings.TrimSpace(parts[2]),
			source:       filepath.Base(path),
		}
		if len(parts) >= 4 {
			record.accessToken = strings.TrimSpace(parts[3])
		}
		records = append(records, record)
	}
	return records, nil
}

func parsePasswordFile(path string) ([]mailboxRecord, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	records := []mailboxRecord{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		email, password, _ := strings.Cut(line, ":")
		if email = emailx.Normalize(email); email != "" {
			records = append(records, mailboxRecord{email: email, password: strings.TrimSpace(password), source: filepath.Base(path)})
		}
	}
	return records, nil
}
