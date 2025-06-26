import { triggerNewAsset } from "./polling";

interface FullyKioskBrowser {
    getDisplayWidth: () => number;
    getDisplayHeight: () => number;
    getFullyVersion: () => string;
    getWebviewVersion: () => string;
    getAndroidVersion: () => string;
    getScreenOrientation: () => number;
    getScreenBrightness: () => number;
    startScreensaver: () => void;
    stopScreensaver: () => void;
    getScreenOn: () => boolean;
    turnScreenOff: (keepAlive?: boolean) => void;
    turnScreenOn: () => void;
    showToast: (message: string) => void;
    setScreenBrightness: (brightness: number) => void;
    getBooleanSetting: (key: string) => string;
    getStringSetting: (key: string) => string;
    setBooleanSetting: (key: string, value: boolean) => void;
    setStringSetting: (key: string, value: string) => void;
    importSettingsFile: (url: string) => void;
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

    private readonly SCREENSAVER_DELAY_MS = 4 * 1000;
    private readonly DEFAULT_MAX_BRIGHTNESS = "225";

    private readonly screensaverBrightness = "screensaverBrightness";
    private readonly screensaverWallpaperURL = "screensaverWallpaperURL";
    private readonly screensaverWallpaperURLBlack = "fully://color#000000";
    private readonly preventSleepWhileScreenOff = "preventSleepWhileScreenOff";
    private readonly screensaverDaydream = "screensaverDaydream";

    private initScreensaverBrightness: string;
    private initScreensaverWallpaperURL: string;
    private initScreensaverDaydream: boolean;

    inSleepMode: boolean = false;

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

    public getVersionInfo(): {
        fully: string;
        webview: string;
        android: string;
    } {
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

    private setScreensaverSettings() {
        if (!this.fully) return;

        // Get initial settings
        this.initScreensaverBrightness = this.fully.getStringSetting(
            this.screensaverBrightness,
        );
        if (Number(this.initScreensaverBrightness) < 1) {
            this.initScreensaverBrightness = this.DEFAULT_MAX_BRIGHTNESS;
        }

        this.initScreensaverWallpaperURL = this.fully.getStringSetting(
            this.screensaverWallpaperURL,
        );
        this.initScreensaverDaydream =
            this.fully.getBooleanSetting(this.screensaverDaydream) === "true";

        // Set settings
        this.fully.setStringSetting(this.screensaverBrightness, "0");
        this.fully.setStringSetting(
            this.screensaverWallpaperURL,
            this.screensaverWallpaperURLBlack,
        );
        this.fully.setBooleanSetting(this.preventSleepWhileScreenOff, true);
    }

    private resetScreensaverSettings() {
        if (!this.fully) return;

        this.fully.setStringSetting(
            this.screensaverBrightness,
            this.initScreensaverBrightness,
        );
        this.fully.setStringSetting(
            this.screensaverWallpaperURL,
            this.initScreensaverWallpaperURL,
        );
        this.fully.setBooleanSetting(
            this.screensaverDaydream,
            this.initScreensaverDaydream,
        );
    }

    public setScreensaverState(enable: boolean): void {
        if (!this.fully) return;

        try {
            if (enable) {
                if (this.inSleepMode) return;

                this.fully.showToast("Entering sleep mode");
                this.inSleepMode = true;

                const fullyRef = this.fully; // capture reference for setTimeout
                setTimeout(() => {
                    try {
                        this.setScreensaverSettings();
                        fullyRef.startScreensaver();
                    } catch (error) {
                        console.error("Error turning screen off:", error);
                    }
                }, this.SCREENSAVER_DELAY_MS);
            } else {
                if (!this.inSleepMode) return;

                this.resetScreensaverSettings();
                this.fully.stopScreensaver();
                this.fully.showToast("Exited sleep mode");
                this.inSleepMode = false;
                triggerNewAsset();
            }
        } catch (error) {
            console.error("Error in Fully Kiosk screen operations:", error);
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
