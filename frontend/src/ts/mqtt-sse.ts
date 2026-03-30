/**
 * mqtt-sse.ts
 * Listens to the /events SSE endpoint and translates incoming MQTT navigation
 * commands into the HTMX events that the kiosk already understands.
 */

import htmx from "htmx.org";

type KioskCommand = "next" | "previous";

/**
 * Initialises the SSE connection to /events.
 * Must be called after the kiosk element is available in the DOM.
 */
export function initMqttSSE(): void {
    const params = new URLSearchParams(window.location.search);
    const client = params.get("client");
    const url = client ? `/events?client=${encodeURIComponent(client)}` : "/events";
    const evtSource = new EventSource(url);

    evtSource.addEventListener("kiosk-command", (e: MessageEvent) => {
        const cmd = (e.data as string).trim() as KioskCommand;

        switch (cmd) {
            case "next": {
                const kioskEl = document.getElementById("kiosk");
                if (kioskEl) {
                    htmx.trigger(kioskEl, "kiosk-new-asset");
                }
                break;
            }
            case "previous": {
                const prevEl = document.getElementById(
                    "navigation-interaction-area--previous-asset",
                );
                if (prevEl) {
                    htmx.trigger(prevEl, "kiosk-prev-asset");
                }
                break;
            }
            default:
                console.warn("immich-kiosk: unknown MQTT command:", cmd);
        }
    });

    evtSource.onerror = () => {
        // Browser will automatically reconnect for EventSource; nothing extra needed.
        console.debug("immich-kiosk: SSE connection error, will retry automatically");
    };
}
