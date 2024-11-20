import { format } from "date-fns/format";

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
}

interface ClockElements {
  main: Element | null;
  date: Element | null;
  time: Element | null;
}

class Clock {
  private config: ClockConfig;
  private elements: ClockElements;
  private intervalId?: number;

  constructor(config: ClockConfig) {
    this.config = config;
    this.elements = this.initializeElements();
  }

  private initializeElements(): ClockElements {
    return {
      main: document.querySelector("clock"),
      date: document.querySelector(".clock--date"),
      time: document.querySelector(".clock--time"),
    };
  }

  private updateDate(now: Date): void {
    if (!this.config.showDate || !this.elements.date) return;
    this.elements.date.innerHTML = format(now, this.config.dateFormat);
  }

  private updateTime(now: Date): void {
    if (!this.config.showTime || !this.elements.time) return;

    const timeFormat =
      this.config.timeFormat === "12"
        ? TIME_FORMATS.TWELVE_HOUR
        : TIME_FORMATS.TWENTY_FOUR_HOUR;

    const formattedTime = format(now, timeFormat);
    this.elements.time.innerHTML =
      this.config.timeFormat === "12"
        ? formattedTime.toLowerCase()
        : formattedTime;
  }

  private render(): void {
    const now = new Date();
    this.updateDate(now);
    this.updateTime(now);
  }

  public start(): void {
    this.render();
    this.intervalId = window.setInterval(
      () => this.render(),
      CLOCK_UPDATE_INTERVAL,
    );
  }

  public stop(): void {
    if (this.intervalId) {
      window.clearInterval(this.intervalId);
    }
  }
}

function initClock(
  kioskShowDate: boolean,
  kioskDateFormat: string,
  kioskShowTime: boolean,
  kioskTimeFormat: TimeFormat,
): Clock {
  const config: ClockConfig = {
    showDate: kioskShowDate,
    dateFormat: kioskDateFormat,
    showTime: kioskShowTime,
    timeFormat: kioskTimeFormat,
  };

  const clock = new Clock(config);
  clock.start();
  return clock;
}

export { initClock, Clock, type TimeFormat };
