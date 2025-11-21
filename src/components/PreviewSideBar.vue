<script setup lang="ts">
import { ref, watch, computed } from "vue";
import {
    X,
    Code,
    Eye,
    Check,
    Copy,
    Earth,
    ArrowLeft,
    LoaderCircle,
} from "lucide-vue-next";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Textarea } from "./ui/textarea";
import { Label } from "./ui/label";
import type { Ref } from "vue";
import { inject } from "vue";
import hljs from "highlight.js/lib/common";
import { toast } from "vue-sonner";
import {
    TooltipProvider,
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "./ui/tooltip";
import { useRouter } from "vue-router";
import { WRAPPER_URL } from "@/lib/globals";
import type { RawArcade } from "@/types";

const {
    isOnline,
    screenWidth,
    showPreviewSidebar,
    previewCode,
    previewLanguage,
    closePreview,
    metadata,
    isCollapsed,
    parsedUserDetails,
    apiCall,
} = inject("globalState") as {
    isOnline: Ref<boolean>;
    isCollapsed: Ref<boolean>;
    metadata: Ref<
        | {
              fileSize: string;
              wordCount: number;
              charCount: number;
          }
        | undefined
    >;
    screenWidth: Ref<number>;
    showPreviewSidebar: Ref<boolean>;
    previewCode: Ref<string>;
    previewLanguage: Ref<string>;
    closePreview: () => void;
    parsedUserDetails: Ref<any>;
    apiCall: (endpoint: string, options: RequestInit) => Promise<any>;
};

const router = useRouter();
const activeTab = ref<"preview" | "code" | "publish">(
    metadata?.value ? "code" : "preview",
);
const previousTab = ref<"preview" | "code">("preview");
const copied = ref(false);
const sidebarWidth = ref<number>(window.innerWidth * 0.5);
const minWidth = 300;
const maxWidth = window.innerWidth * 0.8;
const isResizing = ref(false);
const isPublishing = ref(false);

// Publish form data
const publishForm = ref<{
    label: string;
    description: string;
}>({
    label: "",
    description: "",
});

const formErrors = ref<{
    label?: string;
    description?: string;
}>({});

const isFormValid = computed(() => {
    return (
        publishForm.value.label.trim().length > 0 &&
        publishForm.value.label.trim().length <= 100 &&
        publishForm.value.description.trim().length > 0 &&
        publishForm.value.description.trim().length <= 500
    );
});

// Compute transition direction
const transitionName = computed(() => {
    if (activeTab.value === "publish") return "fade";

    // If switching between preview and code
    if (previousTab.value === "preview" && activeTab.value === "code") {
        return "slide-left";
    } else if (previousTab.value === "code" && activeTab.value === "preview") {
        return "slide-right";
    }
    return "slide-left";
});

// Watch tab changes to track previous tab
watch(activeTab, (newVal, oldVal) => {
    if (oldVal === "preview" || oldVal === "code") {
        previousTab.value = oldVal;
    }
});

const copyToClipboard = async () => {
    try {
        await navigator.clipboard.writeText(previewCode.value);
        copied.value = true;
        setTimeout(() => {
            copied.value = false;
        }, 2000);
    } catch (err) {
        console.error("Failed to copy:", err);
    }
};

const startResize = (e: MouseEvent) => {
    console.log("startResize: " + e);
    isResizing.value = true;
    document.addEventListener("mousemove", resize);
    document.addEventListener("mouseup", stopResize);
};

const resize = (e: MouseEvent) => {
    if (!isResizing.value) return;
    let newWidth = window.innerWidth - e.clientX;
    if (newWidth > minWidth && newWidth < maxWidth) {
        sidebarWidth.value = newWidth;
    }
};

const stopResize = () => {
    isResizing.value = false;
    document.removeEventListener("mousemove", resize);
    document.removeEventListener("mouseup", stopResize);
};

const openPublishTab = async () => {
    if (!isOnline.value) {
        toast.error("You are offline", {
            duration: 3000,
            description: "Please check your internet connection",
        });
        return;
    }

    // Generate a smart label and description from code if empty
    if (!publishForm.value.label) {
        publishForm.value.label = generateSmartLabel(previewCode.value);
    }
    activeTab.value = "publish";
};

const backToPreview = () => {
    activeTab.value = "preview";
    // Clear form errors when going back
    formErrors.value = {};
};

const validateForm = (): boolean => {
    formErrors.value = {};

    if (!publishForm.value.label.trim()) {
        formErrors.value.label = "Label is required";
        return false;
    }

    if (publishForm.value.label.trim().length > 100) {
        formErrors.value.label = "Label must be 100 characters or less";
        return false;
    }

    if (!publishForm.value.description.trim()) {
        formErrors.value.description = "Description is required";
        return false;
    }

    if (publishForm.value.description.trim().length > 500) {
        formErrors.value.description =
            "Description must be 500 characters or less";
        return false;
    }

    return true;
};

const publishToArcade = async () => {
    if (!validateForm()) {
        toast.error("Please fix the form errors", {
            duration: 3000,
        });
        return;
    }

    if (!parsedUserDetails.value?.userId) {
        toast.error("Authentication required", {
            duration: 3000,
            description: "Please log in to publish",
        });
        return;
    }

    isPublishing.value = true;

    try {
        const arcadeData: RawArcade = {
            user_id: parsedUserDetails.value.userId,
            code: previewCode.value,
            label: publishForm.value.label.trim(),
            description: publishForm.value.description.trim(),
            code_type: previewLanguage.value,
            created_at: new Date(),
        };

        const response = await apiCall("/arcades", {
            method: "POST",
            body: JSON.stringify(arcadeData),
        });

        if (response.success) {
            toast.success("Published to Arcade!", {
                duration: 5000,
                description: "Your code is now visible in the Arcade",
            });

            // Reset form and close
            publishForm.value = {
                label: "",
                description: "",
            };
            formErrors.value = {};
            closePreview();
            router.push("/arcade");
        } else {
            throw new Error(response.message || "Failed to publish");
        }
    } catch (error: any) {
        console.error("Publish error:", error);
        toast.error("Failed to publish", {
            duration: 5000,
            description: error.message || "Please try again later",
        });
    } finally {
        isPublishing.value = false;
    }
};

// Helper function to generate a smart label from code
const generateSmartLabel = (code: string): string => {
    // Try to extract title from HTML
    const titleMatch = code.match(/<title>(.*?)<\/title>/i);
    if (titleMatch && titleMatch[1]) {
        return titleMatch[1].slice(0, 100);
    }

    // Try to extract h1
    const h1Match = code.match(/<h1[^>]*>(.*?)<\/h1>/i);
    if (h1Match && h1Match[1]) {
        return h1Match[1].replace(/<[^>]*>/g, "").slice(0, 100);
    }

    // Fallback to generic name with timestamp
    return `Code Preview - ${new Date().toLocaleDateString()}`;
};

const isGeneratingDescription = ref(false);
const generateSmartDescription = async (label: string): Promise<string> => {
    let timeoutId: number | undefined;
    try {
        isGeneratingDescription.value = true;
        publishForm.value.description = "";
        const controller = new AbortController();
        timeoutId = window.setTimeout(() => {
            controller.abort();
            isGeneratingDescription.value = false;
            toast.error("Description generation timed out.", {
                duration: 3000,
            });
        }, 8 * 1000);

        const response = await fetch(`${WRAPPER_URL}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(
                `Provide a description for ${label} in text, less than 200 characters`,
            ),
            signal: controller.signal,
        });

        if (!response.ok) {
            throw new Error(response.statusText);
        }

        const data = await response.json();
        isGeneratingDescription.value = false;
        return data.response;
    } catch (error: any) {
        isGeneratingDescription.value = false;
        if (error.name === "AbortError") {
            console.log("Fetch aborted");
        } else {
            console.error(error);
        }
        return "";
    } finally {
        clearTimeout(timeoutId);
    }
};

watch(
    isCollapsed,
    (newValue) => {
        if (!newValue) {
            sidebarWidth.value = window.innerWidth * 0.4;
            return;
        }
        sidebarWidth.value = window.innerWidth * 0.5;
    },
    {
        immediate: true,
    },
);

// Reset to preview tab when sidebar opens (unless metadata exists)
watch(showPreviewSidebar, (newVal) => {
    if (newVal) {
        activeTab.value = metadata?.value ? "code" : "preview";
        previousTab.value = "preview";
        // Reset publish form
        publishForm.value = {
            label: "",
            description: "",
        };
        formErrors.value = {};
    }
});

watch(
    screenWidth,
    (newVal) => {
        if (newVal > 1120) {
            sidebarWidth.value = window.innerWidth * 0.5;
            return;
        }

        sidebarWidth.value = window.innerWidth * 0.4;
    },
    {
        immediate: true,
    },
);

watch(
    () => publishForm.value.label,
    async (newLabel) => {
        if (newLabel) {
            const smartDescription = await generateSmartDescription(newLabel);
            smartDescription !== "" &&
                (publishForm.value.description = smartDescription);
        }
    },
    {
        immediate: false,
    },
);
</script>

<template>
    <!-- Backdrop for mobile -->
    <Transition
        name="backdrop"
        @after-leave="activeTab = metadata ? 'code' : 'preview'"
    >
        <div
            v-if="showPreviewSidebar"
            class="fixed inset-0 bg-black/50 z-40 md:hidden"
            @click="closePreview"
        />
    </Transition>
    <div
        v-if="showPreviewSidebar"
        class="max-md:hidden group w-2 relative h-full cursor-col-resize -mr-1 z-30 grid place-items-center hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        @mousedown="startResize"
    >
        <div
            class="absolute top-0 bottom-0 right-1 w-[0.5px] bg-gray-300 dark:bg-gray-600 transition-all group-hover:bg-gray-400 dark:group-hover:bg-gray-500 group-hover:w-px group-hover:translate-x-[0.5px]"
        ></div>
        <div
            class="h-6 w-2 relative rounded-full border-[0.5px] border-gray-400 dark:border-gray-500 bg-gray-200 dark:bg-gray-700 shadow transition duration-200 group-hover:border-gray-700 dark:group-hover:border-gray-400 cursor-col-resize"
        ></div>
    </div>
    <!-- Sidebar -->
    <Transition name="slide">
        <div
            v-if="showPreviewSidebar"
            class="fixed top-0 right-0 bottom-0 md:relative md:z-20 w-full max-w-full z-50 flex-shrink-0 md:flex md:items-stretch md:justify-stretch"
            :style="{
                width: screenWidth > 720 ? sidebarWidth + 'px' : '100vw',
            }"
        >
            <!-- Sidebar Content -->
            <div
                class="relative flex flex-col h-full w-full bg-gray-100 dark:bg-gray-900 overflow-hidden shadow-2xl md:shadow-none md:border-l md:border-gray-200 md:dark:border-gray-800"
            >
                <!-- Header -->
                <div
                    class="flex-shrink-0 border-b border-gray-200 dark:border-gray-800"
                >
                    <div class="flex items-center justify-between px-4 py-3">
                        <!-- Back Button (only in publish tab) -->
                        <Button
                            v-if="activeTab === 'publish'"
                            size="sm"
                            variant="ghost"
                            class="h-8 px-2 hover:bg-gray-200 dark:hover:bg-gray-800"
                            @click="backToPreview"
                        >
                            <ArrowLeft class="w-4 h-4 mr-1" />
                            <span class="text-xs">Back</span>
                        </Button>

                        <!-- Tabs (Only show if no metadata and not in publish mode) -->
                        <div
                            v-else-if="!metadata"
                            class="flex items-center gap-1.5 bg-gray-200 dark:bg-gray-800 p-1 rounded-lg"
                        >
                            <button
                                @click="activeTab = 'preview'"
                                :class="[
                                    'px-3 py-1.5 text-xs rounded-md transition-all duration-200 inline-flex items-center gap-1.5 font-medium',
                                    activeTab === 'preview'
                                        ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                ]"
                            >
                                <Eye :size="14" />
                                <span>Preview</span>
                            </button>
                            <button
                                @click="activeTab = 'code'"
                                :class="[
                                    'px-3 py-1.5 text-xs rounded-md transition-all duration-200 inline-flex items-center gap-1.5 font-medium',
                                    activeTab === 'code'
                                        ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                ]"
                            >
                                <Code :size="14" />
                                <span>Code</span>
                            </button>
                        </div>

                        <!-- Title for pasted content -->
                        <div
                            v-else
                            class="flex font-light text-xs items-center gap-2 text-gray-800 dark:text-gray-200"
                        >
                            <p>{{ metadata.fileSize }}</p>
                            •
                            <p>{{ metadata.wordCount }} words</p>
                            •
                            <p>{{ metadata.charCount }} characters</p>
                        </div>

                        <!-- Close Button -->
                        <div class="flex justify-end gap-2">
                            <TooltipProvider>
                                <Tooltip>
                                    <TooltipTrigger as-child>
                                        <Button
                                            v-if="
                                                !metadata &&
                                                activeTab === 'preview'
                                            "
                                            size="sm"
                                            @click="openPublishTab"
                                            :class="[
                                                'px-3 py-1.5 h-8 text-xs font-medium rounded-md transition-all ease-in-out duration-200 inline-flex items-center gap-1.5',
                                                'bg-gray-700 dark:bg-gray-200 hover:bg-gray-600 dark:hover:bg-gray-300 text-gray-200 dark:text-gray-700 shadow-sm',
                                            ]"
                                            :disabled="!isOnline"
                                        >
                                            <Earth class="w-4 h-4" />
                                            Publish
                                        </Button>
                                    </TooltipTrigger>
                                    <TooltipContent
                                        side="left"
                                        :avoid-collisions="true"
                                    >
                                        <p>
                                            This will make it visible to others
                                            in Arcade.
                                        </p>
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>

                            <Button
                                v-if="activeTab !== 'publish'"
                                size="sm"
                                variant="ghost"
                                class="h-8 w-8 p-0 hover:bg-gray-200 dark:hover:bg-gray-800"
                                @click="closePreview"
                                title="Close preview"
                            >
                                <X class="w-4 h-4" />
                            </Button>
                        </div>
                    </div>
                </div>

                <!-- Content Area -->
                <div class="flex-1 overflow-hidden relative">
                    <!-- Publish Form Tab -->
                    <div
                        v-if="activeTab === 'publish'"
                        class="absolute inset-0 overflow-auto transition duration-300 ease-in-out custom-scrollbar bg-white dark:bg-gray-900 p-6"
                    >
                        <div class="max-w-2xl mx-auto space-y-6">
                            <div>
                                <h2
                                    class="text-2xl font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    Publish to Arcade
                                </h2>
                                <p
                                    class="text-sm text-gray-600 dark:text-gray-400"
                                >
                                    Share your creation with the community
                                </p>
                            </div>

                            <div class="space-y-4">
                                <!-- Label Field -->
                                <div class="space-y-2">
                                    <Label
                                        for="arcade-label"
                                        class="text-sm font-medium"
                                    >
                                        Label
                                        <span class="text-red-500">*</span>
                                    </Label>
                                    <Input
                                        id="arcade-label"
                                        v-model="publishForm.label"
                                        placeholder="Give your code a catchy title..."
                                        maxlength="100"
                                        :class="[
                                            'w-full',
                                            formErrors.label
                                                ? 'border-red-500'
                                                : '',
                                        ]"
                                    />
                                    <div
                                        class="flex justify-between items-center"
                                    >
                                        <p
                                            v-if="formErrors.label"
                                            class="text-xs text-red-500"
                                        >
                                            {{ formErrors.label }}
                                        </p>
                                        <p
                                            class="text-xs text-gray-500 dark:text-gray-400 ml-auto"
                                        >
                                            {{ publishForm.label.length }}/100
                                        </p>
                                    </div>
                                </div>

                                <!-- Description Field -->
                                <div class="space-y-2">
                                    <Label
                                        for="arcade-description"
                                        class="text-sm font-medium"
                                    >
                                        Description
                                        <span class="text-red-500">*</span>
                                    </Label>
                                    <Textarea
                                        id="arcade-description"
                                        v-model="publishForm.description"
                                        :placeholder="
                                            isGeneratingDescription
                                                ? 'Generating description...'
                                                : 'Describe what your code does, how it works, or what makes it special...'
                                        "
                                        :rows="isGeneratingDescription ? 2 : 4"
                                        :disabled="isGeneratingDescription"
                                        maxlength="500"
                                        :class="[
                                            'w-full resize-none',
                                            formErrors.description
                                                ? 'border-red-500'
                                                : '',
                                            isGeneratingDescription
                                                ? 'animate-pulse'
                                                : '',
                                        ]"
                                    />
                                    <div
                                        class="flex justify-between items-center"
                                    >
                                        <p
                                            v-if="formErrors.description"
                                            class="text-xs text-red-500"
                                        >
                                            {{ formErrors.description }}
                                        </p>
                                        <p
                                            class="text-xs text-gray-500 dark:text-gray-400 ml-auto"
                                        >
                                            {{
                                                publishForm.description.length
                                            }}/500
                                        </p>
                                    </div>
                                </div>

                                <!-- Code Type Info -->
                                <div
                                    class="py-2 px-4 bg-gray-50 dark:bg-gray-800 rounded-lg"
                                >
                                    <div
                                        class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"
                                    >
                                        <Code class="w-4 h-4" />
                                        <span
                                            class="font-medium text-sm text-gray-700 dark:text-gray-300"
                                            >Code Type:</span
                                        >
                                        <span
                                            class="px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs font-mono"
                                        >
                                            {{ previewLanguage }}
                                        </span>
                                    </div>
                                </div>
                            </div>

                            <!-- Action Buttons -->
                            <div class="flex gap-3 pt-4">
                                <Button
                                    @click="publishToArcade"
                                    :disabled="!isFormValid || isPublishing"
                                    class="flex-1 h-[35px] bg-gray-700 text-gray-200 dark:bg-gray-200 dark:text-gray-700 hover:bg-gray-600 hover:text-gray-300 dark:hover:bg-gray-600 dark:hover:text-gray-300"
                                >
                                    <LoaderCircle
                                        v-if="isPublishing"
                                        class="w-4 h-4 mr-2 animate-spin"
                                    />
                                    <Earth v-else class="w-4 h-4 mr-2" />
                                    {{
                                        isPublishing
                                            ? "Publishing..."
                                            : "Publish to Arcade"
                                    }}
                                </Button>
                                <Button
                                    @click="backToPreview"
                                    variant="outline"
                                    :disabled="isPublishing"
                                    class="h-[35px] px-6"
                                >
                                    Cancel
                                </Button>
                            </div>

                            <!-- Info Text -->
                            <p
                                class="text-xs text-gray-500 dark:text-gray-400 text-center"
                            >
                                By publishing, you agree to share this code
                                publicly in the Arcade.
                            </p>
                        </div>
                    </div>

                    <!-- Pasted Content View (when metadata exists) -->
                    <div
                        v-else-if="metadata"
                        class="absolute inset-0 overflow-auto w-full bg-gray-800 custom-scrollbar"
                    >
                        <div class="relative min-h-full w-full">
                            <!-- Content Display -->
                            <pre class="px-4 py-4 text-sm">
                                <code
                                    :class="`language-${previewLanguage} leading-relaxed text-gray-300 break-words whitespace-pre-wrap`"
                                    v-html="previewCode ? hljs.highlight(previewCode, { language: previewLanguage }).value : 'No code available'"
                                ></code>
                            </pre>
                        </div>
                    </div>

                    <!-- Preview/Code Tabs (when no metadata) -->
                    <template v-else>
                        <div class="relative h-full w-full overflow-hidden">
                            <Transition :name="transitionName">
                                <!-- Preview Tab -->
                                <div
                                    v-if="activeTab === 'preview'"
                                    class="absolute inset-0"
                                    key="preview"
                                >
                                    <iframe
                                        v-if="previewCode"
                                        :srcdoc="previewCode"
                                        class="w-full h-full border-0 bg-gray-100"
                                        sandbox="allow-scripts allow-forms allow-modals allow-popups allow-same-origin"
                                        title="HTML Preview"
                                    />
                                    <div
                                        v-else
                                        class="flex items-center bg-gray-100 dark:bg-gray-800 justify-center h-full text-gray-500 dark:text-gray-400"
                                    >
                                        <div class="text-center">
                                            <Eye
                                                class="w-12 h-12 mx-auto mb-4 opacity-30"
                                            />
                                            <p class="text-sm font-medium">
                                                No preview available
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                <!-- Code Tab -->
                                <div
                                    v-else
                                    class="absolute inset-0 overflow-auto w-full bg-gray-800 custom-scrollbar"
                                    key="code"
                                >
                                    <div class="relative min-h-full w-full">
                                        <!-- Copy Button -->
                                        <div
                                            class="sticky top-0 z-10 flex justify-end p-3"
                                        >
                                            <button
                                                @click="copyToClipboard"
                                                class="inline-flex items-center gap-2 bg-gray-900 dark:bg-gray-200 hover:bg-gray-700 text-white dark:text-gray-800 px-3 py-1.5 rounded-md text-xs font-medium transition-all duration-200 shadow-lg hover:shadow-xl"
                                            >
                                                <Check
                                                    v-if="copied"
                                                    :size="14"
                                                    class="text-green-400 dark:text-green-600"
                                                />
                                                <Copy v-else :size="14" />
                                                <span>{{
                                                    copied
                                                        ? "Copied!"
                                                        : "Copy code"
                                                }}</span>
                                            </button>
                                        </div>

                                        <!-- Code Display -->
                                        <pre class="px-4 pb-4 text-sm"><code
                                            :class="`language-${previewLanguage} leading-relaxed text-gray-300 break-words whitespace-pre-wrap`"
                                            v-html="previewCode ? hljs.highlight(previewCode, { language: previewLanguage }).value : 'No code available'"
                                        ></code></pre>
                                    </div>
                                </div>
                            </Transition>
                        </div>
                    </template>
                </div>

                <!-- Footer -->
                <div
                    class="flex-shrink-0 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 px-4 py-2.5"
                >
                    <div class="flex items-center justify-between text-xs">
                        <!-- For pasted content -->
                        <div
                            v-if="metadata"
                            class="flex items-center gap-4 text-gray-600 dark:text-gray-400"
                        >
                            <span>{{ metadata.wordCount }} words</span>
                            <span>{{ metadata.charCount }} characters</span>
                        </div>

                        <!-- For publish tab -->
                        <div
                            v-else-if="activeTab === 'publish'"
                            class="flex items-center gap-2 text-gray-600 dark:text-gray-400"
                        >
                            <Earth class="w-3.5 h-3.5" />
                            <span>Publishing to Arcade</span>
                        </div>

                        <!-- For code preview -->
                        <template v-else>
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-1.5 h-1.5 rounded-full transition-all duration-300"
                                    :class="
                                        activeTab === 'preview'
                                            ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]'
                                            : 'bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.6)]'
                                    "
                                ></div>
                                <span
                                    class="font-medium text-gray-700 dark:text-gray-300"
                                >
                                    {{
                                        activeTab === "preview"
                                            ? "Live Preview"
                                            : "Source Code"
                                    }}
                                </span>
                            </div>
                            <span
                                class="text-gray-500 dark:text-gray-500 hidden sm:inline"
                            >
                                {{
                                    activeTab === "preview"
                                        ? "Interactive demo"
                                        : `${previewCode.split("\n").length} lines`
                                }}
                            </span>
                        </template>
                    </div>
                </div>
            </div>
        </div>
    </Transition>
</template>

<style scoped>
/* Slide transition for sidebar */
.slide-enter-active,
.slide-leave-active {
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-enter-from {
    transform: translateX(100%);
}

.slide-leave-to {
    transform: translateX(-100%);
}

@media (min-width: 768px) {
    .slide-enter-from,
    .slide-leave-to {
        transform: translateX(0);
        opacity: 0;
    }

    .slide-enter-active,
    .slide-leave-active {
        transition: opacity 0.4s ease-in-out;
    }
}

/* Backdrop transition */
.backdrop-enter-active,
.backdrop-leave-active {
    transition: opacity 0.3s ease;
}

.backdrop-enter-from,
.backdrop-leave-to {
    opacity: 0;
}

/* Fade transition for tab content */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
