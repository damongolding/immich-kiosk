/**
 * Manages the UI sleep mode state and controls device screen through Fully Kiosk API
 * @param {boolean} turnOn - Whether to enter (true) or exit (false) sleep mode
 */
function sleep(turnOn: boolean): void {
  const fk = window.fully;

  const SCREEN_OFF_DELAY_MS = 4 * 1000;

  if (turnOn) {
    document.body.classList.add("sleep");
    if (typeof fk !== "undefined" && fk.getScreenOn?.()) {
      try {
        fk.showToast?.("Entering sleep mode");
        setTimeout(() => fk.turnScreenOff?.(), SCREEN_OFF_DELAY_MS);
      } catch (error) {
        console.error("Error in Fully Kiosk screen operations:", error);
      }
    }
    return;
  }

  document.body.classList.remove("sleep");
  if (typeof fk !== "undefined" && !fk.getScreenOn?.()) {
    try {
      fk.turnScreenOn?.();
      fk.showToast?.("Exited sleep mode");
    } catch (error) {
      console.error("Error in Fully Kiosk screen operations:", error);
    }
  }
}

export { sleep };
