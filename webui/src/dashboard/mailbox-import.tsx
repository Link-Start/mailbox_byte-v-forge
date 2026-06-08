import { useState } from 'react';
import { Plus } from 'lucide-react';
import {
  ActionButtonGroup,
  errorText,
  MailboxProviderAction,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SegmentedControl,
  useAsyncActionRunner,
  useForm
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
  const singleForm = useForm<MailboxImportFormState>({ defaultValues: { email: '', password: '', refresh_token: '', access_token: '' } });
  const batchForm = useForm<MailboxBatchImportFormState>({ defaultValues: { batchText: '' } });
  const importAction = providerAction(capability, MailboxProviderAction.MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX);
  const activeFormId = mode === 'single' ? 'mailbox-import-single' : 'mailbox-import-batch';
  if (!importAction) return null;
  const credentialKinds = importAction.required_credentials || [];

  async function saveSingle(values: MailboxImportFormState) {
    await runImport(async () => {
      const message = await importSingleMailbox(provider, credentialKinds, values);
      singleForm.reset({ email: '', password: '', refresh_token: '', access_token: '' });
      return message;
    });
  }

  async function saveBatch(values: MailboxBatchImportFormState) {
    await runImport(async () => {
      const message = await importMailboxBatch(provider, credentialKinds, values);
      batchForm.reset({ batchText: '' });
      return message;
    });
  }

  async function runImport(importer: () => Promise<string>) {
    await runner.tryRun(`import:${mode}`, async () => {
      onDone(await importer());
    }, { onError: (err) => onError(errorText(err)) });
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[min(460px,100vw)] p-0 sm:max-w-none">
        <SheetHeader className="border-b">
          <SheetTitle>添加邮箱账号</SheetTitle>
          <SheetDescription>{providerDisplayName(capability, provider)} 可附带 provider 支持的凭据。</SheetDescription>
        </SheetHeader>
        <div className="grid gap-3 p-4">
          <SegmentedControl value={mode} options={mailboxImportModeOptions} onChange={setMode} />
          {mode === 'single' ? (
            <SingleMailboxImportForm formId="mailbox-import-single" control={singleForm.control} credentialKinds={credentialKinds} onSubmit={singleForm.handleSubmit(saveSingle)} />
          ) : (
            <BatchMailboxImportForm formId="mailbox-import-batch" control={batchForm.control} placeholder={batchPlaceholder(credentialKinds)} onSubmit={batchForm.handleSubmit(saveBatch)} />
          )}
        </div>
        <SheetFooter className="border-t">
          <ActionButtonGroup className="grid gap-2" actions={footerActions({ busy: busy || runner.busy, form: activeFormId, disabled: submitDisabled(mode, singleForm.watch('email'), batchForm.watch('batchText')), onClose: () => onOpenChange(false) })} />
        </SheetFooter>
      </SheetContent>
    </Sheet>
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
    icon: <Plus />,
    type: 'submit',
    form,
    disabled: busy || disabled,
  }];
}
