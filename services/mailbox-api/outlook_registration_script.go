package main

const outlookOAuthScript = `async ({email, password}) => {
  const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
  const visible = (el) => {
    if (!el) return false;
    const rect = el.getBoundingClientRect();
    const style = getComputedStyle(el);
    return rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none";
  };
  const fill = (el, value) => {
    el.focus();
    const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")?.set;
    if (setter) setter.call(el, value); else el.value = value;
    el.dispatchEvent(new Event("input", {bubbles: true}));
    el.dispatchEvent(new Event("change", {bubbles: true}));
  };
  const clickByText = (pattern) => {
    for (const el of document.querySelectorAll("button,input[type=submit],a,div[role=button]")) {
      const text = (el.innerText || el.textContent || el.value || "").trim();
      if (!el.disabled && visible(el) && pattern.test(text)) {
        el.click();
        return true;
      }
    }
    return false;
  };
  const current = () => ({url: location.href, text: document.body ? document.body.innerText || "" : ""});
  for (let i = 0; i < 180; i++) {
    const state = current();
    if (/[?&](code|error)=/.test(state.url)) return {state: "redirected", url: state.url};
    if (/approve sign in request|help us protect|verify your identity|account.live.com\/abuse/i.test(state.text)) {
      return {state: "needs_manual_verification", url: state.url};
    }
    const emailInput = Array.from(document.querySelectorAll('input[type="email"],input[name="loginfmt"]')).find(visible);
    if (emailInput) {
      fill(emailInput, email);
      await sleep(300);
      clickByText(/^(next|continue)$/i);
      await sleep(1200);
      continue;
    }
    const passwordInput = Array.from(document.querySelectorAll('input[type="password"],input[name="passwd"]')).find(visible);
    if (passwordInput) {
      fill(passwordInput, password);
      await sleep(300);
      clickByText(/^(sign in|next|continue)$/i);
      await sleep(1500);
      continue;
    }
    if (clickByText(/^(no|yes|accept|continue|next)$/i)) {
      await sleep(1200);
      continue;
    }
    await sleep(1000);
  }
  return {state: "timeout", url: location.href};
}`
