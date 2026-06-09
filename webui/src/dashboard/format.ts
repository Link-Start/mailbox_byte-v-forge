export function formatUnix(value: number) {
  return value ? new Date(value * 1000).toLocaleString() : '-';
}

export function formatInboxTime(value: number) {
  if (!value) return '-';
  const date = new Date(value * 1000);
  const now = new Date();
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  if (date.getFullYear() === now.getFullYear()) {
    return date.toLocaleDateString([], { month: 'numeric', day: 'numeric' });
  }
  return date.toLocaleDateString([], { year: 'numeric', month: 'numeric', day: 'numeric' });
}
