import fullyKiosk from "./fullykiosk";

/**
 * Manages the UI sleep mode state and controls device screen through Fully Kiosk API
 * @param {boolean} turnOn - Whether to enter (true) or exit (false) sleep mode
 */
function sleepMode(turnOn: boolean, screensaver: boolean): void {
  if (turnOn) {
    document.body.classList.add("sleep");
  } else {
    document.body.classList.remove("sleep");
  }

  if (screensaver) fullyKiosk.setScreensaverState(turnOn);
}

export { sleepMode };
