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

    private timeoutId: number | null = null;

    public static getInstance(): ImmichFrame {
        if (!ImmichFrame.instance) {
            ImmichFrame.instance = new ImmichFrame();
        }
        return ImmichFrame.instance;
    }

    public async dimScreen(): Promise<void> {
        try {
            await fetch(`${this.BASE_URL}/${this.endpoints.DIM}`, {
                signal: AbortSignal.timeout(5000),
            });
        } catch (error) {
            console.debug("Error dimming ImmichFrame screen:", error);
        }
    }

    public async undimScreen(): Promise<void> {
        try {
            await fetch(`${this.BASE_URL}/${this.endpoints.UNDIM}`, {
                signal: AbortSignal.timeout(5000),
            });
        } catch (error) {
            console.debug("Error undimming ImmichFrame screen:", error);
        }
    }

    public async setScreensaverState(enable: boolean): Promise<void> {
        try {
            if (this.timeoutId) {
                clearTimeout(this.timeoutId);
                this.timeoutId = null;
            }

            if (enable) {
                if (this.inSleepMode) return;

                this.inSleepMode = true;

                this.timeoutId = setTimeout(async () => {
                    await this.dimScreen();
                    this.timeoutId = null;
                }, this.SCREENSAVER_DELAY_MS);
            } else {
                if (!this.inSleepMode) return;

                await this.undimScreen();

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
