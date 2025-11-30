/**
 * @module menu-controls
 * Module for handling kiosk menu interactions and image navigation
 * @description Controls menu behavior and navigation between assets in a kiosk interface
 */

import htmx from "htmx.org";

let disableNavigation: boolean = false;

let nextAssetMenuButton: HTMLElement;
let prevAssetMenuButton: HTMLElement;

let assetOverlayVisible: boolean = false;
let linkOverlayVisible: boolean = false;

const redirectsContainer = document.getElementById(
    "redirects-container",
) as HTMLElement | null;
let redirects: NodeListOf<HTMLAnchorElement> | null;
let currentRedirectIndex = -1;

let allowMoreInfo: boolean;
let infoKeyPress: () => void;
let redirectsKeyPress: () => void;

// Store albums data for sorting
let allAlbums: Record<string, Array<{id: string; albumName: string; assetCount: number}>> = {};
let activeTab: string = "Main Library";
const selectedAlbumIds: Set<string> = new Set();
let currentSort: string = 'name';
let currentSearchQuery: string = '';
let albumsLoaded: boolean = false;
let apiAlbumsPassword: string | null = null;
let passwordTimer: number | null = null;
let countdownInterval: number | null = null;

/**
 * Clears the password timer
 */
function clearPasswordTimer(): void {
    if (passwordTimer !== null) {
        clearTimeout(passwordTimer);
        passwordTimer = null;
    }
    if (countdownInterval !== null) {
        clearInterval(countdownInterval);
        countdownInterval = null;
    }
    // Hide countdown display
    hidePasswordCountdown();
}

/**
 * Handles password expiration after timer runs out
 */
function handlePasswordExpiration(): void {
    // Clear the stored password
    apiAlbumsPassword = null;
    // Clear the timer and interval
    if (passwordTimer !== null) {
        clearTimeout(passwordTimer);
        passwordTimer = null;
    }
    if (countdownInterval !== null) {
        clearInterval(countdownInterval);
        countdownInterval = null;
    }
    // Reset albums loaded state so it will try to load again
    albumsLoaded = false;
    // Hide countdown
    hidePasswordCountdown();
    // Show password prompt again
    showPasswordPrompt();
}

/**
 * Updates the password countdown display
 */
function updatePasswordCountdown(remainingSeconds: number): void {
    const countdownElement = document.getElementById("password-countdown");
    if (countdownElement) {
        // Ensure remainingSeconds is not negative
        const safeSeconds = Math.max(0, remainingSeconds);
        const minutes = Math.floor(safeSeconds / 60);
        const seconds = safeSeconds % 60;
        const formattedSeconds = seconds < 10 ? `0${seconds}` : seconds.toString();
        countdownElement.textContent = `Password expires in ${minutes}:${formattedSeconds}`;
    }
}

/**
 * Shows the password countdown display
 */
function showPasswordCountdown(): void {
    const albumsList = document.getElementById("api-albums-list");
    if (!albumsList) return;
    
    // Check if countdown element already exists
    let countdownElement = document.getElementById("password-countdown");
    if (!countdownElement) {
        countdownElement = document.createElement("div");
        countdownElement.id = "password-countdown";
        countdownElement.className = "password-countdown";
        albumsList.insertBefore(countdownElement, albumsList.firstChild);
    }
    countdownElement.style.display = "block";
}

/**
 * Hides the password countdown display
 */
function hidePasswordCountdown(): void {
    const countdownElement = document.getElementById("password-countdown");
    if (countdownElement) {
        countdownElement.style.display = "none";
    }
}

/**
 * Starts the password expiration timer (1 minute)
 */
function startPasswordTimer(): void {
    // Clear any existing timer
    clearPasswordTimer();
    
    let remainingSeconds = 60;
    
    // Show countdown
    showPasswordCountdown();
    updatePasswordCountdown(remainingSeconds);
    
    // Start countdown interval (updates every second)
    countdownInterval = window.setInterval(() => {
        remainingSeconds--;
        
        // Stop the interval if we've reached 0 or negative
        if (remainingSeconds <= 0) {
            if (countdownInterval !== null) {
                clearInterval(countdownInterval);
                countdownInterval = null;
            }
            return;
        }
        
        updatePasswordCountdown(remainingSeconds);
    }, 1000);
    
    // Start expiration timer
    passwordTimer = window.setTimeout(handlePasswordExpiration, 60000);
}

