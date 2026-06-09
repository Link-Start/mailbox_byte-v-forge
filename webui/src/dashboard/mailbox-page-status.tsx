import {
  Alert,
  AlertDescription,
  Badge,
  compactToast,
  errorText
} from './dashboard-kit';

export function MailboxPageStatus({ total, runningCount, error }: {
  total: number;
  runningCount: number;
  error?: unknown;
}) {
  return (
    <div className="grid gap-2">
      {Boolean(error) && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(errorText(error))}</AlertDescription>
        </Alert>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="secondary">{total} 邮箱</Badge>
        {runningCount > 0 && <Badge variant="outline">运行中 {runningCount}</Badge>}
      </div>
    </div>
  );
}
