import htmx from "htmx.org";
import {
  addFullscreenEventListener,
  fullscreenAPI,
  toggleFullscreen,
} from "./fullscreen";
import {
  initPolling,
  startPolling,
  stopPolling,
  togglePolling,
  pausePolling,
  videoHandler,
} from "./polling";
import { preventSleep } from "./wakelock";
import {
  initMenu,
  disableAssetNavigationButtons,
  enableAssetNavigationButtons,
  toggleAssetOverlay,
} from "./menu";
import { initClock } from "./clock";
import type { TimeFormat } from "./clock";

("use strict");

interface HTMXEvent extends Event {
  preventDefault: () => void;
  detail: {
    successful: boolean;
    parameters: FormData;
    method: string;
  };
}

/**
 * @typedef KioskData Configuration data for the kiosk
 * @property {boolean} debug Enable debug mode
 * @property {boolean} debugVerbose Enable verbose debug logging
 * @property {string} version Version string
 * @property {Record<string, unknown>} params Additional configuration parameters
 * @property {number} refresh Refresh interval in seconds
 * @property {boolean} disableScreensaver Whether to prevent screen sleeping
 * @property {boolean} showDate Whether to display the date
 * @property {string} dateFormat Format string for date display
 * @property {boolean} showTime Whether to display the time
 * @property {TimeFormat} timeFormat Format for time display
 * @property {string} transition Type of transition animation
 * @property {boolean} showMoreInfo Show the more info image overlay
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
const kioskQueries = htmx.findAll(".kiosk-param");
const menu = htmx.find(".navigation") as HTMLElement | null;
const menuInteraction = htmx.find(
  "#navigation-interaction-area--menu",
) as HTMLElement | null;
const menuPausePlayButton = htmx.find(
  ".navigation--play-pause",
) as HTMLElement | null;
const nextImageMenuButton = htmx.find(
  ".navigation--next-asset",
) as HTMLElement | null;
const prevImageMenuButton = htmx.find(
  ".navigation--prev-asset",
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
 * - Clock display with configured date/time settings
 * - Screen sleep prevention if enabled
 * - Service worker registration for offline functionality
 * - Fullscreen capability checking and button setup
 * - Image polling with configured interval
 * - Navigation menu initialization
 * - Event listener registration for interactivity
 * @returns {Promise<void>} Promise that resolves when initialization is complete
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
    initPolling(pollInterval, kiosk, menu);
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
 * @description Toggles fullscreen mode for the document body using the browser's fullscreen API
 * The button state is automatically updated based on fullscreen status
 */
function handleFullscreenClick(): void {
  toggleFullscreen(documentBody, fullscreenButton);
}

/**
 * Handle 'i' key press events
 * @description Controls synchronized polling and image overlay behaviors:
 * - When polling is paused and overlay visible: Resumes polling and hides overlay
 * - When polling active or overlay hidden: Pauses polling if needed and toggles overlay
 * Ensures consistent state between polling and overlay display
 */
function handleInfoKeyPress(): void {
  const isPollingPaused = document.body.classList.contains("polling-paused");
  const hasMoreInfo = document.body.classList.contains("more-info");

  if (isPollingPaused && hasMoreInfo) {
    togglePolling();
    toggleAssetOverlay();
  } else {
    if (!isPollingPaused) {
      togglePolling();
    }
    toggleAssetOverlay();
  }
}

