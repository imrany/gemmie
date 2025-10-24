import type { PlatformError } from "@/types";
import { toast } from "vue-sonner";
import { useApiCall } from "./useApiCall";

// Error queue management
const errorQueue: PlatformError[] = [];
const MAX_QUEUE_SIZE = 50;
const MAX_STORED_ERRORS = 100;
let isReporting = false;
let reportTimeout: any = null;

/**
 * Generate unique error ID
 */
export const generateErrorId = (): string => {
  try {
    return crypto.randomUUID();
  } catch {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }
};

export function usePlatformError() {
  /**
   * Store error in localStorage with queue management
   */
  const storeErrorLocally = (error: PlatformError) => {
    try {
      const stored = localStorage.getItem("platform_errors");
      const errors: PlatformError[] = stored ? JSON.parse(stored) : [];

      // Add new error
      errors.push(error);

      // Keep only the most recent errors
      const recentErrors = errors.slice(-MAX_STORED_ERRORS);

      localStorage.setItem("platform_errors", JSON.stringify(recentErrors));
    } catch (storageError) {
      console.error("Failed to store error locally:", storageError);
      // If storage fails, try to clear old errors and retry
      try {
        localStorage.removeItem("platform_errors");
        localStorage.setItem("platform_errors", JSON.stringify([error]));
      } catch {
        console.error("Critical: Unable to store errors");
      }
    }
  };

  /**
   * Send error reports to server in batch
   */
  const sendErrorBatch = async () => {
    if (isReporting || errorQueue.length === 0) return;

    isReporting = true;
    const errorsToSend = [...errorQueue];
    errorQueue.length = 0; // Clear queue

    try {
      const { apiCall } = useApiCall();
      await apiCall("/errors", {
        method: "POST",
        body: JSON.stringify({
          errors: errorsToSend,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
        }),
      });

      console.log(
        `✅ Successfully reported ${errorsToSend.length} error(s) to server`,
      );
    } catch (reportError: any) {
      console.error("Failed to report errors to server:", reportError);

      // Re-add errors to queue (up to max size) for retry
      const errorsToRequeue = errorsToSend.slice(
        0,
        MAX_QUEUE_SIZE - errorQueue.length,
      );
      errorQueue.unshift(...errorsToRequeue);

      // Schedule retry
      if (reportTimeout) clearTimeout(reportTimeout);
      reportTimeout = setTimeout(() => {
        sendErrorBatch();
      }, 30000); // Retry after 30 seconds
    } finally {
      isReporting = false;
    }
  };

  /**
   * Main error reporting function
   */
  const reportError = (platformError: PlatformError) => {
    const {
      action,
      message,
      description,
      context = {},
      status,
      userId,
      severity = "medium",
    } = platformError;

    // Create complete error object with proper defaults
    const completeError: PlatformError = {
      id: platformError.id || generateErrorId(),
      action,
      message,
      status: status || "unknown",
      userId: userId || "anonymous",
      severity,
      description,
      context,
      createdAt: platformError.createdAt || new Date().toISOString(),
    };

    try {
      // Log to console with context
      const logLevel =
        severity === "critical" || severity === "high" ? "error" : "warn";
      console[logLevel](`Platform Error [${action}]`, {
        id: completeError.id,
        message: completeError.message,
        status: completeError.status,
        description: completeError.description,
        userId: completeError.userId,
        severity: completeError.severity,
        context: completeError.context,
        timestamp: completeError.createdAt,
      });

      // Store locally immediately
      storeErrorLocally(completeError);

      // Add to queue for batch reporting
      if (errorQueue.length < MAX_QUEUE_SIZE) {
        errorQueue.push(completeError);

        // Debounce batch sending
        if (reportTimeout) clearTimeout(reportTimeout);
        reportTimeout = setTimeout(() => {
          sendErrorBatch();
        }, 2000); // Send after 2 seconds of no new errors
      }

      // Show user-friendly toast for critical/high severity errors
      if (severity === "critical" || severity === "high") {
        toast.error(message.split(":")[0].trim() || "Something went wrong", {
          duration: 5000,
          description:
            description || "Our team has been notified. Please try again.",
        });
      } else if (severity === "medium") {
        // Optional: Show less intrusive notification for medium severity
        console.warn(`⚠️ [${action}] ${message}`);
        toast.warning(message.split(":")[0].trim() || "Something went wrong", {
          duration: 5000,
          description:
            description || "Our team has been notified. Please try again.",
        });
      }

      return completeError.id;
    } catch (error: any) {
      // Fallback logging if error reporting itself fails
      console.error("❌ Critical: Failed to report error", {
        originalError: platformError,
        reportingError: error,
      });

      // Last resort: try to show toast
      try {
        toast.error("Critical error occurred", {
          duration: 3000,
          description: "Please refresh the page",
        });
      } catch {
        // If even toast fails, we can't do much more
        console.error("Complete failure in error reporting system");
      }

      return null;
    }
  };

  /**
   * Get stored errors from localStorage
   */
  const getStoredErrors = (): PlatformError[] => {
    try {
      const stored = localStorage.getItem("platform_errors");
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  };

  /**
   * Clear stored errors
   */
  const clearStoredErrors = () => {
    try {
      localStorage.removeItem("platform_errors");
      errorQueue.length = 0;
      console.log("✅ Cleared all stored errors");
    } catch (error) {
      console.error("Failed to clear stored errors:", error);
    }
  };

  /**
   * Manual trigger to send queued errors
   */
  const flushErrors = async () => {
    if (errorQueue.length > 0) {
      await sendErrorBatch();
    }
  };

  return {
    reportError,
    getStoredErrors,
    clearStoredErrors,
    flushErrors,
  };
}

// Optional: Auto-flush errors on page unload
if (typeof window !== "undefined") {
  window.addEventListener("beforeunload", () => {
    const { flushErrors } = usePlatformError();
    flushErrors().catch(console.error);
  });
}
