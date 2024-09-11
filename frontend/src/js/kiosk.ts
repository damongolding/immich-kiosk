import htmx from "htmx.org";
import {
  fullscreenAPI,
  toggleFullscreen,
  addFullscreenEventListener,
} from "./fullscreen";
import { wakeLock } from "./wakelock";

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

let animationFrameId: number | null = null;
let progressBarElement: HTMLElement | null;

let isPaused = false;

let lastPollTime: number | null = null;
let pausedTime: number | null = null;

// Cache DOM elements for better performance
const documentBody = document.body;
const fullscreenButton = htmx.find(
  ".navigation--fullscreen",
) as HTMLElement | null;
const kiosk = htmx.find("#kiosk") as HTMLElement | null;
const menu = htmx.find(".navigation") as HTMLElement | null;
const menuPausePlayButton = htmx.find(
  ".navigation--control",
) as HTMLElement | null;

/**
 * Initialize Kiosk functionality
 */
function init() {
  if (kioskData.debugVerbose) {
    htmx.logAll();
  }

  if (kioskData.disableScreensaver) {
    wakeLock();
  }

  if (!fullscreenAPI.requestFullscreen) {
    fullscreenButton && htmx.remove(fullscreenButton);
  }

  addEventListeners();
}

/**
 * Updates the kiosk display and progress bar
 * @param {number} timestamp - The current timestamp from requestAnimationFrame
 */
function updateKiosk(timestamp: number) {
  if (pausedTime !== null) {
    // Adjust lastPollTime by the duration of the pause
    lastPollTime! += timestamp - pausedTime;
    pausedTime = null;
  }

  const elapsed = timestamp - lastPollTime!;
  const progress = Math.min(elapsed / pollInterval, 1);

  if (progressBarElement) {
    progressBarElement.style.width = `${progress * 100}%`;
  }

  if (elapsed >= pollInterval) {
    if (kiosk) {
      console.log("Trigger new image");
      htmx.trigger(kiosk, "kiosk-new-image");
    }
    lastPollTime = timestamp;
    stopPolling();
    return;
  }

  animationFrameId = requestAnimationFrame(updateKiosk);
}

/**
 * Start the polling process to fetch new images
 */
function startPolling() {
  progressBarElement = htmx.find(".progress--bar") as HTMLElement | null;
  progressBarElement?.classList.remove("progress--bar-paused");
  menuPausePlayButton?.classList.remove("navigation--control--paused");

  lastPollTime = performance.now();
  pausedTime = null;

  animationFrameId = requestAnimationFrame(updateKiosk);
}

/**
 * Stop the polling process
 */
function stopPolling() {
  if (isPaused && animationFrameId === null) return;

  cancelAnimationFrame(animationFrameId as number);

  progressBarElement?.classList.add("progress--bar-paused");
  menuPausePlayButton?.classList.add("navigation--control--paused");
}

function pausePolling() {
  if (isPaused && animationFrameId === null) return;

  cancelAnimationFrame(animationFrameId as number);
  pausedTime = performance.now();

  progressBarElement?.classList.add("progress--bar-paused");
  menuPausePlayButton?.classList.add("navigation--control--paused");
  menu?.classList.remove("navigation-hidden");

  isPaused = true;
}

function resumePolling() {
  if (!isPaused) return;

  animationFrameId = requestAnimationFrame(updateKiosk);

  progressBarElement?.classList.remove("progress--bar-paused");
  menuPausePlayButton?.classList.remove("navigation--control--paused");
  menu?.classList.add("navigation-hidden");

  isPaused = false;
}

/**
 * Toggle the polling state (pause/restart)
 */
function togglePolling() {
  isPaused ? resumePolling() : pausePolling();
}

function handleFullscreenClick() {
  toggleFullscreen(documentBody, fullscreenButton);
}

/**
 * Add event listeners to Kiosk elements
 */
function addEventListeners() {
  // Pause and show menu
  kiosk?.addEventListener("click", togglePolling);
  menuPausePlayButton?.addEventListener("click", togglePolling);

  fullscreenButton?.addEventListener("click", handleFullscreenClick);

  addFullscreenEventListener(fullscreenButton);

  // Server online check. Fires after every AJAX request.
  htmx.on("htmx:afterRequest", function (e: any) {
    const offline = htmx.find("#offline");

    if (e.detail.successful) {
      htmx.removeClass(offline, "offline");
    } else {
      htmx.addClass(offline, "offline");
    }
  });
}

/**
 * Remove first frame
 */
function cleanupFrames() {
  const frames = htmx.findAll(".frame");
  if (frames.length > 3) {
    htmx.remove(frames[0], 3000);
  }
}

// Initialize Kiosk when the DOM is fully loaded
htmx.onLoad(init);

export { startPolling, cleanupFrames };
