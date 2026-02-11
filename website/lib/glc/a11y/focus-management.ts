/**
 * GLC Accessibility Utilities
 *
 * Focus management, keyboard navigation, and ARIA helpers
 */

/**
 * Get all focusable elements within a container
 */
export function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const selector = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(', ');

  return Array.from(container.querySelectorAll<HTMLElement>(selector));
}

/**
 * Trap focus within a container (for modals)
 */
export function createFocusTrap(container: HTMLElement) {
  const focusableElements = getFocusableElements(container);
  const firstElement = focusableElements[0];
  const lastElement = focusableElements[focusableElements.length - 1];

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key !== 'Tab') return;

    if (e.shiftKey) {
      if (document.activeElement === firstElement) {
        e.preventDefault();
        lastElement?.focus();
      }
    } else {
      if (document.activeElement === lastElement) {
        e.preventDefault();
        firstElement?.focus();
      }
    }
  }

  return {
    activate: () => {
      container.addEventListener('keydown', handleKeyDown);
      firstElement?.focus();
    },
    deactivate: () => {
      container.removeEventListener('keydown', handleKeyDown);
    },
  };
}

/**
 * Get screen reader announcement function
 */
export function announce(message: string, priority: 'polite' | 'assertive' = 'polite'): void {
  const announcement = document.createElement('div');
  announcement.setAttribute('aria-live', priority);
  announcement.setAttribute('aria-atomic', 'true');
  announcement.setAttribute('role', 'status');
  announcement.className = 'sr-only';
  announcement.style.cssText = 'position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0;';

  document.body.appendChild(announcement);

  // Delay to ensure screen reader picks up the change
  setTimeout(() => {
    announcement.textContent = message;
    setTimeout(() => {
      document.body.removeChild(announcement);
    }, 1000);
  }, 100);
}

/**
 * Check if reduced motion is preferred
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Get appropriate animation duration based on preferences
 */
export function getAnimationDuration(defaultDuration: number): number {
  return prefersReducedMotion() ? 0 : defaultDuration;
}

/**
 * Generate unique ID for ARIA associations
 */
let idCounter = 0;
export function generateAriaId(prefix: string = 'glc'): string {
  return `${prefix}-${++idCounter}`;
}
