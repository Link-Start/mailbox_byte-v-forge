import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRoot } from 'react-dom/client';
import { Toaster, TooltipProvider } from './dashboard/dashboard-kit';
import { MailboxRouterProvider } from './dashboard/mailbox-router';
import './styles.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5000
    }
  }
});

createRoot(document.getElementById('root')!).render(
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <MailboxRouterProvider />
      <Toaster richColors />
    </TooltipProvider>
  </QueryClientProvider>
);
