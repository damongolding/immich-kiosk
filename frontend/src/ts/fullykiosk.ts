interface FullyKioskBrowser {
  getDisplayWidth: () => number;
  getDisplayHeight: () => number;
  getFullyVersion: () => string;
  getWebviewVersion: () => string;
  getAndroidVersion: () => string;
  getScreenOrientation: () => number;
  getScreenBrightness: () => number;
  getScreenOn: () => boolean;
  turnScreenOff: (keepAlive?: boolean) => void;
  turnScreenOn: () => void;
  showToast: (message: string) => void;
}

// Augment the Window interface
declare global {
  interface Window {
    fully: FullyKioskBrowser | undefined;
  }
}

class FullyKiosk {
  private static instance: FullyKiosk | null = null;
  public readonly fully: FullyKioskBrowser | undefined;

  private readonly SCREEN_OFF_DELAY_MS = 4 * 1000;

  private constructor() {
    this.fully = window.fully;
  }

  public static getInstance(): FullyKiosk {
    if (!FullyKiosk.instance) {
      FullyKiosk.instance = new FullyKiosk();
    }
    return FullyKiosk.instance;
  }

  public getDisplayDimensions(): { width: number; height: number } {
    if (this.fully === undefined) {
      return {
        width: window.innerWidth,
        height: window.innerHeight,
      };
    }
    return {
      width: this.fully.getDisplayWidth(),
      height: this.fully.getDisplayHeight(),
    };
  }

  public getVersionInfo(): { fully: string; webview: string; android: string } {
    if (this.fully === undefined) {
      return {
        fully: "unknown",
        webview: "unknown",
        android: "unknown",
      };
    }
    return {
      fully: this.fully.getFullyVersion(),
      webview: this.fully.getWebviewVersion(),
      android: this.fully.getAndroidVersion(),
    };
  }

  public toggleScreen(turnOff: boolean): void {
    if (this.fully === undefined) return;

    if (turnOff) {
      if (this.fully.getScreenOn()) {
        try {
          this.fully.showToast("Entering sleep mode");
          setTimeout(() => {
            if (this.fully) {
              this.fully.turnScreenOff(true);
            }
          }, this.SCREEN_OFF_DELAY_MS);
        } catch (error) {
          console.error("Error in Fully Kiosk screen operations:", error);
        }
      }
      return;
    }

    if (this.fully.getScreenOn() === false) {
      try {
        this.fully.turnScreenOn();
        this.fully.showToast("Exited sleep mode");
      } catch (error) {
        console.error("Error in Fully Kiosk screen operations:", error);
      }
    }
  }

  public showToast(message: string): void {
    if (this.fully === undefined) return;
    try {
      this.fully.showToast(message);
    } catch (error) {
      console.error("Error in Fully Kiosk toast operations:", error);
    }
  }
}

const fullyKiosk = FullyKiosk.getInstance();

export default fullyKiosk;
