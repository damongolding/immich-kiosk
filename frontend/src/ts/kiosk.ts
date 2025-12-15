import DOMPurify from "dompurify";
import htmx from "htmx.org";
import type { TimeFormat } from "./clock";
import { initClock } from "./clock";
import {
    addFullscreenEventListener,
    fullscreenAPI,
    toggleFullscreen,
} from "./fullscreen";
import fullyKiosk from "./fullykiosk";
import { livePhoto } from "./live-photo";
import {
    disableAssetNavigationButtons,
    enableAssetNavigationButtons,
    initMenu,
    toggleAssetOverlay,
    toggleRedirectsOverlay,
} from "./menu";
import { toggleMute } from "./mute";
import {
    initPolling,
    pausePolling,
    resumePolling,
    startPolling,
    stopPolling,
    togglePolling,
    triggerNewAsset,
    videoHandler,
} from "./polling";
import { sleepMode } from "./sleep";
import { preventSleep } from "./wakelock";

("use strict");

interface HTMXEvent extends Event {
    preventDefault: () => void;
    detail: {
        successful: boolean;
        parameters: FormData;
        method: string;
        pathInfo: {
            requestPath: string;
        };
    };
}

/**
 * Configuration data for managing the kiosk display and behavior
 */
type KioskData = {
    debug: boolean;
    debugVerbose: boolean;
    version: string;
    langCode: string;
    params: Record<string, unknown>;
    duration: number;
    disableNavigation: boolean;
    disableScreensaver: boolean;
    showDate: boolean;
    dateFormat: string;
    showTime: boolean;
    timeFormat: TimeFormat;
    clockSource: "client" | "server";
    transition: string;
    showMoreInfo: boolean;
    showRedirects: boolean;
    livePhotos: boolean;
    livePhotoLoopDelay: number;
    burnInInterval: number;
    burnInDuration: number;
    httpTimeout: number;
};

const MAX_FRAMES: number = 2 as const;

const TIMEOUT_RETRIES: number = 2 as const;
const timeouts: Record<string, number> = {};

// Parse kiosk data from the HTML element
const kioskData: KioskData = JSON.parse(
    document.getElementById("kiosk-data")?.textContent || "{}",
);

// Set polling interval based on the duration rate in kiosk data
const pollInterval = htmx.parseInterval(`${kioskData.duration}s`);

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
const toggleMuteMenuButton = htmx.find(
    ".navigation--mute",
) as HTMLElement | null;
const moreInfoButton = htmx.find(
    ".navigation--more-info",
) as HTMLElement | null;
const linksButton = htmx.find(".navigation--links") as HTMLElement | null;
const offlineSVG = htmx.find("#offline") as HTMLElement | null;

let requestInFlight = false;

let burnInTimerId: number | null = null;

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

    const MILLISECONDS_PER_SECOND = 1000;
    const TIMEOUT_GRACE_FACTOR = 3;

    if (kioskData.httpTimeout <= 0) {
        htmx.config.timeout = 0;
    } else {
        htmx.config.timeout =
            kioskData.httpTimeout *
            MILLISECONDS_PER_SECOND *
            TIMEOUT_GRACE_FACTOR;
    }

    if (
        kioskData.clockSource === "client" &&
        (kioskData.showDate || kioskData.showTime)
    ) {
        initClock(
            kioskData.showDate,
            kioskData.dateFormat,
            kioskData.showTime,
            kioskData.timeFormat,
            kioskData.langCode,
        );
    }

    if (kioskData.disableScreensaver) {
        await preventSleep();
    }

    if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/assets/js/sw.js").then(
            () => {
                console.log("ServiceWorker registration successful");
            },
            (err) => {
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

    initMenu(
        kioskData.disableNavigation,
        nextImageMenuButton as HTMLElement,
        prevImageMenuButton as HTMLElement,
        kioskData.showMoreInfo,
        handleInfoKeyPress,
        handleRedirectsKeyPress,
    );

    addEventListeners();

    if (kioskData.livePhotos) livePhoto(kioskData.livePhotoLoopDelay);

    // Burn-in prevention
    if (kioskData.burnInInterval > 0 && kioskData.burnInDuration > 0)
        burnInCycle();
}

