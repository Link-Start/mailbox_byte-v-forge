import { Copy } from 'lucide-react';
import { Button, Card, buttonHint, formatUnix, maskPreview } from './dashboard-kit';
import type { LatestOtp } from './types';

export function MailboxOtpPanel({ latestOtp, showSecrets, loading, compact, onCopy }: {
  latestOtp: LatestOtp | null;
  showSecrets: boolean;
  loading: boolean;
  compact?: boolean;
  onCopy: (label: string, value: string) => void;
}) {
  const hasOtp = !!latestOtp?.captured;
  const subject = latestOtp?.subject || 'Latest OTP';
  const displaySubject = showSecrets ? subject : maskPreview(subject);
  const sizeClass = compact ? 'min-h-[52px] p-2' : 'min-h-[68px] p-2';
  const codeClass = compact ? 'text-sm' : 'text-base';
  return (
    <Card className={`grid grid-cols-[minmax(0,1fr)_34px] items-center gap-2 shadow-none ${sizeClass} ${hasOtp ? 'border-emerald-200 bg-emerald-50' : 'bg-muted/30'}`} role="status" aria-live="polite">
      <div className="grid min-w-0 gap-1">
        <span className="text-xs font-semibold text-muted-foreground">{loading ? '正在拉取 OTP' : '最近 OTP'}</span>
        <strong className={`truncate leading-tight ${codeClass} ${hasOtp ? 'text-emerald-700' : ''}`}>
          {hasOtp ? '验证码已捕获' : '暂无 OTP'}
        </strong>
        <small className="truncate text-xs text-muted-foreground" title={displaySubject}>
          {hasOtp ? `${formatUnix(latestOtp?.received_at_unix || 0)} · ${displaySubject}` : '点击拉取 OTP 后在这里显示最新验证码'}
        </small>
      </div>
      <Button className="copyButton" {...buttonHint('OTP 通过 SecretRef 保存')} disabled onClick={() => onCopy('OTP', '')}>
        <Copy size={14} />
      </Button>
    </Card>
  );
}
