import { triggerNewAsset } from "./polling";

class ImmichFrame {
  private static instance: ImmichFrame | null = null;

  private readonly SCREENSAVER_DELAY_MS = 4 * 1000;

  private readonly PORT = 53287 as const;

  inSleepMode: boolean = false;
  // scheme: string = "http";

  // constructor() {
  //   // this.scheme = window.location.protocol;
  // }

  public static getInstance(): ImmichFrame {
    if (!ImmichFrame.instance) {
      ImmichFrame.instance = new ImmichFrame();
    }
    return ImmichFrame.instance;
  }

  public dimScreen(): void {
    fetch(`http://localhost:${this.PORT}/dim`).catch((error) =>
      console.error("Error dimming ImmichFrame screen:", error),
    );
  }

  public undimScreen(): void {
    fetch(`http://localhost:${this.PORT}/undim`).catch((error) =>
      console.error("Error undimming ImmichFrame screen:", error),
    );
  }

  public setScreensaverState(enable: boolean): void {
    try {
      if (enable) {
        if (this.inSleepMode) return;

        this.inSleepMode = true;

        const i = this; // capture reference for setTimeout
        setTimeout(() => i.dimScreen(), this.SCREENSAVER_DELAY_MS);
      } else {
        if (!this.inSleepMode) return;

        this.undimScreen();

        this.inSleepMode = false;

        triggerNewAsset();
      }
    } catch (error) {
      console.error("Error in ImmichFrame screen operations:", error);
    }
  }
}

const immichFrame = ImmichFrame.getInstance();

export default immichFrame;
