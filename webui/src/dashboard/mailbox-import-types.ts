import type { MailboxCredentialKind } from '@byte-v-forge/common-ui';

export type MailboxImportFormState = {
  email: string;
  password: string;
  refresh_token: string;
  access_token: string;
};

export type MailboxBatchImportFormState = {
  batchText: string;
};

export type MailboxImportMode = 'single' | 'batch';

export const mailboxImportModeOptions = [
  { value: 'single', label: '单个' },
  { value: 'batch', label: '批量' },
] satisfies { value: MailboxImportMode; label: string }[];

export type MailboxImportPayloadInput = {
  provider: string;
  credentialKinds: MailboxCredentialKind[];
  email: string;
  password: string;
  values?: Partial<MailboxImportFormState>;
};
