"use strict";
var __values = (this && this.__values) || function(o) {
    var s = typeof Symbol === "function" && Symbol.iterator, m = s && o[s], i = 0;
    if (m) return m.call(o);
    if (o && typeof o.length === "number") return {
        next: function () {
            if (o && i >= o.length) o = void 0;
            return { value: o && o[i++], done: !o };
        }
    };
    throw new TypeError(s ? "Object is not iterable." : "Symbol.iterator is not defined.");
};
var __read = (this && this.__read) || function (o, n) {
    var m = typeof Symbol === "function" && o[Symbol.iterator];
    if (!m) return o;
    var i = m.call(o), r, ar = [], e;
    try {
        while ((n === void 0 || n-- > 0) && !(r = i.next()).done) ar.push(r.value);
    }
    catch (error) { e = { error: error }; }
    finally {
        try {
            if (r && !r.done && (m = i["return"])) m.call(i);
        }
        finally { if (e) throw e.error; }
    }
    return ar;
};
/**
 * Immediately Invoked Function Expression (IIFE) to encapsulate the kiosk functionality
 * and avoid polluting the global scope.
 */
(function () {
    var _a;
    // Parse kiosk data from the HTML element
    var kioskData = JSON.parse(((_a = document.getElementById("kiosk-data")) === null || _a === void 0 ? void 0 : _a.textContent) || '{}');
    // Set polling interval based on the refresh rate in kiosk data
    var pollInterval = htmx.parseInterval("".concat(kioskData.refresh, "s"));
    var pollingInterval;
    var lastUpdateTime = 0;
    var animationFrameId;
    var progressBarElement;
    var isPaused = false;
    var isFullscreen = false;
    var triggerSent = false;
    // Cache DOM elements for better performance
    var documentBody = document.body;
    var progressBar = htmx.find(".progress--bar");
    var fullscreenButton = htmx.find(".navigation--fullscreen");
    var kiosk = htmx.find("#kiosk");
    var menu = htmx.find(".navigation");
    var menuPausePlayButton = htmx.find(".navigation--control");
    // Get the appropriate fullscreen API for the current browser
    var fullscreenAPI = getFullscreenAPI();
    /**
     * Initialize Kiosk functionality
     */
    function init() {
        if (!fullscreenAPI.requestFullscreen) {
            fullscreenButton && htmx.remove(fullscreenButton);
        }
        if (!isPaused)
            startPolling();
        addEventListeners();
    }
    /**
     * Updates the kiosk display and progress bar
     * @param {number} timestamp - The current timestamp from requestAnimationFrame
     */
    function updateKiosk(timestamp) {
        // Initialize lastUpdateTime if it's the first update
        if (!lastUpdateTime)
            lastUpdateTime = timestamp;
        // Calculate elapsed time and progress
        var elapsed = timestamp - lastUpdateTime;
        var triggerOffset = 500; // 0.5 second offset
        var progress = Math.min(elapsed / pollInterval, 1);
        // Update progress bar width
        if (progressBarElement) {
            progressBarElement.style.width = "".concat(progress * 100, "%");
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
    function getFullscreenAPI() {
        var e_1, _a;
        var apis = [
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
        try {
            for (var apis_1 = __values(apis), apis_1_1 = apis_1.next(); !apis_1_1.done; apis_1_1 = apis_1.next()) {
                var _b = __read(apis_1_1.value, 4), request = _b[0], exit = _b[1], element = _b[2], enabled = _b[3];
                if (request in document.documentElement) {
                    return {
                        requestFullscreen: request,
                        exitFullscreen: exit,
                        fullscreenElement: element,
                        fullscreenEnabled: enabled,
                    };
                }
            }
        }
        catch (e_1_1) { e_1 = { error: e_1_1 }; }
        finally {
            try {
                if (apis_1_1 && !apis_1_1.done && (_a = apis_1.return)) _a.call(apis_1);
            }
            finally { if (e_1) throw e_1.error; }
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
        var _a, _b;
        if (isFullscreen) {
            (_a = document[fullscreenAPI.exitFullscreen]) === null || _a === void 0 ? void 0 : _a.call(document);
        }
        else {
            (_b = documentBody[fullscreenAPI.requestFullscreen]) === null || _b === void 0 ? void 0 : _b.call(documentBody);
        }
        isFullscreen = !isFullscreen;
        fullscreenButton === null || fullscreenButton === void 0 ? void 0 : fullscreenButton.classList.toggle("navigation--fullscreen-enabled");
    }
    /**
     * Start the polling process to fetch new images
     */
    function startPolling() {
        progressBarElement = htmx.find(".progress--bar");
        progressBarElement === null || progressBarElement === void 0 ? void 0 : progressBarElement.classList.remove("progress--bar-paused");
        menuPausePlayButton === null || menuPausePlayButton === void 0 ? void 0 : menuPausePlayButton.classList.remove("navigation--control--paused");
        lastUpdateTime = 0;
        animationFrameId = requestAnimationFrame(updateKiosk);
    }
    /**
     * Stop the polling process
     */
    function stopPolling() {
        cancelAnimationFrame(animationFrameId);
        progressBarElement === null || progressBarElement === void 0 ? void 0 : progressBarElement.classList.add("progress--bar-paused");
        menuPausePlayButton === null || menuPausePlayButton === void 0 ? void 0 : menuPausePlayButton.classList.add("navigation--control--paused");
    }
    /**
     * Toggle the polling state (pause/restart)
     */
    function togglePolling() {
        isPaused ? startPolling() : stopPolling();
        menu === null || menu === void 0 ? void 0 : menu.classList.toggle("navigation-hidden");
        isPaused = !isPaused;
    }
    /**
     * Add event listeners to Kiosk elements
     */
    function addEventListeners() {
        // Pause and show menu
        kiosk === null || kiosk === void 0 ? void 0 : kiosk.addEventListener("click", togglePolling);
        menuPausePlayButton === null || menuPausePlayButton === void 0 ? void 0 : menuPausePlayButton.addEventListener("click", togglePolling);
        fullscreenButton === null || fullscreenButton === void 0 ? void 0 : fullscreenButton.addEventListener("click", toggleFullscreen);
        document.addEventListener("fullscreenchange", function () {
            isFullscreen = !!document[fullscreenAPI.fullscreenElement];
            fullscreenButton === null || fullscreenButton === void 0 ? void 0 : fullscreenButton.classList.toggle("navigation--fullscreen-enabled", isFullscreen);
        });
    }
    // Initialize Kiosk when the DOM is fully loaded
    document.addEventListener("DOMContentLoaded", init);
})();
//# sourceMappingURL=kiosk.js.map