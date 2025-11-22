import type { ApiResponse, UserDetails } from "@/types";
import { API_BASE_URL } from "@/lib/globals";
import { ref, type Ref } from "vue";

export function useApiCall(globalStateRefs?: {
  parsedUserDetails?: Ref<UserDetails>;
  syncStatus?: Ref<any>;
}) {
  // Use provided refs or create local ones as fallback
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

  async function apiCall<T>(
    endpoint: string,
    options: RequestInit = {},
    retryCount = 0,
  ): Promise<ApiResponse<T>> {
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

      const data: ApiResponse<T> = await response.json();

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

  async function unsecureApiCall<T>(
    endpoint: string,
    options: RequestInit = {},
    retryCount = 0,
  ): Promise<ApiResponse<T>> {
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

      const data: ApiResponse<T> = await response.json();
      if (data && !data.success) {
        throw new Error(data.message || "Request failed");
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

  async function checkInternetConnection(): Promise<
    [boolean, "online" | "offline" | "checking"]
  > {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(`${API_BASE_URL}/health`, {
        method: "GET",
        signal: controller.signal,
        cache: "no-cache",
      });

      clearTimeout(timeoutId);

      const isConnected = response.status < 400;
      const status = isConnected ? "online" : "offline";
      return [isConnected, status];
    } catch (error: any) {
      console.warn("Internet connection check failed:", error);
      return [false, "offline"];
    }
  }

  return { apiCall, unsecureApiCall, checkInternetConnection };
}
