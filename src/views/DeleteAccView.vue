<script lang="ts" setup>
import { inject, onMounted, ref, watch, type Ref } from "vue";
import { toast } from "vue-sonner";
import { API_BASE_URL } from "../utils/globals";
import { useRouter } from "vue-router";
import type { UserDetails } from "@/types";

const username = ref("");
const email = ref("");
const password = ref("");
const confirmText = ref("");
const loading = ref(false);
const errorMsg = ref("");
const successMsg = ref("");
const showConfirm = ref(false);

const globalState = inject("globalState");
const { parsedUserDetails } = globalState as {
    parsedUserDetails: Ref<UserDetails>;
};

const router = useRouter();

// Validate form before showing confirmation
function validateAndShowConfirm() {
    errorMsg.value = "";

    if (!username.value || !email.value || !password.value) {
        errorMsg.value = "Please fill in all fields.";
        return;
    }

    // Verify details match stored credentials
    if (
        parsedUserDetails.value.username !== username.value.trim() ||
        parsedUserDetails.value.email !== email.value.trim()
    ) {
        errorMsg.value = "Entered details do not match your account.";
        return;
    }

    showConfirm.value = true;
}

// Call backend DELETE endpoint
async function handleDeleteAccount() {
    // Final confirmation check
    if (confirmText.value.toLowerCase() !== "delete my account") {
        errorMsg.value = 'Please type "delete my account" to confirm.';
        return;
    }

    loading.value = true;
    errorMsg.value = "";
    successMsg.value = "";

    try {
        const res = await fetch(`${API_BASE_URL}/delete_account`, {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                "X-User-ID": parsedUserDetails.value.userId || "",
            },
            body: JSON.stringify({
                username: username.value.trim(),
                email: email.value.trim(),
                password: password.value,
            }),
        });

        const responseData = await res.json().catch(() => ({}));

        if (!res.ok) {
            throw new Error(
                responseData.message || `HTTP ${res.status}: ${res.statusText}`,
            );
        }

        successMsg.value = "Your account has been permanently deleted.";

        // Clear all stored data
        localStorage.clear();
        sessionStorage.clear();

        // Redirect after a delay
        setTimeout(() => {
            window.location.href = "/";
        }, 2000);
    } catch (err: any) {
        errorMsg.value =
            err.message || "Failed to delete account. Please try again.";
        showConfirm.value = false; // Reset confirmation dialog
        toast.error("Account Deletion Failed", {
            duration: 4000,
            description: errorMsg.value,
        });
    } finally {
        loading.value = false;
    }
}

function cancelDelete() {
    showConfirm.value = false;
    confirmText.value = "";
    errorMsg.value = "";
}

// Lifecycle hooks
onMounted(() => {
    if (parsedUserDetails.value) {
        username.value = parsedUserDetails.value.username || "";
        email.value = parsedUserDetails.value.email || "";
        return;
    }

    router.push("/");
});
</script>

