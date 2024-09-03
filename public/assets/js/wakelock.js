(async () => {
  if ("wakeLock" in navigator) {
    let wakeLock = null;

    // request a wake lock
    const requestWakeLock = async () => {
      try {
        wakeLock = await navigator.wakeLock.request("screen");
      } catch (err) {
        console.error(`${err.name}, ${err.message}`);
      }
    };

    document.addEventListener("visibilitychange", () => {
      if (wakeLock !== null && document.visibilityState === "visible") {
        requestWakeLock();
      }
    });

    await requestWakeLock();
  }
})();
