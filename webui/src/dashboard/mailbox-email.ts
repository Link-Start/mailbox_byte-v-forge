import { normalizeUiEmail } from './email-utils';

export function domainForEmail(email: string) {
  const [, domain = ''] = normalizeUiEmail(email).split('@');
  return domain;
}
