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
    MessageSquare,
    ArrowUp,
} from "lucide-vue-next";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Textarea } from "./ui/textarea";
import { Label } from "./ui/label";
import type { Ref } from "vue";
import { inject } from "vue";
import hljs from "highlight.js/lib/common";
import { toast } from "vue-sonner";
import type {
    ApiResponse,
    RawArcade,
    Chat,
    Message,
    UserDetails,
} from "@/types";
import {
    TooltipProvider,
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "./ui/tooltip";
import { generateChatTitle } from "@/lib/globals";
import type { FunctionalComponent } from "vue";
import { useRoute } from "vue-router";
import MarkdownRenderer from "./ui/markdown/MarkdownRenderer.vue";

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
    isDarkMode,
    createNewChat,
    arcade,
    chats,
    onMessageAdded,
} = inject("globalState") as {
    onMessageAdded: (message: Message, id?: string) => void;
    chats: Ref<Chat[]>;
    arcade: Ref<RawArcade>;
    createNewChat: (firstMessage?: string, chatId?: string) => Promise<string>;
    isDarkMode: Ref<boolean>;
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
    showPreviewSidebar: Ref<string | null>;
    previewCode: Ref<string>;
    previewLanguage: Ref<string>;
    closePreview: () => void;
    parsedUserDetails: Ref<UserDetails>;
    apiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
};

const route = useRoute();
const activeTab = ref<"preview" | "code" | "publish" | "chat">(
    metadata?.value ? "code" : "preview",
);
const previousTab = ref<"preview" | "code" | "chat">("preview");
const copied = ref(false);
const sidebarWidth = ref<number>(window.innerWidth * 0.5);
const minWidth = 300;
const maxWidth = window.innerWidth * 0.8;
const isResizing = ref(false);
const isPublishing = ref(false);

// Chat state
const currentChat = computed(() => {
    if (!route.params.id || !chats.value.length) {
        return undefined;
    }
    return chats.value.find((chat) => chat.id === route.params.id);
});

const currentMessages = computed(() => {
    return currentChat.value?.messages || [];
});

const currentMessage = computed(() => {
    return currentMessages.value.find((message) =>
        message.response.includes(previewCode.value),
    );
});

const chatInput = ref("");
const isSendingMessage = ref(false);
const chatContainer = ref<HTMLDivElement | null>(null);

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

const tabs: {
    label: "preview" | "code" | "chat";
    icon: FunctionalComponent;
}[] = [
    ...(route.path.startsWith("/arcade/")
        ? []
        : [{ label: "preview" as "preview", icon: Eye }]),
    { label: "code", icon: Code },
    ...(route.path.startsWith("/chat/")
        ? []
        : [{ label: "chat" as "chat", icon: MessageSquare }]),
];

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

    const tabOrder = ["preview", "code", "chat"];
    const currentIndex = tabOrder.indexOf(activeTab.value);
    const previousIndex = tabOrder.indexOf(previousTab.value);

    if (currentIndex > previousIndex) {
        return "slide-left";
    } else if (currentIndex < previousIndex) {
        return "slide-right";
    }
    return "slide-left";
});

