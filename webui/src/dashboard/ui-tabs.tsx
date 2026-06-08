import type { ReactNode } from 'react';
import { Tabs } from 'radix-ui';
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
  return <div className="segmentedControl">{options.map((option) => <button type="button" key={option.value} className={option.value === value ? 'active' : ''} onClick={() => onChange(option.value)}>{option.label}</button>)}</div>;
}

function ManagedTabs({ value, onValueChange, tabs, tabsClassName, tabsListClassName }: ManagedTabsProps) {
  return (
    <Tabs.Root value={value} onValueChange={onValueChange} className={cx('tabsRoot', tabsClassName)}>
      <Tabs.List className={cx('tabsList', tabsListClassName)}>{tabs.map((tab) => <Tabs.Trigger key={tab.value} value={tab.value} className={cx('tabsTrigger', tab.triggerClassName)}>{tab.label}</Tabs.Trigger>)}</Tabs.List>
      {tabs.map((tab) => <Tabs.Content key={tab.value} value={tab.value} className={cx('tabsContent', tab.contentClassName)}>{tab.content}</Tabs.Content>)}
    </Tabs.Root>
  );
}
