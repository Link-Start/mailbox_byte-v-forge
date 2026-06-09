import { Inbox, Mail, UsersRound } from 'lucide-react';
import { NavLink, useLocation } from 'react-router';
import { Button, Tooltip, TooltipContent, TooltipTrigger } from './dashboard-kit';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';
import { mailboxAccountsIndexPath, mailboxInboxIndexPath } from './mailbox-route-paths';

const navItems = [
  { id: 'inbox', label: '邮件', to: mailboxInboxIndexPath(), icon: <Inbox className="size-4" /> },
  { id: 'accounts', label: '账户', to: mailboxAccountsIndexPath(), icon: <UsersRound className="size-4" /> },
];

export function MailboxAppNav() {
  const { pathname, search } = useLocation();
  const persistentSearch = persistentMailboxPanelSearch(search);
  return (
    <nav className="mailboxAppNav" aria-label="Mailbox">
      <div className="mailboxAppMark" aria-label="Mailbox"><Mail className="size-5" /></div>
      <div className="mailboxAppNavItems">
        {navItems.map((item) => (
          <Tooltip key={item.id}>
            <TooltipTrigger asChild>
              <Button asChild variant={pathname.startsWith(item.to) ? 'secondary' : 'ghost'} size="icon" className="mailboxAppNavButton">
                <NavLink to={`${item.to}${persistentSearch}`} aria-label={item.label}>{item.icon}</NavLink>
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{item.label}</TooltipContent>
          </Tooltip>
        ))}
      </div>
    </nav>
  );
}
