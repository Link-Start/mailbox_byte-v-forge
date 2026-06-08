import type { ButtonHTMLAttributes, HTMLAttributes } from 'react';
import { Alert as AntAlert, Button as AntButton, Card as AntCard, Empty, List, Tag } from 'antd';
import type { ButtonProps as AntButtonProps } from 'antd';
import { cx } from './text';

export type ButtonProps = Omit<AntButtonProps, 'type' | 'size' | 'variant'> & {
  variant?: 'default' | 'outline' | 'destructive' | 'ghost';
  size?: 'default' | 'sm' | 'icon';
  type?: ButtonHTMLAttributes<HTMLButtonElement>['type'];
};

export function Button({ className, variant = 'default', size = 'default', type = 'button', children, ...props }: ButtonProps) {
  return (
    <AntButton
      {...props}
      htmlType={type}
      type={buttonType(variant)}
      danger={variant === 'destructive'}
      size={size === 'sm' ? 'small' : 'middle'}
      className={cx(size === 'icon' && 'iconButton', className)}
    >
      {children}
    </AntButton>
  );
}

export function Card({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <AntCard className="mailboxCard" styles={{ body: { padding: 0 } }}>
      <div className={className} {...props}>{children}</div>
    </AntCard>
  );
}

export function Badge({ className, variant = 'default', ...props }: HTMLAttributes<HTMLSpanElement> & { variant?: 'default' | 'outline' | 'secondary' }) {
  return <Tag className={cx('mailboxTag', `mailboxTag-${variant}`, className)} {...props} />;
}

export function Alert({ className, variant = 'default', children }: HTMLAttributes<HTMLDivElement> & { variant?: 'default' | 'destructive' }) {
  return <AntAlert className={className} type={variant === 'destructive' ? 'error' : 'info'} showIcon message={children} />;
}

export function AlertDescription({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('text-sm', className)} {...props} />;
}

export function EmptyBlock({ text }: { text: string }) {
  return <Empty className="emptyBlock" image={Empty.PRESENTED_IMAGE_SIMPLE} description={text} />;
}

export function StatusBadge({ status }: { status: string }) {
  return <Badge variant="secondary" className={`statusBadge ${status.toLowerCase().replaceAll('_', '-')}`}>{status || '-'}</Badge>;
}

export function Item({ className, variant, ...props }: HTMLAttributes<HTMLDivElement> & { variant?: 'outline' }) {
  return <List.Item className={cx('uiItem', variant === 'outline' && 'uiItem-outline', className)} {...props} />;
}

export function ItemContent({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('uiItemContent', className)} {...props} />;
}

export function ItemTitle({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('uiItemTitle', className)} {...props} />;
}

export function ItemDescription({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('uiItemDescription', className)} {...props} />;
}

function buttonType(variant: ButtonProps['variant']) {
  if (variant === 'ghost') return 'text';
  if (variant === 'outline' || variant === 'destructive') return 'default';
  return 'primary';
}
