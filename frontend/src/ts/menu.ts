/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 * @description Controls menu behavior and navigation between assets in a kiosk interface
 */

import htmx from "htmx.org";

let nextAssetMenuButton: HTMLElement;
let prevAssetMenuButton: HTMLElement;

let assetOverlayVisible: boolean = false;

/**
 * Disables both next and previous asset navigation buttons
 * @returns {void}
 */
function disableAssetNavigationButtons(): void {
  if (!nextAssetMenuButton || !prevAssetMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.addClass(nextAssetMenuButton, "disabled");
  htmx.addClass(prevAssetMenuButton, "disabled");
}

/**
 * Enables both next and previous asset navigation buttons
 * @returns {void}
 */
function enableAssetNavigationButtons(): void {
  if (!nextAssetMenuButton || !prevAssetMenuButton) {
    console.error("Navigation buttons not initialized");
    return;
  }
  htmx.removeClass(nextAssetMenuButton as Element, "disabled");
  htmx.removeClass(prevAssetMenuButton as Element, "disabled");
}

/**
 * Shows the asset information overlay
 * Only works when polling is paused
 * @returns {void}
 */
function showAssetOverlay(): void {
  if (!document.body) return;
  if (!document.body.classList.contains("polling-paused")) return;
  document.body.classList.add("more-info");
  assetOverlayVisible = true;
}

/**
 * Hides the asset information overlay
 * @returns {void}
 */
function hideAssetOverlay(): void {
  if (!document.body) return;
  document.body.classList.remove("more-info");
  assetOverlayVisible = false;
}

/**
 * Toggles the asset information overlay visibility
 * @returns {void}
 */
function toggleAssetOverlay(): void {
  assetOverlayVisible ? hideAssetOverlay() : showAssetOverlay();
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param nextAssetButton - The next image navigation button element
 * @param prevAssetButton - The previous image navigation button element
 * @throws {Error} If either navigation button is not provided
 * @returns {void}
 */
function initMenu(
  nextAssetButton: HTMLElement,
  prevAssetButton: HTMLElement,
): void {
  if (!nextAssetButton || !prevAssetButton) {
    throw new Error("Both navigation buttons must be provided");
  }
  nextAssetMenuButton = nextAssetButton;
  prevAssetMenuButton = prevAssetButton;
}

export {
  initMenu,
  disableAssetNavigationButtons,
  enableAssetNavigationButtons,
  showAssetOverlay,
  hideAssetOverlay,
  toggleAssetOverlay,
};
