import {
  Form,
  Input,
  MailboxCredentialKind,
  type FormInstance
} from './dashboard-kit';
import type { MailboxBatchImportFormState, MailboxImportFormState } from './mailbox-import-types';

export function SingleMailboxImportForm({ formId, form, credentialKinds, onFinish }: {
  formId: string;
  form: FormInstance<MailboxImportFormState>;
  credentialKinds: MailboxCredentialKind[];
  onFinish: (values: MailboxImportFormState) => void | Promise<void>;
}) {
  return (
    <Form id={formId} form={form} layout="vertical" className="importForm" initialValues={{ email: '', password: '', refresh_token: '', access_token: '' }} onFinish={onFinish}>
      {singleMailboxImportFields(credentialKinds).map((field) => (
        <Form.Item key={field.id} name={field.name} label={field.label} hidden={field.visible === false}>
          <Input id={field.inputId || field.id} type={field.type || 'text'} placeholder={field.placeholder} />
        </Form.Item>
      ))}
    </Form>
  );
}

export function BatchMailboxImportForm({ formId, form, placeholder, onFinish }: {
  formId: string;
  form: FormInstance<MailboxBatchImportFormState>;
  placeholder: string;
  onFinish: (values: MailboxBatchImportFormState) => void | Promise<void>;
}) {
  return (
    <Form id={formId} form={form} layout="vertical" className="importForm" initialValues={{ batchText: '' }} onFinish={onFinish}>
      <Form.Item name="batchText">
        <Input.TextArea className="min-h-32 resize-y" placeholder={placeholder} autoSize={{ minRows: 6 }} />
      </Form.Item>
    </Form>
  );
}

type ImportField = {
  id: string;
  name: keyof MailboxImportFormState;
  label: string;
  placeholder?: string;
  type?: string;
  inputId?: string;
  visible?: boolean;
};

function singleMailboxImportFields(credentialKinds: MailboxCredentialKind[]): ImportField[] {
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
