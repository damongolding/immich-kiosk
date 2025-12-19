/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 * @description Controls menu behavior and navigation between assets in a kiosk interface
 */

import htmx from "htmx.org";

let disableNavigation: boolean = false;

let nextAssetMenuButton: HTMLElement;
let prevAssetMenuButton: HTMLElement;

let assetOverlayVisible: boolean = false;
let linkOverlayVisible: boolean = false;

const redirectsContainer = document.getElementById(
    "redirects-container",
) as HTMLElement | null;
let redirects: NodeListOf<HTMLAnchorElement> | null;
let currentRedirectIndex = -1;

let allowMoreInfo: boolean;
let infoKeyPress: () => void;
let redirectsKeyPress: () => void;

/**
 * Disables both next and previous asset navigation buttons
 * @returns {void}
 */
function disableAssetNavigationButtons(): void {
    if (disableNavigation) return;
    if (!nextAssetMenuButton || !prevAssetMenuButton) {
        console.debug("Navigation buttons not initialized.");
        return;
    }
    htmx.addClass(nextAssetMenuButton, "disabled");
    htmx.addClass(prevAssetMenuButton, "disabled");
}

/**
 * Enables both next and previous asset navigation buttons
 * @returns {void}
 */
function enableAssetNavigationButtons(): void {
    if (disableNavigation) return;
    if (!nextAssetMenuButton || !prevAssetMenuButton) {
        console.error("Navigation buttons not initialized");
        return;
    }
    htmx.removeClass(nextAssetMenuButton, "disabled");
    htmx.removeClass(prevAssetMenuButton, "disabled");
}

/**
 * Shows the asset information overlay
 * Only works when polling is paused
 * @returns {void}
 */
function showAssetOverlay(): void {
    if (!document.body) return;
    if (!document.body.classList.contains("polling-paused")) return;
    hideRedirectsOverlay();
    document.body.classList.add("more-info");
    assetOverlayVisible = true;
}

/**
 * Hides the asset information overlay
 * @returns {void}
 */
function hideAssetOverlay(): void {
    if (!document.body) return;
    document.body.classList.remove("more-info");
    assetOverlayVisible = false;
}

/**
 * Toggles the asset information overlay visibility
 * @returns {void}
 */
function toggleAssetOverlay(): void {
    assetOverlayVisible ? hideAssetOverlay() : showAssetOverlay();
}

function redirectKeyHandler(e: KeyboardEvent) {
    if (!redirects) return;

    switch (e.code) {
        case "ArrowDown":
            e.preventDefault(); // Prevent page scrolling
            currentRedirectIndex =
                (currentRedirectIndex + 1) % redirects.length;
            redirects[currentRedirectIndex].focus();
            break;
        case "ArrowUp":
            e.preventDefault(); // Prevent page scrolling
            currentRedirectIndex =
                (currentRedirectIndex - 1 + redirects.length) %
                redirects.length;
            redirects[currentRedirectIndex].focus();
            break;
        case "KeyI":
            if (!allowMoreInfo) return;
            e.preventDefault();
            infoKeyPress();
            break;
        case "KeyR":
            if (e.ctrlKey || e.metaKey) return;
            e.preventDefault();
            redirectsKeyPress();
            break;
    }
}

/**
 * Shows the links overlay
 * Only works when polling is paused
 * Hides image overlay if visible
 */
function showRedirectsOverlay(): void {
    if (!document.body) return;
    if (!document.body.classList.contains("polling-paused")) return;

    document.addEventListener("keydown", redirectKeyHandler);

    hideAssetOverlay();
    document.body.classList.add("redirects-open");
    linkOverlayVisible = true;
}

/**
 * Hides the links overlay
 */
function hideRedirectsOverlay(): void {
    if (!document.body) return;
    document.body.classList.remove("redirects-open");

    document.removeEventListener("keydown", redirectKeyHandler);

    linkOverlayVisible = false;
}

/**
 * Toggles the links overlay visibility
 */
function toggleRedirectsOverlay(): void {
    linkOverlayVisible ? hideRedirectsOverlay() : showRedirectsOverlay();
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param nextAssetButton - The next image navigation button element
 * @param prevAssetButton - The previous image navigation button element
 * @throws {Error} If either navigation button is not provided
 * @returns {void}
 */
function initMenu(
    disableNav: boolean,
    nextAssetButton: HTMLElement | null,
    prevAssetButton: HTMLElement | null,
    showMoreInfo: boolean,
    handleInfoKeyPress: () => void,
    handleRedirectsKeyPress: () => void,
): void {
    if (disableNav) {
        disableNavigation = disableNav;
        return;
    }

    if (!nextAssetButton || !prevAssetButton) {
        throw new Error("Both navigation buttons must be provided");
    }

    nextAssetMenuButton = nextAssetButton;
    prevAssetMenuButton = prevAssetButton;

    if (redirectsContainer) {
        redirects = redirectsContainer.querySelectorAll("a");
    }

    allowMoreInfo = showMoreInfo;
    infoKeyPress = handleInfoKeyPress;
    redirectsKeyPress = handleRedirectsKeyPress;
}

export {
    initMenu,
    disableAssetNavigationButtons,
    enableAssetNavigationButtons,
    showAssetOverlay,
    hideAssetOverlay,
    toggleAssetOverlay,
    toggleRedirectsOverlay,
};
