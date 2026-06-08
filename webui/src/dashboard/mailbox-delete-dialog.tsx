import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from './dashboard-kit';
import { maskEmail } from './email-utils';
import type { Mailbox } from './types';

export function MailboxDeleteDialog({ mailbox, showSecrets, busy, onCancel, onConfirm }: {
  mailbox?: Mailbox | null;
  showSecrets: boolean;
  busy: boolean;
  onCancel: () => void;
  onConfirm: () => void | Promise<void>;
}) {
  const email = mailbox?.email_address || '';
  const displayEmail = showSecrets ? email : maskEmail(email);
  return (
    <AlertDialog open={!!mailbox} onOpenChange={(open) => { if (!open) onCancel(); }}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除邮箱？</AlertDialogTitle>
          <AlertDialogDescription>
            将删除 {displayEmail || '当前邮箱'} 及已缓存收件箱记录，此操作不可撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>取消</AlertDialogCancel>
          <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90" disabled={busy} onClick={() => void onConfirm()}>
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
