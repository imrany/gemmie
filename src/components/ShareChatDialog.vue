<script lang="ts" setup>
import type { Chat } from "@/types";
import { Check, Copy, Globe, Lock } from "lucide-vue-next";
import { ref, watch } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import DialogBox from "./Dialog/DialogBox.vue";

interface Props {
    show: boolean;
    currentChat?: Chat;
    isLoading?: boolean;
}

interface Emits {
    (e: "close"): void;
    (e: "share", isPrivate: boolean): void;
    (e: "copy", link: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const shareLink = ref<string | null>(null);
const isCopied = ref(false);

function handleShare(isPrivate: boolean) {
    emit("share", isPrivate);
}

function handleCopy() {
    if (shareLink.value) {
        emit("copy", shareLink.value);
        isCopied.value = true;
        setTimeout(() => {
            isCopied.value = false;
        }, 2000);
    }
}

watch(
    () => props.currentChat?.is_private,
    (newValue) => {
        if (!newValue && props.currentChat?.id) {
            shareLink.value = `${window.location.origin}/chat/${props.currentChat.id}`;
        }
    },
    { immediate: true },
);
</script>

<template>
    <DialogBox name="Share Chat" :show="show" :closeModal="() => emit('close')">
        <div class="flex flex-col gap-4 sm:gap-5 p-1">
            <!-- Header -->
            <div class="space-y-1">
                <h2
                    class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100"
                >
                    {{ currentChat?.is_private ? "Share chat" : "Chat shared" }}
                </h2>
                <p
                    class="text-sm sm:text-base text-gray-600 dark:text-gray-400"
                >
                    {{
                        currentChat?.is_private
                            ? "Share this chat with your friends!"
                            : "Anyone who has this link will be able to view this."
                    }}
                </p>
            </div>

            <!-- Privacy Options -->
            <div
                class="rounded-lg border border-gray-300 dark:border-gray-700 overflow-hidden"
            >
                <!-- Private Option -->
                <button
                    @click="handleShare(true)"
                    :disabled="isLoading"
                    class="flex w-full gap-3 sm:gap-4 py-3 px-4 sm:px-6 items-center border-b border-gray-300 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <Lock
                        class="w-5 h-5 flex-shrink-0 text-gray-700 dark:text-gray-300"
                    />
                    <div
                        class="flex items-start flex-col gap-0.5 sm:gap-1 text-left flex-1 min-w-0"
                    >
                        <p
                            class="font-semibold text-sm sm:text-base text-gray-900 dark:text-gray-100"
                        >
                            Private
                        </p>
                        <p
                            class="text-xs sm:text-sm text-gray-600 dark:text-gray-400"
                        >
                            Only you can see this chat.
                        </p>
                    </div>
                    <Check
                        class="ml-auto w-5 h-5 flex-shrink-0 text-green-500"
                        v-if="currentChat?.is_private"
                    />
                </button>

                <!-- Public Option -->
                <button
                    @click="handleShare(false)"
                    :disabled="isLoading"
                    class="flex w-full gap-3 sm:gap-4 py-3 px-4 sm:px-6 items-center hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <Globe
                        class="w-5 h-5 flex-shrink-0 text-gray-700 dark:text-gray-300"
                    />
                    <div
                        class="flex items-start flex-col gap-0.5 sm:gap-1 text-left flex-1 min-w-0"
                    >
                        <p
                            class="font-semibold text-sm sm:text-base text-gray-900 dark:text-gray-100"
                        >
                            Public
                        </p>
                        <p
                            class="text-xs sm:text-sm text-gray-600 dark:text-gray-400"
                        >
                            Anyone can see this chat.
                        </p>
                    </div>
                    <Check
                        class="ml-auto w-5 h-5 flex-shrink-0 text-green-500"
                        v-if="!currentChat?.is_private"
                    />
                </button>
            </div>

            <!-- Share Link Section -->
            <div
                v-if="!currentChat?.is_private"
                class="flex flex-col sm:flex-row gap-2 items-stretch sm:items-center"
            >
                <Input
                    type="text"
                    readonly
                    :defaultValue="shareLink || ''"
                    :value="shareLink || ''"
                    class="flex-1 text-sm border-gray-300 dark:border-gray-700"
                />
                <Button
                    :disabled="isLoading"
                    type="button"
                    @click="handleCopy"
                    size="sm"
                    :class="[
                        'px-3 sm:px-4 whitespace-nowrap',
                        isLoading ? 'cursor-not-allowed' : '',
                    ]"
                >
                    <Copy class="w-4 h-4" v-if="!isCopied" />
                    <Check class="w-4 h-4" v-else />
                    <span class="ml-2 sm:hidden">{{
                        isCopied ? "Copied" : "Copy"
                    }}</span>
                </Button>
            </div>

            <!-- Private Chat Info -->
            <div
                v-else
                class="flex flex-col gap-3 sm:gap-4 w-full items-stretch sm:items-end"
            >
                <p
                    class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 text-center sm:text-left"
                >
                    Don't share personal information or third-party content
                    without permission.
                </p>
                <Button
                    type="button"
                    @click="handleShare(false)"
                    :disabled="isLoading"
                    size="sm"
                    class="w-full sm:w-auto px-4"
                >
                    <span v-if="isLoading">Creating link...</span>
                    <span v-else>Create share link</span>
                </Button>
            </div>
        </div>
    </DialogBox>
</template>
