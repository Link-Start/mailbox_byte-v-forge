import { Badge, cx } from './dashboard-kit';
import { messageHasVerificationSignal, messageSignals, signalKindName, signalLabel } from './mailbox-signal-utils';
import type { InboxMessage } from './types';

export function MessageSignalBadges({ message, className }: { message: InboxMessage; className?: string }) {
  const signals = messageSignals(message);
  if (signals.length === 0 && messageHasVerificationSignal(message)) {
    return <Badge variant="outline" className="border-emerald-200 bg-emerald-50 text-emerald-700">验证码</Badge>;
  }
  if (signals.length === 0) return null;
  return (
    <span className={cx('flex shrink-0 flex-wrap items-center gap-1', className)}>
      {signals.map((signal, index) => {
        const kind = signalKindName(signal.kind);
        return <Badge variant="secondary" key={`${kind}-${signal.label || index}`}>{signalLabel(signal) || kind}</Badge>;
      })}
    </span>
  );
}