/**
 * Add event listeners to Kiosk elements
 * @description Sets up all interactive behaviors and event handling:
 * - Menu interaction for polling control via clicks
 * - Keyboard shortcuts (Space for polling, 'i' for info overlay)
 * - Fullscreen mode toggling via button
 * - Image overlay visibility control
 * - HTMX error handling for offline detection
 * - Server connectivity status monitoring and display
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
        if (e.ctrlKey || e.metaKey) return;
        e.preventDefault();
        handleInfoKeyPress();
        break;
    }
  });

  // Fullscreen
  fullscreenButton?.addEventListener("click", handleFullscreenClick);
  addFullscreenEventListener(fullscreenButton);

  // More info overlay
  moreInfoButton?.addEventListener("click", () => toggleAssetOverlay());

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
 * @description Memory management function that:
 * - Monitors the number of frame elements in the DOM
 * - Removes oldest frame when count exceeds MAX_FRAMES limit
 * - Logs debug info and errors during cleanup process
 * @returns {Promise<void>} Promise that resolves when cleanup completes
 * @throws {Error} If frame removal fails
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
 * @param {HTMXEvent} e Event object that triggered the request
 * @description Request management function that coordinates:
 * - Concurrent request prevention via locking
 * - Polling pause during active requests
 * - Navigation control disabling
 * - Request lock state management
 * @throws {Error} If request lock is already set
 */
function setRequestLock(e: HTMXEvent): void {
  if (requestInFlight) {
    e.preventDefault();
    return;
  }

  pausePolling(false);

  disableAssetNavigationButtons();

  requestInFlight = true;
}

/**
 * Releases the request lock after a request completes
 * @description Request cleanup function that:
 * - Re-enables navigation button controls
 * - Clears the request lock flag
 * - Restores normal kiosk operation state
 */
function releaseRequestLock(): void {
  enableAssetNavigationButtons();

  requestInFlight = false;
}

/**
 * Checks if there are enough history entries to navigate back
 * @param {HTMXEvent} e Event object for the history navigation request
 * @description Navigation safety check that ensures:
 * - Sufficient history depth exists before navigation
 * - No active requests are in progress
 * - Sets request lock when navigation is permitted
 */
function checkHistoryExists(e: HTMXEvent): void {
  const historyItems = htmx.findAll(".kiosk-history--entry");
  if (requestInFlight || historyItems.length < 2) {
    e.preventDefault();
    return;
  }

  setRequestLock(e);
}

/**
 * @typedef {Object} BrowserData
 * @property {number} client_width Window inner width
 * @property {number} client_height Window inner height
 */
type BrowserData = {
  client_width: number;
  client_height: number;
};

/**
 * Get current browser viewport dimensions
 * @returns {BrowserData} Object containing window width and height
 */
function clientData(): BrowserData {
  return {
    client_width: window.innerWidth,
    client_height: window.innerHeight,
  };
}

/**
 * Sanitizes input string by escaping special characters
 * @param {string} value The input string to sanitize
 * @returns {string} Sanitized string with HTML special characters escaped:
 * - < and > removed entirely
 * - & encoded as &amp;
 * - " encoded as &quot;
 * - ' encoded as &#x27;
 * - ` encoded as &#x60;
 * @description Prevents XSS attacks by encoding potentially dangerous characters
 */
function sanitiseInput(value: string): string {
  return value
    .replace(/[<>]/g, "")
    .replace(/[&]/g, "&amp;")
    .replace(/["]/g, "&quot;")
    .replace(/[']/g, "&#x27;")
    .replace(/[`]/g, "&#x60;");
}

// Add kiosk query parameters to HTMX requests
if (kioskQueries.length > 0) {
  document.body.addEventListener("htmx:configRequest", function (e: HTMXEvent) {
    if (!e.detail?.parameters) {
      console.warn("Request parameters object not found");
      return;
    }

    try {
      kioskQueries.forEach((q: HTMLInputElement) => {
        if (!(q instanceof HTMLInputElement)) {
          console.warn(`Element ${q} is not an input`);
          return;
        }

        if (!q.name || !q.value) {
          console.debug(`Skipping invalid input: ${q}`);
          return;
        }

        const sanitizedValue = sanitiseInput(q.value);

        e.detail.parameters.append(q.name, sanitizedValue);
      });
    } catch (error) {
      console.error("Error processing parameters:", error);
    }
  });
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
  init();
});

export {
  cleanupFrames,
  startPolling,
  stopPolling,
  setRequestLock,
  releaseRequestLock,
  checkHistoryExists,
  clientData,
  videoHandler,
};