/**
 * Loads and displays albums from the API
 */
async function loadApiAlbums(): Promise<void> {
    // Prevent loading albums multiple times
    if (albumsLoaded) {
        renderAlbumsList();
        return;
    }
    
    const albumsList = document.getElementById("api-albums-list") as HTMLElement;
    
    if (!albumsList) return;

    try {
        const url = "/api/albums";
        const headers: HeadersInit = {};
        
        if (apiAlbumsPassword) {
            headers["X-Kiosk-Password"] = apiAlbumsPassword;
        }
        
        const response = await fetch(url, {
            headers: headers
        });
        
        if (response.status === 401) {
            const errorData = await response.json().catch(() => ({}));
            if (errorData.code === "password_required") {
                // Password required
                showPasswordPrompt();
                return;
            } else if (errorData.code === "invalid_password") {
                // Wrong password provided
                showPasswordError();
                return;
            }
        }
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const albumsData = await response.json();
        
        // Handle both old (array) and new (map) response formats for backward compatibility
        if (Array.isArray(albumsData)) {
             // Remove duplicates based on album ID
            const uniqueAlbums = albumsData.filter((album: {id: string; albumName: string; assetCount: number}, index: number, self: Array<{id: string; albumName: string; assetCount: number}>) => 
                index === self.findIndex((a) => a.id === album.id)
            );
            allAlbums = { "Main Library": uniqueAlbums };
        } else {
            // Process each user's albums
            allAlbums = {};
            Object.keys(albumsData).forEach(user => {
                const userAlbums = albumsData[user];
                const uniqueAlbums = userAlbums.filter((album: {id: string; albumName: string; assetCount: number}, index: number, self: Array<{id: string; albumName: string; assetCount: number}>) => 
                    index === self.findIndex((a) => a.id === album.id)
                );
                allAlbums[user] = uniqueAlbums;
            });
        }
        
        // Initialize selected albums from URL if first load
        // We check if any albums are loaded across all users
        const totalAlbumsCount = Object.keys(allAlbums).reduce((acc, key) => acc + allAlbums[key].length, 0);
        
        if (totalAlbumsCount > 0 && selectedAlbumIds.size === 0) {
            const urlParams = new URLSearchParams(window.location.search);
            const albumsFromUrl = urlParams.getAll('album');
            albumsFromUrl.forEach(id => {
                selectedAlbumIds.add(id);
            });
        }

        albumsLoaded = true;
        
        renderTabs();
        renderAlbumsList();
        updateApplyButtonState();
        
        // Start the password expiration timer
        startPasswordTimer();
    } catch (error) {
        console.error("Failed to load albums:", error);
        albumsList.innerHTML = '<div class="loading-albums">Failed to load albums</div>';
    }
}

/**
 * Shows password prompt for API albums
 */
function showPasswordPrompt(): void {
    const albumsList = document.getElementById("api-albums-list") as HTMLElement;
    const sortControls = document.querySelector(".albums-sort-controls") as HTMLElement;
    const albumsActions = document.querySelector(".api-albums-actions") as HTMLElement;
    
    if (!albumsList) return;
    
    // Clear any existing timer since we're prompting for password again
    clearPasswordTimer();
    
    // Hide sort controls and actions when showing password prompt
    if (sortControls) sortControls.style.display = "none";
    const searchContainer = document.querySelector(".albums-search-container") as HTMLElement;
    if (searchContainer) searchContainer.style.display = "none";
    const tabsContainer = document.getElementById("albums-tabs");
    if (tabsContainer) tabsContainer.style.display = "none";
    if (albumsActions) albumsActions.style.display = "none";
    
    albumsList.innerHTML = `
        <div class="password-prompt">
            <div class="password-prompt-title">Enter Password</div>
            <input type="password" id="api-albums-password" placeholder="Password" />
            <button id="submit-password-btn">Submit</button>
            <button id="cancel-password-btn">Cancel</button>
        </div>
    `;
    
    const submitBtn = document.getElementById("submit-password-btn") as HTMLButtonElement;
    const cancelBtn = document.getElementById("cancel-password-btn") as HTMLButtonElement;
    const passwordInput = document.getElementById("api-albums-password") as HTMLInputElement;
    
    if (submitBtn && cancelBtn && passwordInput) {
        submitBtn.addEventListener("click", () => {
            const password = passwordInput.value.trim();
            if (password) {
                apiAlbumsPassword = password;
                loadApiAlbums();
            }
        });
        
        cancelBtn.addEventListener("click", () => {
            // Clear timer and reset state when canceling
            clearPasswordTimer();
            apiAlbumsPassword = null;
            albumsLoaded = false;
            albumsList.innerHTML = '<div class="loading-albums">Access cancelled</div>';
        });
        
        passwordInput.addEventListener("keypress", (e) => {
            if (e.key === "Enter") {
                submitBtn.click();
            }
        });
        
        // Focus the password input
        passwordInput.focus();
    }
}

