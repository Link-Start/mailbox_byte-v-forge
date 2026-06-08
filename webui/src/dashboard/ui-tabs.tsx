import type { ReactNode } from 'react';
import { Segmented, Tabs } from 'antd';
import { cx } from './text';

export type TabDescriptor = {
  value: string;
  label: string;
  content: ReactNode;
  triggerClassName?: string;
  contentClassName?: string;
};

type ManagedTabsProps = {
  value: string;
  onValueChange: (value: string) => void;
  tabs: TabDescriptor[];
  tabsClassName?: string;
  tabsListClassName?: string;
  tabsListVariant?: string;
};

export function PanelTabs(props: ManagedTabsProps) {
  return <ManagedTabs {...props} />;
}

export function ContentTabs(props: ManagedTabsProps) {
  return <ManagedTabs {...props} />;
}

export function SegmentedControl<T extends string>({ value, options, onChange }: { value: T; options: { value: T; label: string }[]; onChange: (value: T) => void }) {
  return <Segmented className="segmentedControl" block value={value} options={options} onChange={(next) => onChange(next as T)} />;
}

function ManagedTabs({ value, onValueChange, tabs, tabsClassName, tabsListClassName }: ManagedTabsProps) {
  return (
    <Tabs
      activeKey={value}
      onChange={onValueChange}
      className={cx('tabsRoot', tabsClassName, tabsListClassName)}
      items={tabs.map((tab) => ({
        key: tab.value,
        label: <span className={tab.triggerClassName}>{tab.label}</span>,
        children: <div className={cx('tabsContent', tab.contentClassName)}>{tab.content}</div>
      }))}
    />
  );
}
