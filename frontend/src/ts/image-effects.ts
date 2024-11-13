import htmx from "htmx.org";
import type { KioskData } from "./kiosk";

function foo(kioskData: KioskData) {
  const frames = htmx.findAll(".frame");
  const newFrame = frames[frames.length - 1] as HTMLElement;
  const images = htmx.findAll(newFrame, ".frame--image");

  images.forEach((i: HTMLElement) => calculateBackgroundSize(i, kioskData));
}

function calculateBackgroundSize(el: HTMLElement, kioskData: KioskData) {
  const imageWidth = Number(el.dataset["imagewidth"]);
  const imageHeight = Number(el.dataset["imageheight"]);

  // Calculate aspect ratios
  const imgAspectRatio: number = imageWidth / imageHeight;
  const divAspectRatio: number = el.offsetWidth / el.offsetHeight;

  if (imgAspectRatio > divAspectRatio) {
    // Image is wider than div (relative to their heights)
    // Scale based on height to cover div fully
    el.style.backgroundSize = `auto ${kioskData.imageEffectAmount}%`;
  } else {
    // Image is taller than or equal to div (relative to their widths)
    // Scale based on width to cover div fully
    el.style.backgroundSize = `${kioskData.imageEffectAmount}% auto`;
  }
}

function initImageEffects(kioskData: KioskData) {
  if (kioskData.imageEffect === "pan") {
    htmx.on(htmx.find("#kiosk") as HTMLElement, "htmx:afterSwap", (e) =>
      foo(kioskData),
    );
  }
}

export { initImageEffects };
