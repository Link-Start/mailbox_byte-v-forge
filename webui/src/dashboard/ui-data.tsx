import { Copy } from 'lucide-react';
import { Button } from './ui-basic';
import { mask } from './text';

export type KVDescriptor = {
  id: string;
  label: string;
  value: string;
  copyValue?: string;
  copyDisabled?: boolean;
  masked?: boolean;
  mono?: boolean;
};

export function KVList({ items, onCopy }: { items: KVDescriptor[]; onCopy: (label: string, value: string) => void }) {
  return <dl className="kvList">{items.map((item) => <div className="kvRow" key={item.id}><dt>{item.label}</dt><dd className={item.mono ? 'font-mono' : ''}>{item.masked ? mask(item.value) : item.value}</dd><Button variant="ghost" size="icon" disabled={item.copyDisabled} aria-label={`复制${item.label}`} onClick={() => onCopy(item.label, item.copyValue || item.value)}><Copy size={14} /></Button></div>)}</dl>;
}

export function CursorPager({ itemCount, pageSize, hasNext, loading, onNext }: {
  itemCount: number;
  pageSize: number;
  hasNext?: boolean;
  loading?: boolean;
  onNext: () => void;
}) {
  if (!hasNext && itemCount < pageSize) return null;
  return <div className="cursorPager"><span>{itemCount} 条</span>{hasNext && <Button variant="outline" size="sm" disabled={loading} onClick={onNext}>{loading ? '加载中' : '加载更多'}</Button>}</div>;
}
