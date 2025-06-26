let isFullscreen = false;

// Get the appropriate fullscreen API for the current browser
const fullscreenAPI = getFullscreenAPI();

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
function toggleFullscreen(
    documentBody: HTMLElement,
    fullscreenButton: HTMLElement | null,
) {
    if (isFullscreen) {
        if (fullscreenAPI.exitFullscreen) {
            (document as Document)[fullscreenAPI.exitFullscreen]();
        }
    } else {
        documentBody[fullscreenAPI.requestFullscreen as keyof Document]?.();
    }

    isFullscreen = !isFullscreen;
    fullscreenButton?.classList.toggle("navigation--fullscreen-enabled");
}

/**
 * Add fullscreen event listener
 */
function addFullscreenEventListener(fullscreenButton: HTMLElement | null) {
    document.addEventListener("fullscreenchange", () => {
        isFullscreen =
            !!document[fullscreenAPI.fullscreenElement as keyof string];
        fullscreenButton?.classList.toggle(
            "navigation--fullscreen-enabled",
            isFullscreen,
        );
    });
}

export { addFullscreenEventListener, fullscreenAPI, toggleFullscreen };
