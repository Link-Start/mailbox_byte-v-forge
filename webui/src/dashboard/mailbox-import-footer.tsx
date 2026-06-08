import { Plus } from 'lucide-react';
import { ActionButtonGroup } from './dashboard-kit';
import type { ActionButtonDescriptor } from './dashboard-kit';

export function MailboxImportFooter({ busy, form, disabled, onClose }: {
  busy: boolean;
  form: string;
  disabled: boolean;
  onClose: () => void;
}) {
  return <ActionButtonGroup className="grid gap-2" actions={footerActions({ busy, form, disabled, onClose })} />;
}

function footerActions({ busy, form, disabled, onClose }: {
  busy: boolean;
  form: string;
  disabled: boolean;
  onClose: () => void;
}): ActionButtonDescriptor[] {
  return [{
    id: 'close',
    label: '关闭',
    variant: 'outline',
    onClick: onClose,
  }, {
    id: 'submit',
    label: '添加',
    icon: <Plus />,
    type: 'submit',
    form,
    disabled: busy || disabled,
  }];
}
