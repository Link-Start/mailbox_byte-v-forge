import { useState } from 'react';
import { Plus } from 'lucide-react';
import {
  ActionButtonGroup,
  errorText,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SegmentedControl,
  useAsyncActionRunner,
  useForm
} from '@byte-v-forge/common-ui';
import type { ActionButtonDescriptor } from '@byte-v-forge/common-ui';
import { mailboxProviderConfig, type MailboxProviderTab } from './mailbox-utils';
import { BatchMailboxImportForm, SingleMailboxImportForm } from './mailbox-import-form';
import { importMailboxBatch, importSingleMailbox } from './mailbox-import-submit';
import { mailboxImportModeOptions, type MailboxBatchImportFormState, type MailboxImportFormState, type MailboxImportMode } from './mailbox-import-types';

export function MailboxImportSheet({ open, provider, busy, onOpenChange, onDone, onError }: {
  open: boolean;
  provider: MailboxProviderTab;
  busy: boolean;
  onOpenChange: (open: boolean) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
}) {
  const [mode, setMode] = useState<MailboxImportMode>('single');
  const runner = useAsyncActionRunner();
  const singleForm = useForm<MailboxImportFormState>({ defaultValues: { email: '', password: '', refresh_token: '', access_token: '' } });
  const batchForm = useForm<MailboxBatchImportFormState>({ defaultValues: { batchText: '' } });
  const importConfig = mailboxProviderConfig(provider).import;
  const activeFormId = mode === 'single' ? 'mailbox-import-single' : 'mailbox-import-batch';
  if (!importConfig) return null;
  const config = importConfig;

  async function saveSingle(values: MailboxImportFormState) {
    await runImport(async () => {
      const message = await importSingleMailbox(provider, config.credentialKinds, values);
      singleForm.reset({ email: '', password: '', refresh_token: '', access_token: '' });
      return message;
    });
  }

  async function saveBatch(values: MailboxBatchImportFormState) {
    await runImport(async () => {
      const message = await importMailboxBatch(provider, config.credentialKinds, values);
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
          <SheetDescription>{importConfig.description}</SheetDescription>
        </SheetHeader>
        <div className="grid gap-3 p-4">
          <SegmentedControl value={mode} options={mailboxImportModeOptions} onChange={setMode} />
          {mode === 'single' ? (
            <SingleMailboxImportForm formId="mailbox-import-single" control={singleForm.control} credentialKinds={importConfig.credentialKinds} onSubmit={singleForm.handleSubmit(saveSingle)} />
          ) : (
            <BatchMailboxImportForm formId="mailbox-import-batch" control={batchForm.control} placeholder={importConfig.batchPlaceholder} onSubmit={batchForm.handleSubmit(saveBatch)} />
          )}
        </div>
        <SheetFooter className="border-t">
          <ActionButtonGroup className="grid gap-2" actions={footerActions({ busy: busy || runner.busy, form: activeFormId, disabled: submitDisabled(mode, singleForm.watch('email'), batchForm.watch('batchText')), onClose: () => onOpenChange(false) })} />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
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
