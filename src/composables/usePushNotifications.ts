import { API_BASE_URL } from "@/lib/globals";
import { ref, onMounted } from "vue";
import type { UserDetails, ApiResponse } from "@/types";
import { toast } from "vue-sonner";

interface BackendSubscriptionData {
  auth_key: string;
  created_at: string;
  endpoint: string;
  p256dh_key: string;
  updated_at: string;
  user_agent: string;
  user_id: string;
}

export function usePushNotifications() {
  const parsedUserDetails = ref<UserDetails>(
    JSON.parse(localStorage.getItem("userdetails") || "{}"),
  );
  const isSupported = ref(false);
  const isSubscribed = ref(false);
  const subscription = ref<PushSubscription | null>(null);
  const error = ref<string | null>(null);

  // Your VAPID public key (generate using the Go code)
  const VAPID_PUBLIC_KEY =
    "BCF32vavNQBrxlhVAhTf8CYQVmGOQlg93scirZrtGMYAH0_uS9K0EGAbcolQinWV9Hmw9riZsERvye08CwsRpPg";

  const urlBase64ToUint8Array = (string: string): Uint8Array => {
    try {
      const base64String = string.replace(/-/g, "+").replace(/_/g, "/");
      // Decode the Base64 string into a binary string
      const binaryString = atob(base64String);

      // Create a Uint8Array with the same length as the binary string
      const uint8Array = new Uint8Array(binaryString.length);

      // Populate the Uint8Array with the character codes from the binary string
      for (let i = 0; i < binaryString.length; i++) {
        uint8Array[i] = binaryString.charCodeAt(i);
      }

      return uint8Array;
    } catch (error) {
      console.error("Error decoding Base64 string:", error);
      throw new Error("Failed to decode base64 string");
    }
  };

  const requestPermission = async () => {
    const permission = await Notification.requestPermission();
    if (permission !== "granted") {
      error.value = "Notification permission denied";
      throw new Error("Permission not granted");
    }
    return permission;
  };

  const subscribe = async () => {
    try {
      error.value = null;

      // Request permission
      await requestPermission();

      // Register service worker
      const registration = await navigator.serviceWorker.getRegistration();

      // Wait for service worker to be ready
      await navigator.serviceWorker.ready;

      let sub = await registration?.pushManager.getSubscription();

      if (!sub) {
        // Subscribe to push notifications
        sub = await registration?.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
        });
      }

      subscription.value = sub ?? null;
      isSubscribed.value = true;

      // Send subscription to backend
      const response = await fetch(`${API_BASE_URL}/push/subscribe`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(parsedUserDetails.value?.userId
            ? { "X-User-ID": parsedUserDetails.value.userId }
            : {}),
          "User-Agent": navigator.userAgent,
        },
        body: JSON.stringify(sub?.toJSON()),
      });

      const data: ApiResponse<any> = await response.json();
      if (!response.ok) {
        throw new Error("Failed to subscribe to push notifications");
      }

      if (!data.success) {
        throw new Error(
          data.message || "Failed to subscribe to push notifications",
        );
      }

      console.log("Push notification subscription successful", data.data);
      toast.success(data.message);
      return sub;
    } catch (err: any) {
      console.error("Subscription failed:", err);
      error.value = err.message;
      throw err;
    }
  };

  const unsubscribe = async () => {
    try {
      // Ensure subscription.value is a native PushSubscription type if you intend to call .unsubscribe() on it.
      if (
        !subscription.value ||
        typeof subscription.value.unsubscribe !== "function"
      ) {
        throw new Error(
          "No valid native PushSubscription found to unsubscribe from.",
        );
      }

      //Notify backend using the endpoint associated with the now-inactive subscription
      const subDetails = subscription.value.toJSON(); // Get the standard endpoint, auth, p256dh keys

      // Notify backend
      const response = await fetch(`${API_BASE_URL}/push/unsubscribe`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(parsedUserDetails.value?.userId
            ? { "X-User-ID": parsedUserDetails.value.userId }
            : {}),
        },
        body: JSON.stringify(subDetails),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.message || `HTTP error! status: ${response.status}`,
        );
      }

      const data: ApiResponse<any> = await response.json();

      if (!data.success) {
        throw new Error(
          data.message || "Failed to unsubscribe to push notifications",
        );
      }

      subscription.value = null;
      isSubscribed.value = false;
      toast.success("Unsubscribed from push notifications");
      console.log("Unsubscribed from push notifications", data);
    } catch (err: any) {
      console.error("Unsubscribe failed:", err);
      error.value = err.message;
      throw err;
    }
  };

  const checkSubscription = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/push/subscriptions`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          ...(parsedUserDetails.value?.userId
            ? { "X-User-ID": parsedUserDetails.value.userId }
            : {}),
        },
      });

      const data: ApiResponse<BackendSubscriptionData[]> =
        await response.json();

      if (!response.ok) {
        throw new Error("Failed to check subscription");
      }

      if (!data.success) {
        throw new Error(data.message || "Failed to check subscription");
      }

      if (!data.data) {
        throw new Error(data.message || "No subscription found");
      }

      const subscriptionData = data.data[0];
      const mappedSub: PushSubscription = {
        endpoint: subscriptionData.endpoint,
        expirationTime: null,
        getKey: (key: PushEncryptionKeyName) => {
          if (key === "auth") {
            return urlBase64ToUint8Array(JSON.parse(subscriptionData.auth_key));
          } else if (key === "p256dh") {
            return urlBase64ToUint8Array(
              JSON.parse(subscriptionData.p256dh_key),
            );
          }
          return null;
        },
        options: {} as PushSubscriptionOptions,
        toJSON: () => mappedSub,
        unsubscribe: () => {
          return Promise.reject(
            "Unsubscribe not implemented client side, use the unsubscribe function",
          );
        },
      };
      subscription.value = mappedSub;
      isSubscribed.value = true;
      console.log("Found existing subscription");
    } catch (err) {
      console.error("Failed to check subscription:", err);
      toast.error("Failed to check subscription");
    }
  };

  onMounted(() => {
    isSupported.value = "serviceWorker" in navigator && "PushManager" in window;

    if (isSupported.value) {
      checkSubscription();
    }
  });

  return {
    isSupported,
    isSubscribed,
    subscription,
    error,
    subscribe,
    unsubscribe,
  };
}
