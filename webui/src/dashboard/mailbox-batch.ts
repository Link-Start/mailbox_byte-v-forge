import { MailboxCredentialKind } from './dashboard-kit';

export type MailboxBatchItem = {
  email: string;
  password: string;
};

export function parseMailboxBatch(value: string, credentialKinds: MailboxCredentialKind[]) {
  const items: MailboxBatchItem[] = [];
  const errors: string[] = [];
  const allowPlainEmailBatch = !credentialKinds.includes(MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD);

  value.split(/\r?\n/).forEach((raw, index) => {
    const line = raw.trim();
    if (!line) return;
    const delimiterIndex = line.indexOf('----');
    if (allowPlainEmailBatch && delimiterIndex < 0) {
      items.push({ email: line, password: '' });
      return;
    }
    if (delimiterIndex < 0) {
      errors.push(`第 ${index + 1} 行缺少 ----`);
      return;
    }
    const email = line.slice(0, delimiterIndex).trim();
    const password = line.slice(delimiterIndex + 4).trim();
    if (!email) {
      errors.push(`第 ${index + 1} 行缺少账号`);
      return;
    }
    items.push({ email, password });
  });

  return { items, errors };
}
