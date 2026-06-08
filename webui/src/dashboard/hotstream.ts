import { useEffect, useRef } from 'react';
import { useQueryClient, type QueryClient, type QueryKey } from '@tanstack/react-query';
import { HotStreamControlKind, type HotStreamControlEvent, type HotStreamEvent } from '../proto/byte/v/forge/contracts/observability/v1/hotstream';

type HotStreamFilters = {
  eventTypes?: string[];
  resourceTypes?: string[];
  resourceIds?: string[];
};

type HotStreamInvalidationRule = {
  queryKey: QueryKey;
  eventTypes?: string[];
  resourceTypes?: string[];
  resourceIds?: string[];
};

type HotStreamInvalidationOptions = {
  enabled?: boolean;
  url: string;
  rules: HotStreamInvalidationRule[];
  onEvent?: (event: HotStreamEvent, queryClient: QueryClient) => void;
};

export function useHotStreamInvalidation(options: HotStreamInvalidationOptions) {
  const queryClient = useQueryClient();
  const latest = useRef(options);
  latest.current = options;
  useEffect(() => {
    if (options.enabled === false || !options.url) return;
    const source = new EventSource(options.url);
    const eventHandler = (message: Event) => {
      const event = parseMessage<HotStreamEvent>(message);
      if (!event) return;
      for (const rule of latest.current.rules) if (matchesRule(event, rule)) queryClient.invalidateQueries({ queryKey: rule.queryKey });
      latest.current.onEvent?.(event, queryClient);
    };
    const controlHandler = (message: Event) => {
      const event = parseMessage<HotStreamControlEvent>(message);
      if (event?.kind === HotStreamControlKind.HOT_STREAM_CONTROL_KIND_RESYNC_REQUIRED) {
        for (const rule of latest.current.rules) queryClient.invalidateQueries({ queryKey: rule.queryKey });
      }
    };
    source.addEventListener('hotstream', eventHandler);
    source.addEventListener('hotstream.control', controlHandler);
    source.addEventListener('error', () => latest.current.rules.forEach((rule) => queryClient.invalidateQueries({ queryKey: rule.queryKey })));
    return () => source.close();
  }, [options.enabled, options.url, queryClient]);
}

export function createHotStreamURL(apiBase: string, filters: HotStreamFilters = {}) {
  const params = new URLSearchParams();
  appendAll(params, 'event_type', filters.eventTypes);
  appendAll(params, 'resource_type', filters.resourceTypes);
  appendAll(params, 'resource_id', filters.resourceIds);
  const query = params.toString();
  return `${apiBase.replace(/\/$/, '')}/streams/state${query ? `?${query}` : ''}`;
}

function parseMessage<T>(message: Event): T | null {
  try { return JSON.parse((message as MessageEvent<string>).data) as T; } catch { return null; }
}

function matchesRule(event: HotStreamEvent, rule: HotStreamInvalidationRule) {
  return includes(rule.eventTypes, event.metadata?.type) && includes(rule.resourceTypes, event.resource_type) && includes(rule.resourceIds, event.resource_id);
}

function includes(values: string[] | undefined, value: string | undefined) {
  return !values?.length || values.includes(value || '');
}

function appendAll(params: URLSearchParams, name: string, values?: string[]) {
  for (const value of values || []) if (value.trim()) params.append(name, value.trim());
}
