import fullyKiosk from "./fullykiosk";
import immichFrame from "./immichframe";

/**
 * Toggles UI sleep mode and optionally controls the device screensaver via the Fully Kiosk API.
 *
 * Adds or removes the "sleep" CSS class on the document body based on {@link turnOn}. If {@link screensaver} is true, also enables or disables the device screensaver accordingly.
 *
 * @param turnOn - If true, enters sleep mode; if false, exits sleep mode.
 * @param screensaver - If true, toggles the device screensaver state to match {@link turnOn}.
 */
function sleepMode(turnOn: boolean, screensaver: boolean): void {
  if (turnOn) {
    document.body.classList.add("sleep");
  } else {
    document.body.classList.remove("sleep");
  }

  if (screensaver) {
    fullyKiosk.setScreensaverState(turnOn);
    immichFrame.setScreensaverState(turnOn);
  }
}

export { sleepMode };
