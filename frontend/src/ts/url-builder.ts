import Choices from "choices.js";

function initUrlBuilder(): void {
    const form = document.getElementById("url-builder-form");

    if (!form) return;

    console.log("Initializing URL builder");

    // People
    const peopleMultiSelect = document.getElementById("url-builder-people");
    if (peopleMultiSelect) {
        const _peopleMultiSelectChoices = new Choices(peopleMultiSelect, {
            placeholderValue: "Select people",
            removeItemButton: true,
        });
        peopleMultiSelect.addEventListener(
            'change',
            () => {
                document.querySelector('body')?.dispatchEvent(new Event('multiselect-change'));
            },
            false,
        );
    }

    // Album
    const albumMultiSelect = document.getElementById("url-builder-albums");
    if (albumMultiSelect) {
        const _albumMultiSelectChoices = new Choices(albumMultiSelect, {
            placeholderValue: "Select albums",
            removeItemButton: true,
        });
        albumMultiSelect.addEventListener(
            'change',
            () => {
                document.querySelector('body')?.dispatchEvent(new Event('multiselect-change'));
            },
            false,
        );
    }
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
    initUrlBuilder();
});
