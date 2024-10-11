export const wakeLock = async () => {
  if ("wakeLock" in navigator) {
    let wakeLock: null | WakeLockSentinel = null;

    // request a wake lock
    const requestWakeLock = async () => {
      try {
        wakeLock = await navigator.wakeLock.request("screen");
      } catch (err) {
        if (err.name === "TypeError") {
          // The "screen" parameter is not supported, try without it
          wakeLock = await navigator.wakeLock.request();
        } else {
          console.error(`${err.name}, ${err.message}`);
        }
      }
    };

    document.addEventListener("visibilitychange", () => {
      if (wakeLock !== null && document.visibilityState === "visible") {
        requestWakeLock();
      }
    });

    await requestWakeLock();
  }
};
