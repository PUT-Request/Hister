import GithubSlugger from 'github-slugger';
import type { RequestHandler } from '@sveltejs/kit';
import type { DocsSearchEntry, DocsSearchIndex } from '$lib/docs-search';

const sources = import.meta.glob('../../../content/docs/*.md', {
  eager: true,
  query: '?raw',
  import: 'default',
}) as Record<string, string>;

interface Frontmatter {
  title: string;
  description: string;
  body: string;
}

function unquote(value: string): string {
  const trimmed = value.trim();
  if (trimmed.startsWith("'") && trimmed.endsWith("'")) {
    return trimmed.slice(1, -1).replace(/''/g, "'");
  }
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    try {
      return JSON.parse(trimmed) as string;
    } catch {
      return trimmed.slice(1, -1);
    }
  }
  return trimmed;
}

function parseFrontmatter(source: string): Frontmatter {
  const match = source.match(/^---\s*\n([\s\S]*?)\n---\s*\n?/);
  const rawFrontmatter = match?.[1] ?? '';
  const field = (name: string) => {
    const value = rawFrontmatter.match(new RegExp(`^${name}:\\s*(.+)$`, 'm'))?.[1] ?? '';
    return unquote(value);
  };

  return {
    title: field('title'),
    description: field('description'),
    body: match ? source.slice(match[0].length) : source,
  };
}

function plainText(value: string): string {
  return value
    .replace(/<!--([\s\S]*?)-->/g, ' ')
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/<[^>]+>/g, ' ')
    .replace(/https?:\/\/\S+/g, ' ')
    .replace(/[`*~>|{}[\]();,]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function buildEntries(path: string, source: string): DocsSearchEntry[] {
  const slug = path.split('/').pop()?.replace('.md', '') ?? path;
  const { title, description, body } = parseFrontmatter(source);
  const pageTitle = title || slug.replace(/-/g, ' ');
  const slugger = new GithubSlugger();
  const entries: DocsSearchEntry[] = [];
  let heading: string | null = null;
  let headingId: string | null = null;
  let contentLines: string[] = [];
  let fence: string | null = null;

  const addEntry = () => {
    const content = plainText(contentLines.join('\n'));
    if (!heading && !content && !description) return;
    entries.push({
      url: `/docs/${slug}${headingId ? `#${headingId}` : ''}`,
      pageTitle,
      description,
      heading,
      content: content || description,
    });
  };

  for (const line of body.split('\n')) {
    const fenceMatch = line.match(/^\s*(```+|~~~+)/);
    if (fenceMatch) {
      if (!fence) fence = fenceMatch[1][0];
      else if (fence === fenceMatch[1][0]) fence = null;
      continue;
    }

    const headingMatch = !fence ? line.match(/^\s*#{2,6}\s+(.+?)\s*#*\s*$/) : null;
    if (headingMatch) {
      addEntry();
      heading = plainText(headingMatch[1]);
      headingId = slugger.slug(heading);
      contentLines = [];
      continue;
    }

    contentLines.push(line);
  }

  addEntry();
  return entries;
}

const searchIndex: DocsSearchIndex = {
  entries: Object.entries(sources)
    .flatMap(([path, source]) => buildEntries(path, source))
    .sort((a, b) => a.pageTitle.localeCompare(b.pageTitle) || a.url.localeCompare(b.url)),
};

export const prerender = true;

export const GET: RequestHandler = () =>
  new Response(JSON.stringify(searchIndex), {
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      'Cache-Control': 'public, max-age=3600',
    },
  });
