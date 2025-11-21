import { ref, type Ref } from "vue";
import { toast } from "vue-sonner";
import { useRouter } from "vue-router";
import { API_BASE_URL, validateCredentials } from "@/lib/globals";
import type { ApiResponse, UserDetails } from "@/types";

export interface AuthData {
  username: string;
  email: string;
  password: string;
  agreeToTerms: boolean;
}

export interface AuthConfig {
  loadingDelay?: number;
  redirectDelay?: number;
  minPasswordLength?: number;
  maxPasswordLength?: number;
}

export function useAuth(
  config: AuthConfig = {},
  parsedUserDetails: Ref<UserDetails>,
) {
  const router = useRouter();

  const {
    loadingDelay = 0,
    redirectDelay = 0,
    minPasswordLength = 8,
    maxPasswordLength = 128,
  } = config;

  const authStep = ref(1);
  const authData: Ref<AuthData> = ref({
    username: "",
    email: "",
    password: "",
    agreeToTerms: false,
  });
  const isLoading = ref(false);

  // Validation patterns
  const USERNAME_PATTERN = /^[a-zA-Z0-9_]+$/;
  const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  // const PASSWORD_PATTERN = /^(?=.*[A-Za-z])(?=.*\d).+$/; // At least one letter and one number

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

      if (!response.ok) {
        const errorText = await response.text().catch(() => "Unknown error");
        throw new Error(
          `HTTP ${response.status}: ${response.statusText} - ${errorText}`,
        );
      }

      const data: ApiResponse<T> = await response.json();

      if (!response.ok || !data.success) {
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

  function nextAuthStep() {
    if (authStep.value < 4) {
      authStep.value++;
    }
  }

  function prevAuthStep() {
    if (authStep.value > 1) {
      authStep.value--;
    }
  }

  function validateCurrentStep(): boolean {
    try {
      switch (authStep.value) {
        case 1: {
          const username = authData.value.username?.trim();
          return !!(
            username &&
            username.length >= 2 &&
            username.length <= 50 &&
            USERNAME_PATTERN.test(username)
          );
        }
        case 2: {
          const email = authData.value.email?.trim();
          return !!(
            email &&
            email.length > 0 &&
            email.length <= 100 &&
            EMAIL_PATTERN.test(email)
          );
        }
        case 3: {
          const password = authData.value.password;
          return !!(
            (
              password &&
              password.length >= minPasswordLength &&
              password.length <= maxPasswordLength
            )
            // && PASSWORD_PATTERN.test(password)
          );
        }
        case 4: {
          return authData.value.agreeToTerms;
        }
        default:
          return false;
      }
    } catch (error) {
      console.error("Error validating current step:", error);
      return false;
    }
  }

  function handleValidationError() {
    const errorMessages = {
      1: {
        title: "Invalid Username",
        message:
          "Username must be 2-50 characters and contain only letters, numbers, and underscores",
      },
      2: {
        title: "Invalid Email",
        message: "Please enter a valid email address (e.g., name@example.com)",
      },
      3: {
        title: "Weak Password",
        message: `Password must be ${minPasswordLength}-${maxPasswordLength} characters with at least one letter and one number`,
      },
      4: {
        title: "Terms Required",
        message:
          "Please accept the Terms of Service and Privacy Policy to continue",
      },
    };

    const error = errorMessages[authStep.value as keyof typeof errorMessages];
    if (error) {
      toast.warning(error.title, {
        duration: 4000,
        description: error.message,
        action: {
          label: "Got it",
          onClick: () => {},
        },
      });
    }
  }

  async function handleStepSubmit(e: Event) {
    e.preventDefault();

    if (!validateCurrentStep()) {
      handleValidationError();
      return;
    }

    if (authStep.value < 4) {
      nextAuthStep();
      return;
    }

    await handleFinalAuthStep();
  }

  async function handleFinalAuthStep() {
    try {
      isLoading.value = true;

      if (loadingDelay > 0) {
        await new Promise((resolve) => setTimeout(resolve, loadingDelay));
      }

      // Sanitize data before submission
      const sanitizedData: AuthData = {
        username: authData.value.username.trim(),
        email: authData.value.email.trim().toLowerCase(),
        password: authData.value.password,
        agreeToTerms: authData.value.agreeToTerms,
      };

      const response = (await handleAuth(sanitizedData)) as ApiResponse;

      if (!response) {
        throw new Error("No response received from authentication service");
      }

      if (!response.data || !response.success) {
        throw new Error(
          response.message ||
            "Authentication failed - invalid response structure",
        );
      }

      await handlePostAuthRedirect();
    } catch (err: any) {
      await handleAuthError(err);
    } finally {
      isLoading.value = false;
    }
  }

  async function handleAuth(data: {
    username: string;
    email: string;
    password: string;
    agreeToTerms: boolean;
  }) {
    const { username, email, password, agreeToTerms } = data;

    try {
      const validationError = validateCredentials(
        username,
        email,
        password,
        agreeToTerms,
      );
      if (validationError) {
        throw new Error(validationError);
      }

      let response: ApiResponse;
      let isLogin = false;

      try {
        console.log("Attempting login...");
        response = await unsecureApiCall("/login", {
          method: "POST",
          body: JSON.stringify({
            username,
            email,
            password,
            agree_to_terms: agreeToTerms,
            user_agent: navigator.userAgent,
          }),
        });
        isLogin = true;
      } catch (loginError: any) {
        console.log("Login failed, attempting registration...");

        try {
          response = await unsecureApiCall("/register", {
            method: "POST",
            body: JSON.stringify({
              username,
              email,
              password,
              agree_to_terms: agreeToTerms,
              user_agent: navigator.userAgent,
            }),
          });

          toast.success("Account created successfully!", {
            duration: 3000,
            description: `Welcome ${response.data?.username}!`,
          });
        } catch (registerError: any) {
          if (
            loginError.message?.includes("Connection") ||
            loginError.message?.includes("Network")
          ) {
            throw loginError;
          } else {
            throw registerError;
          }
        }
      }

      if (!response || !response.data) {
        throw new Error("Invalid response from server");
      }

      if (isLogin) {
        toast.success("Welcome back!", {
          duration: 3000,
          description: `Logged in as ${response.data.username}`,
        });
      }

      const userData: UserDetails = {
        userId: response.data.user_id,
        username: response.data.username,
        email: response.data.email,
        createdAt: response.data.created_at,
        sessionId: btoa(email + ":" + password + ":" + username),
        workFunction: response.data.work_function || "",
        preferences: response.data.preferences || "",
        theme: response.data.theme || "system",
        syncEnabled: response.data.sync_enabled,
        phoneNumber: response.data.phone_number || "",
        plan: response.data.plan || "free",
        planName: response.data.plan_name || "",
        amount: response.data.amount || 0,
        duration: response.data.duration || "",
        price: response.data.price || 0,
        responseMode: response.data.response_mode || "light-response",
        expiryTimestamp: response.data.expiry_timestamp || null,
        expireDuration: response.data.expire_duration || "",
        emailVerified: response.data.email_verified || false,
        emailSubscribed: response.data.email_subscribed || true,
        requestCount: response.data.request_count || {
          count: 0,
          timestamp: Date.now(),
        },
      };

      parsedUserDetails.value = userData;
      localStorage.setItem("userdetails", JSON.stringify(userData));

      console.log(
        `Authentication successful for user: ${userData.username} (sync: ${userData.syncEnabled})`,
      );

      return response;
    } catch (error: any) {
      console.error("Authentication error:", error);
      throw error;
    }
  }

  async function handlePostAuthRedirect() {
    // Check URL parameters for redirect intent
    const urlParams = new URLSearchParams(window.location.search);
    const redirectParam = urlParams.get("redirect") || urlParams.get("from");
    const currentRoute = router.currentRoute.value;

    // Whitelist of allowed redirects to prevent open redirect vulnerabilities
    const allowedWholeRedirects = ["upgrade", "chats", "arcade"];
    const allowedPartialRedirects = ["chat/", "arcade/"];
    const isValidRedirect =
      redirectParam &&
      (allowedWholeRedirects.includes(redirectParam) ||
        allowedPartialRedirects.some((prefix) =>
          redirectParam.startsWith(prefix),
        ));

    if (redirectDelay > 0) {
      await new Promise((resolve) => setTimeout(resolve, redirectDelay));
    }

    if (isValidRedirect) {
      console.log(`Redirecting to ${redirectParam} page after authentication`);
      router.push(`/${redirectParam}`);
    } else {
      console.log("Redirecting to home page after authentication");

      // User just logged in
      console.log("✅ User authenticated");

      // Don't navigate if already on a valid chat route
      if (currentRoute.path.startsWith("/chat/")) {
        console.log("Already on chat route, staying here");
        return;
      }

      // Navigate to new chat only if on login/home page
      if (currentRoute.path === "/") {
        console.log("Navigating to new chat");
        router.push("/new");
      }
    }

    resetAuth();
  }

  async function handleAuthError(err: any) {
    console.error("Authentication error:", err);

    const errorMap = {
      timeout: {
        title: "Connection Timeout",
        message:
          "Server took too long to respond. Please check your connection and try again.",
      },
      network: {
        title: "Network Error",
        message:
          "Unable to reach our servers. Please check your internet connection.",
      },
      credentials: {
        title: "Invalid Credentials",
        message: "The username, email, or password you entered is incorrect.",
      },
      duplicate: {
        title: "Account Exists",
        message: "An account with this email or username already exists.",
      },
      server: {
        title: "Server Error",
        message:
          "Our servers are experiencing issues. Please try again in a few minutes.",
      },
      validation: {
        title: "Validation Error",
        message: "Please check your information and try again.",
      },
      ratelimit: {
        title: "Too Many Attempts",
        message: "Please wait a few minutes before trying again.",
      },
      unknown: {
        title: "Authentication Failed",
        message: "An unexpected error occurred. Please try again.",
      },
    };

    let errorType = "unknown";
    const errorMessage = err?.message?.toLowerCase() || "";
    const errorCode = err?.code?.toLowerCase() || "";

    if (errorMessage.includes("timeout") || errorCode === "etimedout") {
      errorType = "timeout";
    } else if (
      errorMessage.includes("network") ||
      errorMessage.includes("fetch") ||
      errorCode === "network_error"
    ) {
      errorType = "network";
    } else if (
      errorMessage.includes("credentials") ||
      errorMessage.includes("unauthorized") ||
      err?.status === 401
    ) {
      errorType = "credentials";
    } else if (
      errorMessage.includes("duplicate") ||
      errorMessage.includes("already exists") ||
      err?.status === 409
    ) {
      errorType = "duplicate";
    } else if (errorMessage.includes("rate limit") || err?.status === 429) {
      errorType = "ratelimit";
    } else if (errorMessage.includes("server") || err?.status >= 500) {
      errorType = "server";
    } else if (errorMessage.includes("validation") || err?.status === 400) {
      errorType = "validation";
    }

    const error = errorMap[errorType as keyof typeof errorMap];

    toast.error(error.title, {
      duration: 6000,
      description: error.message,
      action: {
        label: "Retry",
        onClick: () => {
          authStep.value = 1;
        },
      },
    });
  }

  function updateAuthData(field: keyof AuthData, value: any) {
    authData.value[field] = value as never;
  }

  function resetAuth() {
    authStep.value = 1;
    authData.value = {
      username: "",
      email: "",
      password: "",
      agreeToTerms: false,
    };
    isLoading.value = false;
  }

  return {
    authStep,
    authData,
    isLoading,
    nextAuthStep,
    prevAuthStep,
    validateCurrentStep,
    handleValidationError,
    handleStepSubmit,
    handleAuthError,
    handleFinalAuthStep,
    updateAuthData,
    resetAuth,
  };
}
