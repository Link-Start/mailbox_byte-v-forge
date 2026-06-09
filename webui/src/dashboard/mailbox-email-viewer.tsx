import { useMemo } from 'react';
import DOMPurify from 'dompurify';

const emptyBodyText = '无正文';

type EmailViewerContent =
  | { kind: 'html'; srcDoc: string }
  | { kind: 'text'; text: string };

export function MailboxEmailViewer({ htmlBody, textBody }: {
  htmlBody?: string;
  textBody?: string;
}) {
  const content = useMemo(() => emailViewerContent(htmlBody || '', textBody || ''), [htmlBody, textBody]);
  if (content.kind === 'text') {
    return <pre className="mailboxPlainTextBody">{content.text}</pre>;
  }
  return (
    <iframe
      className="mailboxEmailFrame"
      title="邮件正文"
      srcDoc={content.srcDoc}
      sandbox="allow-popups allow-popups-to-escape-sandbox"
      referrerPolicy="no-referrer"
    />
  );
}

function emailViewerContent(htmlBody: string, textBody: string): EmailViewerContent {
  const html = String(htmlBody || '').trim();
  if (html) {
    const sanitized = sanitizeEmailHTML(html);
    if (hasRenderableHTML(sanitized)) {
      return { kind: 'html', srcDoc: emailFrameDocument(sanitized) };
    }
  }
  const text = cleanTextBody(textBody);
  return { kind: 'text', text: text || emptyBodyText };
}

function emailFrameDocument(bodyHTML: string) {
  return `<!doctype html><html><head><meta charset="utf-8"><base target="_blank">${frameStyle()}</head><body>${bodyHTML}</body></html>`;
}

function sanitizeEmailHTML(value: string) {
  return DOMPurify.sanitize(value, {
    FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button', 'textarea', 'select'],
  });
}

function hasRenderableHTML(value: string) {
  const doc = new DOMParser().parseFromString(`<body>${value}</body>`, 'text/html');
  if (cleanVisibleText(doc.body.textContent || '')) return true;
  return !!doc.body.querySelector('img[src], svg, canvas, video, audio');
}

function cleanVisibleText(value: string) {
  return String(value || '')
    .replace(/[\u200B-\u200D\uFEFF]/g, '')
    .replace(/\u00A0/g, ' ')
    .trim();
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
