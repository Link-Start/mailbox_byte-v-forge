package mailboxpg

const inboxSeenPrimaryKeyStatement = `DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_seen_pkey'
				  AND conrelid = 'mailbox_inbox_seen'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_seen DROP CONSTRAINT mailbox_inbox_seen_pkey;
			END IF;
		END $$`

const inboxMessagesPrimaryKeyStatement = `DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_messages_pkey'
				  AND conrelid = 'mailbox_inbox_messages'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_messages DROP CONSTRAINT mailbox_inbox_messages_pkey;
			END IF;
		END $$`