/**
 * Shows error message for incorrect password
 */
function showPasswordError(): void {
    const albumsList = document.getElementById("api-albums-list") as HTMLElement;
    const sortControls = document.querySelector(".albums-sort-controls") as HTMLElement;
    const albumsActions = document.querySelector(".api-albums-actions") as HTMLElement;
    
    if (!albumsList) return;
    
    // Clear the stored password and timer so user can try again
    apiAlbumsPassword = null;
    clearPasswordTimer();
    
    // Hide sort controls and actions when showing password error
    if (sortControls) sortControls.style.display = "none";
    const searchContainer = document.querySelector(".albums-search-container") as HTMLElement;
    if (searchContainer) searchContainer.style.display = "none";
    const tabsContainer = document.getElementById("albums-tabs");
    if (tabsContainer) tabsContainer.style.display = "none";
    if (albumsActions) albumsActions.style.display = "none";
    
    albumsList.innerHTML = `
        <div class="password-error">
            <div class="password-error-title">Incorrect Password</div>
            <div class="password-error-message">The password you entered is incorrect. Please try again.</div>
            <button id="retry-password-btn">Try Again</button>
            <button id="cancel-password-btn">Cancel</button>
        </div>
    `;
    
    const retryBtn = document.getElementById("retry-password-btn") as HTMLButtonElement;
    const cancelBtn = document.getElementById("cancel-password-btn") as HTMLButtonElement;
    
    if (retryBtn && cancelBtn) {
        retryBtn.addEventListener("click", () => {
            showPasswordPrompt();
        });
        
        cancelBtn.addEventListener("click", () => {
            // Clear timer and reset state when canceling
            clearPasswordTimer();
            albumsLoaded = false;
            albumsList.innerHTML = '<div class="loading-albums">Access cancelled</div>';
        });
    }
}

/**
 * Renders the tabs for switching between users
 */
function renderTabs(): void {
    const tabsContainer = document.getElementById("albums-tabs");
    if (!tabsContainer) return;
    
    // Only show tabs if we have more than one user/library
    const users = Object.keys(allAlbums);
    if (users.length <= 1) {
        tabsContainer.style.display = "none";
        return;
    }
    
    tabsContainer.style.display = "flex";
    tabsContainer.innerHTML = "";
    
    // Sort users so "Main Library" is always first
    users.sort((a, b) => {
        if (a === "Main Library") return -1;
        if (b === "Main Library") return 1;
        return a.localeCompare(b);
    });
    
    // Ensure active tab is valid
    if (!allAlbums[activeTab] && users.length > 0) {
        activeTab = users[0];
    }
    
    users.forEach(user => {
        const tab = document.createElement("div");
        tab.className = `album-tab ${user === activeTab ? "active" : ""}`;
        tab.textContent = user;
        tab.addEventListener("click", () => {
            activeTab = user;
            renderTabs(); // Re-render to update active class
            renderAlbumsList();
        });
        tabsContainer.appendChild(tab);
    });
}

/**
 * Renders the albums list with current sorting
 */
