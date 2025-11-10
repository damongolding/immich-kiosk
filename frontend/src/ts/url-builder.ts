import Choices from "choices.js";
import "htmx.org";

type choicesOptions = Record<string, unknown>;

/**
 * Initializes a Choices.js multiselect on the DOM element with the given id and wires change events.
 *
 * If the element exists, creates a multiselect with a placeholder of `Select {displayName}` and a remove-item button, merging any provided `options` into the defaults. When the selection changes, a `multiselect-change` event is dispatched on the document body.
 *
 * @param elementId - The id of the target DOM element to convert into a multiselect
 * @param displayName - A human-readable name used in the placeholder (rendered as `Select {displayName}`)
 * @param options - Optional Choices.js configuration to merge with the defaults
 */
function initMultiselect(
    elementId: string,
    displayName: string,
    options?: choicesOptions,
) {
    const multiSelect = document.getElementById(elementId);
    if (multiSelect) {
        const choicesOptions: choicesOptions = {
            placeholderValue: `Select ${displayName}`,
            removeItemButton: true,
            ...(options || {}),
        };
        const _multiSelectChoices = new Choices(multiSelect, choicesOptions);
        multiSelect.addEventListener(
            "change",
            () => {
                document
                    .querySelector("body")
                    ?.dispatchEvent(new Event("multiselect-change"));
            },
            false,
        );
    }
}

/**
 * Copy the provided text to the clipboard and show temporary visual feedback on the button.
 *
 * Attempts to write `text` to the system clipboard; if that fails or is unavailable, falls back to a DOM-based copy method.
 *
 * @param btn - The button element to which success or failure classes (`copied`, `copy-failed`) will be applied temporarily
 * @param text - The text to copy to the clipboard
 */
async function copyToClipboard(
    btn: HTMLButtonElement,
    text: string,
): Promise<void> {
    try {
        if (navigator.clipboard) {
            await navigator.clipboard.writeText(text);
        } else {
            const textarea = document.createElement("textarea");
            textarea.value = text;
            document.body.appendChild(textarea);
            textarea.select();
            document.execCommand("copy");
            document.body.removeChild(textarea);
        }
        btn.classList.add("copied");
        setTimeout(() => {
            btn.classList.remove("copied");
        }, 2000);
    } catch (error) {
        console.error("Failed to copy text:", error);
        btn.classList.add("copy-failed");
        setTimeout(() => {
            btn.classList.remove("copy-failed");
        }, 2000);
    }
}

/**
 * Attaches a click handler to the first `.copy` button that copies the text content of `#url-result` to the clipboard and updates the button's UI to reflect success or failure.
 *
 * If either the `.copy` button or the `#url-result` element is not present, the function does nothing.
 */
function initCopyToClipboard(): void {
    const copyButton = document.querySelector<HTMLButtonElement>(".copy");
    if (!copyButton) return;

    copyButton.addEventListener("click", () => {
        const url = document.getElementById("url-result");
        if (!url) return;

        copyToClipboard(copyButton, url.innerText);
    });
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
    initCopyToClipboard();
});

export { initMultiselect };