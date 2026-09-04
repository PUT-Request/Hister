import adapter from '@sveltejs/adapter-static';
import { mdsvex, escapeSvelte } from 'mdsvex';
import rehypeSlug from 'rehype-slug';
import { createHighlighter } from 'shiki';

const theme = 'github-dark';
const langs = [
  'bash',
  'shell',
  'yaml',
  'json',
  'javascript',
  'typescript',
  'html',
  'css',
  'nginx',
  'nix',
  'go',
  'text',
  'plaintext',
  'markdown',
  'dockerfile',
];
const langAliases = { textplain: 'text', caddy: 'text' };
const highlighter = await createHighlighter({ themes: [theme], langs });
const loadedLangs = new Set(highlighter.getLoadedLanguages());

/** @type {import('@sveltejs/kit').Config} */
export default {
  extensions: ['.svelte', '.md', '.svx'],
  preprocess: [
    mdsvex({
      extensions: ['.md', '.svx'],
      rehypePlugins: [rehypeSlug],
      highlight: {
        highlighter: (code, lang) => {
          const resolved = langAliases[lang] || lang || 'text';
          const html = escapeSvelte(
            highlighter.codeToHtml(code, {
              lang: loadedLangs.has(resolved) ? resolved : 'text',
              theme,
            }),
          );
          return `{@html \`${html}\` }`;
        },
      },
    }),
  ],
  kit: {
    adapter: adapter({ pages: 'build', assets: 'build', fallback: '404.html' }),
    prerender: { handleHttpError: 'warn', handleMissingId: 'ignore' },
  },
};
