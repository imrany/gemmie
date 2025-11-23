<script lang="ts" setup>
import type { Chat } from "@/types";
import {
    ref,
    inject,
    computed,
    onMounted,
    onUnmounted,
    nextTick,
    type Ref,
} from "vue";
import { useRouter } from "vue-router";
import ChatDropdown from "./Dropdowns/ChatDropdown.vue";
import ProfileDropdown from "./Dropdowns/ProfileDropdown.vue";
import {
    MessageCircle,
    FilePenLine,
    Ellipsis,
    Clock,
    ChevronDown,
    ChevronUp,
    AlignLeft,
    AlignJustify,
    X,
    Settings,
    HelpCircle,
    ArrowUpRight,
    Plus,
    CircleArrowUp,
    Info,
    Gamepad2,
} from "lucide-vue-next";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import type { FunctionalComponent } from "vue";
import { Button } from "./ui/button";
import { toast } from "vue-sonner";
import { watch } from "vue";

const {
    activeChatMenu,
    toggleChatMenu,
    showProfileMenu,
    handleClickOutside,
    planStatus,
    hideSidebar,
    isSidebarHidden,
    screenWidth,
    currentChatId,
    isCollapsed,
    chats,
    closePreview,
    showPreviewSidebar,
    isOnline,
} = inject("globalState") as {
    isOnline: Ref<boolean>;
    showPreviewSidebar: Ref<boolean>;
    isCollapsed: Ref<boolean>;
    closePreview: () => void;
    chats: Ref<Chat[]>;
    currentChatId: Ref<string>;
    activeChatMenu: Ref<string | null>;
    toggleChatMenu: (chatId: string, event: Event) => void;
    showProfileMenu: Ref<boolean>;
    handleClickOutside: () => void;
    planStatus: Ref<{
        status: string;
        timeLeft: string;
        expiryDate: string;
        isExpired: boolean;
    }>;
    hideSidebar: () => void;
    isSidebarHidden: Ref<boolean>;
    screenWidth: Ref<number>;
};

const props = defineProps<{
    data: {
        chats: Chat[];
        parsedUserDetails: {
            username: string;
            email: string;
            syncEnabled: boolean;
            planName?: string;
            expiryTimestamp?: number;
        };
    };
    functions: {
        clearAllChats: () => void;
        toggleSidebar: () => void;
        logout: () => void;
        createNewChat: () => void;
        deleteChat: (chatId: string) => void;
        renameChat: (chatId: string, newTitle: string) => Promise<string>;
        manualSync: () => void;
    };
}>();

const router = useRouter();
const isRenaming = ref<string | null>(null);
const renameValue = ref("");
const now = ref(Date.now());
const hoveredChatId = ref<string | null>(null);
const isLoading = ref(false);

let timer: number | null = null;

onMounted(() => {
    timer = window.setInterval(() => (now.value = Date.now()), 1000);
});

onUnmounted(() => {
    if (timer) clearInterval(timer);
});

// Computed
const showFullSidebar = computed(
    () => !isCollapsed.value || screenWidth.value < 720,
);
const planColor = computed(() => {
    if (planStatus.value.isExpired)
        return "text-red-600 bg-red-50 dark:bg-red-900/20 dark:text-red-400";
    if (planStatus.value.status === "no-plan") return "text-gray-400 bg-none";
    return "text-green-600 bg-green-50 dark:bg-green-900/20 dark:text-green-400";
});

const sidebarIconClass = computed(() => {
    let icon: FunctionalComponent<any>;

    if (screenWidth.value > 720) {
        if (isCollapsed.value) {
            icon = AlignJustify;
        } else {
            icon = AlignLeft;
        }
    } else {
        icon = X;
    }

    return icon;
});

// Constants
const profileOptions = [
    {
        id: "settings",
        label: "Settings",
        icon: Settings,
        action: () => router.push("/settings/general"),
    },
    {
        id: "help",
        label: "Get help",
        icon: HelpCircle,
        action: () => window.open("mailto:imranmat254@gmail.com", "_blank"),
    },
    {
        id: "upgrade",
        label: props.data.parsedUserDetails?.planName
            ? "Manage plan"
            : "Upgrade plan",
        icon: CircleArrowUp,
        action: () => router.push("/upgrade"),
    },
    { id: "learn", icon: Info, label: "Learn more", action: () => {} },
];

// Methods
function startRename(chatId: string, currentTitle: string) {
    isRenaming.value = chatId;
    renameValue.value = currentTitle;
    activeChatMenu.value = null;

    nextTick(() => {
        const input = document.getElementById(
            `rename-${chatId}`,
        ) as HTMLInputElement;
        input?.focus();
        input?.select();
    });
}

function openWorkplace() {
    window.open("/workplace", "_blank");
    if (screenWidth.value < 720) hideSidebar();
}

