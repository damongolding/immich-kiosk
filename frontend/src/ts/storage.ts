export const storageUtils = {
    /**
     * Get item from localStorage with type safety
     * @param key Storage key
     * @returns Parsed value or null if not found
     */
    get<T>(key: string): T | null {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : null;
        } catch (error) {
            console.error(`Error reading from localStorage [${key}]:`, error);
            return null;
        }
    },

    /**
     * Set item in localStorage with type safety
     * @param key Storage key
     * @param value Value to store
     * @returns boolean indicating success
     */
    set<T>(key: string, value: T): boolean {
        try {
            localStorage.setItem(key, JSON.stringify(value));
            return true;
        } catch (error) {
            console.error(`Error writing to localStorage [${key}]:`, error);
            return false;
        }
    },

    /**
     * Remove item from localStorage
     * @param key Storage key
     */
    remove(key: string): void {
        try {
            localStorage.removeItem(key);
        } catch (error) {
            console.error(`Error removing from localStorage [${key}]:`, error);
        }
    },

    /**
     * Clear all items from localStorage
     */
    clear(): void {
        try {
            localStorage.clear();
        } catch (error) {
            console.error("Error clearing localStorage:", error);
        }
    },
};