function renderAlbumsList(): void {
    const albumsList = document.getElementById("api-albums-list") as HTMLElement;
    const sortControls = document.querySelector(".albums-sort-controls") as HTMLElement;
    const albumsActions = document.querySelector(".api-albums-actions") as HTMLElement;
    
    if (!albumsList) return;
    
    // Check if password has expired (no password but albums were previously loaded)
    const totalAlbumsCount = Object.keys(allAlbums).reduce((acc, key) => acc + allAlbums[key].length, 0);
    if (apiAlbumsPassword === null && albumsLoaded === false && totalAlbumsCount > 0) {
        showPasswordPrompt();
        return;
    }
    
    // Show sort controls and actions when rendering albums
    if (sortControls) sortControls.style.display = "flex";
    const searchContainer = document.querySelector(".albums-search-container") as HTMLElement;
    if (searchContainer) searchContainer.style.display = "flex";
    const tabsContainer = document.getElementById("albums-tabs");
    // Only show tabs container if we have multiple users (handled in renderTabs, but ensure visibility here if needed)
    if (tabsContainer && Object.keys(allAlbums).length > 1) tabsContainer.style.display = "flex";
    if (albumsActions) albumsActions.style.display = "flex";
    
    // Preserve countdown element if it exists
    const existingCountdown = document.getElementById("password-countdown");
    albumsList.innerHTML = "";
    
    // Re-add countdown if it existed
    if (existingCountdown && existingCountdown.style.display !== "none") {
        albumsList.appendChild(existingCountdown);
    }
    
    const currentAlbums = allAlbums[activeTab] || [];

    if (currentAlbums.length === 0) {
        albumsList.innerHTML = '<div class="loading-albums">No albums found</div>';
        return;
    }
    
    // Filter albums based on search query
    let filteredAlbums = currentAlbums;
    if (currentSearchQuery) {
        const query = currentSearchQuery.toLowerCase();
        filteredAlbums = currentAlbums.filter(album => 
            album.albumName.toLowerCase().includes(query)
        );
    }
    
    // Sort albums based on current sort setting
    const sortedAlbums = [...filteredAlbums].sort((a, b) => {
        switch (currentSort) {
            case 'count-asc':
                return a.assetCount - b.assetCount;
            case 'count-desc':
                return b.assetCount - a.assetCount;
            default:
                return a.albumName.localeCompare(b.albumName);
        }
    });
    
    sortedAlbums.forEach((album: {id: string; albumName: string; assetCount: number}) => {
        const item = document.createElement("div");
        item.className = "album-item";
        
        const checkbox = document.createElement("input");
        checkbox.type = "checkbox";
        checkbox.id = `album-${album.id}`;
        checkbox.value = album.id;
        checkbox.checked = selectedAlbumIds.has(album.id);
        checkbox.addEventListener("change", (e) => {
            const target = e.target as HTMLInputElement;
            if (target.checked) {
                selectedAlbumIds.add(album.id);
            } else {
                selectedAlbumIds.delete(album.id);
            }
            updateApplyButtonState();
            updateSelectAllBtn();
        });
        
        const label = document.createElement("label");
        label.htmlFor = `album-${album.id}`;
        
        const nameSpan = document.createElement("span");
        nameSpan.className = "album-name";
        nameSpan.textContent = album.albumName;
        
        const countSpan = document.createElement("span");
        countSpan.className = "album-count";
        countSpan.textContent = `${album.assetCount} assets`;
        
        label.appendChild(nameSpan);
        label.appendChild(countSpan);
        
        item.appendChild(checkbox);
        item.appendChild(label);
        albumsList.appendChild(item);
    });
}

/**
 * Refreshes the albums from the API
 */
function refreshAlbums(): void {
    // Reset albums loaded state to force reload
    albumsLoaded = false;
    // Clear any existing timer since we're refreshing
    clearPasswordTimer();
    // Load albums again
    loadApiAlbums();
}

/**
 * Handles sorting button clicks
 */
function handleSortChange(sortType: string): void {
    currentSort = sortType;
    
    // Update active button state
    const sortButtons = document.querySelectorAll('.sort-btn');
    sortButtons.forEach(btn => {
        btn.classList.remove('active');
        if ((btn as HTMLElement).dataset.sort === sortType) {
            btn.classList.add('active');
        }
    });
    
    // Re-render the list with new sorting
    renderAlbumsList();
}

/**
 * Updates the apply button state based on checkbox selections
 */
function updateApplyButtonState(): void {
    const applyBtn = document.getElementById("apply-albums-btn") as HTMLButtonElement;
    
    if (applyBtn) {
        applyBtn.disabled = selectedAlbumIds.size === 0;
    }
}

/**
 * Applies the selected albums and navigates
 */
function applyApiAlbumsSelection(): void {
    const selectedAlbums = Array.from(selectedAlbumIds);
    
    if (selectedAlbums.length === 0) {
        return;
    }
    
    const albumParams = selectedAlbums.map(id => `album=${encodeURIComponent(id)}`).join("&");
    const url = `/?${albumParams}`;
    
    window.location.href = url;
}

