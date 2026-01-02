// Get the base path for API calls
// This allows the app to work in both subdomain and subfolder configurations
export function getBasePath() {
    if (typeof window !== 'undefined' && window.BASE_PATH !== undefined) {
        return window.BASE_PATH;
    }
    return '';
}

// Create an API URL with the correct base path
export function apiUrl(path) {
    const base = getBasePath();
    // Ensure path starts with /
    const normalizedPath = path.startsWith('/') ? path : `/${path}`;
    return `${base}${normalizedPath}`;
}

// Fetch wrapper that automatically handles base path
export async function apiFetch(path, options = {}) {
    return fetch(apiUrl(path), options);
}
