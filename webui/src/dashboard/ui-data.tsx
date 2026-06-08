import { CopyOutlined } from '@ant-design/icons';
import { Descriptions, Space } from 'antd';
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
  return (
    <Descriptions
      className="kvList"
      bordered
      column={1}
      size="small"
      items={items.map((item) => ({
        key: item.id,
        label: item.label,
        children: (
          <Space className="kvValue" size={8}>
            <span className={item.mono ? 'font-mono' : ''}>{item.masked ? mask(item.value) : item.value}</span>
            <Button variant="ghost" size="icon" disabled={item.copyDisabled} aria-label={`复制${item.label}`} onClick={() => onCopy(item.label, item.copyValue || item.value)}><CopyOutlined /></Button>
          </Space>
        )
      }))}
    />
  );
}

export function CursorPager({ itemCount, pageSize, hasNext, loading, onNext }: {
  itemCount: number;
  pageSize: number;
  hasNext?: boolean;
  loading?: boolean;
  onNext: () => void;
}) {
  if (!hasNext && itemCount < pageSize) return null;
  return <Space className="cursorPager"><span>{itemCount} 条</span>{hasNext && <Button variant="outline" size="sm" disabled={loading} onClick={onNext}>{loading ? '加载中' : '加载更多'}</Button>}</Space>;
}