async function submitRename(chatId: string) {
    isLoading.value = true;
    await props.functions
        .renameChat(chatId, renameValue.value.trim())
        .then((newTitle) => {
            isRenaming.value = null;
            renameValue.value = "";
            console.log("New title:", newTitle);
            isLoading.value = false;
        })
        .catch((error: unknown) => {
            console.log(error);
            toast.error(`Failed to rename this chat`);
            isLoading.value = false;
        });
}

function cancelRename() {
    isRenaming.value = null;
    renameValue.value = "";
}

function handleChatClick(chatId: string) {
    // Don't process if already on this chat
    if (chatId === currentChatId.value && screenWidth.value > 720) return;

    currentChatId.value = chatId;
    router.push(`/chat/${chatId}`);
    if (screenWidth.value < 720) {
        hideSidebar();
    }
    if (showPreviewSidebar.value) closePreview();
}

function handleNavAction(action: () => void) {
    if (screenWidth.value < 720) hideSidebar();
    if (showPreviewSidebar.value) closePreview();
    action();
}

const handleSidebarToggle = () => {
    if (screenWidth.value > 720) {
        props.functions.toggleSidebar();
    } else {
        hideSidebar();
    }
};

const navLinks: {
    label: string;
    description: string;
    path: string;
    icon: FunctionalComponent<any>;
    action: () => void;
}[] = [
    {
        label: "New Chat",
        description: "New Chat",
        icon: Plus,
        path: "/new",
        action: () => handleNavAction(() => props.functions.createNewChat()),
    },
    {
        label: "Chats",
        description: "Recent Chats",
        icon: MessageCircle,
        path: "/chats",
        action: () => handleNavAction(() => router.push("/chats")),
    },
    {
        label: "Arcade",
        description: "Play Games and view other users publications",
        icon: Gamepad2,
        path: "/arcade",
        action: () => handleNavAction(() => router.push("/arcade")),
    },
];

watch(showPreviewSidebar, (newVal) => {
    if (newVal) {
        if (!isCollapsed.value) {
            props.functions.toggleSidebar();
        }
    }
});
</script>

