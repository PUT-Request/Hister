import type { RequestHandler } from '@sveltejs/kit';
import { SITE_URL } from '$lib/config';

interface ContentMetadata {
  date?: string;
  draft?: boolean;
}

interface SitemapEntry {
  path: string;
  lastModified?: string;
}

const docs = import.meta.glob('../../content/docs/*.md', { eager: true });
const posts = import.meta.glob('../../content/posts/*.md', { eager: true });

const staticEntries: SitemapEntry[] = [
  { path: '/' },
  { path: '/datasets' },
  { path: '/docs' },
  { path: '/posts' },
  { path: '/support' },
];

function contentEntries(modules: Record<string, unknown>, routePrefix: string): SitemapEntry[] {
  return Object.entries(modules)
    .map(([path, mod]) => {
      const slug = path.split('/').pop()?.replace('.md', '') ?? path;
      const metadata = (mod as { metadata?: ContentMetadata }).metadata;
      return {
        path: `${routePrefix}/${slug}`,
        lastModified: metadata?.date,
        draft: metadata?.draft ?? false,
      };
    })
    .filter((entry) => !entry.draft)
    .map(({ path, lastModified }) => ({ path, lastModified }));
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

function formatLastModified(value?: string): string | null {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date.toISOString();
}

export const prerender = true;

export const GET: RequestHandler = () => {
  const entries = [
    ...staticEntries,
    ...contentEntries(docs, '/docs'),
    ...contentEntries(posts, '/posts'),
  ];

  const urls = entries
    .map((entry) => {
      const location = escapeXml(new URL(entry.path, SITE_URL).href);
      const lastModified = formatLastModified(entry.lastModified);
      return `  <url>\n    <loc>${location}</loc>${
        lastModified ? `\n    <lastmod>${lastModified}</lastmod>` : ''
      }\n  </url>`;
    })
    .join('\n');

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls}
</urlset>`;

  return new Response(xml, {
    headers: {
      'Content-Type': 'application/xml; charset=utf-8',
      'Cache-Control': 'max-age=3600',
    },
  });
};
