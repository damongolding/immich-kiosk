import Choices from "choices.js";

function initUrlBuilder(): void {
    const form = document.getElementById("url-builder-form");

    if (!form) return;

    console.log("Initializing URL builder");

    // People
    const peopleMultiSelect = document.getElementById("url-builder-people");
    if (peopleMultiSelect) {
        const _peopleMultiSelectChoices = new Choices(peopleMultiSelect, {
            removeItemButton: true,
        });
    }
}

// Initialize Kiosk when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", () => {
    initUrlBuilder();
});