function updateClearSearchBtn() {
    const clearSearchBtn = document.getElementById("clear-search-btn") as HTMLButtonElement;
    if (clearSearchBtn) {
        clearSearchBtn.style.display = currentSearchQuery ? "block" : "none";
    }
}

function getVisibleAlbums() {
    const albums = allAlbums[activeTab] || [];
    if (!currentSearchQuery) return albums;
    
    const query = currentSearchQuery.toLowerCase();
    return albums.filter(album => 
        album.albumName.toLowerCase().includes(query)
    );
}

function updateSelectAllBtn() {
    const selectAllBtn = document.getElementById("select-all-albums-btn") as HTMLButtonElement;
    if (!selectAllBtn) return;
    
    const visibleAlbums = getVisibleAlbums();
    if (visibleAlbums.length === 0) {
        selectAllBtn.textContent = "Select All";
        selectAllBtn.classList.remove("active");
        return;
    }

    const allSelected = visibleAlbums.every(album => selectedAlbumIds.has(album.id));
    
    if (allSelected) {
        selectAllBtn.textContent = "Deselect All";
        selectAllBtn.classList.add("active");
    } else {
        selectAllBtn.textContent = "Select All";
        selectAllBtn.classList.remove("active");
    }
}

function toggleSelectAll() {
    const visibleAlbums = getVisibleAlbums();
    if (visibleAlbums.length === 0) return;

    const allSelected = visibleAlbums.every(album => selectedAlbumIds.has(album.id));

    if (allSelected) {
        // Deselect all visible
        visibleAlbums.forEach(album => {
            selectedAlbumIds.delete(album.id);
        });
    } else {
        // Select all visible
        visibleAlbums.forEach(album => {
            selectedAlbumIds.add(album.id);
        });
    }

    renderAlbumsList();
    updateApplyButtonState();
    updateSelectAllBtn();
}

/**
 * Disables both next and previous asset navigation buttons
 * @returns {void}
 */
function disableAssetNavigationButtons(): void {
    if (disableNavigation) return;
    if (!nextAssetMenuButton || !prevAssetMenuButton) {
        console.debug("Navigation buttons not initialized.");
        return;
    }
    htmx.addClass(nextAssetMenuButton, "disabled");
    htmx.addClass(prevAssetMenuButton, "disabled");
}

/**
 * Enables both next and previous asset navigation buttons
 * @returns {void}
 */
function enableAssetNavigationButtons(): void {
    if (disableNavigation) return;
    if (!nextAssetMenuButton || !prevAssetMenuButton) {
        console.error("Navigation buttons not initialized");
        return;
    }
    htmx.removeClass(nextAssetMenuButton, "disabled");
    htmx.removeClass(prevAssetMenuButton, "disabled");
}

/**
 * Shows the asset information overlay
 * Only works when polling is paused
 * @returns {void}
 */
function showAssetOverlay(): void {
    if (!document.body) return;
    if (!document.body.classList.contains("polling-paused")) return;
    hideRedirectsOverlay();
    document.body.classList.add("more-info");
    assetOverlayVisible = true;
}

/**
 * Hides the asset information overlay
 * @returns {void}
 */
function hideAssetOverlay(): void {
    if (!document.body) return;
    document.body.classList.remove("more-info");
    assetOverlayVisible = false;
}

/**
 * Toggles the asset information overlay visibility
 * @returns {void}
 */
function toggleAssetOverlay(): void {
    assetOverlayVisible ? hideAssetOverlay() : showAssetOverlay();
}

function redirectKeyHandler(e: KeyboardEvent) {
    if (!redirects) return;

    // Ignore key events if typing in an input field
    const target = e.target as HTMLElement;
    if (target.tagName === "INPUT" || target.tagName === "TEXTAREA") {
        return;
    }

    switch (e.code) {
        case "ArrowDown":
            e.preventDefault(); // Prevent page scrolling
            currentRedirectIndex =
                (currentRedirectIndex + 1) % redirects.length;
            redirects[currentRedirectIndex].focus();
            break;
        case "ArrowUp":
            e.preventDefault(); // Prevent page scrolling
            currentRedirectIndex =
                (currentRedirectIndex - 1 + redirects.length) %
                redirects.length;
            redirects[currentRedirectIndex].focus();
            break;
        case "KeyI":
            if (!allowMoreInfo) return;
            e.preventDefault();
            infoKeyPress();
            break;
        case "KeyR":
            if (e.ctrlKey || e.metaKey) return;
            e.preventDefault();
            redirectsKeyPress();
            break;
    }
}

