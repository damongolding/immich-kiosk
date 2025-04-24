import htmx from "htmx.org";

let isGloballyMuted = true;

const muteButton = htmx.find(".navigation--toggle-mute") as HTMLElement | null;

/**
 * Toggles the global mute state and updates all video elements and the button icon.
 */
export function toggleMute() {
  isGloballyMuted = !isGloballyMuted;

  applyMuteStateToVideos();

  if (muteButton) {
    muteButton.classList.toggle("is-muted", isGloballyMuted);
  }
}

// Set initial button state on load
if (muteButton) {
  muteButton.classList.toggle("is-muted", isGloballyMuted);
}

export function applyMuteStateToVideos() {
  const newVideos = htmx.findAll("video") as NodeListOf<HTMLVideoElement>;
  newVideos.forEach((video) => (video.muted = isGloballyMuted));
}

export function getMuteState(): boolean {
  return isGloballyMuted;
}
