<script lang="ts">
  import X from '@lucide/svelte/icons/x';
  import { focusTrap } from '$lib/focus-trap';

  let { contentKey }: { contentKey?: string } = $props();

  let openSrc = $state<string | null>(null);
  let openAlt = $state('');

  function close() {
    openSrc = null;
  }

  function isPreviewImage(target: EventTarget | null): target is HTMLImageElement {
    return (
      target instanceof HTMLImageElement &&
      Boolean(target.closest('.content')) &&
      !target.closest('a')
    );
  }

  function open(img: HTMLImageElement) {
    openSrc = img.src;
    openAlt = img.alt;
  }

  $effect(() => {
    void contentKey;

    function handleClick(e: MouseEvent) {
      if (isPreviewImage(e.target)) {
        open(e.target);
      }
    }

    function handleImageKeydown(e: KeyboardEvent) {
      if ((e.key === 'Enter' || e.key === ' ') && isPreviewImage(e.target)) {
        e.preventDefault();
        open(e.target);
      }
    }

    const images = Array.from(document.querySelectorAll<HTMLImageElement>('.content img')).filter(
      (img) => !img.closest('a'),
    );
    const originalAttributes = images.map((img) => ({
      img,
      role: img.getAttribute('role'),
      tabindex: img.getAttribute('tabindex'),
      label: img.getAttribute('aria-label'),
    }));

    for (const img of images) {
      img.setAttribute('role', 'button');
      img.setAttribute('tabindex', '0');
      img.setAttribute('aria-label', `Open image preview: ${img.alt || 'image'}`);
    }

    document.addEventListener('click', handleClick);
    document.addEventListener('keydown', handleImageKeydown);

    return () => {
      document.removeEventListener('click', handleClick);
      document.removeEventListener('keydown', handleImageKeydown);
      for (const { img, role, tabindex, label } of originalAttributes) {
        if (role === null) img.removeAttribute('role');
        else img.setAttribute('role', role);
        if (tabindex === null) img.removeAttribute('tabindex');
        else img.setAttribute('tabindex', tabindex);
        if (label === null) img.removeAttribute('aria-label');
        else img.setAttribute('aria-label', label);
      }
    };
  });
</script>

{#if openSrc}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 md:p-10"
    onclick={(e) => {
      if (e.target === e.currentTarget) close();
    }}
    onkeydown={(e) => {
      if (e.target === e.currentTarget && (e.key === 'Enter' || e.key === ' ')) close();
    }}
  >
    <div
      use:focusTrap={{ onEscape: close }}
      role="dialog"
      aria-modal="true"
      aria-label="Image preview"
      tabindex="-1"
      class="relative flex max-h-[calc(100dvh-2rem)] max-w-[calc(100dvw-2rem)] items-center justify-center md:max-h-[calc(100dvh-5rem)] md:max-w-[calc(100dvw-5rem)]"
    >
      <img
        src={openSrc}
        alt={openAlt}
        class="block max-h-[calc(100dvh-2rem)] max-w-[calc(100dvw-2rem)] border-[3px] border-white object-contain shadow-[8px_8px_0_rgba(255,255,255,0.15)] md:max-h-[calc(100dvh-5rem)] md:max-w-[calc(100dvw-5rem)]"
      />
      <button
        onclick={close}
        aria-label="Close image preview"
        class="absolute top-2 right-2 flex size-11 cursor-pointer items-center justify-center border-[3px] border-white bg-white text-black shadow-[3px_3px_0_rgba(255,255,255,0.2)] transition-colors hover:bg-gray-100"
      >
        <X size={20} />
      </button>
    </div>
  </div>
{/if}
