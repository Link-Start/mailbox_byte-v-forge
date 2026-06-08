import type { HTMLAttributes } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { cn } from '@/lib/utils';

export { Alert, AlertDescription, Badge, Button, Card };

export function EmptyBlock({ text }: { text: string }) {
  return <div className="emptyBlock">{text}</div>;
}

export function StatusBadge({ status }: { status: string }) {
  return <Badge variant="secondary" className={`statusBadge ${status.toLowerCase().replaceAll('_', '-')}`}>{status || '-'}</Badge>;
}

export function Item({ className, variant, ...props }: HTMLAttributes<HTMLDivElement> & { variant?: 'outline' }) {
  return <div className={cn('uiItem', variant === 'outline' && 'uiItem-outline', className)} {...props} />;
}

export function ItemContent({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('uiItemContent', className)} {...props} />;
}

export function ItemTitle({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('uiItemTitle', className)} {...props} />;
}

export function ItemDescription({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('uiItemDescription', className)} {...props} />;
}
