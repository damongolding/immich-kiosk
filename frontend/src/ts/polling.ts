import htmx from "htmx.org";
import { hideImageOverlay } from "./menu";

let animationFrameId: number | null = null;
let progressBarElement: HTMLElement | null;
let lastPollTime: number | null = null;
let pausedTime: number | null = null;
let isPaused = false;

let pollInterval: number;
let kioskElement: HTMLElement | null;
let menuElement: HTMLElement | null;
let menuPausePlayButton: HTMLElement | null;

function initPolling(
  interval: number,
  kiosk: HTMLElement | null,
  menu: HTMLElement | null,
  pausePlayButton: HTMLElement | null,
) {
  pollInterval = interval;
  kioskElement = kiosk;
  menuElement = menu;
  menuPausePlayButton = pausePlayButton;
}

/**
 * Updates the kiosk display and progress bar
 * @param {number} timestamp - The current timestamp from requestAnimationFrame
 */
function updateKiosk(timestamp: number) {
  if (pausedTime !== null) {
    lastPollTime! += timestamp - pausedTime;
    pausedTime = null;
  }

  const elapsed = timestamp - lastPollTime!;
  const progress = Math.min(elapsed / pollInterval, 1);

  if (progressBarElement) {
    progressBarElement.style.width = `${progress * 100}%`;
  }

  if (elapsed >= pollInterval) {
    htmx.trigger(kioskElement as HTMLElement, "kiosk-new-image");
    lastPollTime = timestamp;
    stopPolling();
    return;
  }

  animationFrameId = requestAnimationFrame(updateKiosk);
}

type ProgressSource = {
  type: "image" | "video";
  startTime?: number;
  duration?: number;
  element?: HTMLVideoElement;
};

let currentProgressSource: ProgressSource | null = null;

function updateProgress(timestamp: number) {
  if (!progressBarElement) return;

  if (!currentProgressSource) return;

  let progress: number;

  if (currentProgressSource.type === "video") {
    if (!currentProgressSource.element) return;
    progress =
      (currentProgressSource.element.currentTime /
        currentProgressSource.element.duration) *
      100;
  } else {
    // image
    if (pausedTime !== null) {
      lastPollTime! += timestamp - pausedTime;
      pausedTime = null;
    }

    const elapsed = timestamp - lastPollTime!;
    progress = Math.min(elapsed / pollInterval, 1) * 100;

    if (elapsed >= pollInterval) {
      htmx.trigger(kioskElement as HTMLElement, "kiosk-new-image");
      lastPollTime = timestamp;
      stopPolling();
      return;
    }
  }

  progressBarElement.style.width = `${progress}%`;
  animationFrameId = requestAnimationFrame(updateProgress);
}

/**
 * Start the polling process to fetch new images
 */
function startPolling() {
  progressBarElement = htmx.find(".progress--bar") as HTMLElement | null;
  progressBarElement?.classList.remove("progress--bar-paused");

  menuElement?.classList.add("navigation-hidden");

  lastPollTime = performance.now();
  pausedTime = null;

  currentProgressSource = {
    type: "image",
    startTime: lastPollTime,
    duration: pollInterval,
  };

  animationFrameId = requestAnimationFrame(updateProgress);

  document.body.classList.remove("polling-paused");
  hideImageOverlay();

  isPaused = false;
}

/**
 * Stop the polling process
 */
function stopPolling() {
  if (isPaused && animationFrameId === null) return;

  cancelAnimationFrame(animationFrameId as number);

  progressBarElement?.classList.add("progress--bar-paused");
}

/**
 * Pause the polling process
 */
function pausePolling(showMenu: boolean = true) {
  if (isPaused && animationFrameId === null) return;

  cancelAnimationFrame(animationFrameId as number);
  pausedTime = performance.now();

  if (currentProgressSource?.type === "video" && video) {
    video.pause();
  }

  progressBarElement?.classList.add("progress--bar-paused");

  if (showMenu) {
    menuElement?.classList.remove("navigation-hidden");
    document.body.classList.add("polling-paused");
  }

  isPaused = true;
}

/**
 * Resume the polling process
 */
function resumePolling(hideOverlay: boolean = false) {
  if (!isPaused) return;

  if (currentProgressSource?.type === "video" && video) {
    video.play();
  } else {
    // For image polling
    currentProgressSource = {
      type: "image",
      startTime: performance.now(),
      duration: pollInterval,
    };
  }

  animationFrameId = requestAnimationFrame(updateProgress);

  progressBarElement?.classList.remove("progress--bar-paused");
  menuElement?.classList.add("navigation-hidden");

  document.body.classList.remove("polling-paused");
  if (hideOverlay) hideImageOverlay();

  isPaused = false;
}

/**
 * Toggle the polling state (pause/restart)
 */
function togglePolling(hideOverlay: boolean = false) {
  isPaused ? resumePolling(hideOverlay) : pausePolling();
}

function nextAsset() {
  htmx.trigger(kioskElement as HTMLElement, "kiosk-new-image");
}

let video: HTMLVideoElement | null = null;

function videoHandler(id: string): void {
  if (!id) return;

  video = document.getElementById(id) as HTMLVideoElement;

  if (!video) {
    console.error("Video element not found");
    return;
  }

  progressBarElement?.classList.remove("progress--bar-paused");

  currentProgressSource = {
    type: "video",
    element: video,
  };

  // Start the progress update
  animationFrameId = requestAnimationFrame(updateProgress);

  video.addEventListener(
    "ended",
    () => {
      if (animationFrameId !== null) {
        cancelAnimationFrame(animationFrameId);
      }
      videoCleanup();
      nextAsset();
    },
    false,
  );
}

// Clean up function when needed
function videoCleanup(): void {
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId);
  }

  lastPollTime = performance.now();

  currentProgressSource = {
    type: "image",
    startTime: lastPollTime,
    duration: pollInterval,
  };
}

export {
  initPolling,
  startPolling,
  pausePolling,
  stopPolling,
  nextAsset,
  resumePolling,
  togglePolling,
  videoHandler,
};
