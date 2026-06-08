import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from 'react';
import { cx } from './text';

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'default' | 'outline' | 'destructive' | 'ghost';
  size?: 'default' | 'sm' | 'icon';
};

export function Button({ className, variant = 'default', size = 'default', ...props }: ButtonProps) {
  return <button className={cx('uiButton', `uiButton-${variant}`, `uiButton-${size}`, className)} {...props} />;
}

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('uiCard', className)} {...props} />;
}

export function Badge({ className, variant = 'default', ...props }: HTMLAttributes<HTMLSpanElement> & { variant?: 'default' | 'outline' | 'secondary' }) {
  return <span className={cx('uiBadge', `uiBadge-${variant}`, className)} {...props} />;
}

export function Alert({ className, variant = 'default', ...props }: HTMLAttributes<HTMLDivElement> & { variant?: 'default' | 'destructive' }) {
  return <div role="alert" className={cx('uiAlert', `uiAlert-${variant}`, className)} {...props} />;
}

export function AlertDescription({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('text-sm', className)} {...props} />;
}

export function EmptyBlock({ text }: { text: string }) {
  return <div className="emptyBlock">{text}</div>;
}

export function StatusBadge({ status }: { status: string }) {
  return <Badge variant="secondary" className={`statusBadge ${status.toLowerCase().replaceAll('_', '-')}`}>{status || '-'}</Badge>;
}

export function Item({ className, variant, ...props }: HTMLAttributes<HTMLDivElement> & { variant?: 'outline' }) {
  return <div className={cx('uiItem', variant === 'outline' && 'uiItem-outline', className)} {...props} />;
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

export function ThemeProvider({ children }: { children: ReactNode; defaultTheme?: string; storageKey?: string }) {
  return <>{children}</>;
}
