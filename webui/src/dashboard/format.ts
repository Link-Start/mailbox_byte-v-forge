export function formatUnix(value: number) {
  return value ? new Date(value * 1000).toLocaleString() : '-';
}
