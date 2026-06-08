import type { HTMLAttributes, ReactNode } from 'react';
import { cx } from './text';

export function RecordList({ className, emptyText, children }: HTMLAttributes<HTMLDivElement> & { emptyText: string }) {
  const hasChildren = Array.isArray(children) ? children.length > 0 : !!children;
  return <div className={cx('recordList', className)}>{hasChildren ? children : <div className="emptyBlock">{emptyText}</div>}</div>;
}

export function RecordCard({ selected, onClick, children }: { selected?: boolean; onClick?: () => void; children: ReactNode }) {
  return <article className={cx('recordCard', selected && 'selected', onClick && 'clickable')} onClick={onClick}>{children}</article>;
}

export function RecordMain(props: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('recordMain', props.className)} {...props} />;
}

export function RecordTop(props: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('recordTop', props.className)} {...props} />;
}

export function RecordMeta(props: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('recordMeta', props.className)} {...props} />;
}

export function RecordActions(props: HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('recordActions', props.className)} {...props} />;
}

export function RecordIdentity({ icon, title, subtitle }: { icon: ReactNode; title: ReactNode; subtitle?: ReactNode }) {
  return <div className="recordIdentity"><span className="recordIcon">{icon}</span><div className="min-w-0"><strong className="recordTitle">{title}</strong>{subtitle && <small>{subtitle}</small>}</div></div>;
}
