import type { ReactNode } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { cn } from '@/lib/utils';

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
  tabsListVariant?: 'default' | 'line' | string;
};

export function PanelTabs(props: ManagedTabsProps) {
  return <ManagedTabs {...props} />;
}

export function ContentTabs(props: ManagedTabsProps) {
  return <ManagedTabs {...props} />;
}

export function SegmentedControl<T extends string>({ value, options, onChange }: { value: T; options: { value: T; label: string }[]; onChange: (value: T) => void }) {
  return <ToggleGroup type="single" value={value} className="segmentedControl" onValueChange={(next) => { if (next) onChange(next as T); }}>{options.map((option) => <ToggleGroupItem key={option.value} value={option.value}>{option.label}</ToggleGroupItem>)}</ToggleGroup>;
}

function ManagedTabs({ value, onValueChange, tabs, tabsClassName, tabsListClassName, tabsListVariant }: ManagedTabsProps) {
  return (
    <Tabs value={value} onValueChange={onValueChange} className={cn('tabsRoot', tabsClassName)}>
      <TabsList variant={tabsListVariant === 'line' ? 'line' : 'default'} className={cn('tabsList', tabsListClassName)}>{tabs.map((tab) => <TabsTrigger key={tab.value} value={tab.value} className={cn('tabsTrigger', tab.triggerClassName)}>{tab.label}</TabsTrigger>)}</TabsList>
      {tabs.map((tab) => <TabsContent key={tab.value} value={tab.value} className={cn('tabsContent', tab.contentClassName)}>{tab.content}</TabsContent>)}
    </Tabs>
  );
}
