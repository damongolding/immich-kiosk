import htmx from "htmx.org";
import { hideAssetOverlay } from "./menu";
import { getMuteState, unmute } from "./mute";
import { storageUtils } from "./storage";

/**
 * Represents a source for progress tracking, either an image or video
 */
interface ProgressSource {
  type: "image" | "video";
  startTime?: number;
  duration?: number;
  element?: HTMLVideoElement;
}

class PollingController {
  private static instance: PollingController;

  private animationFrameId: number | null = null;
  private progressBarElement: HTMLElement | null = null;
  private lastPollTime: number | null = null;
  private pausedTime: number | null = null;
  private isPaused: boolean = false;
  private pollInterval: number = 0;
  private kioskElement: HTMLElement | null = null;
  private menuElement: HTMLElement | null = null;
  private currentProgressSource: ProgressSource | null = null;
  private video: HTMLVideoElement | null = null;
  private playTimeout: number | null;

  private constructor() {
    // Private constructor to enforce singleton pattern
  }

  /**
   * Returns the singleton instance of PollingController
   */
  public static getInstance(): PollingController {
    if (!PollingController.instance) {
      PollingController.instance = new PollingController();
    }
    return PollingController.instance;
  }

  /**
   * Initializes the polling controller with required parameters
   * @param interval - The polling interval in milliseconds
   * @param kiosk - The kiosk element to control
   * @param menu - The menu element to control
   */
  init(interval: number, kiosk: HTMLElement | null, menu: HTMLElement | null) {
    if (!interval || !kiosk) {
      throw new Error("PollingController: Missing required parameters");
    }
    this.pollInterval = interval;
    this.kioskElement = kiosk;
    this.menuElement = menu;
    this.progressBarElement = htmx.find(".progress--bar") as HTMLElement | null;
  }

  /**
   * Updates the progress bar based on current time and source type
   * @param timestamp - Current timestamp from requestAnimationFrame
   */
  private updateProgress = (timestamp: number) => {
    if (!this.currentProgressSource) return;

    if (this.pausedTime !== null) {
      this.lastPollTime! += timestamp - this.pausedTime;
      this.pausedTime = null;
    }

    let progress: number;

    if (this.currentProgressSource.type === "video") {
      if (!this.currentProgressSource.element) return;
      const video = this.currentProgressSource.element;
      progress = video.currentTime / video.duration;
    } else {
      const elapsed = timestamp - this.lastPollTime!;
      progress = Math.min(elapsed / this.pollInterval, 1);

      if (elapsed >= this.pollInterval) {
        this.triggerNewAsset();
        return;
      }
    }

    if (this.progressBarElement) {
      this.progressBarElement.style.transform = `scaleX(${progress}) translateZ(0)`;
      // this.progressBarElement.style.width = `${progress}%`;
    }

    this.animationFrameId = requestAnimationFrame(this.updateProgress);
  };

  /**
   * Triggers a new image to be loaded
   */
  private triggerNewAsset = () => {
    this.stopPolling();
    this.lastPollTime = performance.now();
    htmx.trigger(this.kioskElement as HTMLElement, "kiosk-new-asset");
  };

  /**
   * Handles video end event
   */
  private videoEndedHandler = () => {
    this.videoCleanup();
    this.triggerNewAsset();
  };

  /**
   * Starts the polling process
   */
  startPolling = () => {
    this.progressBarElement?.classList.remove("progress--bar-paused");
    this.menuElement?.classList.add("navigation-hidden");
    this.lastPollTime = performance.now();
    this.pausedTime = null;

    this.currentProgressSource = {
      type: "image",
      startTime: this.lastPollTime,
      duration: this.pollInterval,
    };

    this.animationFrameId = requestAnimationFrame(this.updateProgress);
    document.body.classList.remove("polling-paused");
    hideAssetOverlay();
    this.isPaused = false;
  };

  /**
   * Pauses the polling process
   * @param showMenu - Whether to show the menu when pausing
   */
  pausePolling = (showMenu: boolean = true) => {
    if (this.isPaused && this.animationFrameId === null) return;

    if (this.animationFrameId !== null) {
      cancelAnimationFrame(this.animationFrameId);
      this.animationFrameId = null;
    }

    this.pausedTime = performance.now();

    if (this.currentProgressSource?.type === "video" && this.video) {
      this.video.pause();
    }

    this.progressBarElement?.classList.add("progress--bar-paused");

    if (showMenu) {
      this.menuElement?.classList.remove("navigation-hidden");
      document.body.classList.add("polling-paused");
    }

    this.isPaused = true;
  };

  /**
   * Resumes the polling process
   * @param hideOverlay - Whether to hide the overlay when resuming
   */
  resumePolling = (hideOverlay: boolean = false) => {
    if (!this.isPaused || this.animationFrameId !== null) return;

    if (this.currentProgressSource?.type === "video" && this.video) {
      this.video.play();
    } else {
      this.currentProgressSource = {
        type: "image",
        startTime: performance.now(),
        duration: this.pollInterval,
      };
    }

    this.animationFrameId = requestAnimationFrame(this.updateProgress);
    this.progressBarElement?.classList.remove("progress--bar-paused");
    this.menuElement?.classList.add("navigation-hidden");
    document.body.classList.remove("polling-paused");

    if (hideOverlay) hideAssetOverlay();
    this.isPaused = false;
  };

