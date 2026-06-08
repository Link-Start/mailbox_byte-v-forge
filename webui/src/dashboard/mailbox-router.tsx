import { useMemo } from 'react';
import { createBrowserRouter, Navigate } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { MailboxInboxRoute, MailboxOverviewRoute, MailboxPage } from './mailbox-page';
import { mailboxIndexPath, mailboxStandaloneBasename } from './mailbox-route-paths';

function createMailboxRouter(basename: string) {
  return createBrowserRouter([
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
  ], { basename });
}

export function MailboxRouterProvider({ basename = mailboxStandaloneBasename }: { basename?: string }) {
  const router = useMemo(() => createMailboxRouter(basename), [basename]);
  return <RouterProvider router={router} />;
}
