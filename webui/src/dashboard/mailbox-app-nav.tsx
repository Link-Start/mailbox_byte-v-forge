import { Inbox, UsersRound } from 'lucide-react';
import { NavLink, useLocation } from 'react-router';
import { mailboxAccountsIndexPath, mailboxInboxIndexPath } from './mailbox-route-paths';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';

const navItems = [
  { id: 'inbox', label: '邮件', to: mailboxInboxIndexPath(), icon: <Inbox className="size-4" /> },
  { id: 'accounts', label: '账户', to: mailboxAccountsIndexPath(), icon: <UsersRound className="size-4" /> },
];

export function MailboxAppNav() {
  const { search } = useLocation();
  const persistentSearch = persistentMailboxPanelSearch(search);
  return (
    <nav className="mailboxAppNav" aria-label="Mailbox">
      <div className="mailboxAppMark">M</div>
      <div className="mailboxAppNavItems">
        {navItems.map((item) => (
          <NavLink key={item.id} to={`${item.to}${persistentSearch}`} className={({ isActive }) => `mailboxAppNavItem ${isActive ? 'active' : ''}`} title={item.label}>
            {item.icon}
            <span>{item.label}</span>
          </NavLink>
        ))}
      </div>
    </nav>
  );
}
