import { api, errorText, MailboxCredentialKind } from '@byte-v-forge/common-ui';
import { parseMailboxBatch, type MailboxProviderTab } from './mailbox-utils';
import type { UpsertEmailMailboxRequest, UpsertEmailMailboxResponse } from '../proto/email';
import type { MailboxBatchImportFormState, MailboxImportFormState, MailboxImportPayloadInput } from './mailbox-import-types';

export function mailboxImportPayload({ provider, credentialKinds, email, password, values = {} }: MailboxImportPayloadInput): UpsertEmailMailboxRequest {
  return {
    mailbox: {
      email_address: email,
      password: importCredentialValue(credentialKinds, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD, password),
      refresh_token: importCredentialValue(credentialKinds, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, values.refresh_token),
      access_token: importCredentialValue(credentialKinds, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, values.access_token),
      provider_key: provider,
      auth_status: '',
      last_error: '',
      created_at: 0,
      updated_at: 0,
      latest_signal: undefined,
      domain: ''
    }
  };
}

export async function importSingleMailbox(provider: MailboxProviderTab, credentialKinds: MailboxCredentialKind[], values: MailboxImportFormState) {
  const resp = await upsertMailbox(mailboxImportPayload({ provider, credentialKinds, email: values.email, password: values.password, values }));
  return `邮箱已入池: ${resp.mailbox?.email_address || values.email}`;
}

export async function importMailboxBatch(provider: MailboxProviderTab, credentialKinds: MailboxCredentialKind[], values: MailboxBatchImportFormState) {
  const batch = parseMailboxBatch(values.batchText, credentialKinds);
  if (batch.items.length === 0) {
    throw new Error(batch.errors.length ? `批量入池失败：${batch.errors[0]}` : '没有可入池邮箱');
  }

  let success = 0;
  const failures = [...batch.errors];
  for (const item of batch.items) {
    try {
      await upsertMailbox(mailboxImportPayload({ provider, credentialKinds, email: item.email, password: item.password }));
      success += 1;
    } catch (err) {
      failures.push(`${item.email}: ${errorText(err)}`);
    }
  }
  if (success === 0) {
    throw new Error(`批量入池失败：${failures.slice(0, 3).join('；')}`);
  }
  return `批量入池成功 ${success}${failures.length ? `，失败 ${failures.length}` : ''}`;
}

function importCredentialValue(kinds: MailboxCredentialKind[], kind: MailboxCredentialKind, value?: string) {
  return kinds.includes(kind) ? String(value || '') : '';
}

function upsertMailbox(body: UpsertEmailMailboxRequest) {
  return api<UpsertEmailMailboxResponse>('/api/mailbox/mailboxes', { method: 'POST', body: JSON.stringify(body) });
}
