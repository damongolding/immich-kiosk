import htmx from "htmx.org";

let ticker: number;

function preload() {
  clearTimeout(ticker);

  ticker = setTimeout(() => {
    htmx.trigger(htmx.find("#preload") as HTMLElement, "preload");
  }, 10000);
}

export { preload };
