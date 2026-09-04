interface TimeFilter {
  comparison: string;
  value: string;
}

function escapeRegularExpression(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function timeFilterPattern(field: string): RegExp | null {
  if (!field) return null;
  return new RegExp(
    `(?:^|\\s)${escapeRegularExpression(field)}:(<=|>=|<|>)(\\d+[smhdwSMHDW]|\\d{4}-\\d{2}-\\d{2})(?=\\s|$)`,
    'g',
  );
}

export function shiftISODate(value: string, days: number): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return '';
  const date = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== value) return '';
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

function isValidTimeValue(value: string): boolean {
  return /^\d+[smhdw]$/i.test(value) || shiftISODate(value, 0) !== '';
}

function isInsideQuotedText(text: string, offset: number): boolean {
  return [...text.slice(0, offset).matchAll(/(?<!\\)(?:\\\\)*"/g)].length % 2 === 1;
}

export function timeFilters(text: string, field: string): TimeFilter[] {
  const pattern = timeFilterPattern(field);
  if (!pattern) return [];
  return [...text.matchAll(pattern)]
    .filter((match) => {
      const tokenOffset = (match.index ?? 0) + match[0].indexOf(`${field}:`);
      return !isInsideQuotedText(text, tokenOffset) && isValidTimeValue(match[2]);
    })
    .map((match) => ({ comparison: match[1], value: match[2] }));
}

export function removeTimeFilters(text: string, field: string): string {
  const pattern = timeFilterPattern(field);
  if (!pattern) return text.trim();
  return text
    .replace(pattern, (match, _comparison, value, offset) => {
      const tokenOffset = offset + match.indexOf(`${field}:`);
      return !isInsideQuotedText(text, tokenOffset) && isValidTimeValue(value) ? ' ' : match;
    })
    .replace(/\s+/g, ' ')
    .trim();
}

export function replaceTimeFilters(text: string, field: string, filters: string[]): string {
  if (!field) return text;
  const remaining = removeTimeFilters(text, field);
  const tokens = filters.map((filter) => `${field}:${filter}`);
  return [...(remaining ? [remaining] : []), ...tokens].join(' ');
}

export function customDatesFromQuery(text: string, field: string): { from: string; to: string } {
  let from = '';
  let to = '';
  for (const filter of timeFilters(text, field)) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(filter.value)) continue;
    if (filter.comparison === '>=') from = filter.value;
    if (filter.comparison === '<') to = shiftISODate(filter.value, -1);
  }
  return { from, to };
}
