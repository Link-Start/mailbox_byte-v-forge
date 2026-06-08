import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRoot } from 'react-dom/client';
import { ThemeProvider, TooltipProvider } from './dashboard/dashboard-kit';
import { MailboxPage } from './dashboard/mailbox-page';
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
    <ThemeProvider defaultTheme="system" storageKey="mailbox-dashboard-theme">
      <TooltipProvider>
        <MailboxPage />
      </TooltipProvider>
    </ThemeProvider>
  </QueryClientProvider>
);
