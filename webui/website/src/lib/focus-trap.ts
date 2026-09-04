const focusableSelector = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

function getFocusableElements(node: HTMLElement): HTMLElement[] {
  return Array.from(node.querySelectorAll<HTMLElement>(focusableSelector)).filter(
    (element) =>
      element.getClientRects().length > 0 && element.getAttribute('aria-hidden') !== 'true',
  );
}

interface FocusTrapOptions {
  onEscape?: () => void;
}

function trapTabKey(node: HTMLElement, event: KeyboardEvent) {
  const focusableElements = getFocusableElements(node);
  if (focusableElements.length === 0) {
    event.preventDefault();
    node.focus();
    return;
  }

  const first = focusableElements[0];
  const last = focusableElements[focusableElements.length - 1];
  const activeElement = document.activeElement;
  const movingBeforeStart = event.shiftKey && activeElement === first;
  const movingPastEnd = !event.shiftKey && activeElement === last;

  if (!movingBeforeStart && !movingPastEnd && node.contains(activeElement)) return;

  event.preventDefault();
  (event.shiftKey ? last : first).focus();
}

export function focusTrap(node: HTMLElement, options: FocusTrapOptions = {}) {
  let currentOptions = options;
  const previousFocus =
    document.activeElement instanceof HTMLElement ? document.activeElement : null;
  const focusFrame = requestAnimationFrame(() => {
    const [first] = getFocusableElements(node);
    (first ?? node).focus();
  });

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && currentOptions.onEscape) {
      event.preventDefault();
      event.stopPropagation();
      currentOptions.onEscape();
      return;
    }

    if (event.key === 'Tab') trapTabKey(node, event);
  }

  node.addEventListener('keydown', handleKeydown);

  return {
    update(nextOptions: FocusTrapOptions) {
      currentOptions = nextOptions;
    },
    destroy() {
      cancelAnimationFrame(focusFrame);
      node.removeEventListener('keydown', handleKeydown);
      if (previousFocus?.isConnected) previousFocus.focus();
    },
  };
}
