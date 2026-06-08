import type { HTMLAttributes, ReactNode } from 'react';
import { Drawer } from 'antd';
import { cx } from './text';

export function WorkspacePanel({ children }: { children: ReactNode }) {
  return <main className="workspacePanel">{children}</main>;
}

export function AccountManagementFrame({ title, icon, actions, children }: { title: string; icon?: ReactNode; actions?: ReactNode; children: ReactNode }) {
  return <section className="accountFrame"><header className="panelHeader"><div className="panelTitle">{icon}{title}</div>{actions}</header><div className="panelBody">{children}</div></section>;
}

export function AppDrawer({ open, title, description, icon, children, footer, size = 'default', bodyClassName, onOpenChange }: {
  open: boolean;
  title: string;
  description?: string;
  icon?: ReactNode;
  size?: 'wide' | 'default';
  bodyClassName?: string;
  footer?: ReactNode;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}) {
  return (
    <Drawer
      open={open}
      size={size === 'wide' ? 736 : 460}
      title={<DrawerTitle icon={icon} title={title} description={description} />}
      footer={footer}
      onClose={() => onOpenChange(false)}
      destroyOnHidden
    >
      <div className={cx('drawerBody', bodyClassName)}>{children}</div>
    </Drawer>
  );
}

export function SheetFooter({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('sheetFooter', className)} {...props} />;
}

function DrawerTitle({ icon, title, description }: { icon?: ReactNode; title: string; description?: string }) {
  return (
    <div className="drawerTitle">
      <span className="drawerTitleMain">{icon}{title}</span>
      {description && <small>{description}</small>}
    </div>
  );
}
