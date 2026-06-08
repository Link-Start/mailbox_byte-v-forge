import {
  Alert,
  AlertDescription,
  Badge,
  Card,
  compactToast,
  errorText
} from './dashboard-kit';

export function MailboxPageStatus({ total, providerCount, runningCount, showSecrets, error }: {
  total: number;
  providerCount: number;
  runningCount: number;
  showSecrets: boolean;
  error?: unknown;
}) {
  return (
    <div className="grid gap-2">
      {Boolean(error) && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(errorText(error))}</AlertDescription>
        </Alert>
      )}
      <Card className="flex flex-wrap items-center gap-2 p-3 shadow-none">
        <Badge variant="secondary">邮箱 {total}</Badge>
        <Badge variant="secondary">Provider {providerCount}</Badge>
        {runningCount > 0 && <Badge variant="outline">运行中 {runningCount}</Badge>}
        <Badge variant={showSecrets ? 'outline' : 'secondary'}>{showSecrets ? '显示邮箱地址' : '隐私模式'}</Badge>
      </Card>
    </div>
  );
}
