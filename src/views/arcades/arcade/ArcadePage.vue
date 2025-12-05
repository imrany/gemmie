<script setup lang="ts">
import type { ApiResponse, RawArcade, UserDetails } from "@/types";
import { ref, computed, nextTick } from "vue";
import { inject } from "vue";
import { watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { toast } from "vue-sonner";
import {
    Loader2,
    Trash2,
    Edit,
    Download,
    Share2,
    Copy,
    X,
    Code2,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import LoadingState from "@/components/LoadingState.vue";
import ErrorState from "@/components/ErrorState.vue";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { Ref } from "vue";
import DialogBox from "@/components/Dialog/DialogBox.vue";
import OverallLayout from "@/layout/OverallLayout.vue";

const route = useRoute();
const router = useRouter();
const arcadeId = ref(route.params.id as string);

const {
    unsecureApiCall,
    apiCall,
    parsedUserDetails,
    openPreview,
    loadChat,
    updateExpandedArray,
    loadChatDrafts,
    arcade,
} = inject("globalState") as {
    arcade: Ref<RawArcade>;
    updateExpandedArray: () => void;
    loadChatDrafts: () => void;
    loadChat: (id: string) => Promise<[boolean, string]>;
    openPreview: (
        code: string,
        language: string,
        data?: {
            fileSize: string;
            wordCount: number;
            charCount: number;
        },
    ) => void;
    unsecureApiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
    apiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
    parsedUserDetails: Ref<UserDetails>;
};

const isLoading = ref(true);
const error = ref<string | null>(null);
const iframeLoaded = ref(false);
const showErrorSection = ref(false);

// Dialog states
const showDeleteDialog = ref(false);
const showRenameDialog = ref(false);
const isDeleting = ref(false);
const isRenaming = ref(false);

// Rename form
const renameForm = ref({
    label: "",
    description: "",
});

const renameErrors = ref<{
    label?: string;
    description?: string;
}>({});

// Check if current user owns the arcade
const isOwner = computed(() => {
    return arcade.value?.user_id === parsedUserDetails.value?.userId;
});

async function fetchArcade() {
    isLoading.value = true;
    error.value = null;
    iframeLoaded.value = false;

    try {
        const res = await unsecureApiCall<RawArcade>(
            `/arcades/${arcadeId.value.trim()}`,
            {
                method: "GET",
            },
        );

        if (!res.success) {
            throw new Error(res.message || "Failed to load arcade content");
        }

        if (res.data) {
            arcade.value = res.data;
        } else {
            throw new Error("No arcade data found");
        }
    } catch (err: any) {
        console.error("Fetch arcade error:", err);
        if (err.message?.includes("Failed to fetch")) {
            error.value =
                "Failed to connect to the server. Please check your internet connection.";
        } else if (err.message?.includes("404")) {
            error.value =
                "Arcade not found. It may have been removed or the ID is incorrect.";
        } else {
            error.value = err.message || "Failed to load arcade content";
        }
    } finally {
        isLoading.value = false;
    }
}

function handleIframeLoad() {
    iframeLoaded.value = true;
}

function handleIframeError() {
    error.value = "Failed to load content in preview";
    toast.error("Preview failed to load");
}

// Rename functionality
function openRenameDialog() {
    if (arcade.value) {
        renameForm.value = {
            label: arcade.value.label,
            description: arcade.value.description,
        };
        renameErrors.value = {};
        showRenameDialog.value = true;
    }
}

function validateRenameForm(): boolean {
    renameErrors.value = {};

    if (!renameForm.value.label.trim()) {
        renameErrors.value.label = "Label is required";
        return false;
    }

    if (renameForm.value.label.trim().length > 100) {
        renameErrors.value.label = "Label must be 100 characters or less";
        return false;
    }

    if (!renameForm.value.description.trim()) {
        renameErrors.value.description = "Description is required";
        return false;
    }

    if (renameForm.value.description.trim().length > 500) {
        renameErrors.value.description =
            "Description must be 500 characters or less";
        return false;
    }

    return true;
}

async function handleRename() {
    if (!validateRenameForm()) {
        return;
    }

    isRenaming.value = true;

    try {
        const res = await apiCall<RawArcade>(`/arcades/${arcadeId.value}`, {
            method: "PUT",
            body: JSON.stringify({
                label: renameForm.value.label.trim(),
                description: renameForm.value.description.trim(),
            }),
        });

        if (!res.success) {
            throw new Error(res.message || "Failed to update arcade");
        }

        if (res.data) {
            toast.success(res.message);
            // Update local data
            if (arcade.value) {
                arcade.value.label = res.data.label.trim();
                arcade.value.description = res.data.description.trim();
            }
        }

        showRenameDialog.value = false;
    } catch (err: any) {
        console.error("Rename error:", err);
        toast.error("Failed to update arcade", {
            description: err.message,
        });
    } finally {
        isRenaming.value = false;
    }
}

// Delete functionality
function openDeleteDialog() {
    showDeleteDialog.value = true;
}

async function handleDelete() {
    isDeleting.value = true;

    try {
        const res = await apiCall<void>(`/arcades/${arcadeId.value}`, {
            method: "DELETE",
        });

        if (!res.success) {
            throw new Error(res.message || "Failed to delete arcade");
        }

        toast.success("Arcade deleted successfully!");

        // Navigate back to arcade list
        router.push("/arcade");
    } catch (err: any) {
        console.error("Delete error:", err);
        toast.error("Failed to delete arcade", {
            description: err.message,
        });
    } finally {
        isDeleting.value = false;
        showDeleteDialog.value = false;
    }
}

// Download functionality
function handleDownload() {
    if (!arcade.value) return;

    const blob = new Blob([arcade.value.code], { type: "text/html" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${arcade.value.label.replace(/[^a-z0-9]/gi, "_")}.html`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast.success("Download started!");
}

// Share functionality
async function handleShare() {
    const url = window.location.href;

    if (navigator.share) {
        try {
            await navigator.share({
                title: arcade.value?.label,
                text: arcade.value?.description,
                url: url,
            });
            toast.success("Shared successfully!");
        } catch (err: any) {
            if (err.name !== "AbortError") {
                // Fallback to copy
                copyLink();
            }
        }
    } else {
        // Fallback to copy
        copyLink();
    }
}

async function copyLink() {
    try {
        await navigator.clipboard.writeText(window.location.href);
        toast.success("Link copied to clipboard!");
    } catch (err) {
        toast.error("Failed to copy link");
    }
}

const showFloatActionBtn = ref(true);
function handleClose() {
    showFloatActionBtn.value = false;
}

function openLoginPage() {
    handleClose();
    window.open("/", "blank");
}

watch(
    arcadeId,
    async (newId) => {
        if (!newId) return; // Guard clause

        arcadeId.value = newId; // Ensure arcadeId is updated first
        await fetchArcade();

        const [success, message] = await loadChat(newId);
        if (success) {
            showErrorSection.value = false;
            updateExpandedArray();
            nextTick(() => {
                loadChatDrafts();
            });
            console.log(message);
        } else {
            console.warn(message);
            showErrorSection.value = true;
        }
    },
    {
        immediate: true,
    },
);
</script>

<template>
    <div class="w-screen h-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
        <OverallLayout>
            <div class="h-full w-full">
                <!-- Loading State -->
                <LoadingState
                    description="Please wait while we fetch your content"
                    label="Loading Arcade Content"
                    v-if="isLoading"
                />

                <!-- Error State -->
                <ErrorState
                    v-else-if="error"
                    @retry="fetchArcade"
                    back-button-text="Back to Arcade"
                    :error="error"
                />

                <!-- Content State -->
                <div v-else-if="arcade?.code" class="relative w-full h-full">
                    <!-- Loading overlay for iframe -->
                    <Transition name="fade">
                        <div
                            v-if="!iframeLoaded"
                            class="absolute inset-0 flex items-center justify-center bg-white dark:bg-gray-900 z-10"
                        >
                            <div class="text-center">
                                <Loader2
                                    class="w-10 h-10 text-blue-500 animate-spin mx-auto mb-3"
                                />
                                <p
                                    class="text-sm text-gray-500 dark:text-gray-400"
                                >
                                    Rendering preview...
                                </p>
                            </div>
                        </div>
                    </Transition>

                    <!-- Floating Action Buttons -->
                    <div
                        v-if="showFloatActionBtn"
                        class="fixed bottom-6 right-6 z-20"
                    >
                        <!-- Owner Menu -->
                        <DropdownMenu v-if="isOwner">
                            <DropdownMenuTrigger as-child>
                                <Button
                                    size="sm"
                                    class="h-8 w-40 flex flex-col justify-center gap-0.5 bg-gray-900 dark:bg-gray-800 hover:bg-gray-800 dark:hover:bg-gray-700 rounded-md shadow-lg hover:shadow-xl transition-all relative"
                                >
                                    <div
                                        class="absolute top-1 right-1 z-10 cursor-pointer group"
                                        @click.stop.prevent="handleClose"
                                    >
                                        <X
                                            class="w-0.5 h-0.5 text-gray-400 group-hover:text-white transition-colors"
                                        />
                                    </div>
                                    <div
                                        class="flex items-center justify-center gap-1 mt-0.5"
                                    >
                                        <p
                                            class="text-xs flex gap-1 font-medium text-gray-400 leading-none"
                                        >
                                            Edit with
                                            <span
                                                class="text-white flex items-center gap-2 font-semibold"
                                            >
                                                <img
                                                    src="/logo-light.svg"
                                                    alt="Gemmie Logo"
                                                    class="w-3 h-3"
                                                />
                                                <span>Gemmie</span>
                                            </span>
                                        </p>
                                    </div>
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" class="w-52">
                                <DropdownMenuItem
                                    @click="openRenameDialog"
                                    class="cursor-pointer"
                                >
                                    <Edit class="w-4 h-4 mr-2" />
                                    <span>Edit Details</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    @click="
                                        () => {
                                            if (arcade)
                                                openPreview(
                                                    arcade?.code,
                                                    arcade?.code_type || 'html',
                                                    undefined,
                                                );
                                        }
                                    "
                                    class="cursor-pointer"
                                >
                                    <Code2 class="w-4 h-4 mr-2" />
                                    <span>Update Code</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    @click="handleDownload"
                                    class="cursor-pointer"
                                >
                                    <Download class="w-4 h-4 mr-2" />
                                    <span>Download</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    @click="handleShare"
                                    class="cursor-pointer"
                                >
                                    <Share2 class="w-4 h-4 mr-2" />
                                    <span>Share</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                    @click="copyLink"
                                    class="cursor-pointer"
                                >
                                    <Copy class="w-4 h-4 mr-2" />
                                    <span>Copy Link</span>
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    @click="openDeleteDialog"
                                    class="cursor-pointer text-red-600 dark:text-red-400 focus:text-red-600 dark:focus:text-red-400"
                                >
                                    <Trash2 class="w-4 h-4 mr-2" />
                                    <span>Delete</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>

                        <!-- Non-Owner Share Button -->
                        <Button
                            v-else
                            @click="openLoginPage"
                            size="sm"
                            class="h-8 w-40 flex flex-col justify-center gap-0.5 bg-gray-900 dark:bg-gray-800 hover:bg-gray-800 dark:hover:bg-gray-700 rounded-md shadow-lg hover:shadow-xl transition-all relative"
                        >
                            <div
                                class="absolute top-1 right-1 z-10 cursor-pointer group"
                                @click.stop.prevent="handleClose"
                            >
                                <X
                                    class="w-0.5 h-0.5 text-gray-400 group-hover:text-white transition-colors"
                                />
                            </div>
                            <div
                                class="flex items-center justify-center gap-1 mt-0.5"
                            >
                                <p
                                    class="text-xs flex gap-1 font-medium text-gray-400 leading-none"
                                >
                                    Edit with
                                    <span
                                        class="text-white flex items-center gap-2 font-semibold"
                                    >
                                        <img
                                            src="/logo-light.svg"
                                            alt="Gemmie Logo"
                                            class="w-3 h-3"
                                        />
                                        <span>Gemmie</span>
                                    </span>
                                </p>
                            </div>
                        </Button>
                    </div>

                    <!-- Iframe -->
                    <iframe
                        :srcdoc="arcade.code"
                        @load="handleIframeLoad"
                        @error="handleIframeError"
                        class="w-full h-full border-0 bg-white dark:bg-gray-900"
                        sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox allow-modals allow-forms allow-pointer-lock allow-same-origin allow-top-navigation allow-top-navigation-by-user-activation"
                        title="Arcade Content Preview"
                        referrerpolicy="no-referrer"
                    />
                </div>

                <!-- No Content State -->
                <ErrorState
                    v-else
                    @retry="fetchArcade"
                    back-button-text="Back to Arcade"
                    error="This arcade entry doesn't have any content to display."
                />

                <!-- Rename Dialog -->
                <DialogBox
                    :close-modal="() => (showRenameDialog = false)"
                    name="rename-dialog"
                    :show="showRenameDialog"
                >
                    <div class="sm:max-w-[500px]">
                        <div>
                            <p
                                class="text-xl font-bold dark:text-gray-100 text-gray-900"
                            >
                                Edit Arcade Details
                            </p>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                Update the label and description for your
                                arcade.
                            </p>
                        </div>
                        <form @submit.prevent="handleRename">
                            <div
                                class="space-y-4 py-4 dark:text-gray-100 text-gray-900"
                            >
                                <!-- Label Field -->
                                <div class="space-y-2">
                                    <Label
                                        for="edit-label"
                                        class="text-sm font-medium"
                                    >
                                        Label
                                        <span class="text-red-500">*</span>
                                    </Label>
                                    <Input
                                        id="edit-label"
                                        v-model="renameForm.label"
                                        placeholder="Enter arcade label..."
                                        maxlength="100"
                                        :class="[
                                            'w-full resize-none border-none ring-[1px] ring-gray-800 dark:ring-gray-200 outline-none focus:border-none focus-visible:ring-gray-700 dark:focus-visible:ring-gray-300',
                                            renameErrors.label
                                                ? 'border-red-500'
                                                : '',
                                        ]"
                                    />
                                    <div
                                        class="flex justify-between items-center"
                                    >
                                        <p
                                            v-if="renameErrors.label"
                                            class="text-xs text-red-500"
                                        >
                                            {{ renameErrors.label }}
                                        </p>
                                        <p
                                            class="text-xs text-gray-500 dark:text-gray-200 ml-auto"
                                        >
                                            {{ renameForm.label.length }}/100
                                        </p>
                                    </div>
                                </div>

                                <!-- Description Field -->
                                <div class="space-y-2">
                                    <Label
                                        for="edit-description"
                                        class="text-sm font-medium"
                                    >
                                        Description
                                        <span class="text-red-500">*</span>
                                    </Label>
                                    <Textarea
                                        id="edit-description"
                                        v-model="renameForm.description"
                                        placeholder="Enter arcade description..."
                                        rows="4"
                                        maxlength="500"
                                        :class="[
                                            'w-full resize-none border-none ring-[1px] ring-gray-800 dark:ring-gray-200 outline-none focus:border-none focus-visible:ring-gray-700 dark:focus-visible:ring-gray-300',
                                            renameErrors.description
                                                ? 'border-red-500'
                                                : '',
                                        ]"
                                    />
                                    <div
                                        class="flex justify-between items-center"
                                    >
                                        <p
                                            v-if="renameErrors.description"
                                            class="text-xs text-red-500"
                                        >
                                            {{ renameErrors.description }}
                                        </p>
                                        <p
                                            class="text-xs text-gray-500 dark:text-gray-400 ml-auto"
                                        >
                                            {{
                                                renameForm.description.length
                                            }}/500
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <div class="flex items-center justify-end gap-2">
                                <Button
                                    type="button"
                                    variant="outline"
                                    class="bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200 h-[40px]"
                                    @click="showRenameDialog = false"
                                    :disabled="isRenaming"
                                >
                                    Cancel
                                </Button>
                                <Button
                                    class="bg-gray-900 dark:bg-gray-100 text-gray-100 dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-200 hover:text-gray-100 dark:hover:text-gray-900 h-[40px]"
                                    type="submit"
                                    :disabled="isRenaming"
                                >
                                    <Loader2
                                        v-if="isRenaming"
                                        class="w-4 h-4 mr-2 animate-spin"
                                    />
                                    {{
                                        isRenaming
                                            ? "Saving..."
                                            : "Save Changes"
                                    }}
                                </Button>
                            </div>
                        </form>
                    </div>
                </DialogBox>

                <!-- Delete Confirmation Dialog -->
                <DialogBox
                    name="delete-confirmation"
                    :close-modal="() => (showDeleteDialog = false)"
                    :show="showDeleteDialog"
                >
                    <div class="sm:max-w-[425px]">
                        <div class="space-y-2">
                            <div
                                class="text-xl font-bold dark:text-gray-100 text-gray-900"
                            >
                                Delete Arcade
                            </div>
                            <div
                                class="text-sm text-gray-500 dark:text-gray-400"
                            >
                                Are you sure you want to delete this arcade?
                                This action cannot be undone.
                            </div>
                        </div>
                        <div class="py-4">
                            <div
                                class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4"
                            >
                                <p
                                    class="text-sm text-red-800 dark:text-red-200 font-medium"
                                >
                                    {{ arcade?.label }}
                                </p>
                                <p
                                    class="text-xs text-red-600 dark:text-red-400 mt-1"
                                >
                                    This will permanently delete your arcade and
                                    all its content.
                                </p>
                            </div>
                        </div>
                        <div class="flex items-center justify-end space-x-2">
                            <Button
                                variant="outline"
                                @click="showDeleteDialog = false"
                                class="bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200 h-[40px]"
                                :disabled="isDeleting"
                            >
                                Cancel
                            </Button>
                            <Button
                                variant="destructive"
                                @click="handleDelete"
                                :disabled="isDeleting"
                            >
                                <Loader2
                                    v-if="isDeleting"
                                    class="w-4 h-4 mr-2 animate-spin"
                                />
                                {{
                                    isDeleting ? "Deleting..." : "Delete Arcade"
                                }}
                            </Button>
                        </div>
                    </div>
                </DialogBox>
            </div>
        </OverallLayout>
    </div>
</template>

<style scoped>
/* Fade transition */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
