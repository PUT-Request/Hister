import type { SearchSortCapabilities } from '$lib/search-schema';

function escapeRegularExpression(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function sortDirectivePattern(sort: SearchSortCapabilities): RegExp | null {
  if (!sort.field || sort.options.length === 0) return null;
  const values = sort.options
    .map((option) => option.value)
    .sort((left, right) => right.length - left.length)
    .map(escapeRegularExpression)
    .join('|');
  return new RegExp(`(?:^|\\s)${escapeRegularExpression(sort.field)}:(${values})(?=\\s|$)`, 'g');
}

function isInsideQuotedText(text: string, offset: number): boolean {
  return [...text.slice(0, offset).matchAll(/(?<!\\)(?:\\\\)*"/g)].length % 2 === 1;
}

export function sortDirectiveFromQuery(text: string, sort: SearchSortCapabilities): string | null {
  const pattern = sortDirectivePattern(sort);
  if (!pattern) return null;
  const matches = [...text.matchAll(pattern)].filter((match) => {
    const tokenOffset = (match.index ?? 0) + match[0].indexOf(`${sort.field}:`);
    return !isInsideQuotedText(text, tokenOffset);
  });
  return matches.at(-1)?.[1] ?? null;
}

export function sortValueFromQuery(text: string, sort: SearchSortCapabilities): string {
  const directive = sortDirectiveFromQuery(text, sort);
  if (directive === null) return '';
  return sort.options.find((option) => option.value === directive)?.default ? '' : directive;
}

export function removeSortDirectives(text: string, sort: SearchSortCapabilities): string {
  const pattern = sortDirectivePattern(sort);
  if (!pattern) return text.trim();
  return text
    .replace(pattern, (match, _sort, offset) => {
      const tokenOffset = offset + match.indexOf(`${sort.field}:`);
      return isInsideQuotedText(text, tokenOffset) ? match : '';
    })
    .trim();
}

export function replaceSortDirective(
  text: string,
  value: string,
  sort: SearchSortCapabilities,
): string {
  const definition = sort.options.find((option) => option.value === value);
  if (value && !definition) return text;
  const remaining = removeSortDirectives(text, sort);
  if (!value || definition?.default) return remaining;
  return `${remaining ? `${remaining} ` : ''}${sort.field}:${value}`;
}