<template>
    <div
        id="side_nav"
        :class="[
            // Mobile styles
            screenWidth <= 720
                ? isSidebarHidden
                    ? 'w-0 opacity-0 -translate-x-full'
                    : 'w-full opacity-100 translate-x-0 bg-gray-100 dark:bg-gray-800'
                : '',

            // Desktop styles
            screenWidth > 720
                ? isCollapsed
                    ? 'w-[60px]'
                    : 'w-[270px] bg-gray-100 dark:bg-gray-800'
                : '',

            // Base styles
            'border-r z-40 fixed top-0 left-0 bottom-0 flex flex-col',
            'dark:border-gray-700',

            // Animation classes
            'transition-all duration-300 ease-in-out transform select-none',
        ]"
        @click="handleClickOutside"
    >
        <!-- Scrollable area -->
        <div class="flex-1 overflow-y-auto custom-scrollbar px-3">
            <!-- Top Header -->
            <div class="flex items-center py-3">
                <p
                    v-if="showFullSidebar"
                    class="text-gray-700 dark:text-gray-300 text-xl max-md:text-2xl font-semibold tracking-wide select-none"
                >
                    Gemmie
                </p>

                <div class="flex ml-auto gap-2 items-center justify-center">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger as-child>
                                <button
                                    @click="handleSidebarToggle"
                                    class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors"
                                >
                                    <component
                                        :is="sidebarIconClass"
                                        class="text-gray-500 dark:text-gray-400 w-4 h-4"
                                    />
                                </button>
                            </TooltipTrigger>
                            <TooltipContent
                                v-if="!showFullSidebar"
                                side="right"
                                :avoid-collisions="true"
                            >
                                <p>Toggle Sidebar</p>
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
            </div>

            <!-- Navigation Menu -->
            <div
                v-if="props.data.parsedUserDetails.username"
                class="mb-4 mt-2 max-md:text-lg flex flex-col gap-1 text-sm"
            >
                <div v-for="navlink in navLinks" :key="navlink.label">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger as-child>
                                <Button
                                    variant="ghost"
                                    @click="navlink.action"
                                    :class="[
                                        'w-full group flex items-center gap-2 h-[40px] rounded-lg transition-colors',
                                        showFullSidebar
                                            ? 'justify-start'
                                            : 'justify-center',
                                        router.currentRoute.value.path ===
                                        navlink.path
                                            ? 'bg-gray-200 dark:bg-gray-700/50 px-2'
                                            : 'px-2',
                                        !showFullSidebar &&
                                        navlink.path === '/new'
                                            ? 'hover:bg-inherit dark:hover:bg-inherit'
                                            : ' hover:bg-gray-200 dark:hover:bg-gray-700/50 px-2',
                                        navlink.path === '/new' && !isOnline
                                            ? 'disabled:cursor-not-allowed'
                                            : '',
                                    ]"
                                    :disabled="
                                        navlink.path === '/new' &&
                                        !isOnline &&
                                        !isLoading
                                    "
                                >
                                    <div
                                        :class="[
                                            navlink.path === '/new'
                                                ? 'text-white dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full p-[4px] group-hover:scale-110 transition-transform duration-300'
                                                : '',
                                            'text-gray-500 dark:text-gray-400',
                                        ]"
                                    >
                                        <component
                                            :is="navlink.icon"
                                            class="w-5 h-5"
                                        />
                                    </div>
                                    <span
                                        v-if="showFullSidebar"
                                        class="dark:text-gray-200"
                                        >{{ navlink.label }}</span
                                    >
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent
                                v-if="!showFullSidebar"
                                side="right"
                                :avoid-collisions="true"
                            >
                                <p>{{ navlink.description }}</p>
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>

                <!-- workplace button -->
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger as-child>
                            <Button
                                variant="ghost"
                                :class="[
                                    'w-full flex items-center h-[40px] hover:bg-gray-200 dark:hover:bg-gray-700/50 rounded-lg px-2 transition-colors',
                                    showFullSidebar
                                        ? 'justify-start'
                                        : 'justify-center',
                                ]"
                                :disabled="screenWidth < 720"
                                @click="openWorkplace"
                            >
                                <div class="flex items-center gap-2 flex-grow">
                                    <FilePenLine
                                        class="text-gray-500 dark:text-gray-400 w-5 h-5"
                                    />
                                    <span
                                        v-if="showFullSidebar"
                                        class="dark:text-gray-200"
                                        >Workplace</span
                                    >
                                </div>
                                <ArrowUpRight
                                    v-if="showFullSidebar"
                                    class="text-gray-500 dark:text-gray-400 w-4 h-4"
                                />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent
                            v-if="!showFullSidebar"
                            side="right"
                            :avoid-collisions="true"
                        >
                            <p>Workplace</p>
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            </div>

            <!-- Recent Chats -->
            <div
                v-if="
                    chats &&
                    chats.length &&
                    props.data.parsedUserDetails.username &&
                    showFullSidebar
                "
                class="flex flex-col pb-4"
            >
                <p
                    class="text-sm text-gray-800 dark:text-gray-200 mb-3 tracking-wider"
                >
                    Recents
                </p>
                <div class="flex flex-col gap-1">
                    <Button
                        variant="ghost"
                        v-for="chat in chats"
                        :key="chat.id"
                        @mouseenter="hoveredChatId = chat.id"
                        @mouseleave="hoveredChatId = null"
                        :class="[
                            'group hover:bg-gray-200 px-0 py-0 dark:hover:bg-gray-700/50 text-sm justify-start w-full flex items-center rounded-md relative transition-all duration-150',
                            'h-[35px]',
                            chat.id === currentChatId
                                ? 'bg-gray-200 dark:bg-gray-700/80'
                                : 'hover:bg-gray-200 dark:hover:bg-gray-700/50',
                        ]"
                    >
                        <!-- Chat content area -->
                        <div
                            :disabled="isLoading"
                            @click="() => handleChatClick(chat.id)"
                            :class="[
                                'flex items-center h-full flex-grow px-2 py-[3px] cursor-pointer overflow-hidden',
                                chat.id === currentChatId
                                    ? 'font-medium'
                                    : 'font-normal',
                            ]"
                        >
                            <div
                                v-if="
                                    isRenaming === chat.id &&
                                    !chat.is_read_only &&
                                    !isLoading
                                "
                                class="flex-grow"
                                @click.stop
                            >
                                <input
                                    :id="`rename-${chat.id}`"
                                    v-model="renameValue"
                                    @keyup.enter="submitRename(chat.id)"
                                    @keyup.escape="cancelRename"
                                    @blur="submitRename(chat.id)"
                                    class="w-full px-2 py-[3px] text-sm bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border-[1px] border-blue-500 dark:border-blue-400 rounded-md focus:outline-none focus:ring-[1px] focus:ring-blue-500 dark:focus:ring-blue-400"
                                />
                            </div>
                            <div
                                v-else
                                class="flex-grow flex items-center gap-2 min-w-0"
                            >
                                <span
                                    :class="[
                                        'truncate',
                                        chat.id === currentChatId
                                            ? 'text-gray-800 dark:text-gray-200'
                                            : 'text-gray-500 dark:text-gray-500',
                                    ]"
                                    :title="
                                        isLoading && isRenaming === chat.id
                                            ? 'Renaming...'
                                            : chat.title.includes(
                                                    '#pastedText#',
                                                )
                                              ? chat.title.split(
                                                    '#pastedText#',
                                                )[1]
                                              : chat.title || 'Untitled Chat'
                                    "
                                >
                                    {{
                                        isLoading && isRenaming === chat.id
                                            ? "Renaming..."
                                            : chat.title.includes(
                                                    "#pastedText#",
                                                )
                                              ? chat.title.split(
                                                    "#pastedText#",
                                                )[1]
                                              : chat.title || "Untitled Chat"
                                    }}
                                </span>
                            </div>
                        </div>

                        <!-- Menu button with smooth transition -->
                        <button
                            v-show="
                                ((hoveredChatId === chat.id ||
                                    activeChatMenu === chat.id) &&
                                    !chat.is_read_only &&
                                    !isLoading) ||
                                !isOnline
                            "
                            @click.stop="toggleChatMenu(chat.id, $event)"
                            :class="[
                                'text-inherit group flex-shrink-0 flex items-center justify-center h-full px-2',
                            ]"
                        >
                            <div
                                class="rounded-md p-1.5 group-hover:bg-gray-300 dark:group-hover:bg-gray-600 transition-colors"
                            >
                                <Ellipsis
                                    class="w-4 h-4 text-gray-600 dark:text-gray-300"
                                />
                            </div>
                        </button>

                        <ChatDropdown
                            :data="{ activeChatMenu, chat, screenWidth }"
                            :functions="{
                                deleteChat: props.functions.deleteChat,
                                startRename,
                                hideSidebar,
                            }"
                        />
                    </Button>
                </div>
            </div>
        </div>

        <!-- Fixed Bottom User Profile -->
        <div
            :class="[
                screenWidth > 720
                    ? showFullSidebar
                        ? 'border-t'
                        : ''
                    : isSidebarHidden
                      ? 'hidden'
                      : '',
                'border-gray-200 dark:border-gray-700 p-3 sticky bottom-0',
                isCollapsed ? '' : 'bg-gray-100 dark:bg-gray-800',
            ]"
        >
            <!-- Plan Status -->
            <div
                v-if="
                    props.data.parsedUserDetails.username &&
                    planStatus.status === 'active' &&
                    showFullSidebar
                "
                class="mb-2 px-2 py-1 text-xs rounded transition-colors"
                :class="planColor"
            >
                <div class="flex items-center justify-between">
                    <span class="font-normal">{{ planStatus.timeLeft }}</span>
                    <Clock class="w-4 h-4" />
                </div>
            </div>

            <!-- Profile -->
            <div
                class="flex items-center justify-between cursor-pointer mr-1"
                @click.stop="
                    () => {
                        if (isCollapsed && screenWidth > 720) {
                            handleSidebarToggle();
                        }
                        showProfileMenu = !showProfileMenu;
                    }
                "
            >
                <div class="flex items-center gap-2">
                    <div
                        class="w-[35px] h-[35px] flex justify-center items-center bg-gray-300 dark:bg-gray-700 rounded-full relative"
                    >
                        <span
                            class="text-sm font-medium text-gray-800 max-md:text-lg dark:text-gray-200"
                        >
                            {{
                                props.data.parsedUserDetails.username
                                    .toUpperCase()
                                    .slice(0, 2)
                            }}
                        </span>
                        <!-- Plan status indicator -->
                        <div
                            v-if="planStatus.isExpired"
                            class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 dark:bg-red-400 rounded-full border-2 border-white dark:border-gray-900"
                        ></div>
                        <div
                            v-else-if="planStatus.status === 'active'"
                            class="absolute -top-1 -right-1 w-3 h-3 bg-green-500 dark:bg-green-400 rounded-full border-2 border-white dark:border-gray-900"
                        ></div>
                    </div>
                    <div v-if="showFullSidebar">
                        <p
                            class="text-base font-medium max-md:text-lg dark:text-gray-200"
                        >
                            {{ props.data.parsedUserDetails.username }}
                        </p>
                        <p
                            v-if="props.data.parsedUserDetails.planName"
                            class="text-xs text-gray-500 font-normal dark:text-gray-400"
                        >
                            {{ props.data.parsedUserDetails.planName }}
                        </p>
                    </div>
                </div>
                <div v-if="showFullSidebar">
                    <ChevronUp
                        v-if="showProfileMenu"
                        class="w-4 h-4 dark:text-gray-300"
                    />
                    <ChevronDown v-else class="w-4 h-4 dark:text-gray-300" />
                </div>
            </div>

            <!-- Profile Dropdown -->
            <ProfileDropdown
                :data="{
                    planColor,
                    profileOptions,
                    showProfileMenu,
                }"
                :functions="{
                    handleNavAction,
                    logout: props.functions.logout,
                }"
            />
        </div>
    </div>
</template>
