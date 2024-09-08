import type { htmx } from "htmx.org";

("use strict");

/**
 * Immediately Invoked Function Expression (IIFE) to encapsulate the kiosk functionality
 * and avoid polluting the global scope.
 */
(() => {
  // Parse kiosk data from the HTML element
  const kioskData = JSON.parse(
    document.getElementById("kiosk-data")?.textContent || "{}",
  );

  // Set polling interval based on the refresh rate in kiosk data
  const pollInterval = htmx.parseInterval(`${kioskData.refresh}s`);
  let pollingInterval: number;

  let lastUpdateTime = 0;
  let animationFrameId: number;
  let progressBarElement: HTMLElement | null;

  let isPaused = false;
  let isFullscreen = false;
  let triggerSent = false;

  // Cache DOM elements for better performance
  const documentBody = document.body;
  const progressBar = htmx.find(".progress--bar") as HTMLElement | null;
  const fullscreenButton = htmx.find(
    ".navigation--fullscreen",
  ) as HTMLElement | null;
  const kiosk = htmx.find("#kiosk") as HTMLElement | null;
  const menu = htmx.find(".navigation") as HTMLElement | null;
  const menuPausePlayButton = htmx.find(
    ".navigation--control",
  ) as HTMLElement | null;

  // Get the appropriate fullscreen API for the current browser
  const fullscreenAPI = getFullscreenAPI();

  /**
   * Initialize Kiosk functionality
   */
  function init() {
    if (!fullscreenAPI.requestFullscreen) {
      fullscreenButton && htmx.remove(fullscreenButton);
    }

    if (!isPaused) startPolling();

    addEventListeners();
  }

  /**
   * Updates the kiosk display and progress bar
   * @param {number} timestamp - The current timestamp from requestAnimationFrame
   */
  function updateKiosk(timestamp: number) {
    // Initialize lastUpdateTime if it's the first update
    if (!lastUpdateTime) lastUpdateTime = timestamp;

    // Calculate elapsed time and progress
    const elapsed = timestamp - lastUpdateTime;
    const triggerOffset = 500; // 0.5 second offset
    const progress = Math.min(elapsed / pollInterval, 1);

    // Update progress bar width
    if (progressBarElement) {
      progressBarElement.style.width = `${progress * 100}%`;
    }

    // Trigger new image 1 second before the interval has passed
    if (elapsed >= pollInterval - triggerOffset && !triggerSent) {
      console.log("Trigger new image");
      htmx.trigger(kiosk, "kiosk-new-image");
      triggerSent = true;
    }

    // Reset progress bar and lastUpdateTime when the full interval has passed
    if (elapsed >= pollInterval) {
      if (progressBarElement) {
        progressBarElement.style.width = "0%";
      }
      lastUpdateTime = timestamp;
      triggerSent = false;
    }

    // Schedule the next update
    animationFrameId = requestAnimationFrame(updateKiosk);
  }

  /**
   * Determine the correct fullscreen API methods for the current browser
   * @returns {Object} An object containing the appropriate fullscreen methods
   */
  function getFullscreenAPI(): {
    requestFullscreen: string | null;
    exitFullscreen: string | null;
    fullscreenElement: string | null;
    fullscreenEnabled: string | null;
  } {
    const apis = [
      [
        "requestFullscreen",
        "exitFullscreen",
        "fullscreenElement",
        "fullscreenEnabled",
      ],
      [
        "mozRequestFullScreen",
        "mozCancelFullScreen",
        "mozFullScreenElement",
        "mozFullScreenEnabled",
      ],
      [
        "webkitRequestFullscreen",
        "webkitExitFullscreen",
        "webkitFullscreenElement",
        "webkitFullscreenEnabled",
      ],
      [
        "msRequestFullscreen",
        "msExitFullscreen",
        "msFullscreenElement",
        "msFullscreenEnabled",
      ],
    ];

    for (const [request, exit, element, enabled] of apis) {
      if (request in document.documentElement) {
        return {
          requestFullscreen: request,
          exitFullscreen: exit,
          fullscreenElement: element,
          fullscreenEnabled: enabled,
        };
      }
    }

    return {
      requestFullscreen: null,
      exitFullscreen: null,
      fullscreenElement: null,
      fullscreenEnabled: null,
    };
  }

  /**
   * Toggle fullscreen mode
   */
  function toggleFullscreen() {
    if (isFullscreen) {
      document[fullscreenAPI.exitFullscreen as keyof Document]?.();
    } else {
      documentBody[fullscreenAPI.requestFullscreen as keyof HTMLElement]?.();
    }

    isFullscreen = !isFullscreen;
    fullscreenButton?.classList.toggle("navigation--fullscreen-enabled");
  }

  /**
   * Start the polling process to fetch new images
   */
  function startPolling() {
    progressBarElement = htmx.find(".progress--bar") as HTMLElement | null;
    progressBarElement?.classList.remove("progress--bar-paused");

    menuPausePlayButton?.classList.remove("navigation--control--paused");

    lastUpdateTime = 0;
    animationFrameId = requestAnimationFrame(updateKiosk);
  }

  /**
   * Stop the polling process
   */
  function stopPolling() {
    cancelAnimationFrame(animationFrameId);
    progressBarElement?.classList.add("progress--bar-paused");
    menuPausePlayButton?.classList.add("navigation--control--paused");
  }

  /**
   * Toggle the polling state (pause/restart)
   */
  function togglePolling() {
    isPaused ? startPolling() : stopPolling();
    menu?.classList.toggle("navigation-hidden");
    isPaused = !isPaused;
  }

  /**
   * Add event listeners to Kiosk elements
   */
  function addEventListeners() {
    // Pause and show menu
    kiosk?.addEventListener("click", togglePolling);
    menuPausePlayButton?.addEventListener("click", togglePolling);

    fullscreenButton?.addEventListener("click", toggleFullscreen);
    document.addEventListener("fullscreenchange", () => {
      isFullscreen =
        !!document[fullscreenAPI.fullscreenElement as keyof Document];
      fullscreenButton?.classList.toggle(
        "navigation--fullscreen-enabled",
        isFullscreen,
      );
    });
  }

  // Initialize Kiosk when the DOM is fully loaded
  document.addEventListener("DOMContentLoaded", init);
})();