function burnInCycle() {
    if (burnInTimerId !== null) {
        clearTimeout(burnInTimerId);
        burnInTimerId = null;
    }
    const runBurnInCycle = () => {
        document.body.classList.add("burn-in-dim");
        console.debug("Burn-in cycle started");

        setTimeout(() => {
            document.body.classList.remove("burn-in-dim");
            console.debug("Burn-in cycle ended");

            burnInTimerId = setTimeout(
                runBurnInCycle,
                kioskData.burnInInterval * 1000,
            );
        }, kioskData.burnInDuration * 1000);
    };

    burnInTimerId = setTimeout(runBurnInCycle, kioskData.burnInInterval * 1000);
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
 * Handles toggling of overlays with associated polling state management
 * @param overlayClass - CSS class name that identifies the overlay
 * @param toggleFn - Function to toggle the overlay visibility
 * @description Manages overlay state by:
 * - Checking current polling and overlay states
 * - Toggling polling if needed based on overlay state
 * - Toggling the overlay visibility
 */
function handleOverlayToggle(overlayClass: string, toggleFn: () => void): void {
    const isPollingPaused = document.body.classList.contains("polling-paused");
    const isOverlayOpen = document.body.classList.contains(overlayClass);

    if (isPollingPaused && isOverlayOpen) {
        togglePolling();
        toggleFn();
    } else {
        if (!isPollingPaused) {
            togglePolling();
        }
        toggleFn();
    }
}

/**
 * Handles keyboard shortcut for toggling image info overlay
 * @description Toggles the image info overlay and manages polling state
 * when 'i' key is pressed
 */
function handleInfoKeyPress(): void {
    handleOverlayToggle("more-info", toggleAssetOverlay);
}

/**
 * Handles keyboard shortcut for toggling redirects overlay
 * @description Toggles the redirects overlay and manages polling state
 * when 'r' key is pressed
 */
function handleRedirectsKeyPress(): void {
    handleOverlayToggle("redirects-open", toggleRedirectsOverlay);
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
 * - Mute button handling
 */
function addEventListeners(): void {
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
    htmx.on("htmx:afterRequest", (e: HTMXEvent) => {
        if (!offlineSVG) {
            console.error("offline svg missing");
            return;
        }

        if (e.detail.successful) {
            htmx.removeClass(offlineSVG, "offline");
            timeouts[e.detail.pathInfo.requestPath] = 0;
        } else {
            htmx.addClass(offlineSVG, "offline");
        }
    });

    htmx.on("htmx:timeout", (e: HTMXEvent) => {
        let currentTimeout = timeouts[e.detail.pathInfo.requestPath];

        currentTimeout =
            currentTimeout === undefined || Number.isNaN(currentTimeout)
                ? 1
                : currentTimeout + 1;

        timeouts[e.detail.pathInfo.requestPath] = currentTimeout;

        if (currentTimeout > TIMEOUT_RETRIES) {
            window.location.reload();
        }
    });

    if (kioskData.disableNavigation) {
        console.log("Navigation disabled");
        return;
    }

    // Pause/resume polling and show/hide menu
    menuInteraction?.addEventListener("click", () => togglePolling());
    menuPausePlayButton?.addEventListener("click", () => togglePolling());

    document.addEventListener("keydown", (e) => {
        if (e.target !== document.body) return;

        switch (e.code) {
            case "KeyP":
                if (!e.shiftKey) {
                    // Regular P
                    e.preventDefault();
                    pausePolling(true);
                } else {
                    // Shift + P
                    e.preventDefault();
                    resumePolling(true);
                }
                break;

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

            case "KeyR":
                if (!kioskData.showRedirects) return;
                if (e.ctrlKey || e.metaKey) return;
                e.preventDefault();
                handleRedirectsKeyPress();
                break;

            case "KeyM":
                if (!toggleMuteMenuButton) return;
                e.preventDefault();
                toggleMute();
                break;
        }
    });

    // Fullscreen
    fullscreenButton?.addEventListener("click", handleFullscreenClick);
    addFullscreenEventListener(fullscreenButton);

    // Toggle mute
    toggleMuteMenuButton?.addEventListener("click", toggleMute);

    // More info overlay
    moreInfoButton?.addEventListener("click", () => toggleAssetOverlay());

    // Links overlay
    linksButton?.addEventListener("click", () => toggleRedirectsOverlay());
}

/**
 * Performs DOM cleanup to prevent memory leaks
 * @description Performs two cleanup operations:
 * 1. Removes any script tags in the kiosk element after 1 second delay
 * 2. Removes oldest frame element when total frames exceed MAX_FRAMES
 *
 * This helps maintain optimal performance by preventing too many
 * elements accumulating in the DOM over time.
 *
 * @returns {Promise<void>} Resolves when cleanup is complete
 * @throws {Error} If frame removal operation fails
 */
async function cleanupFrames(): Promise<void> {
    const kioskScripts = htmx.findAll(kiosk as HTMLElement, "script");
    if (kioskScripts?.length) {
        kioskScripts.forEach((s) => {
            htmx.remove(s, 1000);
        });
    }

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
    const historyItems = htmx.findAll(
        ".kiosk-history--entry",
    ) as NodeListOf<HTMLInputElement>;
    if (
        requestInFlight ||
        historyItems.length < 2 ||
        (historyItems.length > 0 && historyItems[0].value[0] === "*")
    ) {
        e.preventDefault();
        return;
    }

    setRequestLock(e);
}

/**
 * Browser viewport dimensions
 * @description Contains current window dimensions:
 * - client_width: Width of the browser viewport in pixels
 * - client_height: Height of the browser viewport in pixels
 */
type BrowserData = {
    client_width: number;
    client_height: number;
    fully_version?: string;
    fully_webview_version?: string;
    fully_android_version?: string;
    fully_screen_orientation?: number;
    fully_screen_brightness?: number;
};

/**
 * Returns the current browser viewport dimensions and, if available, Fully Kiosk Browser details.
 *
 * When debug mode is enabled and Fully Kiosk Browser integration is present, the returned object includes version, orientation, and brightness information from the browser.
 *
 * @returns An object containing the viewport width and height, and optionally Fully Kiosk Browser details.
 */
function clientData(): BrowserData {
    const data: BrowserData = {
        client_width: fullyKiosk.getDisplayDimensions().width,
        client_height: fullyKiosk.getDisplayDimensions().height,
    };

    if (kioskData.debug && fullyKiosk.fully !== undefined) {
        data.fully_version = fullyKiosk.fully.getFullyVersion();
        data.fully_webview_version = fullyKiosk.fully.getWebviewVersion();
        data.fully_android_version = fullyKiosk.fully.getAndroidVersion();
        data.fully_screen_orientation = fullyKiosk.fully.getScreenOrientation();
        data.fully_screen_brightness = fullyKiosk.fully.getScreenBrightness();
    }

    return data;
}

// Add kiosk query parameters to HTMX requests
if (kioskQueries.length > 0) {
    document.body.addEventListener("htmx:configRequest", (e: HTMXEvent) => {
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

                const sanitizedValue = DOMPurify.sanitize(q.value);

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
    triggerNewAsset,
    cleanupFrames,
    startPolling,
    stopPolling,
    setRequestLock,
    releaseRequestLock,
    checkHistoryExists,
    clientData,
    videoHandler,
    sleepMode,
};
