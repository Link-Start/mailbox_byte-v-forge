import type { ReactNode } from 'react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';

export { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle };

export function WorkspacePanel({ children }: { children: ReactNode }) {
  return <main className="workspacePanel">{children}</main>;
}

export function AccountManagementFrame({ title, icon, actions, children }: { title?: string; icon?: ReactNode; actions?: ReactNode; children: ReactNode }) {
  const hasTitle = Boolean(title || icon);
  return <section className="accountFrame"><header className="panelHeader">{hasTitle ? <div className="panelTitle">{icon}{title}</div> : <span />}{actions}</header><div className="panelBody">{children}</div></section>;
}

export function AppDrawer({ open, title, description, icon, children, onOpenChange }: {
  open: boolean;
  title: string;
  description?: string;
  icon?: ReactNode;
  size?: 'wide' | 'default';
  bodyClassName?: string;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}) {
  return <Sheet open={open} onOpenChange={onOpenChange}><SheetContent className="appDrawer"><SheetHeader><SheetTitle>{icon}{title}</SheetTitle>{description && <SheetDescription>{description}</SheetDescription>}</SheetHeader><div className="drawerBody">{children}</div></SheetContent></Sheet>;
}
