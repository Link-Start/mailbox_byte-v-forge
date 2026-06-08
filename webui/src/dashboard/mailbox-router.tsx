import { createBrowserRouter, Navigate } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { MailboxInboxRoute, MailboxOverviewRoute, MailboxPage } from './mailbox-page';
import { mailboxIndexPath, mailboxRouterBasename } from './mailbox-route-paths';

const router = createBrowserRouter([
  {
    path: mailboxIndexPath(),
    Component: MailboxPage,
    children: [
      { index: true, element: null },
      { path: 'mailboxes/:mailboxEmail', Component: MailboxOverviewRoute },
      { path: 'mailboxes/:mailboxEmail/inbox', Component: MailboxInboxRoute }
    ]
  },
  { path: '*', element: <Navigate replace to={mailboxIndexPath()} /> }
], { basename: mailboxRouterBasename });

export function MailboxRouterProvider() {
  return <RouterProvider router={router} />;
}
