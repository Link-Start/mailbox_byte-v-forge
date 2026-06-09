import { useMemo } from 'react';
import DOMPurify from 'dompurify';

export function MailboxEmailViewer({ htmlBody, textBody }: {
  htmlBody?: string;
  textBody?: string;
}) {
  const srcDoc = useMemo(() => emailFrameDocument(htmlBody || '', textBody || ''), [htmlBody, textBody]);
  return (
    <iframe
      className="mailboxEmailFrame"
      title="邮件正文"
      srcDoc={srcDoc}
      sandbox="allow-popups allow-popups-to-escape-sandbox"
      referrerPolicy="no-referrer"
    />
  );
}

function emailFrameDocument(htmlBody: string, textBody: string) {
  return `<!doctype html><html><head><meta charset="utf-8"><base target="_blank">${frameStyle()}</head><body>${emailBodyHTML(htmlBody, textBody)}</body></html>`;
}

function emailBodyHTML(htmlBody: string, textBody: string) {
  const html = String(htmlBody || '').trim();
  if (html) return DOMPurify.sanitize(html, {
    FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button', 'textarea', 'select'],
  });
  const text = cleanTextBody(textBody);
  return `<pre class="plainText">${escapeHTML(text || '-')}</pre>`;
}

function frameStyle() {
  return `<style>
    :root{color-scheme:light;background:#fff;color:#111827;font:14px/1.58 ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
    html,body{margin:0;min-height:100%;background:#fff}
    body{box-sizing:border-box;padding:0 2px 24px;overflow-wrap:anywhere}
    *{max-width:100%;box-sizing:border-box}
    table{width:100%!important;table-layout:auto;border-collapse:collapse}
    img{max-width:100%;height:auto}
    a{color:#2563eb;text-decoration:underline}
    p,ul,ol,blockquote,table{margin-block:0 12px}
    .plainText{margin:0;white-space:pre-wrap;font:inherit;color:inherit}
  </style>`;
}

function cleanTextBody(value: string) {
  return String(value || '')
    .replace(/\r\n/g, '\n')
    .replace(/([^\s<][^<\n]{0,120})<https?:\/\/[^>\s]+>/g, '$1')
    .replace(/<https?:\/\/[^>\s]+>/g, '')
    .replace(/\n{3,}/g, '\n\n')
    .trim();
}

function escapeHTML(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
