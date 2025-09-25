<script lang="ts" setup>
import { inject, onMounted, ref, watch, type Ref } from "vue";
import { toast } from "vue-sonner";
import { API_BASE_URL } from "../utils/globals";
import { useRouter } from "vue-router";

const username = ref("");
const email = ref("");
const password = ref("");
const confirmText = ref("");
const loading = ref(false);
const errorMsg = ref("");
const successMsg = ref("");
const showConfirm = ref(false);

const globalState = inject("globalState");
const {
    parsedUserDetails,
    isAuthenticated
}=globalState as {
    parsedUserDetails: Ref<any>,
    isAuthenticated: Ref<boolean>
}

const router = useRouter();
// Read stored credentials
const storedUser = JSON.parse(localStorage.getItem("userdetails") || "{}");

// Validate form before showing confirmation
function validateAndShowConfirm() {
    errorMsg.value = "";

    if (!username.value || !email.value || !password.value) {
        errorMsg.value = "Please fill in all fields.";
        return;
    }

    // Verify details match stored credentials
    if (
        storedUser.username !== username.value.trim() ||
        storedUser.email !== email.value.trim()
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
                "X-User-ID": storedUser.user_id || "",
            },
            body: JSON.stringify({
                username: username.value.trim(),
                email: email.value.trim(),
                password: password.value,
            }),
        });

        const responseData = await res.json().catch(() => ({}));

        if (!res.ok) {
            throw new Error(responseData.message || `HTTP ${res.status}: ${res.statusText}`);
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
        errorMsg.value = err.message || "Failed to delete account. Please try again.";
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
  } else if (isAuthenticated.value) {
    router.push('/')
  }
  if (!isAuthenticated.value) {
    router.push('/login')
  }
})

watch(isAuthenticated, (val) => {
  if (val === false) {
    router.push('/')
  }
})

</script>

<template>
    <div class="flex items-center font-light justify-center min-h-screen">
        <div class="w-full max-w-md max-lg:px-4 bg-none p-6">
            <h1 v-if="!showConfirm" class="text-xl font-semibold text-red-600 mb-2">Delete My Account</h1>
            <p v-if="!showConfirm" class="text-sm text-gray-600 mb-6">
                Please confirm your credentials before deleting your account. <br />
                <b class="font-semibold text-red-700">This action is irreversible.</b>
            </p>

            <!-- Error / Success messages -->
            <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                <p class="text-sm text-red-700">{{ errorMsg }}</p>
            </div>

            <div v-if="successMsg" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
                <p class="text-sm text-green-700">{{ successMsg }}</p>
            </div>

            <!-- Main Form -->
            <form v-if="!showConfirm" @submit.prevent="validateAndShowConfirm" class="flex flex-col gap-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Username</label>
                    <input v-model="username" type="text" placeholder="Enter your username"
                        class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
                        :disabled="loading" required />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
                    <input v-model="email" type="email" placeholder="Enter your email"
                        class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
                        :disabled="loading" required />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
                    <input v-model="password" type="password" placeholder="Enter your password"
                        class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
                        :disabled="loading" required />
                </div>

                <div class="flex gap-3">
                    <button type="button" @click="$router.back()"
                        class="flex-1 bg-gray-300 hover:bg-gray-400 text-gray-700 py-2 px-2 rounded-md transition duration-200 disabled:opacity-50">
                        Cancel
                    </button>

                    <button type="submit" :disabled="loading"
                        class="flex-1 flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 text-white py-2 px-2 rounded-md transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium">
                        Continue
                    </button>
                </div>
            </form>

            <!-- Confirmation Dialog -->
            <div v-if="showConfirm" class="space-y-4">
                <div class="bg-yellow-50 border border-yellow-200 rounded-md p-4">
                    <h3 class="font-semibold text-yellow-800 mb-2">⚠️ Final Confirmation Required</h3>
                    <p class="text-sm text-yellow-700 mb-3">
                        You are about to permanently delete your account. This will:
                    </p>
                    <ul class="text-sm text-yellow-700 list-disc list-inside space-y-1 mb-3">
                        <li>Delete all your chat data</li>
                        <li>Remove your account information</li>
                        <li>Cannot be undone</li>
                    </ul>
                    <p class="text-sm font-medium text-yellow-800">
                        Type "delete my account" below to confirm:
                    </p>
                </div>

                <input v-model="confirmText" type="text" placeholder="Type: delete my account"
                    class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
                    :disabled="loading" />

                <div class="flex gap-3">
                    <button @click="cancelDelete" :disabled="loading"
                        class="flex-1 bg-gray-300 hover:bg-gray-400 text-gray-700 py-2 px-2 rounded-md transition duration-200 disabled:opacity-50">
                        Cancel
                    </button>

                    <button @click="handleDeleteAccount"
                        :disabled="loading || confirmText.toLowerCase() !== 'delete my account'"
                        class="flex-1 flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 text-white py-2 px-2 rounded-md transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium">
                        <span v-if="loading" class="animate-spin">⟳</span>
                        <span>{{ loading ? "Deleting..." : "Delete Account" }}</span>
                    </button>
                </div>
            </div>
            <p class="text-xs text-gray-500 mt-2">
                Your account data will be permanently deleted and cannot be recovered.
            </p>
        </div>
    </div>
</template>