/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 */

import htmx from "htmx.org";
import { pausePolling } from "./polling";

/** Flag indicating if a new image is currently being loaded */
let gettingNewImage = false;

/** Reference to the main kiosk container element */
let kioskElement: HTMLElement | null;

/**
 * Initializes the menu controls and sets up event handlers
 * @param kiosk - The kiosk container element
 * @param menu - The menu container element
 * @param pausePlayButton - The pause/play button element
 */
function initMenu(kiosk: HTMLElement | null) {
  kioskElement = kiosk;

  htmx.on(kiosk as HTMLElement, "htmx:afterSettle", function (e: any) {
    gettingNewImage = false;
  });
}

/**
 * Handles click event for loading the next image
 * Triggers a kiosk-new-image event if no image is currently loading
 */
function handleNextImageClick() {
  if (gettingNewImage) return;

  pausePolling(false);
  htmx.trigger(kioskElement as HTMLElement, "kiosk-new-image");
  gettingNewImage = true;
}

/**
 * Handles click event for loading the previous image
 * Makes AJAX request to load previous image if history exists and no image is loading
 */
function handlePrevImageClick() {
  const historyItems = htmx.findAll(".kiosk-history--entry");
  if (gettingNewImage || historyItems.length < 2) return;

  pausePolling(false);
  try {
    htmx.ajax("post", "/image/previous", {
      source: "#kiosk",
      values: htmx.values(htmx.find("#kiosk-history") as Element, "post"),
    });
  } catch (e) {
    console.log(e);
    return;
  }

  gettingNewImage = true;
}

export { initMenu, handlePrevImageClick, handleNextImageClick };
