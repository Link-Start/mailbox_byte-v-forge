import { Card, formatUnix } from './dashboard-kit';
import type { LatestOtp } from './types';

export function MailboxOtpPanel({ latestOtp, loading, compact }: {
  latestOtp: LatestOtp | null;
  loading: boolean;
  compact?: boolean;
}) {
  const hasOtp = !!latestOtp?.detected;
  const statusText = latestOtp?.secret_resolvable ? '已保存' : hasOtp ? '已检测' : '暂无';
  const subject = latestOtp?.subject || 'Latest OTP';
  const sizeClass = compact ? 'min-h-[52px] p-2' : 'min-h-[68px] p-2';
  const codeClass = compact ? 'text-sm' : 'text-base';
  return (
    <Card className={`grid items-center shadow-none ${sizeClass} ${hasOtp ? 'border-emerald-200 bg-emerald-50' : 'bg-muted/30'}`} role="status" aria-live="polite">
      <div className="grid min-w-0 gap-1">
        <span className="text-xs font-semibold text-muted-foreground">{loading ? '拉取中' : 'OTP'}</span>
        <strong className={`truncate leading-tight ${codeClass} ${hasOtp ? 'text-emerald-700' : ''}`}>
          {statusText}
        </strong>
        {hasOtp && <small className="truncate text-xs text-muted-foreground" title={subject}>{formatUnix(latestOtp?.received_at_unix || 0)} · {subject}</small>}
      </div>
    </Card>
  );
}
