import { ref, type Ref } from "vue";
import { toast } from "vue-sonner";
import { useRouter } from "vue-router";

export interface AuthData {
  username: string;
  email: string;
  password: string;
  agreeToTerms: boolean;
}

export function useAuth() {
  const router = useRouter();

  const authStep = ref(1);
  const authData: Ref<AuthData> = ref({
    username: "",
    email: "",
    password: "",
    agreeToTerms: false,
  });
  const isLoading = ref(false);

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
          return !!(username && username.length >= 2 && username.length <= 50);
        }
        case 2: {
          const email = authData.value.email?.trim();
          return !!(
            email &&
            email.length > 0 &&
            email.length <= 100 &&
            /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
          );
        }
        case 3: {
          const password = authData.value.password;
          return !!(password && password.length > 7 && password.length < 25);
        }
        case 4: {
          const agreeToTerms = authData.value.agreeToTerms;
          return agreeToTerms;
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
        message:
          "Password must be at least 7 characters with a mix of letters and numbers",
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

  async function handleStepSubmit(
    e: Event,
    handleAuth: (data: AuthData) => Promise<any>,
  ) {
    e.preventDefault();

    // Early validation with improved error handling
    if (!validateCurrentStep()) {
      handleValidationError();
      return;
    }

    // Handle step progression vs final submission
    if (authStep.value < 4) {
      nextAuthStep();
      return;
    }

    // Final step - create session
    await handleFinalAuthStep(handleAuth);
  }

  async function handleFinalAuthStep(
    handleAuth: (data: AuthData) => Promise<any>,
  ) {
    try {
      isLoading.value = true;

      // Add a small delay to show the loading state
      await new Promise((resolve) => setTimeout(resolve, 500));

      const response = await handleAuth(authData.value);

      // Validate response structure
      if (!response) {
        throw new Error("No response received from authentication service");
      }

      if (response.error) {
        throw new Error(response.error);
      }

      if (!response.data || !response.success) {
        throw new Error("Authentication failed - invalid response structure");
      }

      // Success handling
      await handleAuthSuccess();
    } catch (err: any) {
      await handleAuthError(err);
    } finally {
      isLoading.value = false;
    }
  }

  async function handleAuthSuccess() {
    // Handle redirect logic
    await handlePostAuthRedirect();

    // Reset form state
    authStep.value = 1;
    authData.value = {
      username: "",
      email: "",
      password: "",
      agreeToTerms: false,
    };
  }

  async function handlePostAuthRedirect() {
    // Check multiple sources for upgrade redirect
    const previousRoute = document.referrer;
    const urlParams = new URLSearchParams(window.location.search);
    const isFromUpgrade =
      previousRoute.includes("/upgrade") ||
      (urlParams.has("from") && urlParams.get("from") === "upgrade") ||
      (urlParams.has("redirect") && urlParams.get("redirect") === "upgrade");

    if (isFromUpgrade) {
      console.log("Redirecting to upgrade page after authentication");
      // Small delay for better UX flow
      await new Promise((resolve) => setTimeout(resolve, 1000));
      router.push("/upgrade");
    } else {
      console.log("Redirecting to home page after authentication");
      // Small delay for better UX flow
      await new Promise((resolve) => setTimeout(resolve, 1000));
      router.push("/");
    }
  }

  async function handleAuthError(err: any) {
    console.error("Authentication error:", err);

    // Map specific error types to user-friendly messages
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
      server: {
        title: "Server Error",
        message:
          "Our servers are experiencing issues. Please try again in a few minutes.",
      },
      validation: {
        title: "Validation Error",
        message: "Please check your information and try again.",
      },
      unknown: {
        title: "Authentication Failed",
        message: "An unexpected error occurred. Please try again.",
      },
    };

    // Determine error type from error object
    let errorType = "unknown";
    const errorMessage = err?.message?.toLowerCase() || "";

    if (errorMessage.includes("timeout")) {
      errorType = "timeout";
    } else if (
      errorMessage.includes("network") ||
      errorMessage.includes("fetch")
    ) {
      errorType = "network";
    } else if (
      errorMessage.includes("credentials") ||
      errorMessage.includes("unauthorized")
    ) {
      errorType = "credentials";
    } else if (errorMessage.includes("server") || err?.status >= 500) {
      errorType = "server";
    } else if (errorMessage.includes("validation")) {
      errorType = "validation";
    }

    const error = errorMap[errorType as keyof typeof errorMap];

    toast.error(error.title, {
      duration: 6000,
      description: error.message,
      action: {
        label: "Retry",
        onClick: () => {
          // Reset to step 1 for retry
          authStep.value = 1;
        },
      },
    });
  }

  function updateAuthData(field: keyof AuthData, value: any) {
    authData.value[field] = value as never;
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
    handleAuthSuccess,
    handleAuthError,
    handleFinalAuthStep,
    updateAuthData,
  };
}
