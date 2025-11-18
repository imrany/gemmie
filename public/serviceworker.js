const CACHE_VERSION = "v0.14.0";
const staticCacheName = `site-static-${CACHE_VERSION}`;
const dynamicCache = `site-dynamic-${CACHE_VERSION}`;

const assets = [
  "/index.html",
  "/manifest.json",
  "/logo.svg",
  "/sounds/bell-notification.wav",
  // Add other critical assets here
];

// Installing service worker
self.addEventListener("install", (evt) => {
  console.log("Service Worker installing... Version:", CACHE_VERSION);

  evt.waitUntil(
    caches
      .open(staticCacheName)
      .then((cache) => {
        console.log("Caching assets...");
        return cache.addAll(assets);
      })
      .then(() => {
        console.log("Assets cached successfully");
        // DON'T call self.skipWaiting() here
        // This keeps the new service worker in "waiting" state
        // until the user approves the update
      })
      .catch((error) => {
        console.error("Failed to cache assets:", error);
      }),
  );
});

// Activating service worker
self.addEventListener("activate", (evt) => {
  console.log("Service Worker activating... Version:", CACHE_VERSION);
  evt.waitUntil(
    caches
      .keys()
      .then((keys) => {
        // Only delete caches that don't match current version
        const deletePromises = keys
          .filter((key) => !key.includes(CACHE_VERSION))
          .map((key) => {
            console.log("Deleting old cache:", key);
            return caches.delete(key);
          });

        return Promise.all(deletePromises);
      })
      .then(() => {
        console.log("Service Worker activated");
        return self.clients.claim(); // Take control of all clients
      }),
  );
});

// cache limit function with async/await
const limitCacheSize = async (name, size) => {
  try {
    const cache = await caches.open(name);
    const keys = await cache.keys();

    if (keys.length > size) {
      await cache.delete(keys[0]);
      // Recursively limit cache size
      await limitCacheSize(name, size);
    }
  } catch (error) {
    console.error("Error limiting cache size:", error);
  }
};

// Helper function to check if request should be cached
const shouldCacheRequest = (url) => {
  // Skip external requests
  if (url.origin !== location.origin) {
    return false;
  }

  // Skip API endpoints
  if (url.pathname.startsWith("/api")) {
    return false;
  }

  // Skip share endpoints (these are SEO routes, don't cache them)
  if (url.pathname.startsWith("/share")) {
    return false;
  }

  // Skip WebSocket connections
  if (url.protocol === "ws:" || url.protocol === "wss:") {
    return false;
  }

  return true;
};

// fetch event with better caching strategies
self.addEventListener("fetch", (evt) => {
  // Skip non-HTTP requests and extension requests
  if (
    !evt.request.url.startsWith("http") ||
    evt.request.url.includes("extension://")
  ) {
    return;
  }

  const url = new URL(evt.request.url);
  const shouldCache = shouldCacheRequest(url);

  // For requests that shouldn't be cached, just pass through
  if (!shouldCache) {
    evt.respondWith(fetch(evt.request));
    return;
  }

  evt.respondWith(
    caches
      .match(evt.request)
      .then((cacheRes) => {
        // If found in cache, return cached version
        if (cacheRes) {
          return cacheRes;
        }

        // Fetch from network
        return fetch(evt.request).then((fetchRes) => {
          // Only cache successful responses
          if (fetchRes.status === 200) {
            const responseClone = fetchRes.clone();

            caches.open(dynamicCache).then(async (cache) => {
              await cache.put(evt.request, responseClone);
              await limitCacheSize(dynamicCache, 15);
            });
          }

          return fetchRes;
        });
      })
      .catch((error) => {
        console.error("Fetch failed:", error);

        // Fallback strategies
        if (evt.request.destination === "document") {
          // For HTML pages, serve the main page
          return caches.match("/index.html");
        }

        // For images, return a placeholder
        if (evt.request.destination === "image") {
          return caches.match("/logo.svg");
        }

        // For other resources, you could return a default image, etc.
        return new Response("Network error occurred", {
          status: 503,
          statusText: "Service Unavailable",
        });
      }),
  );
});

// push notification handler
self.addEventListener("push", (event) => {
  console.log("Push event received:", event);
  const data = event.data.json();

  const notificationOptions = {
    body: data.body,
    icon: data.icon || "/favicon.svg",
    badge: data.badge || "/logo.svg",
    vibrate: [200, 100, 200],
    silent: false,
    data: { link: data.url },
  };

  // Only add sound if browser supports it
  if ("sound" in Notification.prototype) {
    notificationOptions.sound = "/sounds/bell-notification.wav";
  }

  self.registration.showNotification(data.title, notificationOptions);
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const link = event.notification.data.link;
  if (link) {
    event.waitUntil(
      clients
        .matchAll({ type: "window", includeUncontrolled: true })
        .then((windowClients) => {
          const matchingClient = windowClients.find(
            (client) => client.url === link,
          );
          if (matchingClient) {
            return matchingClient.focus();
          } else {
            return clients.openWindow(link);
          }
        }),
    );
  }
});

// Handle notification close events
self.addEventListener("notificationclose", (event) => {
  console.log("Notification closed:", event);
  // You can track analytics or perform cleanup here
});

// Background sync (if needed)
self.addEventListener("sync", (event) => {
  if (event.tag === "background-sync") {
    event.waitUntil(
      Promise.resolve().then(() => {
        console.log("Background sync triggered");
        // Actual sync operations here
      }),
    );
  }
});

// Handle service worker updates
// This is called when user clicks "Update Now"
self.addEventListener("message", (event) => {
  if (event.data && event.data.type === "SKIP_WAITING") {
    console.log(
      "Received SKIP_WAITING message - user approved update, activating now",
    );
    // This makes the waiting service worker become the active one
    self.skipWaiting();
  }
});
