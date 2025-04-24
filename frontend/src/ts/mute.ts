import htmx from "htmx.org";

import { muteVideo, unmuteVideo } from "./polling";

let isMuted = true;

const muteButton = htmx.find(".navigation--mute") as HTMLElement | null;

/**
 * Mutes video audio globally and updates the mute button UI state.
 */
function mute() {
  muteVideo();
  muteButton?.classList.add("is-muted");
  isMuted = true;
}

/**
 * Unmutes video audio globally and updates the mute button UI state.
 */
function unmute() {
  unmuteVideo();
  muteButton?.classList.remove("is-muted");
  isMuted = false;
}

/**
 * Toggles the global mute state and updates all video elements and the button icon.
 */
function toggleMute() {
  isMuted ? unmute() : mute();
}

/**
 * Returns the current global mute state.
 * @returns boolean indicating if audio is currently muted
 */
function getMuteState(): boolean {
  return isMuted;
}

export { toggleMute, getMuteState };
