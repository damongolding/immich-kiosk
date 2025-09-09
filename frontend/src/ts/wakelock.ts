export const preventSleep = async () => {
    let wakeLock: null | WakeLockSentinel = null;

    const requestWakeLock = async () => {
        if ("wakeLock" in navigator) {
            try {
                wakeLock = await navigator.wakeLock.request("screen");
                wakeLock.addEventListener("release", () => {
                    wakeLock = null;
                });
            } catch (err) {
                if (err instanceof TypeError) {
                    try {
                        // The "screen" parameter is not supported, try without it
                        wakeLock = await navigator.wakeLock.request();
                        wakeLock.addEventListener("release", () => {
                            wakeLock = null;
                        });
                    } catch (genericErr) {
                        console.error(
                            "Failed to acquire Wake Lock:",
                            genericErr,
                        );
                    }
                } else {
                    console.error("Error acquiring Wake Lock:", err);
                }
            }
        }
    };

    const handleVisibilityChange = async () => {
        if (document.visibilityState === "visible") {
            await requestWakeLock();
        }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    // Initial wake lock request
    await requestWakeLock();

    // Return a cleanup function
    return () => {
        document.removeEventListener(
            "visibilitychange",
            handleVisibilityChange,
        );
        wakeLock?.release();
    };
};
