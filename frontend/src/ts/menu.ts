/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 */

import htmx from "htmx.org";

let nextImageMenuButton: HTMLElement;
let prevImageMenuButton: HTMLElement;

function disableImageNavigationButtons() {
  htmx.addClass(nextImageMenuButton as Element, "disabled");
  htmx.addClass(prevImageMenuButton as Element, "disabled");
}

function enableImageNavigationButtons() {
  htmx.removeClass(nextImageMenuButton as Element, "disabled");
  htmx.removeClass(prevImageMenuButton as Element, "disabled");
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param kiosk - The kiosk container element
 * @param menu - The menu container element
 * @param pausePlayButton - The pause/play button element
 */
function initMenu(nextImageButton: HTMLElement, prevImageButton: HTMLElement) {
  nextImageMenuButton = nextImageButton;
  prevImageMenuButton = prevImageButton;
}

export {
  initMenu,
  disableImageNavigationButtons,
  enableImageNavigationButtons,
};
