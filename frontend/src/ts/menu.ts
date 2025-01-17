/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 */

import htmx from "htmx.org";

let nextImageMenuButton: HTMLElement;
let prevImageMenuButton: HTMLElement;

let imageOverlayVisible: boolean = false;
let linkOverlayVisible: boolean = false;

const redirectsContainer = document.getElementById(
  "redirects-container",
) as HTMLElement | null;
let redirects: NodeListOf<HTMLAnchorElement> | null;
let currentRedirectIndex = -1;

/**
 * Disables the image navigation buttons by adding a 'disabled' class
 * Logs an error if buttons are not properly initialized
 */
function disableImageNavigationButtons(): void {
  if (!nextImageMenuButton || !prevImageMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.addClass(nextImageMenuButton, "disabled");
  htmx.addClass(prevImageMenuButton, "disabled");
}

/**
 * Enables the image navigation buttons by removing the 'disabled' class
 * Logs an error if buttons are not properly initialized
 */
function enableImageNavigationButtons(): void {
  if (!nextImageMenuButton || !prevImageMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.removeClass(nextImageMenuButton as Element, "disabled");
  htmx.removeClass(prevImageMenuButton as Element, "disabled");
}

/**
 * Shows the image information overlay
 * Only works when polling is paused
 * Hides links overlay if visible
 */
function showImageOverlay(): void {
  if (!document.body) return;
  if (!document.body.classList.contains("polling-paused")) return;
  hideRedirectsOverlay();
  document.body.classList.add("more-info");
  imageOverlayVisible = true;
}

/**
 * Hides the image information overlay
 */
function hideImageOverlay(): void {
  if (!document.body) return;
  document.body.classList.remove("more-info");
  imageOverlayVisible = false;
}

/**
 * Toggles the image information overlay visibility
 */
function toggleImageOverlay(): void {
  imageOverlayVisible ? hideImageOverlay() : showImageOverlay();
}

function redirectKeyHandler(e: KeyboardEvent) {
  if (redirects) {
    if (e.key === "ArrowDown") {
      e.preventDefault(); // Prevent page scrolling
      currentRedirectIndex = (currentRedirectIndex + 1) % redirects.length;
      redirects[currentRedirectIndex].focus();
    } else if (e.key === "ArrowUp") {
      e.preventDefault(); // Prevent page scrolling
      currentRedirectIndex =
        (currentRedirectIndex - 1 + redirects.length) % redirects.length;
      redirects[currentRedirectIndex].focus();
    }
  }
}

/**
 * Shows the links overlay
 * Only works when polling is paused
 * Hides image overlay if visible
 */
function showRedirectsOverlay(): void {
  if (!document.body) return;
  if (!document.body.classList.contains("polling-paused")) return;

  document.addEventListener("keydown", redirectKeyHandler);

  hideImageOverlay();
  document.body.classList.add("redirects-open");
  linkOverlayVisible = true;
}

/**
 * Hides the links overlay
 */
function hideRedirectsOverlay(): void {
  if (!document.body) return;
  document.body.classList.remove("redirects-open");

  document.removeEventListener("keydown", redirectKeyHandler);

  linkOverlayVisible = false;
}

/**
 * Toggles the links overlay visibility
 */
function toggleRedirectsOverlay(): void {
  linkOverlayVisible ? hideRedirectsOverlay() : showRedirectsOverlay();
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param nextImageButton - The next image navigation button element
 * @param prevImageButton - The previous image navigation button element
 */
function initMenu(
  nextImageButton: HTMLElement,
  prevImageButton: HTMLElement,
): void {
  if (!nextImageButton || !prevImageButton) {
    throw new Error("Both navigation buttons must be provided");
  }
  nextImageMenuButton = nextImageButton;
  prevImageMenuButton = prevImageButton;

  if (redirectsContainer) {
    redirects = redirectsContainer.querySelectorAll("a");
  }
}

export {
  initMenu,
  disableImageNavigationButtons,
  enableImageNavigationButtons,
  showImageOverlay,
  hideImageOverlay,
  toggleImageOverlay,
  toggleRedirectsOverlay,
};
