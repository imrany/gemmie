import type { UserDetails } from "@/types";
import { API_BASE_URL } from "@/utils/globals";
import { ref, type Ref } from "vue";

export function useApiCall(globalStateRefs?: {
  isUserOnline?: Ref<boolean>;
  connectionStatus?: Ref<string>;
  parsedUserDetails?: Ref<UserDetails>;
  syncStatus?: Ref<any>;
}) {
  // Use provided refs or create local ones as fallback
  const isUserOnline = globalStateRefs?.isUserOnline || ref(navigator.onLine);
  const connectionStatus =
    globalStateRefs?.connectionStatus || ref<string>("online");
  const parsedUserDetails =
    globalStateRefs?.parsedUserDetails ||
    ref<UserDetails>(
      (() => {
        try {
          const userDetails: any = localStorage.getItem("userdetails");
          return userDetails ? JSON.parse(userDetails) : null;
        } catch (error) {
          console.error("Invalid user details in localStorage:", error);
          localStorage.removeItem("userdetails");
          return null;
        }
      })(),
    );
  const syncStatus =
    globalStateRefs?.syncStatus ||
    ref<any>({
      lastSync: null,
      syncing: false,
      hasUnsyncedChanges: false,
      lastError: null,
      retryCount: 0,
      maxRetries: 3,
    });

  async function apiCall(
    endpoint: string,
    options: RequestInit = {},
    retryCount = 0,
  ): Promise<any> {
    if (!isUserOnline.value) {
      const isActuallyOnline = await checkInternetConnection();
      if (!isActuallyOnline) {
        throw new Error(
          "No internet connection. Please check your network and try again.",
        );
      }
    }

    const maxRetries = 3;
    const retryDelay = Math.pow(2, retryCount) * 1000;

    try {
      if (
        !parsedUserDetails.value?.userId &&
        !endpoint.includes("/login") &&
        !endpoint.includes("/register")
      ) {
        throw new Error("User not authenticated");
      }

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 30000);

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...(parsedUserDetails.value?.userId
            ? { "X-User-ID": parsedUserDetails.value.userId }
            : {}),
          ...options.headers,
        },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      if (!data.success) {
        throw new Error(data.message || "API request failed");
      }

      syncStatus.value.retryCount = 0;
      syncStatus.value.lastError = null;

      return data;
    } catch (error: any) {
      console.error(`API Error on ${endpoint}:`, error);

      if (error.name === "AbortError") {
        throw new Error("Request timeout - please try again");
      }

      if (
        (error.name === "NetworkError" ||
          error.name === "TypeError" ||
          error.name === "TimeoutError") &&
        retryCount < maxRetries
      ) {
        console.log(
          `Retrying ${endpoint} in ${retryDelay}ms (attempt ${retryCount + 1}/${maxRetries})`,
        );

        await new Promise((resolve) => setTimeout(resolve, retryDelay));
        return apiCall(endpoint, options, retryCount + 1);
      }

      if (endpoint.includes("/sync")) {
        syncStatus.value.lastError = error.message;
        syncStatus.value.retryCount = retryCount;
      }

      throw error;
    }
  }

  async function unsecureApiCall(
    endpoint: string,
    options: RequestInit = {},
    retryCount = 0,
  ): Promise<any> {
    const maxRetries = 2;
    const retryDelay = Math.pow(2, retryCount) * 1000;

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000);

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const errorText = await response.text().catch(() => "Unknown error");
        throw new Error(
          `HTTP ${response.status}: ${response.statusText} - ${errorText}`,
        );
      }

      const data = await response.json();

      if (!data.success) {
        throw new Error(data.message || "API request failed");
      }

      return data;
    } catch (error: any) {
      console.error(
        `Unsecure API Error on ${endpoint} (attempt ${retryCount + 1}):`,
        error,
      );

      if (error.name === "AbortError") {
        throw new Error("Request timeout - please try again");
      }

      if (
        (error.name === "NetworkError" ||
          error.name === "TypeError" ||
          error.message?.includes("Failed to fetch")) &&
        retryCount < maxRetries
      ) {
        console.log(
          `Retrying ${endpoint} in ${retryDelay}ms (attempt ${retryCount + 1}/${maxRetries})`,
        );

        await new Promise((resolve) => setTimeout(resolve, retryDelay));
        return unsecureApiCall(endpoint, options, retryCount + 1);
      }

      throw error;
    }
  }

  async function checkInternetConnection(): Promise<boolean> {
    try {
      connectionStatus.value = "checking";

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(`${API_BASE_URL}/health`, {
        method: "GET",
        signal: controller.signal,
        cache: "no-cache",
      });

      clearTimeout(timeoutId);

      const isConnected = response.status < 400;
      isUserOnline.value = isConnected;
      connectionStatus.value = isConnected ? "online" : "offline";

      return isConnected;
    } catch (error) {
      console.warn("Internet connection check failed:", error);
      isUserOnline.value = false;
      connectionStatus.value = "offline";
      return false;
    }
  }

  return { apiCall, unsecureApiCall, checkInternetConnection };
}
