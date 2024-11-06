import htmx from "htmx.org";
import {
  addFullscreenEventListener,
  fullscreenAPI,
  toggleFullscreen,
} from "./fullscreen";
import { initPolling, startPolling, togglePolling } from "./polling";
import { preventSleep } from "./wakelock";
import { handleNextImageClick, initMenu } from "./menu";

("use strict");

type KioskData = {
  debug: boolean;
  debugVerbose: boolean;
  version: string;
  params: Record<string, unknown>;
  refresh: number;
  disableScreensaver: boolean;
};

// Parse kiosk data from the HTML element
const kioskData: KioskData = JSON.parse(
  document.getElementById("kiosk-data")?.textContent || "{}",
);

// Set polling interval based on the refresh rate in kiosk data
const pollInterval = htmx.parseInterval(`${kioskData.refresh}s`);

// Cache DOM elements for better performance
const documentBody = document.body;
const fullscreenButton = htmx.find(
  ".navigation--fullscreen",
) as HTMLElement | null;
const fullScreenButtonSeperator = htmx.find(
  ".navigation--fullscreen-separator",
) as HTMLElement | null;
const kiosk = htmx.find("#kiosk") as HTMLElement | null;
const menu = htmx.find(".navigation") as HTMLElement | null;
const menuInteraction = htmx.find(
  "#navigation-interaction-area--menu",
) as HTMLElement | null;
const menuPausePlayButton = htmx.find(
  ".navigation--control",
) as HTMLElement | null;

/**
 * Initialize Kiosk functionality
 */
async function init() {
  if (kioskData.debugVerbose) {
    htmx.logAll();
  }

  if (kioskData.disableScreensaver) {
    await preventSleep();
  }

  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/assets/js/sw.js").then(
      function (registration) {
        console.log("ServiceWorker registration successful");
      },
      function (err) {
        console.log("ServiceWorker registration failed: ", err);
      },
    );
  }

  if (!fullscreenAPI.requestFullscreen) {
    fullscreenButton && htmx.remove(fullscreenButton);
    fullScreenButtonSeperator && htmx.remove(fullScreenButtonSeperator);
  }

  if (pollInterval) {
    initPolling(pollInterval, kiosk, menu, menuPausePlayButton);
  } else {
    console.error("Could not start polling");
  }

  initMenu(kiosk, menu, menuPausePlayButton);

  addEventListeners();
}

function handleFullscreenClick() {
  toggleFullscreen(documentBody, fullscreenButton);
}

/**
 * Add event listeners to Kiosk elements
 */
function addEventListeners() {
  // Pause/resume polling and show/hide menu
  menuInteraction?.addEventListener("click", togglePolling);
  menuPausePlayButton?.addEventListener("click", togglePolling);

  // Fullscreen
  fullscreenButton?.addEventListener("click", handleFullscreenClick);
  addFullscreenEventListener(fullscreenButton);

  // Menu
  const nextImage = htmx.find("#navigation-interaction-area--next-image");
  nextImage?.addEventListener("click", handleNextImageClick);

  // Server online check. Fires after every AJAX request.
  htmx.on("htmx:afterRequest", function (e: any) {
    const offlineSVG = htmx.find("#offline");

    if (!offlineSVG) {
      console.error("offline svg missing");
      return;
    }

    if (e.detail.successful) {
      htmx.removeClass(offlineSVG, "offline");
    } else {
      htmx.addClass(offlineSVG, "offline");
    }
  });
}

/**
 * Remove first frame
 */
function cleanupFrames() {
  const frames = htmx.findAll(".frame");
  if (frames.length > 3) {
    htmx.remove(frames[0]);
  }
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
  init();
});

export { cleanupFrames, startPolling };
