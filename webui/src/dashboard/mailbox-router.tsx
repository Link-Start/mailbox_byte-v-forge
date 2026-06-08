import { createBrowserRouter, Navigate } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { MailboxPage } from './mailbox-page';
import { mailboxIndexPath, mailboxRouterBasename } from './mailbox-route-paths';

const router = createBrowserRouter([
  { index: true, element: <MailboxPage /> },
  { path: 'mailboxes/:mailboxEmail', element: <MailboxPage detailTab="overview" /> },
  { path: 'mailboxes/:mailboxEmail/inbox', element: <MailboxPage detailTab="inbox" /> },
  { path: '*', element: <Navigate replace to={mailboxIndexPath()} /> }
], { basename: mailboxRouterBasename });

export function MailboxRouterProvider() {
  return <RouterProvider router={router} />;
}