// Watch tab changes to track previous tab
watch(activeTab, async (newVal, oldVal) => {
    if (oldVal === "preview" || oldVal === "code" || oldVal === "chat") {
        previousTab.value = oldVal;
    }
    if (newVal === "chat" && route.params.id) {
        const arcadeChat = await createNewChat(
            arcade.value?.label || "Arcade Chat",
            route.params.id.toString(),
        );
        console.log("Arcade Chat created:", arcadeChat);
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

    if (!publishForm.value.label) {
        publishForm.value.label = generateSmartLabel(previewCode.value);
    }
    activeTab.value = "publish";
};

const backToPreview = () => {
    activeTab.value = "preview";
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
            created_at: new Date().toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                year: "numeric",
            }),
        };

        const response = await apiCall<string>("/arcades", {
            method: "POST",
            body: JSON.stringify(arcadeData),
        });

        if (response.success) {
            let alert = toast.success("Published to Arcade!", {
                duration: 5000,
                description: "Your code is now visible in the Arcade",
            });

            publishForm.value = {
                label: "",
                description: "",
            };
            formErrors.value = {};
            closePreview();
            toast.dismiss(alert);
            alert = toast.loading("Opening Arcade...", {
                cancel: {
                    label: "Cancel",
                    onClick: () => {
                        toast.dismiss(alert);
                    },
                },
            });

            setTimeout(() => {
                window.open(`/arcade/${response.data}`, "_blank");
                toast.dismiss(alert);
            }, 4 * 1000);
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

//  Update arcade code in database
const updateArcadeCode = async (newCode: string): Promise<boolean> => {
    if (!arcade.value?.id) {
        console.warn("No arcade ID available to update");
        return false;
    }

    try {
        const response = await apiCall<RawArcade>(
            `/arcades/${arcade.value.id}`,
            {
                method: "PUT",
                body: JSON.stringify({
                    code: newCode,
                    label: arcade.value.label,
                    description: arcade.value.description,
                }),
            },
        );

        if (response.success) {
            // Update local arcade data
            if (arcade.value) {
                arcade.value.code = newCode;
            }
            return true;
        } else {
            throw new Error(response.message || "Failed to update arcade");
        }
    } catch (error: any) {
        console.error("Update arcade error:", error);
        toast.error("Failed to save code changes", {
            duration: 3000,
            description: error.message,
        });
        return false;
    }
};

const sendChatMessage = async () => {
    if (!chatInput.value.trim() || isSendingMessage.value) return;

    if (!isOnline.value) {
        toast.error("You are offline", {
            duration: 3000,
            description: "Please check your internet connection",
        });
        return;
    }

    const userMessage = chatInput.value.trim();
    const fabricatedPrompt = `You are a code editing assistant. The user has the following code:\n\n\`\`\`${previewLanguage.value}\n${previewCode.value}\n\`\`\`\n\nUser request: ${userMessage}\n\nProvide a brief explanation of the changes you're making, then provide ONLY the full updated code in a single code block. Keep your explanation concise and focused on what changed.`;

    isSendingMessage.value = true;
    const originalInput = chatInput.value;
    chatInput.value = "Thinking...";

    // Find the chat to submit to
    const submissionChatId = arcade.value?.id || route.params.id?.toString();
    const submissionChat = chats.value.find(
        (chat) => chat.id === submissionChatId,
    );

    if (!submissionChat) {
        toast.error("Chat not found", {
            duration: 3000,
            description: "Please refresh and try again",
        });
        isSendingMessage.value = false;
        chatInput.value = originalInput;
        return;
    }

    // Initialize messages array if needed
    if (!Array.isArray(submissionChat.messages)) {
        submissionChat.messages = [];
    }

    // Create temporary message
    const tempMessage: Message = {
        id: `temp_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
        chat_id: submissionChatId!,
        created_at: new Date().toISOString(),
        prompt: userMessage,
        response: "",
        references: [],
        model: "gemini-pro",
    };

    submissionChat.messages.push(tempMessage);
    const tempMessageIndex = submissionChat.messages.length - 1;
    submissionChat.last_message_at = new Date().toISOString();

    // Scroll to bottom after adding temp message
    setTimeout(() => scrollToBottom(), 50);

    // Update chat title if first message
    if (
        submissionChat.messages.length === 1 ||
        submissionChat.title === "New Chat"
    ) {
        submissionChat.title = generateChatTitle(userMessage);
    }

    try {
        const response = await apiCall<{ Response: string; Prompt: string }>(
            `/genai`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(fabricatedPrompt),
            },
        );
        chatInput.value = "";
        if (!response.success) {
            throw new Error(response.message);
        }

        const data = response.data;

        if (!data) {
            throw new Error("No response data received");
        }

        // Update the temporary message with full response
        const updatedMessage: Message = {
            id: `msg_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
            chat_id: submissionChatId!,
            created_at: new Date().toISOString(),
            prompt: userMessage,
            response: data.Response,
            references: [],
            model: "gemini-pro",
        };

        submissionChat.messages[tempMessageIndex] = updatedMessage;
        submissionChat.last_message_at = new Date().toISOString();

        // Extract code from response and update preview
        const codeBlockRegex = /```[\w]*\n([\s\S]*?)```/g;
        const matches = [...updatedMessage.response.matchAll(codeBlockRegex)];

        if (matches.length > 0) {
            const newCode = matches[0][1].trim();
            if (newCode) {
                previewCode.value = newCode;

                // Save the updated code to the arcade
                const updateSuccess = await updateArcadeCode(newCode);

                if (updateSuccess) {
                    toast.success("Code updated!", {
                        duration: 3000,
                        description: "Changes saved to arcade",
                    });
                } else {
                    toast.warning("Code updated locally", {
                        duration: 3000,
                        description: "Failed to sync with arcade",
                    });
                }
            }
        }

        // Trigger onMessageAdded callback
        onMessageAdded(updatedMessage, arcade.value?.id);

        // Scroll to bottom after response
        setTimeout(() => scrollToBottom(), 100);
    } catch (error: any) {
        console.error("Chat error:", error);

        // Remove temporary message on error
        if (submissionChat.messages[tempMessageIndex]?.id === tempMessage.id) {
            submissionChat.messages.splice(tempMessageIndex, 1);
        }

        // Restore input on error
        chatInput.value = originalInput;

        let errorMessage = error.message || "Unknown error";
        let description = "Please try again or check your connection";

        if (error.name === "AbortError" || error.message.includes("abort")) {
            errorMessage = "Request Timeout";
            description =
                "The code generation took too long. Try a simpler request";
        } else if (error.message.includes("Failed to fetch")) {
            errorMessage = "Network Error";
            description = "Please check your internet connection";
        } else if (error.message.includes("timeout")) {
            errorMessage = "Request Timeout";
            description = "The request took too long. Please try again";
        }

        toast.error(`Failed to send message: ${errorMessage}`, {
            duration: 5000,
            description,
        });
    } finally {
        isSendingMessage.value = false;
    }
};

