import type { HTMLAttributes } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { cn } from '@/lib/utils';

export {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';

export { Alert, AlertDescription, Badge, Button, Card, Input };

export function EmptyBlock({ text }: { text: string }) {
  return <Card className="emptyBlock shadow-none">{text}</Card>;
}

export function StatusBadge({ status }: { status: string }) {
  return <Badge variant="secondary" className={`statusBadge ${status.toLowerCase().replaceAll('_', '-')}`}>{status || '-'}</Badge>;
}

export function Item({ className, variant, ...props }: HTMLAttributes<HTMLDivElement> & { variant?: 'outline' }) {
  return <Card className={cn('uiItem shadow-none', variant === 'outline' && 'uiItem-outline', className)} {...props} />;
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
