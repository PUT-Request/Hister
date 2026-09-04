import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { cpSync, mkdirSync } from 'fs';
import { resolve } from 'path';
import { defineConfig } from 'vite';

function copyDatasetJsons() {
  return {
    name: 'copy-dataset-jsons',
    apply: 'build' as const,
    closeBundle() {
      const src = resolve('src/content/datasets');
      const dest = resolve('build/datasets');
      mkdirSync(dest, { recursive: true });
      cpSync(src, dest, { recursive: true });
    },
  };
}

export default defineConfig({
  plugins: [tailwindcss(), sveltekit(), copyDatasetJsons()],
  ssr: {
    noExternal: ['@hister/components', 'bits-ui', 'svelte-toolbelt', 'runed', 'svelte-sonner'],
  },
  build: {
    rolldownOptions: {
      checks: { pluginTimings: false },
    },
  },
});
