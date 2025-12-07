const CACHE_VERSION = "v0.30.4";
const staticCacheName = `site-static-${CACHE_VERSION}`;
const dynamicCache = `site-dynamic-${CACHE_VERSION}`;

const assets = [
  "/index.html",
  "/manifest.json",
  "/logo.svg",
  "/sounds/bell-notification.wav",
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
        return self.clients.claim();
      }),
  );
});

// cache limit function
const limitCacheSize = async (name, size) => {
  try {
    const cache = await caches.open(name);
    const keys = await cache.keys();

    if (keys.length > size) {
      await cache.delete(keys[0]);
      await limitCacheSize(name, size);
    }
  } catch (error) {
    console.error("Error limiting cache size:", error);
  }
};

// Helper function to check if request should be cached
const shouldCacheRequest = (url) => {
  if (url.origin !== location.origin) return false;
  if (url.pathname.startsWith("/api")) return false;
  if (url.pathname.startsWith("/share")) return false;
  if (url.protocol === "ws:" || url.protocol === "wss:") return false;
  return true;
};

// fetch event
self.addEventListener("fetch", (evt) => {
  if (
    !evt.request.url.startsWith("http") ||
    evt.request.url.includes("extension://")
  ) {
    return;
  }

  const url = new URL(evt.request.url);
  const shouldCache = shouldCacheRequest(url);

  if (!shouldCache) {
    evt.respondWith(
      fetch(evt.request).catch(
        () => new Response("Network error", { status: 503 }),
      ),
    );
    return;
  }

  evt.respondWith(
    caches.match(evt.request).then((cacheRes) => {
      if (cacheRes) {
        return cacheRes;
      }

      return fetch(evt.request)
        .then((fetchRes) => {
          if (fetchRes && fetchRes.status === 200) {
            const responseClone = fetchRes.clone();
            caches.open(dynamicCache).then((cache) => {
              cache.put(evt.request, responseClone);
              limitCacheSize(dynamicCache, 15);
            });
          }
          return fetchRes;
        })
        .catch((error) => {
          console.error("Fetch failed:", error);

          if (evt.request.destination === "document") {
            return caches.match("/index.html").then(
              (res) =>
                res ||
                new Response("<h1>Offline</h1>", {
                  headers: { "Content-Type": "text/html" },
                }),
            );
          }

          return new Response("Network error occurred", {
            status: 503,
            statusText: "Service Unavailable",
          });
        });
    }),
  );
});

// Push notification handler
self.addEventListener("push", (event) => {
  console.log("Push event received:", event);

  // Check if event has data
  if (!event.data) {
    console.error("Push event has no data");
    return;
  }

  let data;
  try {
    data = event.data.json();
    console.log("Push data parsed:", data);
  } catch (error) {
    console.error("Failed to parse push data:", error);
    // Try as text
    data = {
      title: "Notification",
      body: event.data.text() || "You have a new notification",
    };
  }

  const notificationOptions = {
    body: data.body || "You have a new notification",
    icon: data.icon || "/favicon.svg",
    badge: data.badge || "/logo.svg",
    image: data.image,
    actions: data.actions || [],
    vibrate: [200, 100, 200],
    silent: false,
    tag: data.tag || "default-tag",
    requireInteraction: data.requireInteraction || false,
    data: data.data || {},
  };

  console.log("Showing notification:", data.title, notificationOptions);

  event.waitUntil(
    self.registration.showNotification(
      data.title || "Notification",
      notificationOptions,
    ),
  );
});

self.addEventListener("notificationclick", (event) => {
  console.log("Notification clicked:", event.notification);
  event.notification.close();

  const link = event.notification.data.url || "/";
  const action = event.action;

  if (action) {
    console.log("Action clicked:", action);
  }

  event.waitUntil(
    // eslint-disable-next-line no-undef
    clients
      .matchAll({ type: "window", includeUncontrolled: true })
      .then((windowClients) => {
        const matchingClient = windowClients.find(
          (client) => client.url === link,
        );
        if (matchingClient) {
          return matchingClient.focus();
        } else {
          // eslint-disable-next-line no-undef
          return clients.openWindow(link);
        }
      }),
  );
});

// Handle notification close events
self.addEventListener("notificationclose", (event) => {
  console.log("Notification closed:", event);
});

// Background sync (if needed)
self.addEventListener("sync", (event) => {
  if (event.tag === "background-sync") {
    event.waitUntil(
      Promise.resolve().then(() => {
        console.log("Background sync triggered");
      }),
    );
  }
});

// Handle service worker updates
// This is called when user clicks "Update Now"
self.addEventListener("message", (event) => {
  if (event.data && event.data.type === "SKIP_WAITING") {
    console.log("Received SKIP_WAITING message - activating now");
    self.skipWaiting();
  }
});
