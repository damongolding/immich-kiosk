/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 */

import htmx from "htmx.org";

let nextImageMenuButton: HTMLElement;
let prevImageMenuButton: HTMLElement;

let imageOverlayVisible: boolean = false;
let linkOverlayVisible: boolean = false;

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
  hideLinksOverlay();
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

/**
 * Shows the links overlay
 * Only works when polling is paused
 * Hides image overlay if visible
 */
function showLinksOverlay(): void {
  if (!document.body) return;
  if (!document.body.classList.contains("polling-paused")) return;
  hideImageOverlay();
  document.body.classList.add("links");
  linkOverlayVisible = true;
}

/**
 * Hides the links overlay
 */
function hideLinksOverlay(): void {
  if (!document.body) return;
  document.body.classList.remove("links");
  linkOverlayVisible = false;
}

/**
 * Toggles the links overlay visibility
 */
function toggleLinksOverlay(): void {
  linkOverlayVisible ? hideLinksOverlay() : showLinksOverlay();
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
}

export {
  initMenu,
  disableImageNavigationButtons,
  enableImageNavigationButtons,
  showImageOverlay,
  hideImageOverlay,
  toggleImageOverlay,
  toggleLinksOverlay,
};
