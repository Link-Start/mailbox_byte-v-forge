import type { HTMLAttributes, ReactNode } from 'react';
import { Dialog } from 'radix-ui';
import { X } from 'lucide-react';
import { Button } from './ui-basic';
import { cx } from './text';

export function WorkspacePanel({ children }: { children: ReactNode }) {
  return <main className="workspacePanel">{children}</main>;
}

export function AccountManagementFrame({ title, icon, actions, children }: { title: string; icon?: ReactNode; actions?: ReactNode; children: ReactNode }) {
  return <section className="accountFrame"><header className="panelHeader"><div className="panelTitle">{icon}{title}</div>{actions}</header><div className="panelBody">{children}</div></section>;
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

export const Sheet = Dialog.Root;

export function SheetContent({ className, children }: HTMLAttributes<HTMLDivElement>) {
  return <Dialog.Portal><Dialog.Overlay className="sheetOverlay" /><Dialog.Content className={cx('sheetContent', className)}>{children}<Dialog.Close asChild><Button className="sheetClose" variant="ghost" aria-label="关闭"><X size={16} /></Button></Dialog.Close></Dialog.Content></Dialog.Portal>;
}

export function SheetHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('sheetHeader', className)} {...props} />;
}

export function SheetFooter({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('sheetFooter', className)} {...props} />;
}

export function SheetTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return <Dialog.Title className={cx('sheetTitle', className)} {...props} />;
}

export function SheetDescription({ className, ...props }: HTMLAttributes<HTMLParagraphElement>) {
  return <Dialog.Description className={cx('sheetDescription', className)} {...props} />;
}
