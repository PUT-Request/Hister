export function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function buildUrlSkipPattern(url: string): string {
  return escapeRegex(url) + '$';
}

export function buildDomainSkipPattern(url: string): string {
  try {
    return escapeRegex(new URL(url).origin);
  } catch {
    return escapeRegex(url);
  }
}
