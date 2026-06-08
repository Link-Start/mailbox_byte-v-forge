import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SegmentedControl
} from './dashboard-kit';
import { providerDisplayName } from './mailbox-provider-capabilities';
import type { MailboxProviderTab } from './mailbox-provider-config';
import { MailboxImportFooter } from './mailbox-import-footer';
import { BatchMailboxImportForm, SingleMailboxImportForm } from './mailbox-import-form';
import { useMailboxImportState, batchMailboxImportFormID, singleMailboxImportFormID } from './mailbox-import-state';
import { mailboxImportModeOptions } from './mailbox-import-types';
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
  const state = useMailboxImportState({ provider, capability, busy, onDone, onError });
  if (!state.available) return null;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[min(460px,100vw)] p-0 sm:max-w-none">
        <SheetHeader className="border-b">
          <SheetTitle>添加邮箱</SheetTitle>
          <SheetDescription>{providerDisplayName(capability, provider)}</SheetDescription>
        </SheetHeader>
        <div className="grid gap-3 p-4">
          <SegmentedControl value={state.mode} options={mailboxImportModeOptions} onChange={state.setMode} />
          {state.mode === 'single' ? (
            <SingleMailboxImportForm formId={singleMailboxImportFormID} control={state.singleForm.control} credentialKinds={state.credentialKinds} onSubmit={state.singleForm.handleSubmit(state.saveSingle)} />
          ) : (
            <BatchMailboxImportForm formId={batchMailboxImportFormID} control={state.batchForm.control} placeholder={state.batchPlaceholder} onSubmit={state.batchForm.handleSubmit(state.saveBatch)} />
          )}
        </div>
        <SheetFooter className="border-t">
          <MailboxImportFooter busy={state.submitting} form={state.activeFormId} disabled={state.submitDisabled} onClose={() => onOpenChange(false)} />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
