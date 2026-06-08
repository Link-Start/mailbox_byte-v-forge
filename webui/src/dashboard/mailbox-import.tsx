import { useState } from 'react';
import { PlusOutlined } from '@ant-design/icons';
import {
  ActionButtonGroup,
  AppDrawer,
  errorText,
  Form,
  MailboxProviderAction,
  SegmentedControl,
  SheetFooter,
  useAsyncActionRunner
} from './dashboard-kit';
import type { ActionButtonDescriptor, MailboxCredentialKind } from './dashboard-kit';
import { providerAction, providerDisplayName, type MailboxProviderTab } from './mailbox-utils';
import { BatchMailboxImportForm, SingleMailboxImportForm } from './mailbox-import-form';
import { importMailboxBatch, importSingleMailbox } from './mailbox-import-submit';
import { mailboxImportModeOptions, type MailboxBatchImportFormState, type MailboxImportFormState, type MailboxImportMode } from './mailbox-import-types';
import type { MailboxProviderCapability } from './types';

export function MailboxImportSheet({ open, provider, capability, busy, onOpenChange, onDone, onError }: {
  open: boolean;
  provider: MailboxProviderTab;
  capability?: MailboxProviderCapability;
  busy: boolean;
  onOpenChange: (open: boolean) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
}) {
  const [mode, setMode] = useState<MailboxImportMode>('single');
  const runner = useAsyncActionRunner();
  const [singleForm] = Form.useForm<MailboxImportFormState>();
  const [batchForm] = Form.useForm<MailboxBatchImportFormState>();
  const importAction = providerAction(capability, MailboxProviderAction.MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX);
  const activeFormId = mode === 'single' ? 'mailbox-import-single' : 'mailbox-import-batch';
  const singleEmail = Form.useWatch('email', singleForm) || '';
  const batchText = Form.useWatch('batchText', batchForm) || '';
  if (!importAction) return null;
  const credentialKinds = importAction.required_credentials || [];

  async function saveSingle(values: MailboxImportFormState) {
    await runImport(async () => {
      const message = await importSingleMailbox(provider, credentialKinds, values);
      singleForm.resetFields();
      return message;
    });
  }

  async function saveBatch(values: MailboxBatchImportFormState) {
    await runImport(async () => {
      const message = await importMailboxBatch(provider, credentialKinds, values);
      batchForm.resetFields();
      return message;
    });
  }

  async function runImport(importer: () => Promise<string>) {
    await runner.tryRun(`import:${mode}`, async () => {
      onDone(await importer());
    }, { onError: (err) => onError(errorText(err)) });
  }

  return (
    <AppDrawer
      open={open}
      title="添加邮箱账号"
      description={`${providerDisplayName(capability, provider)} 可附带 provider 支持的凭据。`}
      footer={<SheetFooter><ActionButtonGroup className="drawerActions" actions={footerActions({ busy: busy || runner.busy, form: activeFormId, disabled: submitDisabled(mode, singleEmail, batchText), onClose: () => onOpenChange(false) })} /></SheetFooter>}
      onOpenChange={onOpenChange}
    >
      <div className="importDrawerBody">
        <SegmentedControl value={mode} options={mailboxImportModeOptions} onChange={setMode} />
        {mode === 'single' ? (
          <SingleMailboxImportForm formId="mailbox-import-single" form={singleForm} credentialKinds={credentialKinds} onFinish={saveSingle} />
        ) : (
          <BatchMailboxImportForm formId="mailbox-import-batch" form={batchForm} placeholder={batchPlaceholder(credentialKinds)} onFinish={saveBatch} />
        )}
      </div>
    </AppDrawer>
  );
}

function batchPlaceholder(credentialKinds: MailboxCredentialKind[]) {
  return credentialKinds.length > 0 ? 'account@example.com----password' : 'account@example.com';
}

function submitDisabled(mode: MailboxImportMode, singleEmail: string, batchText: string) {
  return mode === 'single' ? !singleEmail.trim() : !batchText.trim();
}

function footerActions({ busy, form, disabled, onClose }: {
  busy: boolean;
  form: string;
  disabled: boolean;
  onClose: () => void;
}): ActionButtonDescriptor[] {
  return [{
    id: 'close',
    label: '关闭',
    variant: 'outline',
    onClick: onClose,
  }, {
    id: 'submit',
    label: '添加',
    icon: <PlusOutlined />,
    type: 'submit',
    form,
    disabled: busy || disabled,
  }];
}
