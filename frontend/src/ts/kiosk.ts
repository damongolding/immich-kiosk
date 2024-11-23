import htmx from "htmx.org";
import {
  addFullscreenEventListener,
  fullscreenAPI,
  toggleFullscreen,
} from "./fullscreen";
import {
  initPolling,
  startPolling,
  togglePolling,
  pausePolling,
} from "./polling";
import { preventSleep } from "./wakelock";
import {
  initMenu,
  disableImageNavigationButtons,
  enableImageNavigationButtons,
  toggleImageOverlay,
} from "./menu";
import { initClock } from "./clock";
import type { TimeFormat } from "./clock";

("use strict");

interface HTMXEvent extends Event {
  preventDefault: () => void;
  detail: {
    successful: boolean;
  };
}

/**
 * Type definition for kiosk configuration data
 */
type KioskData = {
  debug: boolean;
  debugVerbose: boolean;
  version: string;
  params: Record<string, unknown>;
  refresh: number;
  disableScreensaver: boolean;
  showDate: boolean;
  dateFormat: string;
  showTime: boolean;
  timeFormat: TimeFormat;
};

let maxFrames: boolean;

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
  ".navigation--play-pause",
) as HTMLElement | null;
const nextImageMenuButton = htmx.find(
  ".navigation--next-image",
) as HTMLElement | null;
const prevImageMenuButton = htmx.find(
  ".navigation--prev-image",
) as HTMLElement | null;
const moreInfoButton = htmx.find(
  ".navigation--more-info",
) as HTMLElement | null;

let requestInFlight = false;

/**
 * Initialize Kiosk functionality
 * Sets up debugging, screensaver prevention, service worker registration,
 * fullscreen capability, polling, menu and event listeners
 */
async function init(): Promise<void> {
  if (kioskData.debugVerbose) {
    htmx.logAll();
  }

  maxFrames = kioskData.transition === "cross-fade" ? 3 : 1;

  initClock(
    kioskData.showDate,
    kioskData.dateFormat,
    kioskData.showTime,
    kioskData.timeFormat,
  );

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

  if (nextImageMenuButton && prevImageMenuButton) {
    initMenu(
      nextImageMenuButton as HTMLElement,
      prevImageMenuButton as HTMLElement,
    );
  } else {
    console.error("Menu buttons not found");
  }
  addEventListeners();
}

/**
 * Handler for fullscreen button clicks
 * Toggles fullscreen mode for the document body
 */
function handleFullscreenClick(): void {
  toggleFullscreen(documentBody, fullscreenButton);
}

/**
 * Handle 'i' key press events
 * Controls polling and image overlay states based on current document status
 * @description If both polling is paused and more info is shown, resumes polling and hides overlay.
 * Otherwise, ensures polling is paused and toggles the overlay visibility.
 */
function handleInfoKeyPress(): void {
  const isPollingPaused = document.body.classList.contains("polling-paused");
  const hasMoreInfo = document.body.classList.contains("more-info");

  if (isPollingPaused && hasMoreInfo) {
    togglePolling();
    toggleImageOverlay();
  } else {
    if (!isPollingPaused) {
      togglePolling();
    }
    toggleImageOverlay();
  }
}

/**
 * Add event listeners to Kiosk elements
 * Sets up listeners for:
 * - Menu interaction and polling control
 * - Fullscreen functionality
 * - Navigation between images
 * - Server connection status monitoring
 */
function addEventListeners(): void {
  // Pause/resume polling and show/hide menu
  menuInteraction?.addEventListener("click", () => togglePolling());
  menuPausePlayButton?.addEventListener("click", () => togglePolling());
  document.addEventListener("keydown", (e) => {
    if (e.target !== document.body) return;

    switch (e.code) {
      case "Space":
        e.preventDefault();
        togglePolling(true);
        break;
      case "KeyI":
        e.preventDefault();
        handleInfoKeyPress();
        break;
    }
  });

  // Fullscreen
  fullscreenButton?.addEventListener("click", handleFullscreenClick);
  addFullscreenEventListener(fullscreenButton);

  // More info overlay
  moreInfoButton?.addEventListener("click", () => toggleImageOverlay());

  // Server online check. Fires after every AJAX request.
  htmx.on("htmx:afterRequest", function (e: HTMXEvent) {
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
 * Remove first frame from the DOM when there are more than maxFrames
 * Used to prevent memory issues from accumulating frames
 */
function cleanupFrames(): void {
  const frames = htmx.findAll(".frame");
  if (frames.length > maxFrames) {
    htmx.remove(frames[0]);
  }
}

/**
 * Sets a lock to prevent concurrent requests
 * @param e - Event object that triggered the request
 * @description Prevents multiple simultaneous requests by checking and setting a lock flag.
 * Also pauses polling and disables navigation buttons while request is in flight.
 */
function setRequestLock(e: HTMXEvent): void {
  if (requestInFlight) {
    e.preventDefault();
    return;
  }

  pausePolling(false);

  disableImageNavigationButtons();

  requestInFlight = true;
}

/**
 * Releases the request lock after a request completes
 * @description Re-enables navigation buttons and marks request as complete by unsetting
 * the requestInFlight flag.
 */
function releaseRequestLock(): void {
  enableImageNavigationButtons();

  requestInFlight = false;
}

/**
 * Checks if there are enough history entries to navigate back
 * @param e - Event object for the history navigation request
 * @description Prevents navigation when there is an active request or insufficient history.
 * If navigation is allowed, sets the request lock.
 */
function checkHistoryExists(e: HTMXEvent): void {
  const historyItems = htmx.findAll(".kiosk-history--entry");
  if (requestInFlight || historyItems.length < 2) {
    e.preventDefault();
    return;
  }

  setRequestLock(e);
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
  init();
});

export {
  cleanupFrames,
  startPolling,
  setRequestLock,
  releaseRequestLock,
  checkHistoryExists,
};
