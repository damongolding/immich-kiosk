/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 */

import htmx from "htmx.org";

let nextImageMenuButton: HTMLElement;
let prevImageMenuButton: HTMLElement;

let imageOverlayVisable = false;

function disableImageNavigationButtons() {
  if (!nextImageMenuButton || !prevImageMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.addClass(nextImageMenuButton, "disabled");
  htmx.addClass(prevImageMenuButton, "disabled");
}

function enableImageNavigationButtons() {
  if (!nextImageMenuButton || !prevImageMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.removeClass(nextImageMenuButton as Element, "disabled");
  htmx.removeClass(prevImageMenuButton as Element, "disabled");
}

function showImageOverlay() {
  if (!document.body.classList.contains("polling-paused")) return;
  console.log("on");
  document.body.classList.add("more-info");
  imageOverlayVisable = true;
}

function hideImageOverlay() {
  console.log("off");
  document.body.classList.remove("more-info");
  imageOverlayVisable = false;
}

function toggleImageOverlay() {
  imageOverlayVisable ? hideImageOverlay() : showImageOverlay();
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param nextImageButton - The next image navigation button element
 * @param prevImageButton - The previous image navigation button element
 */
function initMenu(nextImageButton: HTMLElement, prevImageButton: HTMLElement) {
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
};
