import htmx from "htmx.org";

import { muteVideo, unmuteVideo } from "./polling";
import { storageUtils } from "./storage";

let isMuted = true;

const muteButton = htmx.find(".navigation--mute") as HTMLElement | null;

export type VideoMuteStatus = {
    muted: boolean;
    currentVideoMuted: boolean | null;
    hasCurrentVideo: boolean;
    userActivated: boolean | null;
};

/**
 * Mutes video audio globally and updates the mute button UI state.
 */
function mute() {
    muteVideo();
    muteButton?.classList.add("is-muted");
    isMuted = true;
    storageUtils.set("kioskVideoIsMuted", isMuted);
}

/**
 * Unmutes video audio globally and updates the mute button UI state.
 */
function unmute() {
    unmuteVideo();
    muteButton?.classList.remove("is-muted");
    isMuted = false;
    storageUtils.set("kioskVideoIsMuted", isMuted);
}

/**
 * Sets the global mute state and returns the resulting state.
 */
function setMuteState(muted: boolean): boolean {
    muted ? mute() : unmute();
    return getMuteState();
}

/**
 * Toggles the global mute state and updates all video elements and the button icon.
 */
function toggleMute(): boolean {
    return setMuteState(!isMuted);
}

/**
 * Returns the current global mute state.
 * @returns boolean indicating if audio is currently muted
 */
function getMuteState(): boolean {
    return isMuted;
}

/**
 * Returns the current mute preference and the active video element's mute state.
 */
function getMuteStatus(): VideoMuteStatus {
    const currentVideo = document.querySelector("video") as
        | HTMLVideoElement
        | null;

    return {
        muted: getMuteState(),
        currentVideoMuted: currentVideo?.muted ?? null,
        hasCurrentVideo: currentVideo !== null,
        userActivated: navigator.userActivation?.hasBeenActive ?? null,
    };
}

/**
 * Exposes a small browser API for external kiosk controllers.
 */
function registerVideoMuteApi() {
    window.immichKiosk = {
        ...window.immichKiosk,
        video: {
            ...window.immichKiosk?.video,
            getStatus: getMuteStatus,
            getMuted: getMuteState,
            setMuted: setMuteState,
            toggleMuted: toggleMute,
        },
    };
}

export {
    toggleMute,
    getMuteState,
    getMuteStatus,
    registerVideoMuteApi,
    setMuteState,
    unmute,
};
