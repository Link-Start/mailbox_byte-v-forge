import { api, errorText, MailboxAuthStatus, MailboxCredentialKind } from './dashboard-kit';
import { parseMailboxBatch, type MailboxProviderTab } from './mailbox-utils';
import type { MailboxBatchImportFormState, MailboxImportFormState, MailboxImportPayloadInput } from './mailbox-import-types';
import type { UpsertEmailMailboxRequest, UpsertEmailMailboxResponse } from './types';

export function mailboxImportPayload({ provider, credentialKinds, email, password, values = {} }: MailboxImportPayloadInput): UpsertEmailMailboxRequest {
  return {
    mailbox: {
      email_address: email,
      provider_key: provider,
      credentials: importCredentials(credentialKinds, { password, refresh_token: values.refresh_token, access_token: values.access_token }),
      auth_status: MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNSPECIFIED,
      last_error: '',
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

function importCredentials(kinds: MailboxCredentialKind[], values: { password?: string; refresh_token?: string; access_token?: string }) {
  return [{
    kind: MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD,
    value: values.password,
  }, {
    kind: MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN,
    value: values.refresh_token,
  }, {
    kind: MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN,
    value: values.access_token,
  }]
    .filter((item) => kinds.includes(item.kind) && String(item.value || '').trim())
    .map((item) => ({ kind: item.kind, value: String(item.value || '').trim() }));
}

function upsertMailbox(body: UpsertEmailMailboxRequest) {
  return api<UpsertEmailMailboxResponse>('/api/mailbox/mailboxes', { method: 'POST', body: JSON.stringify(body) });
}
