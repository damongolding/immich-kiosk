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
 * @property debug - Enable debug mode
 * @property debugVerbose - Enable verbose debug logging
 * @property version - Version string
 * @property params - Additional configuration parameters
 * @property refresh - Refresh interval in seconds
 * @property disableScreensaver - Whether to prevent screen sleeping
 * @property showDate - Whether to display the date
 * @property dateFormat - Format string for date display
 * @property showTime - Whether to display the time
 * @property timeFormat - Format for time display
 * @property transition - Type of transition animation
 * @property showMoreInfo - Show the more info image overlay
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
  transition: string;
  showMoreInfo: boolean;
};

const MAX_FRAMES: number = 2 as const;

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
const offlineSVG = htmx.find("#offline") as HTMLElement | null;

let requestInFlight = false;

/**
 * Initialize Kiosk functionality
 * @description Sets up kiosk by configuring:
 * - Debug logging if verbose mode enabled
 * - Clock display
 * - Screen sleep prevention
 * - Service worker registration
 * - Fullscreen capability
 * - Image polling
 * - Navigation menu
 * - Event listeners
 * @returns Promise<void>
 */
async function init(): Promise<void> {
  if (kioskData.debugVerbose) {
    htmx.logAll();
  }

  if (kioskData.showDate || kioskData.showTime) {
    initClock(
      kioskData.showDate,
      kioskData.dateFormat,
      kioskData.showTime,
      kioskData.timeFormat,
    );
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
 * @description Toggles fullscreen mode for the document body using the fullscreen API
 */
function handleFullscreenClick(): void {
  toggleFullscreen(documentBody, fullscreenButton);
}

/**
 * Handle 'i' key press events
 * @description Controls polling and image overlay states:
 * - If polling is paused and overlay shown: resumes polling and hides overlay
 * - Otherwise: ensures polling is paused and toggles overlay visibility
 * This allows for synchronized control of polling and overlay display
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
 * @description Configures interactive behavior by setting up:
 * - Menu click handlers for polling control
 * - Keyboard shortcuts (space and 'i' keys)
 * - Fullscreen toggle functionality
 * - Image overlay controls
 * - HTMX error handling for offline states
 * - Server connectivity monitoring
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
        if (!kioskData.showMoreInfo) return;
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

  // Unable to send ajax. probably offline.
  htmx.on("htmx:sendError", () => {
    releaseRequestLock();

    if (!offlineSVG) {
      console.error("offline svg missing");
      return;
    }

    htmx.addClass(offlineSVG, "offline");
  });

  // Server online check. Fires after every AJAX request.
  htmx.on("htmx:afterRequest", function (e: HTMXEvent) {
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
 * @description Manages frame count to prevent memory issues:
 * - Checks current number of frames in DOM
 * - Removes oldest frame if count exceeds maxFrames limit
 * This helps maintain smooth transitions while preventing memory bloat
 */
async function cleanupFrames(): Promise<void> {
  const frames = htmx.findAll(".frame");
  if (!frames?.length) {
    console.debug("No frames found to clean up");
    return;
  }

  if (frames.length > MAX_FRAMES) {
    try {
      htmx.remove(frames[0]);
    } catch (error) {
      console.error("Failed to remove frame:", error);
    }
  }
}

/**
 * Sets a lock to prevent concurrent requests
 * @param e - Event object that triggered the request
 * @description Request management that:
 * - Prevents multiple simultaneous requests
 * - Pauses polling during request processing
 * - Disables navigation controls
 * - Sets request lock flag
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
 * @description Request cleanup that:
 * - Re-enables navigation controls
 * - Clears request lock flag
 * This restores normal kiosk operation after request processing
 */
function releaseRequestLock(): void {
  enableImageNavigationButtons();

  requestInFlight = false;
}

/**
 * Checks if there are enough history entries to navigate back
 * @param e - Event object for the history navigation request
 * @description Navigation safety check that:
 * - Verifies sufficient history depth exists
 * - Prevents navigation during active requests
 * - Sets request lock if navigation is allowed
 */
function checkHistoryExists(e: HTMXEvent): void {
  const historyItems = htmx.findAll(".kiosk-history--entry");
  if (requestInFlight || historyItems.length < 2) {
    e.preventDefault();
    return;
  }

  setRequestLock(e);
}

type BrowserData = {
  client_width: number;
  client_height: number;
};

function clientData(): BrowserData {
  return {
    client_width: window.innerWidth,
    client_height: window.innerHeight,
  };
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
  clientData,
};
