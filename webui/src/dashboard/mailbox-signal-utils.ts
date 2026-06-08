import { normalizeUiEmail } from './email-utils';
import type { EmailSignal, InboxMessage, InboxResult, LatestOtp } from './types';

export function latestOtpForInboxResult(result: InboxResult | null, email: string): LatestOtp | null {
  const target = normalizeUiEmail(email);
  if (!result || !target) return null;
  const candidates: LatestOtp[] = [];
  for (const message of result.messages || []) {
    const matchesTarget = normalizeUiEmail(message.mailbox_email) === target ||
      (message.recipients || []).some((recipient) => normalizeUiEmail(recipient) === target);
    const refID = verificationRefForMessage(message);
    if (matchesTarget && messageHasVerificationSignal(message)) {
      candidates.push({ detected: true, secret_resolvable: refID !== '', ref_id: refID, subject: message.subject, received_at_unix: message.received_at_unix });
    }
  }
  candidates.sort((a, b) => b.received_at_unix - a.received_at_unix);
  return candidates[0] || null;
}

export function verificationRefForMessage(message: InboxMessage): string {
  const primary = signalSecretID(message.primary_signal, 'otp');
  if (primary) return primary;
  for (const signal of message.signals || []) {
    const refID = signalSecretID(signal, 'otp');
    if (refID) return refID;
  }
  return '';
}

export function messageHasVerificationSignal(message: InboxMessage): boolean {
  return messageSignals(message).some((signal) => signalKindName(signal.kind) === 'otp');
}

export function messageSignals(message: InboxMessage): EmailSignal[] {
  const signals = [...(message.signals || [])];
  if (message.primary_signal && !signals.some((signal) => signal === message.primary_signal)) signals.unshift(message.primary_signal);
  const seen = new Set<string>();
  return signals.filter((signal) => {
    const key = `${signalKindName(signal.kind)}:${signalSecretID(signal)}:${signal.label || ''}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return signalKindName(signal.kind) !== 'unknown';
  });
}

export function signalKindName(kind: EmailSignal['kind']): 'otp' | 'unknown' {
  if (typeof kind === 'number') {
    if (kind === 1) return 'otp';
    return 'unknown';
  }
  const value = String(kind || '').toLowerCase();
  if (value.includes('otp')) return 'otp';
  return 'unknown';
}

export function signalLabel(signal: EmailSignal): string {
  const label = String(signal.label || '').trim();
  if (label) return label;
  if (signalKindName(signal.kind) === 'otp') return '验证码';
  return '-';
}

export function signalSecretID(signal: EmailSignal | undefined, expectedKind?: 'otp') {
  if (expectedKind && (!signal || signalKindName(signal.kind) !== expectedKind)) return '';
  return String(signal?.secret_ref?.secret_id || '').trim();
}

export function signalHasSecretRef(signal: EmailSignal | undefined, expectedKind?: 'otp'): boolean {
  return signalSecretID(signal, expectedKind) !== '';
}
