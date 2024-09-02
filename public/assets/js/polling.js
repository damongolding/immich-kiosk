const pollInterval = htmx.parseInterval(`${kioskData.refresh}s`);
let pollingInterval;
let isPaused = false;

function startPolling() {
  const progressBar = htmx.find(".progress--bar");

  // Remove the paused class to resume animation
  htmx.removeClass(progressBar, "progress--bar-paused");

  // Restart the CSS animation
  progressBar.style.animation = "none";
  progressBar.offsetHeight; // Trigger reflow
  progressBar.style.animation = "";

  // Start the polling interval
  pollingInterval = setInterval(() => {
    htmx.trigger("#kiosk", "new-image");
  }, pollInterval);
}

function stopPolling() {
  // Add the paused class to pause the animation
  const progressBar = htmx.find(".progress--bar");
  htmx.addClass(progressBar, "progress--bar-paused");

  // Clear the interval to stop polling
  clearInterval(pollingInterval);
}

htmx.on("click", () => {
  if (isPaused) {
    startPolling();
  } else {
    stopPolling();
  }

  // Toggle the paused state
  isPaused = !isPaused;
});

htmx.on("DOMContentLoaded", () => {
  if (!isPaused) {
    startPolling();
  }
});
