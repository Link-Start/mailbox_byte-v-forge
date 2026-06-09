export function buttonHint(label: string) {
  return { title: label, 'aria-label': label, 'data-tooltip': label };
}

export function short(value: string, size = 8) {
  if (!value) return '-';
  return value.length > size ? `${value.slice(0, size)}…` : value;
}

export function mask(value: string) {
  return value ? '••••••••' : '-';
}

export function compactToast(value: string) {
  const text = String(value || '');
  return text.length > 150 ? `${text.slice(0, 150)}...` : text;
}

export function uniqueStrings(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

export function cx(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ');
}
