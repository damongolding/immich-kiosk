export const wakeLock = async () => {
  if ("wakeLock" in navigator) {
    let wakeLock: null | WakeLockSentinel = null;
    // request a wake lock
    try {
      wakeLock = await navigator.wakeLock.request("screen");
      wakeLock.addEventListener("release", () => {
        console.log("Screen Wake Lock released:", wakeLock?.released);
      });
    } catch (err) {
      if (err.name === "TypeError") {
        // The "screen" parameter is not supported, try without it
        wakeLock = await navigator.wakeLock.request();
        wakeLock.addEventListener("release", () => {
          console.log("Screen Wake Lock released:", wakeLock?.released);
        });
      } else {
        console.error(`${err.name}, ${err.message}`);
      }
    }
  }
};
