import { triggerNewAsset } from "./polling";

class ImmichFrame {
  private static instance: ImmichFrame | null = null;

  private readonly SCREENSAVER_DELAY_MS = 4 * 1000;

  private readonly PORT = 53287 as const;
  private readonly BASE_URL = `http://localhost:${this.PORT}`;

  private readonly endpoints = {
    DIM: "dim",
    UNDIM: "undim",
    NEXT: "next",
    PREVIOUS: "previous",
    PAUSE: "pause",
    SETTINGS: "settings",
  };

  inSleepMode: boolean = false;

  public static getInstance(): ImmichFrame {
    if (!ImmichFrame.instance) {
      ImmichFrame.instance = new ImmichFrame();
    }
    return ImmichFrame.instance;
  }

  public dimScreen(): void {
    fetch(`${this.BASE_URL}/${this.endpoints.DIM}`).catch((error) =>
      console.error("Error dimming ImmichFrame screen:", error),
    );
  }

  public undimScreen(): void {
    fetch(`${this.BASE_URL}/${this.endpoints.UNDIM}`).catch((error) =>
      console.error("Error undimming ImmichFrame screen:", error),
    );
  }

  public setScreensaverState(enable: boolean): void {
    try {
      if (enable) {
        if (this.inSleepMode) return;

        this.inSleepMode = true;

        setTimeout(() => this.dimScreen(), this.SCREENSAVER_DELAY_MS);
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