<template>
    <div
        class="flex items-center font-light justify-center min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-300"
    >
        <div class="w-full max-w-md max-lg:px-4 p-6">
            <h1
                v-if="!showConfirm"
                class="text-xl font-semibold text-red-600 dark:text-red-500 mb-2"
            >
                Delete My Account
            </h1>
            <p
                v-if="!showConfirm"
                class="text-sm text-gray-600 dark:text-gray-400 mb-6"
            >
                Please confirm your credentials before deleting your account.
                <br />
                <b class="font-semibold text-red-700 dark:text-red-400"
                    >This action is irreversible.</b
                >
            </p>

            <!-- Error / Success messages -->
            <div
                v-if="errorMsg"
                class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md transition-colors duration-200"
            >
                <p class="text-sm text-red-700 dark:text-red-400">
                    {{ errorMsg }}
                </p>
            </div>

            <div
                v-if="successMsg"
                class="mb-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md transition-colors duration-200"
            >
                <p class="text-sm text-green-700 dark:text-green-400">
                    {{ successMsg }}
                </p>
            </div>

            <!-- Main Form -->
            <form
                v-if="!showConfirm"
                @submit.prevent="validateAndShowConfirm"
                class="flex flex-col gap-4"
            >
                <div>
                    <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >Username</label
                    >
                    <input
                        v-model="username"
                        type="text"
                        placeholder="Enter your username"
                        class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 transition-colors duration-200"
                        :disabled="loading"
                        required
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >Email</label
                    >
                    <input
                        v-model="email"
                        type="email"
                        placeholder="Enter your email"
                        class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 transition-colors duration-200"
                        :disabled="loading"
                        required
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >Password</label
                    >
                    <input
                        v-model="password"
                        type="password"
                        placeholder="Enter your password"
                        class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 transition-colors duration-200"
                        :disabled="loading"
                        required
                    />
                </div>

                <div class="flex gap-3">
                    <button
                        type="button"
                        @click="$router.back()"
                        class="flex-1 bg-gray-300 dark:bg-gray-700 hover:bg-gray-400 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 py-2 px-2 rounded-md transition-colors duration-200 disabled:opacity-50 font-medium"
                    >
                        Cancel
                    </button>

                    <button
                        type="submit"
                        :disabled="loading"
                        class="flex-1 flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600 text-white py-2 px-2 rounded-md transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                    >
                        <i
                            v-if="loading"
                            class="pi pi-spin pi-spinner text-sm"
                        ></i>
                        <span>{{ loading ? "Verifying..." : "Continue" }}</span>
                    </button>
                </div>
            </form>

            <!-- Confirmation Dialog -->
            <div v-if="showConfirm" class="space-y-4">
                <div
                    class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-4 transition-colors duration-200"
                >
                    <div class="flex items-start gap-3">
                        <div
                            class="flex-shrink-0 w-6 h-6 bg-yellow-100 dark:bg-yellow-800 rounded-full flex items-center justify-center"
                        >
                            <i
                                class="pi pi-exclamation-triangle text-yellow-600 dark:text-yellow-400 text-sm"
                            ></i>
                        </div>
                        <div>
                            <h3
                                class="font-semibold text-yellow-800 dark:text-yellow-400 mb-2"
                            >
                                Final Confirmation Required
                            </h3>
                            <p
                                class="text-sm text-yellow-700 dark:text-yellow-300 mb-3"
                            >
                                You are about to permanently delete your
                                account. This will:
                            </p>
                            <ul
                                class="text-sm text-yellow-700 dark:text-yellow-300 list-disc list-inside space-y-1 mb-3"
                            >
                                <li>Delete all your chat data and history</li>
                                <li>
                                    Remove your account information permanently
                                </li>
                                <li>Cancel any active subscriptions</li>
                                <li>Cannot be undone or recovered</li>
                            </ul>
                            <p
                                class="text-sm font-medium text-yellow-800 dark:text-yellow-400"
                            >
                                Type "delete my account" below to confirm:
                            </p>
                        </div>
                    </div>
                </div>

                <input
                    v-model="confirmText"
                    type="text"
                    placeholder="Type: delete my account"
                    class="w-full border border-gray-300 dark:border-gray-600 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 transition-colors duration-200"
                    :disabled="loading"
                />

                <div class="flex gap-3">
                    <button
                        @click="cancelDelete"
                        :disabled="loading"
                        class="flex-1 bg-gray-300 dark:bg-gray-700 hover:bg-gray-400 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 py-2 px-2 rounded-md transition-colors duration-200 disabled:opacity-50 font-medium"
                    >
                        Cancel
                    </button>

                    <button
                        @click="handleDeleteAccount"
                        :disabled="
                            loading ||
                            confirmText.toLowerCase() !== 'delete my account'
                        "
                        :class="
                            confirmText.toLowerCase() === 'delete my account'
                                ? 'bg-red-600 hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600'
                                : 'bg-red-400 dark:bg-red-700 cursor-not-allowed'
                        "
                        class="flex-1 flex items-center justify-center gap-2 text-white py-2 px-2 rounded-md transition-colors duration-200 disabled:opacity-50 font-medium"
                    >
                        <i
                            v-if="loading"
                            class="pi pi-spin pi-spinner text-sm"
                        ></i>
                        <span>{{
                            loading ? "Deleting..." : "Delete Account"
                        }}</span>
                    </button>
                </div>
            </div>

            <!-- Footer note -->
            <p
                class="text-xs text-gray-500 dark:text-gray-400 mt-4 text-center"
            >
                Your account data will be permanently deleted and cannot be
                recovered.
                <br />This process may take a few moments to complete.
            </p>
        </div>
    </div>
</template>