  handleVideoError = (e: Event) => {
    console.error("Video playback error:", e);
    this.videoCleanup();
    this.triggerNewAsset();
  };

  // Function to clear timeout when video starts playing
  handlePlayStart = () => {
    const listener = () => {
      if (this.playTimeout) {
        clearTimeout(this.playTimeout);
      }
      this.video?.removeEventListener("playing", listener);
    };
    return listener;
  };

  muteVideo = () => {
    if (!this.video) return;
    this.video.muted = true;
  };

  unmuteVideo = () => {
    if (!this.video) return;
    this.video.muted = false;
  };

  /**
   * Handles video playback
   * @param id - The ID of the video element to handle
   */
  videoHandler = (id: string) => {
    if (!id) {
      console.error("No video ID provided");
      this.triggerNewAsset();
      return;
    }

    this.video = document.getElementById(id) as HTMLVideoElement;
    if (!this.video) {
      console.error("Video element not found");
      return;
    }

    if (navigator.userActivation?.hasBeenActive && localStorage) {
      const kioskVideoIsMuted = storageUtils.get<boolean>("kioskVideoIsMuted");
      if (kioskVideoIsMuted !== null && !kioskVideoIsMuted) {
        unmute();
      }
    } else {
      this.video.muted = getMuteState();
    }

    // Setup timeout to check if video starts playing
    this.playTimeout = setTimeout(() => {
      if (this.video && (this.video.paused || this.video.currentTime === 0)) {
        console.error("Video failed to start playing within timeout period");
        this.handleVideoTimeout();
      }
    }, 5000); // 5 seconds timeout

    // Add listener for when video starts playing
    this.video.addEventListener("playing", this.handlePlayStart(), {
      once: true,
    });

    this.progressBarElement?.classList.remove("progress--bar-paused");
    this.menuElement?.classList.add("navigation-hidden");
    this.lastPollTime = performance.now();
    this.pausedTime = null;

    this.currentProgressSource = {
      type: "video",
      element: this.video,
    };

    this.animationFrameId = requestAnimationFrame(this.updateProgress);
    document.body.classList.remove("polling-paused");
    hideAssetOverlay();

    if (!this.video?.paused) {
      this.video.play().catch((error) => {
        console.error("Video playback error:", error);
        if (this.playTimeout) {
          clearTimeout(this.playTimeout);
        }
        this.handleVideoError(error);
      });
    }

    this.video.addEventListener("error", this.handleVideoError, { once: true });
    this.video.addEventListener("ended", this.videoEndedHandler, {
      once: true,
    });

    this.isPaused = false;
  };

  private handleVideoTimeout = () => {
    console.error("Video playback timeout");

    // Cleanup current video
    this.videoCleanup();

    // Move to next asset
    this.triggerNewAsset();
  };

  /**
   * Cleans up video resources
   */
  private videoCleanup = () => {
    this.video?.removeEventListener("ended", this.videoEndedHandler);
    this.video?.removeEventListener("error", this.handleVideoError);

    this.progressBarElement?.classList.add("progress--bar-paused");

    this.video?.pause();
    this.video = null;

    this.currentProgressSource = null;

    if (this.animationFrameId) {
      cancelAnimationFrame(this.animationFrameId);
      this.animationFrameId = null;
    }
  };

  /**
   * Stops the polling process
   */
  stopPolling = () => {
    if (this.isPaused && this.animationFrameId === null) return;

    if (this.animationFrameId !== null) {
      cancelAnimationFrame(this.animationFrameId);
      this.animationFrameId = null;
    }

    this.progressBarElement?.classList.add("progress--bar-paused");
  };

  /**
   * Toggles between polling states
   * @param hideOverlay - Whether to hide overlay when resuming
   */
  togglePolling = (hideOverlay: boolean = false) => {
    this.isPaused ? this.resumePolling(hideOverlay) : this.pausePolling();
  };

  /**
   * Advances to the next asset
   */
  nextAsset = () => {
    this.triggerNewAsset();
  };
}

const pollingController = PollingController.getInstance();

export const initPolling = (
  interval: number,
  kiosk: HTMLElement | null,
  menu: HTMLElement | null,
) => pollingController.init(interval, kiosk, menu);

export const startPolling = () => pollingController.startPolling();
export const pausePolling = (showMenu?: boolean) =>
  pollingController.pausePolling(showMenu);
export const stopPolling = () => pollingController.stopPolling();
export const nextAsset = () => pollingController.nextAsset();
export const resumePolling = (hideOverlay?: boolean) =>
  pollingController.resumePolling(hideOverlay);
export const togglePolling = (hideOverlay?: boolean) =>
  pollingController.togglePolling(hideOverlay);
export const videoHandler = (id: string) => pollingController.videoHandler(id);
export const muteVideo = () => pollingController.muteVideo();
export const unmuteVideo = () => pollingController.unmuteVideo();
