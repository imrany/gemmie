<script setup lang="ts">
import type { ApiResponse, RawArcade, UserDetails } from "@/types";
import { ref, computed } from "vue";
import { inject } from "vue";
import { onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { toast } from "vue-sonner";
import {
    Loader2,
    AlertCircle,
    ArrowLeft,
    MoreVertical,
    Trash2,
    Edit,
    Download,
    Share2,
    Copy,
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
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { Ref } from "vue";

const route = useRoute();
const router = useRouter();
const arcadeId = route.params.id as string;

const { unsecureApiCall, apiCall, parsedUserDetails } = inject(
    "globalState",
) as {
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

const arcade = ref<RawArcade>();
const isLoading = ref(true);
const error = ref<string | null>(null);
const iframeLoaded = ref(false);

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
            `/arcades/${arcadeId.trim()}`,
            {
                method: "GET",
            },
        );

        if (!res.success) {
            throw new Error(res.message || "Failed to load arcade content");
        }

        if (res.data) {
            arcade.value = res.data;
            console.log(arcade);
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

function goBack() {
    router.push("/arcade");
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
        const res = await apiCall<RawArcade>(`/arcades/${arcadeId}`, {
            method: "PUT",
            body: JSON.stringify({
                label: renameForm.value.label.trim(),
                description: renameForm.value.description.trim(),
            }),
        });

        console.log(res);
        if (!res.success) {
            throw new Error(res.message || "Failed to update arcade");
        }

        toast.success("Arcade updated successfully!");

        // Update local data
        if (arcade.value) {
            arcade.value.label = renameForm.value.label.trim();
            arcade.value.description = renameForm.value.description.trim();
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
        const res = await apiCall<void>(`/arcades/${arcadeId}`, {
            method: "DELETE",
        });

        if (!res.success) {
            throw new Error(res.message || "Failed to delete arcade");
        }

        console.log(res);

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

onMounted(async () => {
    await fetchArcade();
});
</script>

<template>
    <div class="w-screen h-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
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
                        <p class="text-sm text-gray-500 dark:text-gray-400">
                            Rendering preview...
                        </p>
                    </div>
                </div>
            </Transition>

            <!-- Floating Action Button (only for owner) -->
            <div v-if="isOwner" class="fixed bottom-6 right-6 z-20">
                <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                        <Button
                            size="lg"
                            class="h-14 w-14 rounded-full shadow-lg hover:shadow-xl transition-shadow"
                        >
                            <MoreVertical class="w-5 h-5" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-48">
                        <DropdownMenuItem
                            @click="openRenameDialog"
                            class="cursor-pointer"
                        >
                            <Edit class="w-4 h-4 mr-2" />
                            <span>Edit Details</span>
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
            </div>

            <!-- Share Button (for non-owners) -->
            <div v-else class="fixed bottom-6 right-6 z-20">
                <Button
                    @click="handleShare"
                    size="lg"
                    class="h-14 px-6 rounded-full shadow-lg hover:shadow-xl transition-shadow"
                >
                    <Share2 class="w-5 h-5 mr-2" />
                    Share
                </Button>
            </div>

            <!-- Iframe -->
            <iframe
                :srcdoc="arcade.code"
                @load="handleIframeLoad"
                @error="handleIframeError"
                class="w-full h-full border-0 bg-white dark:bg-gray-900"
                sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
                title="Arcade Content Preview"
                referrerpolicy="no-referrer"
            />
        </div>

        <!-- No Content State -->
        <div
            v-else
            class="flex flex-col items-center justify-center h-full gap-4 px-4"
        >
            <AlertCircle class="w-16 h-16 text-gray-400 dark:text-gray-600" />
            <div class="text-center">
                <h2
                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                >
                    No Content Available
                </h2>
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
                    This arcade entry doesn't have any content to display.
                </p>
                <Button
                    @click="goBack"
                    variant="outline"
                    class="inline-flex items-center gap-2"
                >
                    <ArrowLeft class="w-4 h-4" />
                    Back to Arcade
                </Button>
            </div>
        </div>

        <!-- Rename Dialog -->
        <Dialog v-model:open="showRenameDialog">
            <DialogContent class="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Edit Arcade Details</DialogTitle>
                    <DialogDescription>
                        Update the label and description for your arcade.
                    </DialogDescription>
                </DialogHeader>
                <div class="space-y-4 py-4">
                    <!-- Label Field -->
                    <div class="space-y-2">
                        <Label for="edit-label" class="text-sm font-medium">
                            Label <span class="text-red-500">*</span>
                        </Label>
                        <Input
                            id="edit-label"
                            v-model="renameForm.label"
                            placeholder="Enter arcade label..."
                            maxlength="100"
                            :class="[
                                renameErrors.label ? 'border-red-500' : '',
                            ]"
                        />
                        <div class="flex justify-between items-center">
                            <p
                                v-if="renameErrors.label"
                                class="text-xs text-red-500"
                            >
                                {{ renameErrors.label }}
                            </p>
                            <p
                                class="text-xs text-gray-500 dark:text-gray-400 ml-auto"
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
                            Description <span class="text-red-500">*</span>
                        </Label>
                        <Textarea
                            id="edit-description"
                            v-model="renameForm.description"
                            placeholder="Enter arcade description..."
                            rows="4"
                            maxlength="500"
                            :class="[
                                'resize-none',
                                renameErrors.description
                                    ? 'border-red-500'
                                    : '',
                            ]"
                        />
                        <div class="flex justify-between items-center">
                            <p
                                v-if="renameErrors.description"
                                class="text-xs text-red-500"
                            >
                                {{ renameErrors.description }}
                            </p>
                            <p
                                class="text-xs text-gray-500 dark:text-gray-400 ml-auto"
                            >
                                {{ renameForm.description.length }}/500
                            </p>
                        </div>
                    </div>
                </div>
                <DialogFooter>
                    <Button
                        variant="outline"
                        @click="showRenameDialog = false"
                        :disabled="isRenaming"
                    >
                        Cancel
                    </Button>
                    <Button @click="handleRename" :disabled="isRenaming">
                        <Loader2
                            v-if="isRenaming"
                            class="w-4 h-4 mr-2 animate-spin"
                        />
                        {{ isRenaming ? "Saving..." : "Save Changes" }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>

        <!-- Delete Confirmation Dialog -->
        <Dialog v-model:open="showDeleteDialog">
            <DialogContent class="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Delete Arcade</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to delete this arcade? This action
                        cannot be undone.
                    </DialogDescription>
                </DialogHeader>
                <div class="py-4">
                    <div
                        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4"
                    >
                        <p
                            class="text-sm text-red-800 dark:text-red-200 font-medium"
                        >
                            {{ arcade?.label }}
                        </p>
                        <p class="text-xs text-red-600 dark:text-red-400 mt-1">
                            This will permanently delete your arcade and all its
                            content.
                        </p>
                    </div>
                </div>
                <DialogFooter>
                    <Button
                        variant="outline"
                        @click="showDeleteDialog = false"
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
                        {{ isDeleting ? "Deleting..." : "Delete Arcade" }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
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
