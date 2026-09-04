const DEFAULT_SERVER_URL = 'http://127.0.0.1:4433/';

async function ensureDefaultServerURL(): Promise<string> {
  const data = await chrome.storage.local.get(['histerURL']);
  if (data['histerURL']) {
    return data['histerURL'];
  }
  await chrome.storage.local.set({ histerURL: DEFAULT_SERVER_URL });
  return DEFAULT_SERVER_URL;
}

export { DEFAULT_SERVER_URL, ensureDefaultServerURL };
