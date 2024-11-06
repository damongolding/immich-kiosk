import htmx from "htmx.org";
import { pausePolling } from "./polling";

let gettingNewImage = false;

let kioskElement: HTMLElement | null;
let menuElement: HTMLElement | null;
let menuPausePlayButton: HTMLElement | null;

function initMenu(
  kiosk: HTMLElement | null,
  menu: HTMLElement | null,
  pausePlayButton: HTMLElement | null,
) {
  kioskElement = kiosk;
  menuElement = menu;
  menuPausePlayButton = pausePlayButton;

  htmx.on(kiosk as HTMLElement, "htmx:afterSettle", function (e: any) {
    gettingNewImage = false;
  });
}

function handleNextImageClick() {
  if (gettingNewImage) return;

  pausePolling();
  htmx.trigger(kioskElement as HTMLElement, "kiosk-new-image");
  gettingNewImage = true;
}

export { initMenu, handleNextImageClick };
