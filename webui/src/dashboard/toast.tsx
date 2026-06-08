import { useCallback } from 'react';
import { toast } from 'sonner';
import { errorText } from './http';
import { compactToast } from './text';

export function useToastMessage() {
  const showToast = useCallback((kind: 'ok' | 'error', text: string) => {
    const content = compactToast(text);
    if (kind === 'error') {
      toast.error(content);
      return;
    }
    toast.success(content);
  }, []);
  const showOK = useCallback((text: string) => showToast('ok', text), [showToast]);
  const showError = useCallback((err: unknown) => showToast('error', errorText(err)), [showToast]);
  const copyValue = useCallback(async (label: string, value: string) => {
    const copied = await copyText(value);
    showToast(copied ? 'ok' : 'error', `${label}${copied ? '已复制' : '复制失败'}`);
  }, [showToast]);
  return { showToast, showOK, showError, copyValue };
}

async function copyText(value: string) {
  if (!value) return false;
  try {
    await navigator.clipboard.writeText(value);
    return true;
  } catch {
    return false;
  }
}
