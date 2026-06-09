import { AlertCircle } from 'lucide-react';
import type { MailboxOperation } from './types';

export function MailboxErrorMeta({ error }: { error?: string }) {
  if (!error) return null;
  return (
    <div className="recordMeta grid-cols-1">
      <span className="flex min-w-0 items-center gap-1 truncate text-xs text-destructive" title={error}>
        <AlertCircle className="size-3.5 shrink-0" />
        {error}
      </span>
    </div>
  );
}

export function MailboxOperationMeta({ operation }: { operation?: MailboxOperation }) {
  if (!operation) return null;
  return (
    <div className="recordMeta grid-cols-1">
      <span className="truncate text-xs text-muted-foreground" title={operation.operation_id}>
        运行中 · {operation.last_step || operation.action || operation.status}
      </span>
    </div>
  );
}
