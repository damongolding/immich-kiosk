import { format } from "date-fns/format";
import type { Locale } from "date-fns/locale";
import { locales } from "./locales";

const CLOCK_UPDATE_INTERVAL = 5000;

const TIME_FORMATS = {
    TWELVE_HOUR: "h:mma" as const,
    TWENTY_FOUR_HOUR: "HH:mm" as const,
} as const;

type TimeFormat = "12" | "24";

interface ClockConfig {
    showDate: boolean;
    dateFormat: string;
    showTime: boolean;
    timeFormat: TimeFormat;
    langCode: string;
}

interface ClockElements {
    main: HTMLElement | null;
    date: HTMLElement | null;
    time: HTMLElement | null;
}

class Clock {
    private config: ClockConfig;
    private elements: ClockElements;
    private intervalId?: number;
    private langCode: string;
    private lang: Locale;

    constructor(config: ClockConfig) {
        this.config = config;
        this.elements = this.initialiseElements();
        this.langCode = this.config.langCode;
        this.lang = this.initialiseLang();
    }

    private initialiseLang(): Locale {
        const DEFAULT_LANG = "enGB";

        if (!this.langCode) {
            console.error(
                `No language code provided, defaulting to ${DEFAULT_LANG}`,
            );
            return locales[DEFAULT_LANG];
        }

        // Check for exact match
        if (locales[this.langCode]) {
            console.log(`Using language ${this.langCode}`);
            return locales[this.langCode];
        }

        const splitCode = this.langCode.split("_");

        if (splitCode.length > 1) {
            const joinedCode = splitCode[0] + splitCode[1];
            if (locales[joinedCode]) {
                console.log(`Using base language ${joinedCode}`);
                return locales[joinedCode];
            }
        }

        // Try base language code without region
        const baseCode = splitCode[0];
        if (locales[baseCode]) {
            console.log(`Using base language ${baseCode}`);
            return locales[baseCode];
        }

        // Fallback to default
        console.error(
            `Language ${this.langCode} not found, defaulting to ${DEFAULT_LANG}`,
        );

        return locales[DEFAULT_LANG];
    }

    private initialiseElements(): ClockElements {
        const main = document.getElementById("clock");
        if (!main) {
            console.warn(
                "Clock element not found - this is expected if UI is disabled",
            );
            return {
                main: null,
                date: null,
                time: null,
            };
        }
        return {
            main: main as HTMLElement,
            date: document.querySelector(".clock--date"),
            time: document.querySelector(".clock--time"),
        };
    }

    private updateDate(now: Date): void {
        if (!this.config.showDate || !this.elements.date) return;
        try {
            this.elements.date.innerHTML = format(now, this.config.dateFormat, {
                locale: this.lang,
            });
        } catch (error) {
            console.error("Error formatting date:", error);
            this.elements.date.innerHTML = now.toLocaleDateString();
        }
    }

    private updateTime(now: Date): void {
        if (!this.config.showTime || !this.elements.time) return;

        const timeFormat =
            this.config.timeFormat === "12"
                ? TIME_FORMATS.TWELVE_HOUR
                : TIME_FORMATS.TWENTY_FOUR_HOUR;

        try {
            const formattedTime = format(now, timeFormat, {
                locale: this.lang,
            });
            this.elements.time.innerHTML =
                this.config.timeFormat === "12"
                    ? formattedTime.toLowerCase()
                    : formattedTime;
        } catch (error) {
            console.error("Error formatting time:", error);
            this.elements.time.innerHTML = now.toLocaleTimeString();
        }
    }

    private render(): void {
        const now = new Date();
        this.updateDate(now);
        this.updateTime(now);
    }

    public start(): void {
        if (this.intervalId) {
            this.stop();
        }
        this.render();
        this.intervalId = window.setInterval(
            () => this.render(),
            CLOCK_UPDATE_INTERVAL,
        );
    }

    public stop(): void {
        if (this.intervalId) {
            window.clearInterval(this.intervalId);
            this.intervalId = undefined;
        }
    }
}

function initClock(
    kioskShowDate: boolean,
    kioskDateFormat: string,
    kioskShowTime: boolean,
    kioskTimeFormat: TimeFormat,
    kioskLangCode: string,
): Clock {
    const config: ClockConfig = {
        showDate: kioskShowDate,
        dateFormat: kioskDateFormat,
        showTime: kioskShowTime,
        timeFormat: kioskTimeFormat,
        langCode: kioskLangCode,
    };

    const clock = new Clock(config);
    clock.start();

    const handleUnload = () => clock.stop();
    window.addEventListener("unload", handleUnload);

    return Object.assign(clock, {
        cleanup: () => window.removeEventListener("unload", handleUnload),
    });
}

export { initClock, Clock, type TimeFormat };
