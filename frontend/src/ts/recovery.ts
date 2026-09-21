import { stopPolling } from "./polling";

const RETRY_MS = 3000 as const;

let recoveryInFlight = false;

let recoveryModeActive = false;

/**
 * Probes `/health` and, once it answers, reloads onto a fresh kiosk page.
 * @description Only ever called while the offline overlay is up, i.e. after we
 * have already confirmed the server was unreachable. Reloading here is safe
 * because the navigation is guaranteed to succeed.
 */
function probeHealthAndRecover(): void {
    var controller = new AbortController();
    var timeoutID = setTimeout(() => {
        controller.abort();
    }, RETRY_MS);

    fetch("/health", {
        method: "GET",
        cache: "no-store",
        signal: controller.signal,
    })
        .then((response) => {
            if (response.ok) {
                window.location.reload();
            }
        })
        .catch(() => {
            /* still down — keep waiting */
        })
        .then(() => {
            clearTimeout(timeoutID);
            recoveryInFlight = false;
        });
}

function onOfflineVisibilityChange(): void {
    // iOS suspends timers while an installed PWA is backgrounded; re-probe as
    // soon as it comes back to the foreground.
    if (!document.hidden) {
        probeHealthAndRecover();
    }
}

/**
 * Covers the (still rendered) kiosk with a full-screen "unavailable" overlay and
 * starts polling `/health` until the server is back.
 * @description We deliberately do NOT navigate or reload while the server is
 * down. An installed iOS PWA in standalone mode does not route a failed
 * top-level navigation through the service worker, so `location.reload()` during
 * an outage lands on Safari's native "cannot open page" screen and the
 * home-screen app is stuck until it is force-quit. Keeping the current document
 * alive and only reloading once `/health` answers avoids that entirely; the
 * service worker's offline page stays as the safety net for a cold start.
 */
function enterRecoveryMode(): void {
    if (recoveryModeActive) return;
    recoveryModeActive = true;

    stopPolling();

    window.reconnectAnimation?.();

    const el = document.getElementById("kiosk-offline-overlay");
    if (el) {
        el.style.display = "flex";
        el.offsetHeight;
        el.classList.add("visible");
    }

    window.setInterval(probeHealthAndRecover, 3000);
    document.addEventListener("visibilitychange", onOfflineVisibilityChange);
    window.addEventListener("pageshow", probeHealthAndRecover);

    probeHealthAndRecover();
}

function tryRecoveryMode(): void {
    if (recoveryInFlight) return;
    recoveryInFlight = true;

    fetch("/health", { method: "GET", cache: "no-store" })
        .then((response) => {
            if (!response.ok) {
                enterRecoveryMode();
            }
        })
        .catch(() => enterRecoveryMode());

    recoveryInFlight = false;
}

export { tryRecoveryMode };
