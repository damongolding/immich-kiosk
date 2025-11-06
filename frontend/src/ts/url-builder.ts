import Choices from "choices.js";
import "htmx.org";

function initUrlBuilder(): void {
    const form = document.getElementById("url-builder-form");

    if (!form) return;

    // People
    const peopleMultiSelect = document.getElementById("url-builder-people");
    if (peopleMultiSelect) {
        const _peopleMultiSelectChoices = new Choices(peopleMultiSelect, {
            placeholderValue: "Select people",
            removeItemButton: true,
        });
        peopleMultiSelect.addEventListener(
            "change",
            () => {
                document
                    .querySelector("body")
                    ?.dispatchEvent(new Event("multiselect-change"));
            },
            false,
        );
    }

    // Album
    const albumMultiSelect = document.getElementById("url-builder-album");
    if (albumMultiSelect) {
        const _albumMultiSelectChoices = new Choices(albumMultiSelect, {
            placeholderValue: "Select albums",
            removeItemButton: true,
        });
        albumMultiSelect.addEventListener(
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

async function copyToClipboard(text: string): Promise<void> {
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
    } catch (error) {
        console.error("Failed to copy text:", error);
    }
}

function initCopyToClipboard(): void {
    const copyButton = document.querySelector<HTMLButtonElement>(".copy");
    if (!copyButton) return;

    copyButton.addEventListener("click", () => {
        const url = document.getElementById("url-result");
        if (!url) return;

        copyToClipboard(url.innerText);
    });
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
    initUrlBuilder();
    initCopyToClipboard();
});
