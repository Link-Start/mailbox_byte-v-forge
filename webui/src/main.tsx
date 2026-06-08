import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App, ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { createRoot } from 'react-dom/client';
import { MailboxPage } from './dashboard/mailbox-page';
import 'antd/dist/reset.css';
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
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorBgBase: '#f6f3ed',
          colorBgContainer: '#fffefa',
          colorBorder: '#ded8cd',
          colorPrimary: '#7c5c3f',
          colorText: '#26231f',
          colorTextSecondary: '#7d756b',
          borderRadius: 12,
          fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
        }
      }}
    >
      <App>
        <MailboxPage />
      </App>
    </ConfigProvider>
  </QueryClientProvider>
);
