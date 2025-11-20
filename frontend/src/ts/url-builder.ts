import Choices from "choices.js";
import "htmx.org";

type choicesOptions = Record<string, unknown>;

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

function initCopyToClipboard(): void {
    const copyButton = document.querySelector<HTMLButtonElement>(".copy");
    if (!copyButton) return;

    copyButton.addEventListener("click", () => {
        const url = document.getElementById("url-result--url");
        if (!url) return;

        copyToClipboard(copyButton, url.innerText);
    });
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
    initCopyToClipboard();
});

export { initMultiselect };
