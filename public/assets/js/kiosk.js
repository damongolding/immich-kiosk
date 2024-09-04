"use strict";

(() => {
  const kioskData = JSON.parse(
    document.getElementById("kiosk-data").textContent,
  );

  const pollInterval = htmx.parseInterval(`${kioskData.refresh}s`);
  let pollingInterval;

  let isPaused = false;
  let isFullscreen = false;

  const documentBody = document.body;
  const progressBar = htmx.find(".progress--bar");
  const fullscreenButton = htmx.find(".navigation--fullscreen");
  const menu = htmx.find(".navigation");

  // Utility functions for fullscreen actions
  function toggleFullscreen() {
    if (isFullscreen) {
      exitFullscreen();
    } else {
      enterFullscreen();
    }

    isFullscreen = !isFullscreen;
    if (fullscreenButton) {
      htmx.toggleClass(fullscreenButton, "navigation--fullscreen-enabled");
    }
  }

  function enterFullscreen() {
    if (documentBody.requestFullscreen) {
      documentBody.requestFullscreen();
    } else if (documentBody.mozRequestFullScreen) {
      documentBody.mozRequestFullScreen();
    } else if (documentBody.webkitRequestFullscreen) {
      documentBody.webkitRequestFullscreen();
    } else if (documentBody.msRequestFullscreen) {
      documentBody.msRequestFullscreen();
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
    if (progressBar) {
      htmx.removeClass(progressBar, "progress--bar-paused");
      resetAnimation(progressBar);
    }

    pollingInterval = setInterval(() => {
      htmx.trigger("#kiosk", "new-image");
    }, pollInterval);
  }

  function stopPolling() {
    if (progressBar) {
      htmx.addClass(progressBar, "progress--bar-paused");
    }

    clearInterval(pollingInterval);
  }

  function resetAnimation(element) {
    element.style.animation = "none";
    element.offsetHeight; // Trigger reflow
    element.style.animation = "";
  }

  // Event listeners
  htmx.on("#kiosk", "click", () => {
    if (menu) {
      if (isPaused) {
        startPolling();
      } else {
        stopPolling();
      }
      htmx.toggleClass(menu, "navigation-hidden");
    }

    isPaused = !isPaused;
  });

  htmx.on(".navigation--fullscreen", "click", toggleFullscreen);

  // Start polling on page load
  document.addEventListener("DOMContentLoaded", () => {
    if (!isPaused) {
      startPolling();
    }
  });
})();
