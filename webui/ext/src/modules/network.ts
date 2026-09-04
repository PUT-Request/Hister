const FETCHABLE_FAVICON_PROTOCOLS = new Set(['http:', 'https:']);

function isFetchableFaviconURL(rawURL: string): boolean {
  try {
    const url = new URL(rawURL);
    return FETCHABLE_FAVICON_PROTOCOLS.has(url.protocol);
  } catch {
    return false;
  }
}

async function fetchFavicon(url: string): Promise<string> {
  if (!isFetchableFaviconURL(url)) {
    return '';
  }
  const response = await fetch(url);
  let iconBytes = await response.blob();
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      resolve(typeof reader.result === 'string' ? reader.result : '');
    };
    reader.onerror = () => resolve('');
    reader.readAsDataURL(iconBytes);
  });
}

const unauthorizedStatuses = new Set([401, 403]);

type StoredAuth = {
  cookieHeader: string;
  accessToken: string;
};

async function getStoredAuth(): Promise<StoredAuth> {
  return new Promise((resolve) => {
    chrome.storage.local.get(['histerCookies', 'histerToken'], (data) => {
      resolve({
        cookieHeader: data['histerCookies'] || '',
        accessToken: data['histerToken'] || '',
      });
    });
  });
}

function serializeCookies(cookies: chrome.cookies.Cookie[]): string {
  return cookies.map((cookie) => `${cookie.name}=${cookie.value}`).join('; ');
}

function normalizeCookieHeader(cookieHeader: string): string {
  return cookieHeader
    .split(';')
    .map((cookie) => cookie.trim())
    .filter(Boolean)
    .sort()
    .join(';');
}

async function getBrowserCookies(url: string): Promise<string> {
  return new Promise((resolve, reject) => {
    chrome.cookies.getAll({ url }, (cookies) => {
      const error = chrome.runtime.lastError;
      if (error) {
        reject(new Error(error.message ?? 'Failed to read browser cookies'));
        return;
      }
      resolve(serializeCookies(cookies));
    });
  });
}

async function syncServerCookies(url: string): Promise<string> {
  const cookieHeader = await getBrowserCookies(url);
  await chrome.storage.local.set({ histerCookies: cookieHeader });
  return cookieHeader;
}

async function refreshServerCookies(
  url: string,
  rejectedCookieHeader: string,
): Promise<string | null> {
  try {
    const cookieHeader = await getBrowserCookies(url);
    if (normalizeCookieHeader(cookieHeader) === normalizeCookieHeader(rejectedCookieHeader)) {
      return null;
    }
    await chrome.storage.local.set({ histerCookies: cookieHeader });
    return cookieHeader;
  } catch (_) {
    return null;
  }
}

async function fetchWithCookies(
  url: string,
  method: string,
  headers: Record<string, string>,
  body: BodyInit | undefined,
  cookieHeader: string,
): Promise<Response> {
  const requestHeaders = { ...headers };
  const hasCustomCookieHeader = Object.keys(requestHeaders).some(
    (name) => name.toLowerCase() === 'cookie',
  );
  if (cookieHeader && !hasCustomCookieHeader) {
    requestHeaders['Cookie'] = cookieHeader;
  }
  return fetch(url, {
    method,
    headers: requestHeaders,
    body,
    credentials: 'include',
  });
}

async function fetchAPI(
  url: string,
  options: {
    method?: string;
    body?: unknown;
    formData?: Record<string, string>;
    customHeaders?: { name: string; value: string }[];
    accessToken?: string;
  } = {},
): Promise<Response> {
  const storedAuth = await getStoredAuth();
  const cookieHeader = storedAuth.cookieHeader;
  const accessToken = (
    options.accessToken === undefined ? storedAuth.accessToken : options.accessToken
  ).trim();
  const headers: Record<string, string> = {};

  if (options.body !== undefined) {
    headers['Content-type'] = 'application/json; charset=UTF-8';
  } else if (options.formData !== undefined) {
    headers['Content-type'] = 'application/x-www-form-urlencoded';
  }
  for (const h of options.customHeaders ?? []) {
    if (h.name) headers[h.name] = h.value || '';
  }
  if (accessToken) {
    for (const name of Object.keys(headers)) {
      if (name.toLowerCase() === 'x-access-token') {
        delete headers[name];
      }
    }
    headers['X-Access-Token'] = accessToken;
  }

  let fetchBody: BodyInit | undefined;
  if (options.body !== undefined) {
    fetchBody = JSON.stringify(options.body);
  } else if (options.formData !== undefined) {
    fetchBody = new URLSearchParams(options.formData).toString();
  }

  const method = options.method ?? (fetchBody !== undefined ? 'POST' : 'GET');
  const response = await fetchWithCookies(url, method, headers, fetchBody, cookieHeader);
  if (!unauthorizedStatuses.has(response.status)) {
    return response;
  }

  const refreshedCookieHeader = await refreshServerCookies(url, cookieHeader);
  if (refreshedCookieHeader === null) {
    return response;
  }
  return fetchWithCookies(url, method, headers, fetchBody, refreshedCookieHeader);
}

async function sendPageData(url, doc, customHeaders = []) {
  try {
    doc['favicon'] = await fetchFavicon(doc.faviconURL);
  } catch (e) {
    doc['favicon'] = '';
  }
  return sendResult(url, doc, customHeaders);
}

async function sendResult(url, res, customHeaders = []) {
  return fetchAPI(url, { body: res, customHeaders });
}

async function sendPDFData(
  url: string,
  doc: Record<string, unknown>,
  pdfBase64: string,
  customHeaders: { name: string; value: string }[] = [],
): Promise<Response> {
  return fetchAPI(url, { body: { document: doc, pdf: pdfBase64 }, customHeaders });
}

export { fetchAPI, sendPageData, sendResult, sendPDFData, syncServerCookies };
