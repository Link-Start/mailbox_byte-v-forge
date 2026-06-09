import { useMemo } from 'react';
import { createBrowserRouter, Navigate } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { MailboxInboxRoute, MailboxIndexRoute, MailboxOverviewRoute, MailboxPage } from './mailbox-page';
import { mailboxInboxIndexPath, mailboxIndexPath, mailboxStandaloneBasename } from './mailbox-route-paths';

function createMailboxRouter(basename: string) {
  return createBrowserRouter([
    {
      path: mailboxIndexPath(),
      Component: MailboxPage,
      children: [
        { index: true, element: <Navigate replace to={mailboxInboxIndexPath()} /> },
        { path: 'inbox', Component: MailboxIndexRoute },
        { path: 'inbox/:mailboxEmail', Component: MailboxInboxRoute },
        { path: 'accounts/:mailboxEmail', Component: MailboxOverviewRoute }
      ]
    },
    { path: '*', element: <Navigate replace to={mailboxInboxIndexPath()} /> }
  ], { basename });
}

export function MailboxRouterProvider({ basename = mailboxStandaloneBasename }: { basename?: string }) {
  const router = useMemo(() => createMailboxRouter(basename), [basename]);
  return <RouterProvider router={router} />;
}
