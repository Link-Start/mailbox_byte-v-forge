import { useCallback, useMemo, useState } from 'react';

export type AsyncActionHooks<T> = {
  onSuccess?: (value: T) => void | Promise<void>;
  onError?: (error: unknown) => void | Promise<void>;
  onSettled?: () => void | Promise<void>;
};

export function actionTargetStateKey(actionID: string, target?: unknown) {
  return actionStateKey(actionID, target);
}

export function activeActionTargets(activeKeys: Iterable<string>, actionID: string) {
  return Array.from(activeKeys, (key) => activeActionTarget(key, actionID)).filter(Boolean);
}

export function hasActiveAction(activeKeys: Iterable<string>, actionID: string) {
  const prefix = actionStateKey(actionID);
  if (!prefix) return false;
  for (const key of activeKeys) {
    const normalized = actionStateKey(key);
    if (normalized === prefix || normalized.startsWith(`${prefix}:`)) return true;
  }
  return false;
}

export function useAsyncActionRunner(initialKeys: Iterable<string> = []) {
  const [activeKeys, setActiveKeys] = useState(() => new Set(Array.from(initialKeys, actionStateKey).filter(Boolean)));
  const keys = useMemo(() => new Set(activeKeys), [activeKeys]);
  const setActive = useCallback((key: string, active: boolean) => setActiveKeys((prev) => setKeyActive(prev, key, active)), []);
  const run = useCallback(async <T,>(key: string, action: () => T | Promise<T>, hooks: AsyncActionHooks<T> = {}) => {
    const normalized = actionStateKey(key);
    if (normalized) setActive(normalized, true);
    try {
      const value = await action();
      await hooks.onSuccess?.(value);
      return value;
    } catch (error) {
      await hooks.onError?.(error);
      throw error;
    } finally {
      if (normalized) setActive(normalized, false);
      await hooks.onSettled?.();
    }
  }, [setActive]);
  const tryRun = useCallback(async <T,>(key: string, action: () => T | Promise<T>, hooks: AsyncActionHooks<T> = {}) => {
    try { return { ok: true as const, value: await run(key, action, hooks) }; }
    catch (error) { return { ok: false as const, error }; }
  }, [run]);
  return { activeKeys: keys as ReadonlySet<string>, busy: activeKeys.size > 0, run, tryRun };
}

function activeActionTarget(activeKey: string, actionID: string) {
  const prefix = actionStateKey(actionID);
  const key = actionStateKey(activeKey);
  if (!prefix || !key || key === prefix) return '';
  return key.startsWith(`${prefix}:`) ? key.slice(prefix.length + 1) : '';
}

function actionStateKey(...parts: unknown[]) {
  return parts.map((value) => String(value ?? '').trim()).filter(Boolean).join(':');
}

function setKeyActive(prev: Set<string>, key: string, active: boolean) {
  const normalized = actionStateKey(key);
  if (!normalized || prev.has(normalized) === active) return prev;
  const next = new Set(prev);
  if (active) next.add(normalized); else next.delete(normalized);
  return next;
}
