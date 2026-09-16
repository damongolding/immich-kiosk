// Immich Kiosk service worker.
//
// Its job is to keep the installed PWA usable when the Kiosk server is
// unreachable *or* reachable but not responding. iOS "Add to Home Screen" apps
// otherwise get permanently stuck on Safari's native error page ("Safari could
// not open the page because the server stopped responding") and the home-screen
// shortcut has to be deleted and recreated.
//
// This script must be registered with an explicit scope of "/" (the server
// sends a `Service-Worker-Allowed: /` header) so it can intercept top-level
// navigations and not just requests under /assets/js/.

var CACHE = "immich-kiosk-v3";
var OFFLINE_URL = "/assets/offline.html";

// How long to wait for a navigation response before giving up and showing the
// offline page. "server stopped responding" is a hang, not a refused
// connection, so fetch() would otherwise sit here until iOS' own (much longer)
// network timeout fires and paints its native error page.
var NAV_TIMEOUT_MS = 5000;

// Last-resort offline page, embedded so a navigation is *always* answered with a
// real document even when the Cache Storage entry is missing (first launch
// happened while the server was down, iOS evicted the cache, install failed,
// ...). Keep this in sync with frontend/public/assets/offline.html.
var FALLBACK_HTML =
    '<!doctype html><html lang="en"><head><meta charset="utf-8">' +
    '<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">' +
    '<meta name="apple-mobile-web-app-capable" content="yes">' +
    '<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">' +
    '<meta name="theme-color" content="#000000">' +
    "<title>Immich Kiosk</title>" +
    "<style>html,body{margin:0;height:100%;background:#000;color:#8a8a8a;" +
    'font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;' +
    "display:flex;align-items:center;justify-content:center;text-align:center;" +
    "-webkit-user-select:none;user-select:none}" +
    ".w{padding:2rem;max-width:22rem}h1{font-size:1rem;font-weight:600;color:#c9c9c9;margin:0 0 .5rem}" +
    "p{font-size:.8rem;line-height:1.5;margin:0}" +
    ".d{width:.6rem;height:.6rem;border-radius:50%;background:#c23b3b;display:inline-block;" +
    "margin-right:.5rem;vertical-align:middle;animation:p 1.6s ease-in-out infinite}" +
    "@keyframes p{0%,100%{opacity:.25}50%{opacity:1}}</style></head>" +
    '<body><div class="w"><h1><span class="d"></span>Kiosk unavailable</h1>' +
    "<p>Trying to reconnect automatically&hellip;</p></div>" +
    "<script>(function(){function c(){fetch('/health',{cache:'no-store'})" +
    ".then(function(r){if(r&&r.ok)location.reload()}).catch(function(){})}" +
    "setInterval(c,3000);addEventListener('online',c);" +
    "addEventListener('visibilitychange',function(){if(!document.hidden)c()});" +
    "addEventListener('pageshow',c);c()})();<\/script></body></html>";

// Resolve to the cached offline page, or the embedded fallback if it is not
// there. Never rejects.
function offlinePage() {
    return caches
        .match(OFFLINE_URL, { ignoreSearch: true })
        .then(function (cached) {
            return cached || fallbackResponse();
        })
        .catch(fallbackResponse);
}

function fallbackResponse() {
    return new Response(FALLBACK_HTML, {
        headers: { "Content-Type": "text/html; charset=utf-8" },
    });
}

self.addEventListener("install", function (event) {
    event.waitUntil(
        caches
            .open(CACHE)
            .then(function (cache) {
                return cache.add(OFFLINE_URL);
            })
            .catch(function () {
                // Server was down while the worker installed. The embedded
                // FALLBACK_HTML keeps navigations covered until the next
                // successful load refreshes the cache.
            }),
    );
    self.skipWaiting();
});

self.addEventListener("activate", function (event) {
    event.waitUntil(
        caches
            .keys()
            .then(function (keys) {
                return Promise.all(
                    keys
                        .filter(function (key) {
                            return key !== CACHE;
                        })
                        .map(function (key) {
                            return caches.delete(key);
                        }),
                );
            })
            .then(function () {
                return self.clients.claim();
            }),
    );
});

self.addEventListener("fetch", function (event) {
    var request = event.request;

    // Only step in for top-level navigations. Everything else (assets, XHR,
    // the offline page's own /health polling) goes straight to the network.
    if (request.mode !== "navigate") {
        return;
    }

    event.respondWith(
        new Promise(function (resolve) {
            var done = false;

            function settle(value) {
                if (done) {
                    return;
                }
                done = true;
                resolve(value);
            }

            var timer = setTimeout(function () {
                offlinePage().then(settle);
            }, NAV_TIMEOUT_MS);

            fetch(request).then(
                function (response) {
                    clearTimeout(timer);
                    // A crashing/restarting container behind a reverse proxy
                    // answers navigations with 502/503/504 (or 500 while the Go
                    // server is still booting) rather than failing outright.
                    // Treat any server-side error as "not ready yet" and show
                    // the self-retrying offline page. 4xx (auth prompts, bad
                    // paths) is passed through so the user still sees it.
                    if (response.status >= 500) {
                        offlinePage().then(settle);
                        return;
                    }
                    // Refresh the cached offline page opportunistically so a
                    // later cold launch has an up-to-date copy to fall back on.
                    // (offline.html itself, not this navigation response.)
                    caches.open(CACHE).then(function (cache) {
                        cache.add(OFFLINE_URL).catch(function () {});
                    });
                    settle(response);
                },
                function () {
                    clearTimeout(timer);
                    offlinePage().then(settle);
                },
            );
        }),
    );
});
