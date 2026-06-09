export function formatEmailList(values: string[] | undefined) {
  const list = values || [];
  if (list.length === 0) return '-';
  return list.join(', ');
}

export function normalizeUiEmail(value: string) {
  return String(value || '').trim().toLowerCase();
}

export function canonicalUiEmail(value: string) {
  const normalized = normalizeUiEmail(value);
  const [local, domain] = normalized.split('@');
  if (!local || !domain) return normalized;
  return `${local.split('+')[0]}@${domain}`;
}

export function emailDisplayName(value: string) {
  const normalized = normalizeUiEmail(value);
  const [local] = normalized.split('@');
  return local || normalized || '-';
}

export function emailInitial(value: string) {
  const first = emailDisplayName(value).replace(/[^a-z0-9]/gi, '').slice(0, 1);
  return (first || 'M').toUpperCase();
}
