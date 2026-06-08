import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Button } from './ui-basic';
import { cn } from '@/lib/utils';

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
  return <div className={cn('actionButtonGroup', className)}>{actions.map((action) => <ActionButton key={action.id} action={action} />)}</div>;
}

export function ToolbarActionButtons({ actions }: { actions: ToolbarActionDescriptor[] }) {
  return <div className="toolbarActions">{actions.map((action) => <ActionButton key={action.id} action={action} iconOnly />)}</div>;
}

export function RecordActionButtons({ actions }: { actions: RowActionDescriptor[] }) {
  return <div className="rowActionButtons">{actions.map((action) => <ActionButton key={action.id} action={{ ...action, variant: action.kind === 'danger' ? 'destructive' : action.variant }} iconOnly />)}</div>;
}

function ActionButton({ action, iconOnly }: { action: ActionButtonDescriptor; iconOnly?: boolean }) {
  const button = (
    <Button type={action.type || 'button'} form={action.form} variant={action.variant} size={iconOnly ? 'icon' : 'default'} disabled={action.disabled} title={action.label} aria-label={iconOnly ? action.label : undefined} onClick={(event) => { event.stopPropagation(); action.onClick?.(); }}>
      {action.icon}{!iconOnly && action.label}
    </Button>
  );
  return iconOnly ? <Tooltip><TooltipTrigger asChild>{button}</TooltipTrigger><TooltipContent>{action.label}</TooltipContent></Tooltip> : button;
}
