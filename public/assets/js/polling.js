const pollInterval = htmx.parseInterval(`${kioskData.refresh}s`);
let pollingInterval;
let isPaused = false;
let isFullscreen = false;

// Utility functions for fullscreen actions
function toggleFullscreen() {
  const body = document.body;

  if (isFullscreen) {
    exitFullscreen();
  } else {
    enterFullscreen();
  }

  // Toggle the fullscreen state and button class
  isFullscreen = !isFullscreen;
  htmx.toggleClass(
    htmx.find(".navigation--fullscreen"),
    "navigation--fullscreen-enabled",
  );
}

function enterFullscreen() {
  const body = document.body;

  if (body.requestFullscreen) {
    body.requestFullscreen();
  } else if (body.mozRequestFullScreen) {
    body.mozRequestFullScreen();
  } else if (body.webkitRequestFullscreen) {
    body.webkitRequestFullscreen();
  } else if (body.msRequestFullscreen) {
    body.msRequestFullscreen();
  }
}

async function exitFullscreen() {
  if (document.exitFullscreen) {
    await document.exitFullscreen();
  } else if (document.mozCancelFullScreen) {
    await document.mozCancelFullScreen();
  } else if (document.webkitExitFullscreen) {
    await document.webkitExitFullscreen();
  } else if (document.msExitFullscreen) {
    await document.msExitFullscreen();
  }
}

// Functions to manage polling and progress bar animation
function startPolling() {
  const progressBar = htmx.find(".progress--bar");

  htmx.removeClass(progressBar, "progress--bar-paused");
  resetAnimation(progressBar);

  pollingInterval = setInterval(() => {
    htmx.trigger("#kiosk", "new-image");
  }, pollInterval);
}

function stopPolling() {
  htmx.addClass(htmx.find(".progress--bar"), "progress--bar-paused");
  clearInterval(pollingInterval);
}

function resetAnimation(element) {
  element.style.animation = "none";
  element.offsetHeight; // Trigger reflow
  element.style.animation = "";
}

// Event listeners
htmx.on("#kiosk", "click", () => {
  const menu = htmx.find(".navigation");

  isPaused ? startPolling() : stopPolling();
  htmx.toggleClass(menu, "navigation-hidden");

  isPaused = !isPaused;
});

htmx.on(".navigation--fullscreen", "click", toggleFullscreen);

// Start polling on page load
htmx.on("DOMContentLoaded", () => {
  if (!isPaused) startPolling();
});
