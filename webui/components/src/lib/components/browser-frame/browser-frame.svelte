<script lang="ts">
  import type { Snippet } from 'svelte';
  import { cn } from '../../utils.js';

  let {
    bg = '#f5f5f5',
    tabBarClass = 'bg-brutal-card border-b-2 border-brutal-border',
    tabBarHeight = 'h-11',
    class: className = '',
    children,
  }: {
    bg?: string;
    tabBarClass?: string;
    tabBarHeight?: string;
    class?: string;
    children: Snippet;
  } = $props();
</script>

<div
  class={cn(
    'border-brutal-border w-full overflow-hidden border-[3px] shadow-[8px_8px_0_var(--brutal-shadow)]',
    className,
  )}
  style="background-color: {bg};"
>
  <div class={cn('flex items-center gap-2 px-4', tabBarHeight, tabBarClass)} aria-hidden="true">
    <span class="h-3 w-3 rounded-full bg-[#FF5F56]"></span>
    <span class="h-3 w-3 rounded-full bg-[#FFBD2E]"></span>
    <span class="h-3 w-3 rounded-full bg-[#27C93F]"></span>
  </div>
  {@render children()}
</div>
