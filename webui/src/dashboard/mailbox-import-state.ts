import { useState } from 'react';
import {
  errorText,
  MailboxProviderAction,
  useAsyncActionRunner,
  useForm
} from './dashboard-kit';
import type { MailboxCredentialKind } from './dashboard-kit';
import { providerAction } from './mailbox-provider-capabilities';
import type { MailboxProviderTab } from './mailbox-provider-config';
import { importMailboxBatch, importSingleMailbox } from './mailbox-import-submit';
import type { MailboxBatchImportFormState, MailboxImportFormState, MailboxImportMode } from './mailbox-import-types';
import type { MailboxProviderCapability } from './types';

export const singleMailboxImportFormID = 'mailbox-import-single';
export const batchMailboxImportFormID = 'mailbox-import-batch';

export function useMailboxImportState({ provider, capability, busy, onDone, onError }: {
  provider: MailboxProviderTab;
  capability?: MailboxProviderCapability;
  busy: boolean;
  onDone: (message: string) => void;
  onError: (message: string) => void;
}) {
  const [mode, setMode] = useState<MailboxImportMode>('single');
  const runner = useAsyncActionRunner();
  const singleForm = useForm<MailboxImportFormState>({ defaultValues: { email: '', password: '', refresh_token: '', access_token: '' } });
  const batchForm = useForm<MailboxBatchImportFormState>({ defaultValues: { batchText: '' } });
  const importAction = providerAction(capability, MailboxProviderAction.MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX);
  const credentialKinds = importAction?.required_credentials || [];

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

  return {
    available: !!importAction,
    mode,
    setMode,
    singleForm,
    batchForm,
    credentialKinds,
    batchPlaceholder: batchPlaceholder(credentialKinds),
    activeFormId: mode === 'single' ? singleMailboxImportFormID : batchMailboxImportFormID,
    submitting: busy || runner.busy,
    submitDisabled: submitDisabled(mode, singleForm.watch('email'), batchForm.watch('batchText')),
    saveSingle,
    saveBatch
  };
}

function batchPlaceholder(credentialKinds: MailboxCredentialKind[]) {
  return credentialKinds.length > 0 ? 'account@example.com----password' : 'account@example.com';
}

function submitDisabled(mode: MailboxImportMode, singleEmail: string, batchText: string) {
  return mode === 'single' ? !singleEmail.trim() : !batchText.trim();
}
