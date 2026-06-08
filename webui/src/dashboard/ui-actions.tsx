import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Space, Tooltip } from 'antd';
import { Button } from './ui-basic';
import { cx } from './text';

export type ActionButtonDescriptor = {
  id: string;
  label: string;
  icon?: ReactNode;
  variant?: 'default' | 'outline' | 'destructive' | 'ghost';
  type?: ButtonHTMLAttributes<HTMLButtonElement>['type'];
  form?: string;
  disabled?: boolean;
  onClick?: () => void;
};

export type ToolbarActionDescriptor = ActionButtonDescriptor;
export type RowActionDescriptor = ActionButtonDescriptor & { kind?: 'danger' };

export function ActionButtonGroup({ actions, className }: { actions: ActionButtonDescriptor[]; className?: string }) {
  return <Space className={cx('actionButtonGroup', className)} wrap>{actions.map((action) => <ActionButton key={action.id} action={action} />)}</Space>;
}

export function ToolbarActionButtons({ actions }: { actions: ToolbarActionDescriptor[] }) {
  return <Space className="toolbarActions" size={6}>{actions.map((action) => <ActionButton key={action.id} action={action} iconOnly />)}</Space>;
}

export function RecordActionButtons({ actions }: { actions: RowActionDescriptor[] }) {
  return <Space className="rowActionButtons" size={6}>{actions.map((action) => <ActionButton key={action.id} action={{ ...action, variant: action.kind === 'danger' ? 'destructive' : action.variant }} iconOnly />)}</Space>;
}

function ActionButton({ action, iconOnly }: { action: ActionButtonDescriptor; iconOnly?: boolean }) {
  const button = (
    <Button type={action.type || 'button'} form={action.form} variant={action.variant} disabled={action.disabled} title={action.label} aria-label={iconOnly ? action.label : undefined} onClick={(event) => { event.stopPropagation(); action.onClick?.(); }}>
      {action.icon}{!iconOnly && action.label}
    </Button>
  );
  return iconOnly ? <Tooltip title={action.label}>{button}</Tooltip> : button;
}
