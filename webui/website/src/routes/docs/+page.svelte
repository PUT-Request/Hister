<script lang="ts">
  import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
  import Globe from '@lucide/svelte/icons/globe';
  import SettingsIcon from '@lucide/svelte/icons/settings';
  import * as Card from '@hister/components/ui/card';
  import DocsSearch from '$lib/DocsSearch.svelte';
  import Seo from '$lib/Seo.svelte';

  let { data } = $props();
  const featuredSlugs = new Set(['configuration']);
  const featuredDoc = $derived(data.docs.find((doc) => doc.slug === 'configuration'));
</script>

<Seo
  title="Documentation | Hister"
  description="Learn how to install, configure, and use Hister across the web, terminal, browser extensions, API, and MCP."
  path="/docs"
/>

<section class="mx-auto max-w-4xl px-6 py-12 md:px-12">
  <h1
    class="font-space text-4xl font-black tracking-[-1px] text-(--text-primary) uppercase md:text-5xl"
  >
    Documentation
  </h1>
  <div class="mt-6 mb-10">
    <DocsSearch />
  </div>

  <Card.Root href="/docs/configuration" color="hister-indigo" class="bg-hister-indigo/10 mb-10 p-6">
    <div class="flex items-center gap-4">
      <div
        class="border-brutal-border bg-hister-indigo flex size-11 shrink-0 items-center justify-center border-[2px] text-white"
      >
        <SettingsIcon aria-hidden="true" size={22} />
      </div>
      <div class="min-w-0 flex-1">
        <h2
          class="font-space text-xl font-extrabold tracking-[0.5px] text-(--text-primary) sm:text-2xl"
        >
          Configuration Reference
        </h2>
        {#if featuredDoc?.description}
          <p class="font-inter mt-2 text-sm leading-relaxed text-(--text-secondary)">
            {featuredDoc.description}
          </p>
        {/if}
      </div>
      <ArrowRightIcon aria-hidden="true" size={20} class="text-hister-indigo shrink-0" />
    </div>
  </Card.Root>

  {#each data.categories as category}
    <div class="mb-8">
      <div class="mb-3 flex items-center gap-2">
        <div class="h-2.5 w-2.5 bg-hister-{category.color}"></div>
        <span
          class="font-space text-xs font-bold tracking-[2px] text-[var(--text-secondary)] uppercase"
        >
          {category.name}
        </span>
      </div>

      <ul class="m-0 flex list-none flex-col gap-3 p-0">
        {#each category.docs as doc}
          {#if !featuredSlugs.has(doc.slug)}
            <li>
              <Card.Root href="/docs/{doc.slug}" class="bg-brutal-card p-5">
                <h2
                  class="font-space text-lg font-extrabold tracking-[0.5px] text-(--text-primary)"
                >
                  {doc.title}
                </h2>
                {#if doc.description}
                  <p class="font-inter mt-2 text-sm leading-relaxed text-(--text-secondary)">
                    {doc.description}
                  </p>
                {/if}
              </Card.Root>
            </li>
          {/if}
        {/each}
      </ul>
    </div>
  {/each}

  <div class="mb-8">
    <div class="mb-3 flex items-center gap-2">
      <div class="bg-hister-indigo h-2.5 w-2.5"></div>
      <span class="font-space text-xs font-bold tracking-[2px] text-(--text-secondary) uppercase">
        Browser Extensions
      </span>
    </div>

    <ul class="m-0 flex list-none flex-col gap-3 p-0">
      <li>
        <Card.Root
          href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb"
          target="_blank"
          rel="noopener noreferrer"
          class="bg-brutal-card flex-row items-center gap-4 p-5"
        >
          <Globe size={20} class="shrink-0 text-(--text-secondary)" />
          <h2 class="font-space text-lg font-extrabold tracking-[0.5px] text-(--text-primary)">
            Chrome Extension
          </h2>
        </Card.Root>
      </li>
      <li>
        <Card.Root
          href="https://addons.mozilla.org/en-US/firefox/addon/hister/"
          target="_blank"
          rel="noopener noreferrer"
          class="bg-brutal-card flex-row items-center gap-4 p-5"
        >
          <Globe size={20} class="shrink-0 text-(--text-secondary)" />
          <h2 class="font-space text-lg font-extrabold tracking-[0.5px] text-(--text-primary)">
            Firefox Add-on
          </h2>
        </Card.Root>
      </li>
    </ul>
  </div>
</section>