/**
 * Shows the links overlay
 * Only works when polling is paused
 * Hides image overlay if visible
 */
function showRedirectsOverlay(): void {
    if (!document.body) return;
    if (!document.body.classList.contains("polling-paused")) return;

    document.addEventListener("keydown", redirectKeyHandler);

    hideAssetOverlay();
    document.body.classList.add("redirects-open");
    linkOverlayVisible = true;
}

/**
 * Hides the links overlay
 */
function hideRedirectsOverlay(): void {
    if (!document.body) return;
    document.body.classList.remove("redirects-open");

    document.removeEventListener("keydown", redirectKeyHandler);

    linkOverlayVisible = false;
    
    // Clear password timer when closing the overlay
    clearPasswordTimer();
}

/**
 * Toggles the links overlay visibility
 */
function toggleRedirectsOverlay(): void {
    linkOverlayVisible ? hideRedirectsOverlay() : showRedirectsOverlay();
}

/**
 * Initializes the menu controls and sets up event handlers
 * @param nextAssetButton - The next image navigation button element
 * @param prevAssetButton - The previous image navigation button element
 * @throws {Error} If either navigation button is not provided
 * @returns {void}
 */
function initMenu(
    disableNav: boolean,
    nextAssetButton: HTMLElement | null,
    prevAssetButton: HTMLElement | null,
    showMoreInfo: boolean,
    handleInfoKeyPress: () => void,
    handleRedirectsKeyPress: () => void,
): void {
    if (disableNav) {
        disableNavigation = disableNav;
        return;
    }

    if (!nextAssetButton || !prevAssetButton) {
        throw new Error("Both navigation buttons must be provided");
    }

    nextAssetMenuButton = nextAssetButton;
    prevAssetMenuButton = prevAssetButton;

    if (redirectsContainer) {
        redirects = redirectsContainer.querySelectorAll("a");
        
        // Load API albums when redirects menu is initialized
        loadApiAlbums();
        
        // Add event listener for apply albums button
        const applyBtn = document.getElementById("apply-albums-btn") as HTMLButtonElement;
        if (applyBtn) {
            applyBtn.addEventListener("click", applyApiAlbumsSelection);
        }
        
        // Add event listeners for sort and refresh buttons
        const sortButtons = document.querySelectorAll('.sort-btn');
        sortButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const button = e.target as HTMLElement;
                const sortType = button.dataset.sort;
                const isRefresh = button.id === 'refresh-albums-btn';
                
                if (isRefresh) {
                    refreshAlbums();
                } else if (sortType) {
                    handleSortChange(sortType);
                }
            });
        });

        // Add event listeners for search
        const searchInput = document.getElementById('album-search-input') as HTMLInputElement;
        const clearSearchBtn = document.getElementById('clear-search-btn') as HTMLButtonElement;
        const selectAllBtn = document.getElementById("select-all-albums-btn") as HTMLButtonElement;

        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                const target = e.target as HTMLInputElement;
                currentSearchQuery = target.value.trim();
                renderAlbumsList();
                updateClearSearchBtn();
                updateSelectAllBtn();
            });
        }

        if (clearSearchBtn) {
            clearSearchBtn.addEventListener('click', () => {
                currentSearchQuery = '';
                if (searchInput) searchInput.value = '';
                renderAlbumsList();
                updateClearSearchBtn();
                updateSelectAllBtn();
            });
        }
        
        if (selectAllBtn) {
            selectAllBtn.addEventListener("click", toggleSelectAll);
        }
    }

    allowMoreInfo = showMoreInfo;
    infoKeyPress = handleInfoKeyPress;
    redirectsKeyPress = handleRedirectsKeyPress;

    if (nextAssetMenuButton) {
        nextAssetMenuButton.addEventListener("click", () => {
            hideRedirectsOverlay();
        });
    }

    if (prevAssetMenuButton) {
        prevAssetMenuButton.addEventListener("click", () => {
            hideRedirectsOverlay();
        });
    }
}

export {
    initMenu,
    disableAssetNavigationButtons,
    enableAssetNavigationButtons,
    showAssetOverlay,
    hideAssetOverlay,
    toggleAssetOverlay,
    toggleRedirectsOverlay,
};
