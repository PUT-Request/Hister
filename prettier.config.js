/** @type {import("prettier").Config} */
export default {
  plugins: [
    'prettier-plugin-svelte',
    'prettier-plugin-tailwindcss', // MUST COME LAST
  ],
  overrides: [{ files: '*.svelte', options: { parser: 'svelte' } }],
  printWidth: 100,
  tabWidth: 2,
  useTabs: false,
  semi: true,
  singleQuote: true,
  trailingComma: 'all',
};