const scrollToBottom = () => {
    if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
};

const generateSmartLabel = (code: string): string => {
    const titleMatch = code.match(/<title>(.*?)<\/title>/i);
    if (titleMatch && titleMatch[1]) {
        return titleMatch[1].slice(0, 100);
    }

    const h1Match = code.match(/<h1[^>]*>(.*?)<\/h1>/i);
    if (h1Match && h1Match[1]) {
        return h1Match[1].replace(/<[^>]*>/g, "").slice(0, 100);
    }

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
        }, 20 * 1000);

        const response = await apiCall<{ Response: string; Prompt: string }>(
            `/genai`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(
                    `Provide a description for ${label} in text, less than 200 characters`,
                ),
                signal: controller.signal,
            },
        );

        if (!response.success) {
            throw new Error(response.message);
        }

        const data = response.data;
        if (!data) {
            throw new Error("No data received");
        }
        isGeneratingDescription.value = false;
        return data?.Response;
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

watch(showPreviewSidebar, (newVal) => {
    if (newVal) {
        if (route.path.startsWith("/arcade/")) {
            activeTab.value = "chat";
            previousTab.value = "code";
        } else {
            activeTab.value = metadata?.value ? "code" : "preview";
            previousTab.value = "preview";
        }

        publishForm.value = {
            label: "",
            description: "",
        };
        formErrors.value = {};
        chatInput.value = "";
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
            v-if="showPreviewSidebar && currentMessage"
            class="fixed inset-0 bg-black/50 z-40 md:hidden"
            @click="closePreview"
        />
    </Transition>
    <div
        v-if="showPreviewSidebar && currentMessage"
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
            v-if="showPreviewSidebar && currentMessage"
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
                    :class="[
                        'flex-shrink-0 border-b border-gray-200 dark:border-gray-800',
                        activeTab === 'publish' ? 'dark:bg-gray-800' : '',
                    ]"
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

                        <!-- Tabs -->
                        <div
                            v-else-if="!metadata"
                            class="flex items-center gap-1.5 bg-gray-200 dark:bg-gray-800 p-1 rounded-lg"
                        >
                            <button
                                v-for="tab in tabs"
                                :key="tab.label"
                                @click="activeTab = tab.label"
                                :class="[
                                    'px-3 py-1.5 text-xs rounded-md transition-all duration-200 inline-flex items-center gap-1.5 font-medium',
                                    activeTab === tab.label
                                        ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                ]"
                            >
                                <component :is="tab.icon" :size="14" />
                                <span class="capitalize">{{ tab.label }}</span>
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
                                    class="text-2xl font-semibold text-gray-800 dark:text-white mb-2"
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
                                            'w-full resize-none border-none ring-[1px] ring-gray-200 dark:ring-gray-800 outline-none focus:border-none focus-visible:ring-gray-300 dark:focus-visible:ring-gray-700',
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
                                            'w-full resize-none border-none ring-[1px] ring-gray-200 dark:ring-gray-800 outline-none focus:border-none focus-visible:ring-gray-300 dark:focus-visible:ring-gray-700',
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
                                    class="h-[35px] px-6 dark:bg-gray-700 bg-gray-200 dark:hover:bg-gray-600 hover:bg-gray-300 dark:text-gray-200 dark:hover:text-gray-300 text-gray-700 hover:text-gray-600"
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

                    <!-- Preview/Code/Chat Tabs (when no metadata) -->
                    <template v-else>
                        <div class="relative h-full w-full overflow-hidden">
                            <Transition :name="transitionName">
                                <!-- Preview Tab -->
                                <div
                                    v-if="
                                        activeTab === 'preview' &&
                                        !route.path.startsWith('/arcade/')
                                    "
                                    class="absolute inset-0"
                                    key="preview"
                                >
                                    <iframe
                                        v-if="previewCode"
                                        :srcdoc="previewCode"
                                        class="w-full h-full border-0 bg-gray-100"
                                        title="HTML Preview"
                                        sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox allow-top-navigation-by-user-activation allow-modals allow-forms allow-pointer-lock allow-downloads"
                                        referrerpolicy="no-referrer"
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
                                    v-else-if="activeTab === 'code'"
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

                                <!-- Chat Tab -->
                                <div
                                    v-else-if="
                                        activeTab === 'chat' &&
                                        route.path.startsWith('/arcade')
                                    "
                                    class="absolute inset-0 flex flex-col bg-white dark:bg-gray-900"
                                    key="chat"
                                >
                                    <!-- Chat Messages -->
                                    <div
                                        ref="chatContainer"
                                        class="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-4"
                                    >
                                        <!-- Welcome Message -->
                                        <div
                                            v-if="currentMessages.length === 0"
                                            class="flex items-center justify-center h-full text-center"
                                        >
                                            <div class="max-w-md space-y-3">
                                                <div
                                                    class="mx-auto flex items-center justify-center"
                                                >
                                                    <img
                                                        :src="
                                                            parsedUserDetails?.theme ===
                                                                'dark' ||
                                                            (parsedUserDetails?.theme ===
                                                                'system' &&
                                                                isDarkMode)
                                                                ? '/logo-light.svg'
                                                                : '/logo.svg'
                                                        "
                                                        alt="Gemmie Logo"
                                                        class="w-[60px] h-[60px] rounded-md"
                                                    />
                                                </div>
                                                <h3
                                                    class="text-lg font-semibold text-gray-900 dark:text-white"
                                                >
                                                    Code Editing Assistant
                                                </h3>
                                                <p
                                                    class="text-sm text-gray-600 dark:text-gray-400"
                                                >
                                                    Ask me to modify, improve,
                                                    or explain your code. I'll
                                                    help you make changes and
                                                    update the preview
                                                    automatically.
                                                </p>
                                                <div
                                                    class="text-xs text-gray-500 dark:text-gray-500 space-y-1"
                                                >
                                                    <p>Try asking:</p>
                                                    <ul
                                                        class="text-left space-y-1 ml-4"
                                                    >
                                                        <li>
                                                            "Add a dark mode
                                                            toggle"
                                                        </li>
                                                        <li>
                                                            "Change the color
                                                            scheme to blue"
                                                        </li>
                                                        <li>
                                                            "Add hover effects
                                                            to buttons"
                                                        </li>
                                                        <li>
                                                            "Make it responsive"
                                                        </li>
                                                    </ul>
                                                </div>
                                            </div>
                                        </div>

                                        <!-- Chat Messages List -->
                                        <div
                                            v-for="msg in currentMessages"
                                            :key="msg.id"
                                            :class="[
                                                'flex flex-col gap-1',
                                                'justify-start',
                                            ]"
                                        >
                                            <div :class="['max-w-[85%]']">
                                                <div
                                                    class="flex items-start justify-center gap-2 font-medium bg-gray-100 dark:bg-gray-800 text-black dark:text-gray-100 px-4 rounded-2xl prose prose-sm dark:prose-invert chat-bubble w-fit max-w-full"
                                                >
                                                    <!-- Avatar container -->
                                                    <div
                                                        class="flex-shrink-0 py-3"
                                                    >
                                                        <div
                                                            class="flex items-center justify-center w-7 h-7 text-gray-100 dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full text-sm font-semibold"
                                                        >
                                                            {{
                                                                parsedUserDetails.username
                                                                    .toUpperCase()
                                                                    .slice(0, 2)
                                                            }}
                                                        </div>
                                                    </div>
                                                    <div class="flex-1 min-w-0">
                                                        <MarkdownRenderer
                                                            class="break-words text-sm whitespace-pre-wrap overflow-x-hidden"
                                                            :content="
                                                                msg.prompt || ''
                                                            "
                                                        />
                                                    </div>
                                                </div>
                                            </div>
                                            <div :class="['max-w-[85%]']">
                                                <MarkdownRenderer
                                                    class="break-words text-xs whitespace-pre-wrap overflow-x-hidden"
                                                    :content="
                                                        msg.response.replace(
                                                            /```[\w]*\n[\s\S]*?```/g,
                                                            '',
                                                        ) || ''
                                                    "
                                                />
                                            </div>
                                        </div>

                                        <!-- Typing Indicator -->
                                        <div
                                            v-if="isSendingMessage"
                                            class="flex justify-start"
                                        >
                                            <div
                                                class="bg-gray-100 dark:bg-gray-800 rounded-lg px-4 py-3"
                                            >
                                                <div class="flex gap-1">
                                                    <div
                                                        class="w-2 h-2 rounded-full bg-gray-400 dark:bg-gray-600 animate-bounce"
                                                        style="
                                                            animation-delay: 0ms;
                                                        "
                                                    ></div>
                                                    <div
                                                        class="w-2 h-2 rounded-full bg-gray-400 dark:bg-gray-600 animate-bounce"
                                                        style="
                                                            animation-delay: 150ms;
                                                        "
                                                    ></div>
                                                    <div
                                                        class="w-2 h-2 rounded-full bg-gray-400 dark:bg-gray-600 animate-bounce"
                                                        style="
                                                            animation-delay: 300ms;
                                                        "
                                                    ></div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Chat Input -->
                                    <div
                                        class="flex-shrink-0 border-t border-gray-200 dark:border-gray-800 p-4"
                                    >
                                        <form
                                            @submit.prevent="sendChatMessage"
                                            class="flex gap-2"
                                        >
                                            <Input
                                                id="prompt"
                                                v-model="chatInput"
                                                placeholder="Ask me to edit your code..."
                                                :disabled="isSendingMessage"
                                                :class="[
                                                    'flex-1',
                                                    'w-full resize-none border-none ring-[1px] ring-gray-200 dark:ring-gray-800 outline-none focus:border-none focus-visible:ring-gray-300 dark:focus-visible:ring-gray-700',
                                                ]"
                                                @keydown.enter.prevent="
                                                    sendChatMessage
                                                "
                                            />
                                            <Button
                                                type="submit"
                                                :disabled="
                                                    !chatInput.trim() ||
                                                    isSendingMessage ||
                                                    !isOnline
                                                "
                                                size="icon"
                                                class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-white dark:text-white bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-800 disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-400 flex-shrink-0 shadow-sm"
                                            >
                                                <LoaderCircle
                                                    v-if="isSendingMessage"
                                                    class="w-4 h-4 sm:w-5 sm:h-5 animate-spin"
                                                />
                                                <ArrowUp
                                                    v-else
                                                    class="w-4 h-4 sm:w-5 sm:h-5"
                                                />
                                            </Button>
                                        </form>
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

                        <!-- For chat tab -->
                        <div
                            v-else-if="activeTab === 'chat'"
                            class="flex items-center gap-2 text-gray-600 dark:text-gray-400"
                        >
                            <MessageSquare class="w-3.5 h-3.5" />
                            <span>{{ currentMessages.length }} messages</span>
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

/* Slide transitions for tab content */
.slide-left-enter-active,
.slide-left-leave-active,
.slide-right-enter-active,
.slide-right-leave-active {
    transition:
        transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
        opacity 0.3s ease;
}

/* Sliding from left (moving forward) */
.slide-left-enter-from {
    transform: translateX(100%);
    opacity: 0;
}

.slide-left-leave-to {
    transform: translateX(-100%);
    opacity: 0;
}

/* Sliding from right (moving backward) */
.slide-right-enter-from {
    transform: translateX(-100%);
    opacity: 0;
}

.slide-right-leave-to {
    transform: translateX(100%);
    opacity: 0;
}

/* Fade transition */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
