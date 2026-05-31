import type { FormEventHandler } from 'react';
import {
  ControlledInputFieldList,
  ControlledTextareaField,
  MailboxCredentialKind,
  type Control,
  type ControlledInputFieldDescriptor
} from '@byte-v-forge/common-ui';
import type { MailboxBatchImportFormState, MailboxImportFormState } from './mailbox-import-types';

export function SingleMailboxImportForm({ formId, control, credentialKinds, onSubmit }: {
  formId: string;
  control: Control<MailboxImportFormState>;
  credentialKinds: MailboxCredentialKind[];
  onSubmit: FormEventHandler<HTMLFormElement>;
}) {
  return (
    <form id={formId} className="grid gap-2" onSubmit={onSubmit}>
      <ControlledInputFieldList control={control} fields={singleMailboxImportFields(credentialKinds)} />
    </form>
  );
}

export function BatchMailboxImportForm({ formId, control, placeholder, onSubmit }: {
  formId: string;
  control: Control<MailboxBatchImportFormState>;
  placeholder: string;
  onSubmit: FormEventHandler<HTMLFormElement>;
}) {
  return (
    <form id={formId} onSubmit={onSubmit}>
      <ControlledTextareaField
        control={control}
        name="batchText"
        className="min-h-32 resize-y"
        placeholder={placeholder}
      />
    </form>
  );
}

function singleMailboxImportFields(credentialKinds: MailboxCredentialKind[]): ControlledInputFieldDescriptor<MailboxImportFormState>[] {
  return [{
    id: 'email',
    name: 'email',
    label: '邮箱',
    placeholder: '邮箱',
    inputId: 'mailbox-import-email',
  }, {
    id: 'password',
    name: 'password',
    label: '密码',
    placeholder: '邮箱密码，可空',
    type: 'password',
    inputId: 'mailbox-import-password',
    visible: credentialKinds.includes(MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD),
  }, {
    id: 'refresh-token',
    name: 'refresh_token',
    label: 'Refresh token',
    placeholder: 'Refresh token，可空',
    type: 'password',
    inputId: 'mailbox-import-refresh-token',
    visible: credentialKinds.includes(MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN),
  }, {
    id: 'access-token',
    name: 'access_token',
    label: 'Access token',
    placeholder: 'Access token，可空',
    type: 'password',
    inputId: 'mailbox-import-access-token',
    visible: credentialKinds.includes(MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN),
  }];
}
