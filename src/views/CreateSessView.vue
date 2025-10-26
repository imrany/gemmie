<script lang="ts" setup>
import { computed, inject, ref, type Ref, onMounted, onUnmounted } from "vue";
import type { Chat, UserDetails } from "@/types";
import DemoAccountModal from "@/components/Modals/DemoAccountModal.vue";
import {
    ArrowLeft,
    BookText,
    BriefcaseBusiness,
    ChartColumnBig,
    Check,
    Circle,
    Clock,
    CloudLightningIcon,
    Code,
    Database,
    Home,
    Info,
    Map,
    MessageCircle,
    Pencil,
    RefreshCw,
    ScrollText,
    Shield,
    Star,
    Sun,
} from "lucide-vue-next";

const {
    isAuthenticated,
    isDarkMode,
    isDemoMode,
    parsedUserDetails,
    handleAuth,
} = inject("globalState") as {
    isDarkMode: Ref<boolean>;
    isDemoMode: Ref<boolean>;
    isAuthenticated: Ref<boolean>;
    parsedUserDetails: Ref<UserDetails>;
    handleAuth: (data: {
        username: string;
        email: string;
        password: string;
        agreeToTerms: boolean;
    }) => any;
};

const props = defineProps<{
    data: {
        chats: Chat[];
        currentChatId: string;
        isCollapsed?: boolean;
        parsedUserDetails: {
            username: string;
        } | null;
        screenWidth: number;
        syncStatus: {
            lastSync: Date | null;
            syncing: boolean;
            hasUnsyncedChanges: boolean;
        };
        isLoading: boolean;
        authStep: number;
        showCreateSession: boolean;
        authData: {
            username: string;
            email: string;
            password: string;
            agreeToTerms: boolean;
        };
        currentMessages: any[];
    };
    functions: {
        validateCurrentStep: () => boolean;
        setShowInput: () => void;
        clearAllChats: () => void;
        toggleSidebar: () => void;
        logout: () => void;
        handleAuthSuccess: (response: any) => Promise<void>;
        createNewChat: () => void;
        switchToChat: (chatId: string) => void;
        deleteChat: (chatId: string) => void;
        renameChat: (chatId: string, newTitle: string) => void;
        manualSync: () => void;
        handleStepSubmit: (e: Event) => void;
        setShowCreateSession: (value: boolean) => void;
        prevAuthStep: () => void;
        updateAuthData: (
            data: Partial<{
                username: string;
                email: string;
                password: string;
                agreeToTerms: boolean;
            }>,
        ) => void;
    };
}>();

const showDemoModal = ref(false);
// For demo purpose (auto filled inputs)
const demoAccount = {
    email: "demo@example.com",
    password: "demo1234",
    username: "demo",
};

const openDemoModal = () => {
    showDemoModal.value = true;
    stopAutoSlide(); // Pause carousel
};

const closeDemoModal = () => {
    showDemoModal.value = false;
    startAutoSlide(); // Resume carousel
};

const startDemoMode = async () => {
    try {
        isDemoMode.value = true;
        // Auto-fill all fields with demo data
        props.functions.updateAuthData({
            username: demoAccount.username,
            email: demoAccount.email,
            password: demoAccount.password,
            agreeToTerms: true,
        });
        closeDemoModal();
        // creates sessions directly
        const response = await handleAuth(props.data.authData);

        // Validate response structure
        if (!response) {
            throw new Error("No response received from authentication service");
        }

        if (response.error) {
            throw new Error(response.error);
        }

        if (!response.data || !response.success) {
            throw new Error(
                "Authentication failed - invalid response structure",
            );
        }

        // Success handling
        await props.functions.handleAuthSuccess(response);
        isDemoMode.value = true;
    } catch (error: any) {
        console.log("Creating demo session error: ", error);
    }
};

const createRealAccount = () => {
    isDemoMode.value = false;
    // Clear any pre-filled data
    props.functions.updateAuthData({
        username: "",
        email: "",
        password: "",
        agreeToTerms: false,
    });
    closeDemoModal();
    props.functions.setShowCreateSession(true);
};

// Carousel state
const currentSlide = ref(0);
const autoSlideInterval = ref<number | null>(null);

// Touch swipe state
const touchStartX = ref(0);
const touchEndX = ref(0);
const minSwipeDistance = 50; // Minimum distance for a swipe to be detected

// Sample chat suggestions - you can customize these
const chatSuggestions = ref([
    {
        title: "Creative Writing",
        description: "Get help with stories, poems, and creative content",
        icon: Pencil,
        color: "from-purple-500 to-pink-500",
        prompt: "Help me write a creative story about...",
    },
    {
        title: "Code Assistant",
        description: "Debug code, learn programming concepts, and get help",
        icon: Code,
        color: "from-blue-500 to-cyan-500",
        prompt: "Can you help me debug this code?",
    },
    {
        title: "Learning & Research",
        description: "Explore topics, get explanations, and expand knowledge",
        icon: BookText,
        color: "from-green-500 to-teal-500",
        prompt: "Explain this topic in simple terms...",
    },
    {
        title: "Problem Solving",
        description: "Work through challenges and find solutions",
        icon: Sun,
        color: "from-yellow-500 to-orange-500",
        prompt: "I need help solving this problem...",
    },
    {
        title: "Data Analysis",
        description: "Analyze data, create charts, and find insights",
        icon: ChartColumnBig,
        color: "from-indigo-500 to-purple-500",
        prompt: "Help me analyze this data...",
    },
    {
        title: "Business Strategy",
        description: "Strategic planning, market analysis, and business advice",
        icon: BriefcaseBusiness,
        color: "from-orange-500 to-red-500",
        prompt: "I need business advice on...",
    },
]);

const features = ref([
    {
        title: "Privacy First",
        description:
            "Your conversations stay private and secure with end-to-end encryption",
        icon: Shield,
        color: "text-green-500",
    },
    {
        title: "Local Storage",
        description:
            "All data is stored locally on your device for maximum privacy",
        icon: Database,
        color: "text-blue-500",
    },
    {
        title: "Cross-Device Sync",
        description: "Seamlessly access your chats from any device, anywhere",
        icon: RefreshCw,
        color: "text-purple-500",
    },
    {
        title: "Always Available",
        description: "24/7 AI assistance at your fingertips, even offline",
        icon: Clock,
        color: "text-indigo-500",
    },
    {
        title: "Fast Response",
        description:
            "Lightning-fast AI responses powered by advanced algorithms",
        icon: CloudLightningIcon,
        color: "text-yellow-500",
    },
    {
        title: "Smart Memory",
        description:
            "Remembers context across conversations for better assistance",
        icon: BookText,
        color: "text-pink-500",
    },
]);

const tips = ref([
    {
        title: "Be Specific",
        description: "The more details you provide, the better I can help you",
        icon: Pencil,
        example:
            "Instead of 'help with code', try 'debug my React component that won't render'",
    },
    {
        title: "Ask Follow-ups",
        description: "Don't hesitate to ask for clarification or more details",
        icon: MessageCircle,
        example:
            "Can you explain that in simpler terms? or What about edge cases?",
    },
    {
        title: "Use Examples",
        description: "Provide examples of what you're working with",
        icon: ScrollText,
        example:
            "Here's my current code... or This is the error I'm getting...",
    },
    {
        title: "Set Context",
        description:
            "Let me know your skill level and what you're trying to achieve",
        icon: Map,
        example: "I'm a beginner in Python and want to build a web scraper",
    },
]);

const carouselSlides = computed(() => [
    { id: "welcome", title: "Welcome", icon: Home },
    { id: "suggestions", title: "Chat Ideas", icon: MessageCircle },
    { id: "features", title: "Features", icon: Star },
    { id: "tips", title: "Tips", icon: Info },
]);

// Carousel functions
const nextSlide = () => {
    currentSlide.value = (currentSlide.value + 1) % carouselSlides.value.length;
};

const prevSlide = () => {
    currentSlide.value =
        currentSlide.value === 0
            ? carouselSlides.value.length - 1
            : currentSlide.value - 1;
};

const goToSlide = (index: number) => {
    currentSlide.value = index;
};

const startAutoSlide = () => {
    if (autoSlideInterval.value) return;
    autoSlideInterval.value = window.setInterval(nextSlide, 8000); // 8 seconds
};

const stopAutoSlide = () => {
    if (autoSlideInterval.value) {
        clearInterval(autoSlideInterval.value);
        autoSlideInterval.value = null;
    }
};

// Touch event handlers for swipe
const handleTouchStart = (event: TouchEvent) => {
    touchStartX.value = event.touches[0].clientX;
    touchEndX.value = 0;
    stopAutoSlide();
};

const handleTouchMove = (event: TouchEvent) => {
    touchEndX.value = event.touches[0].clientX;
};

const handleTouchEnd = () => {
    if (!touchStartX.value || !touchEndX.value) return;

    const distance = touchStartX.value - touchEndX.value;
    const isLeftSwipe = distance > minSwipeDistance;
    const isRightSwipe = distance < -minSwipeDistance;

    if (isLeftSwipe) {
        nextSlide();
    } else if (isRightSwipe) {
        prevSlide();
    }

    // Restart auto-slide after a delay
    setTimeout(startAutoSlide, 3000);
};

// Handle input updates
const handleUsernameInput = (event: Event) => {
    const value = (event.target as HTMLInputElement).value.trim();
    props.functions.updateAuthData({ username: value });
};

const handleEmailInput = (event: Event) => {
    const value = (event.target as HTMLInputElement).value.trim();
    props.functions.updateAuthData({ email: value });
};

const handlePasswordInput = (event: Event) => {
    const value = (event.target as HTMLInputElement).value.trim();
    props.functions.updateAuthData({ password: value });
};

const handleTermsToggle = (event: Event) => {
    const checked = (event.target as HTMLInputElement).checked;
    props.functions.updateAuthData({ agreeToTerms: checked });
};

// Layout logic
const isDesktop = computed(() => props.data.screenWidth >= 720);
const isMobile = computed(() => props.data.screenWidth < 720);

// Lifecycle
onMounted(() => {
    startAutoSlide();

    const previousRoute = document.referrer;
    const isFromUpgrade =
        previousRoute.includes("/upgrade") ||
        window.location.search.includes("from=upgrade");
    if (isFromUpgrade) {
        props.functions.setShowCreateSession(true);
    } else {
        // ✅ Show demo modal on first visit
        // Only show if user is not authenticated
        if (!isAuthenticated.value) {
            setTimeout(() => {
                openDemoModal();
            }, 1500); // Show after 1.5 seconds
        }
    }
});

onUnmounted(() => {
    stopAutoSlide();
});
</script>

<template>
    <div
        class="flex flex-col items-center justify-center min-h-screen w-screen bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 transition-colors duration-300"
    >
        <DemoAccountModal
            :showDemoModal="showDemoModal"
            :createRealAccount="createRealAccount"
            :startDemoMode="startDemoMode"
            :closeDemoModal="closeDemoModal"
        />
        <!-- Desktop Layout: Side by side -->
        <div
            v-if="isDesktop"
            class="flex scale-90 gap-10 items-center justify-center h-full w-full max-w-7xl mx-auto px-8"
        >
            <!-- LEFT SECTION: Carousel -->
            <div
                class="flex-1 max-w-2xl"
                @mouseenter="stopAutoSlide"
                @mouseleave="startAutoSlide"
            >
                <!-- Carousel Container -->
                <div
                    class="relative h-[600px] overflow-hidden transition-colors duration-300"
                >
                    <!-- Slide 1: Welcome -->
                    <div
                        :class="
                            currentSlide === 0
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 0
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-12 flex flex-col items-center justify-center"
                    >
                        <div class="text-center">
                            <img
                                :src="
                                    isDarkMode ? '/logo-light.svg' : '/logo.svg'
                                "
                                alt="Gemmie Logo"
                                class="w-[60px] h-[60px] mx-auto mb-5 rounded-md"
                            />

                            <h2
                                class="text-4xl font-bold text-gray-900 dark:text-white mb-4"
                            >
                                Welcome to Gemmie
                            </h2>
                            <p
                                class="text-gray-600 dark:text-gray-300 leading-relaxed mb-3 max-w-lg text-base"
                            >
                                Experience privacy-first conversations with
                                advanced AI. Your data stays secure, local, and
                                synced across all your devices.
                            </p>
                            <div
                                class="flex items-center justify-center gap-12"
                            >
                                <div class="flex flex-col items-center gap-3">
                                    <div
                                        class="w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <Shield
                                            class="w-6 h-6 text-green-600 dark:text-green-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Private</span
                                    >
                                </div>
                                <div class="flex flex-col items-center gap-3">
                                    <div
                                        class="w-10 h-10 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <Database
                                            class="w-6 h-6 text-blue-600 dark:text-blue-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Local</span
                                    >
                                </div>
                                <div class="flex flex-col items-center gap-3">
                                    <div
                                        class="w-10 h-10 bg-purple-100 dark:bg-purple-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <RefreshCw
                                            class="w-6 h-6 text-purple-600 dark:text-purple-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Synced</span
                                    >
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Slide 2: Chat Suggestions -->
                    <div
                        :class="
                            currentSlide === 1
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 1
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8"
                    >
                        <div class="text-center mb-2">
                            <h2
                                class="text-3xl font-bold text-gray-900 dark:text-white mb-2"
                            >
                                What can I help with?
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-lg">
                                Try one of these popular conversation starters
                            </p>
                        </div>

                        <div
                            class="grid grid-cols-2 gap-6 h-96 overflow-x-hidden overflow-y-auto no-scrollbar"
                        >
                            <button
                                v-for="suggestion in chatSuggestions"
                                :key="suggestion.title"
                                class="group px-6 py-3 rounded-xl bg-white/80 dark:bg-gray-700/80 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/95 dark:hover:bg-gray-700/95 hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] text-left"
                            >
                                <div
                                    :class="`w-10 h-10 rounded-lg bg-gradient-to-r ${suggestion.color} flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`"
                                >
                                    <component
                                        :is="suggestion.icon"
                                        :class="`w-4 h-4 text-white`"
                                    />
                                </div>
                                <h3
                                    class="font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    {{ suggestion.title }}
                                </h3>
                                <p
                                    class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed"
                                >
                                    {{ suggestion.description }}
                                </p>
                            </button>
                        </div>
                    </div>

                    <!-- Slide 3: Features -->
                    <div
                        :class="
                            currentSlide === 2
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 2
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8"
                    >
                        <div class="text-center mb-8">
                            <h2
                                class="text-3xl font-bold text-gray-900 dark:text-white mb-3"
                            >
                                Why Choose Gemmie?
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-lg">
                                Built with your privacy and convenience in mind
                            </p>
                        </div>

                        <div
                            class="space-y-2 h-96 overflow-y-auto no-scrollbar"
                        >
                            <div
                                v-for="feature in features"
                                :key="feature.title"
                                class="flex items-start gap-4 py-3 px-6 rounded-xl bg-white/80 dark:bg-gray-700/80 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/95 dark:hover:bg-gray-700/95 transition-all duration-300"
                            >
                                <div
                                    class="flex-shrink-0 w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-600 flex items-center justify-center"
                                >
                                    <component
                                        :is="feature.icon"
                                        :class="`w-4 h-4 ${feature.color} dark:${feature.color}`"
                                    />
                                    <i></i>
                                </div>
                                <div>
                                    <h3
                                        class="font-semibold text-gray-900 dark:text-white mb-2"
                                    >
                                        {{ feature.title }}
                                    </h3>
                                    <p
                                        class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed"
                                    >
                                        {{ feature.description }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Slide 4: Tips -->
                    <div
                        :class="
                            currentSlide === 3
                                ? 'translate-x-0 opacity-100'
                                : '-translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8"
                    >
                        <div class="text-center mb-8">
                            <h2
                                class="text-3xl font-bold text-gray-900 dark:text-white mb-3"
                            >
                                Pro Tips
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-lg">
                                Get the most out of your AI conversations
                            </p>
                        </div>

                        <div
                            class="space-y-4 h-96 overflow-y-auto no-scrollbar"
                        >
                            <div
                                v-for="tip in tips"
                                :key="tip.title"
                                class="py-3 px-6 rounded-xl bg-white/80 dark:bg-gray-700/80 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/95 dark:hover:bg-gray-700/95 transition-all duration-300"
                            >
                                <div class="flex items-start gap-4 mb-3">
                                    <div
                                        class="flex-shrink-0 w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center"
                                    >
                                        <component
                                            :is="tip.icon"
                                            :class="`w-4 h-4 text-blue-600 dark:text-blue-400`"
                                        />
                                    </div>
                                    <div>
                                        <h3
                                            class="font-semibold text-gray-900 dark:text-white mb-2"
                                        >
                                            {{ tip.title }}
                                        </h3>
                                        <p
                                            class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed"
                                        >
                                            {{ tip.description }}
                                        </p>
                                    </div>
                                </div>
                                <div
                                    class="ml-14 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg"
                                >
                                    <p
                                        class="text-xs text-blue-700 dark:text-blue-300 italic"
                                    >
                                        {{ tip.example }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Slide Indicators -->
                    <div
                        class="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-3"
                    >
                        <button
                            v-for="(slide, index) in carouselSlides"
                            :key="slide.id"
                            @click="goToSlide(index)"
                            :class="
                                currentSlide === index
                                    ? 'bg-black dark:bg-white shadow-lg scale-110'
                                    : 'bg-gray-100 dark:bg-gray-600 hover:bg-white dark:hover:bg-gray-500'
                            "
                            class="group w-10 h-10 rounded-full transition-all duration-300 flex items-center justify-center"
                            :title="slide.title"
                        >
                            <component
                                :is="slide.icon"
                                :class="[
                                    'w-4 h-4',
                                    currentSlide === index
                                        ? `text-white dark:text-gray-900`
                                        : `text-gray-500 dark:text-gray-300`,
                                ]"
                            />
                        </button>
                    </div>
                </div>
            </div>

            <!-- RIGHT SECTION: Auth Form -->
            <div class="flex-1 max-w-md">
                <!-- Multi-step Auth Form -->
                <div
                    class="text-sm relative overflow-hidden p-8 transition-colors duration-300"
                >
                    <!-- Progress indicator -->
                    <div class="flex justify-center mb-8">
                        <div class="flex items-center space-x-2">
                            <div
                                v-for="step in 4"
                                :key="step"
                                :class="
                                    step <= props.data.authStep
                                        ? 'bg-blue-600 dark:bg-blue-500'
                                        : 'bg-gray-300 dark:bg-gray-600'
                                "
                                class="w-3 h-3 rounded-full transition-colors duration-300"
                            ></div>
                        </div>
                    </div>

                    <!-- Multi-step form container -->
                    <div class="relative h-96">
                        <!-- Step 1: Username -->
                        <div
                            :class="
                                props.data.authStep === 1
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 1
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-8">
                                <h2
                                    class="text-2xl font-semibold text-gray-900 dark:text-white mb-3"
                                >
                                    Welcome!
                                </h2>
                                <p class="text-gray-600 dark:text-gray-300">
                                    Let's start by creating your username
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-6"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3"
                                    >
                                        Choose a username
                                    </label>
                                    <input
                                        v-model="props.data.authData.username"
                                        required
                                        type="text"
                                        placeholder="johndoe"
                                        class="border border-gray-300 dark:border-gray-600 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.username &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handleUsernameInput"
                                    />
                                    <p
                                        class="text-xs text-gray-500 dark:text-gray-400 mt-2"
                                    >
                                        This will be your display name
                                    </p>
                                </div>

                                <button
                                    type="submit"
                                    :disabled="
                                        !props.functions.validateCurrentStep
                                    "
                                    class="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed text-white rounded-xl px-6 py-3 font-semibold transition-all duration-300 transform hover:scale-[1.02] shadow-lg"
                                >
                                    Continue
                                </button>
                            </form>
                        </div>

                        <!-- Step 2: Email -->
                        <div
                            :class="
                                props.data.authStep === 2
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 2
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-8">
                                <h2
                                    class="text-2xl font-semibold text-gray-900 dark:text-white mb-3"
                                >
                                    Hi {{ props.data.authData.username }}!
                                </h2>
                                <p class="text-gray-600 dark:text-gray-300">
                                    What's your email address?
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-6"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3"
                                    >
                                        Email address
                                    </label>
                                    <input
                                        v-model="props.data.authData.email"
                                        required
                                        type="email"
                                        placeholder="johndoe@example.com"
                                        class="border border-gray-300 dark:border-gray-600 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.email &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handleEmailInput"
                                    />
                                    <p
                                        class="text-xs text-gray-500 dark:text-gray-400 mt-2"
                                    >
                                        Used for session identification only
                                    </p>
                                </div>

                                <div class="flex gap-4">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-xl px-4 py-3 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.functions.validateCurrentStep
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-xl px-4 py-3 font-medium text-white transition-all duration-200"
                                    >
                                        Continue
                                    </button>
                                </div>
                            </form>
                        </div>

                        <!-- Step 3: Password -->
                        <div
                            :class="
                                props.data.authStep === 3
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 3
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-8">
                                <h2
                                    class="text-2xl font-semibold text-gray-900 dark:text-white mb-3"
                                >
                                    Almost there!
                                </h2>
                                <p class="text-gray-600 dark:text-gray-300">
                                    Create a secure password
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-6"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3"
                                    >
                                        Password
                                    </label>
                                    <input
                                        v-model="props.data.authData.password"
                                        required
                                        type="password"
                                        placeholder="Enter a secure password"
                                        minlength="8"
                                        class="border border-gray-300 dark:border-gray-600 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.password &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handlePasswordInput"
                                    />
                                    <div class="mt-3">
                                        <div
                                            class="flex items-center gap-2 text-xs"
                                        >
                                            <div
                                                :class="
                                                    props.data.authData.password
                                                        .length >= 8
                                                        ? 'text-green-600 dark:text-green-400'
                                                        : 'text-gray-400 dark:text-gray-500'
                                                "
                                                class="flex items-center gap-1"
                                            >
                                                <Check
                                                    v-if="
                                                        props.data.authData
                                                            .password.length >=
                                                        8
                                                    "
                                                    class="w-4 h-4"
                                                />
                                                <Circle
                                                    class="w-4 h-4"
                                                    v-else
                                                />
                                                <span
                                                    >At least 8 characters</span
                                                >
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex gap-4">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-xl px-4 py-3 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.functions.validateCurrentStep
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-xl px-4 py-3 font-medium text-white transition-all duration-200"
                                    >
                                        Continue
                                    </button>
                                </div>
                            </form>
                        </div>

                        <!-- Step 4: Terms & Conditions -->
                        <div
                            :class="
                                props.data.authStep === 4
                                    ? 'translate-x-0 opacity-100'
                                    : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-8">
                                <h2
                                    class="text-2xl font-semibold text-gray-900 dark:text-white mb-3"
                                >
                                    One last step
                                </h2>
                                <p class="text-gray-600 dark:text-gray-300">
                                    Please review and accept our terms
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-6"
                            >
                                <!-- Terms and Conditions Checkboxes -->
                                <div class="space-y-4">
                                    <div
                                        class="border border-gray-200 dark:border-gray-600 rounded-xl p-4 bg-gray-50 dark:bg-gray-700/50"
                                    >
                                        <div class="flex items-start gap-3">
                                            <input
                                                id="agree-terms"
                                                v-model="
                                                    props.data.authData
                                                        .agreeToTerms
                                                "
                                                type="checkbox"
                                                required
                                                class="mt-1 h-4 w-4 text-blue-600 dark:text-blue-500 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700"
                                                @change="handleTermsToggle"
                                            />
                                            <label
                                                for="agree-terms"
                                                class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed cursor-pointer"
                                            >
                                                I agree to the
                                                <router-link
                                                    to="/legal/terms"
                                                    class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline font-medium"
                                                >
                                                    Terms of Service
                                                </router-link>
                                                and
                                                <router-link
                                                    to="/legal/privacy"
                                                    class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline font-medium"
                                                >
                                                    Privacy Policy
                                                </router-link>
                                                <span class="text-red-500"
                                                    >*</span
                                                >
                                            </label>
                                        </div>
                                    </div>

                                    <!-- Key points about terms -->
                                    <div
                                        class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl p-4"
                                    >
                                        <div class="flex items-start gap-3">
                                            <div
                                                class="flex-shrink-0 w-6 h-6 bg-blue-100 dark:bg-blue-800 rounded-full flex items-center justify-center"
                                            >
                                                <Info
                                                    class="w-4 h-4 text-blue-600 dark:text-blue-400"
                                                />
                                            </div>
                                            <div
                                                class="text-xs text-blue-800 dark:text-blue-200 space-y-2"
                                            >
                                                <p>
                                                    <strong
                                                        >Key highlights:</strong
                                                    >
                                                </p>
                                                <ul
                                                    class="list-disc list-inside space-y-1 ml-2"
                                                >
                                                    <li>
                                                        Your data remains
                                                        private and encrypted
                                                    </li>
                                                    <li>
                                                        We don't sell your
                                                        personal information
                                                    </li>
                                                    <li>
                                                        You can delete your
                                                        account and data anytime
                                                    </li>
                                                    <li>
                                                        Local storage with
                                                        optional cloud sync
                                                    </li>
                                                </ul>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex gap-4">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-xl px-4 py-3 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.data.authData.agreeToTerms ||
                                            props.data.isLoading
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-xl px-4 py-3 font-medium text-white transition-all duration-200"
                                    >
                                        <RotateCw
                                            v-if="props.data.isLoading"
                                            class="animate-spin w-4 h-4"
                                        />
                                        <Check v-else class="w-4 h-4" />
                                        <span>{{
                                            props.data.isLoading
                                                ? "Creating..."
                                                : "Create Session"
                                        }}</span>
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>

                    <!-- Footer note -->
                    <div class="text-center mt-6">
                        <div
                            class="flex text-xs flex-wrap gap-4 justify-center leading-relaxed text-gray-600 dark:text-gray-400"
                        >
                            <router-link
                                to="/legal/privacy"
                                class="hover:text-blue-600 dark:hover:text-blue-400"
                            >
                                Privacy Policy
                            </router-link>
                            <router-link
                                to="/legal/terms"
                                class="hover:text-blue-600 dark:hover:text-blue-400"
                            >
                                Terms of Service
                            </router-link>
                            <span>© 2025 Gemmie. All rights reserved.</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Mobile Layout: Vertical stack with carousel -->
        <div
            v-if="isMobile"
            class="flex flex-col gap-8 items-center justify-center min-h-screen w-full max-w-full px-4 py-6 overflow-y-auto overflow-x-hidden"
        >
            <!-- Mobile Carousel (always shown) -->
            <div
                v-if="!props.data.showCreateSession"
                class="w-full max-w-sm"
                @touchstart="handleTouchStart"
                @touchmove="handleTouchMove"
                @touchend="handleTouchEnd"
                @touchcancel="handleTouchEnd"
            >
                <div
                    class="relative h-[440px] overflow-hidden duration-300 touch-pan-y"
                >
                    <!-- Mobile Slide 1: Welcome -->
                    <div
                        :class="
                            currentSlide === 0
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 0
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-6 flex flex-col items-center justify-center"
                    >
                        <div class="text-center">
                            <img
                                :src="
                                    parsedUserDetails?.theme === 'dark' ||
                                    (parsedUserDetails?.theme === 'system' &&
                                        isDarkMode)
                                        ? '/logo-light.svg'
                                        : '/logo.svg'
                                "
                                alt="Gemmie Logo"
                                class="w-[60px] h-[60px] mx-auto mb-5 rounded-md"
                            />

                            <h2
                                class="text-2xl font-bold text-gray-900 dark:text-white mb-2"
                            >
                                Welcome to Gemmie
                            </h2>
                            <p
                                class="text-gray-600 dark:text-gray-300 leading-relaxed mb-6 text-sm"
                            >
                                Experience privacy-first conversations with
                                advanced AI.
                            </p>
                            <div
                                class="flex items-center justify-center gap-6 text-xs"
                            >
                                <div class="flex flex-col items-center gap-1">
                                    <div
                                        class="w-8 h-8 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <Shield
                                            class="w-5 h-5 text-green-600 dark:text-green-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Private</span
                                    >
                                </div>
                                <div class="flex flex-col items-center gap-1">
                                    <div
                                        class="w-8 h-8 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <Database
                                            class="w-5 h-5 text-blue-600 dark:text-blue-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Local</span
                                    >
                                </div>
                                <div class="flex flex-col items-center gap-1">
                                    <div
                                        class="w-8 h-8 bg-purple-100 dark:bg-purple-900/30 rounded-full flex items-center justify-center"
                                    >
                                        <RefreshCw
                                            class="w-5 h-5 text-purple-600 dark:text-purple-400"
                                        />
                                    </div>
                                    <span
                                        class="text-gray-700 dark:text-gray-300 font-medium"
                                        >Synced</span
                                    >
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Mobile Slide 2: Chat Suggestions -->
                    <div
                        :class="
                            currentSlide === 1
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 1
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4"
                    >
                        <div class="text-center mb-4">
                            <h2
                                class="text-xl font-bold text-gray-900 dark:text-white mb-2"
                            >
                                What can I help with?
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-sm">
                                Try these conversation starters
                            </p>
                        </div>

                        <div
                            class="grid grid-cols-2 gap-3 h-72 overflow-y-auto no-scrollbar"
                        >
                            <button
                                v-for="suggestion in chatSuggestions"
                                :key="suggestion.title"
                                class="group p-3 rounded-lg bg-white/70 dark:bg-gray-700/70 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/90 dark:hover:bg-gray-700/90 hover:shadow-lg transition-all duration-300 transform hover:scale-[1.02] text-left"
                            >
                                <div
                                    :class="`w-8 h-8 rounded-lg bg-gradient-to-r ${suggestion.color} flex items-center justify-center mb-2 group-hover:scale-110 transition-transform`"
                                >
                                    <component
                                        :is="suggestion.icon"
                                        :class="`w-4 h-4 text-white text-sm`"
                                    />
                                </div>
                                <h3
                                    class="font-semibold text-gray-900 dark:text-white text-xs mb-1"
                                >
                                    {{ suggestion.title }}
                                </h3>
                                <p
                                    class="text-xs text-gray-600 dark:text-gray-300 leading-tight"
                                >
                                    {{ suggestion.description }}
                                </p>
                            </button>
                        </div>
                    </div>

                    <!-- Mobile Slide 3: Features -->
                    <div
                        :class="
                            currentSlide === 2
                                ? 'translate-x-0 opacity-100'
                                : currentSlide > 2
                                  ? '-translate-x-full opacity-0'
                                  : 'translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4"
                    >
                        <div class="text-center mb-4">
                            <h2
                                class="text-xl font-bold text-gray-900 dark:text-white mb-2"
                            >
                                Why Choose Gemmie?
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-sm">
                                Built with your privacy in mind
                            </p>
                        </div>

                        <div
                            class="space-y-3 h-72 overflow-y-auto no-scrollbar"
                        >
                            <div
                                v-for="feature in features"
                                :key="feature.title"
                                class="flex items-start gap-3 p-3 rounded-lg bg-white/70 dark:bg-gray-700/70 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/90 dark:hover:bg-gray-700/90 transition-all duration-300 hover:shadow-md"
                            >
                                <div
                                    class="flex-shrink-0 w-8 h-8 rounded-lg bg-gray-100 dark:bg-gray-600 flex items-center justify-center"
                                >
                                    <component
                                        :is="feature.icon"
                                        :class="`w-4 h-4 ${feature.color} dark:${feature.color} text-sm`"
                                    />
                                </div>
                                <div>
                                    <h3
                                        class="font-semibold text-gray-900 dark:text-white text-xs mb-1"
                                    >
                                        {{ feature.title }}
                                    </h3>
                                    <p
                                        class="text-xs text-gray-600 dark:text-gray-300 leading-tight"
                                    >
                                        {{ feature.description }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Mobile Slide 4: Tips -->
                    <div
                        :class="
                            currentSlide === 3
                                ? 'translate-x-0 opacity-100'
                                : '-translate-x-full opacity-0'
                        "
                        class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4"
                    >
                        <div class="text-center mb-4">
                            <h2
                                class="text-xl font-bold text-gray-900 dark:text-white mb-2"
                            >
                                Pro Tips
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300 text-sm">
                                Get the most out of conversations
                            </p>
                        </div>

                        <div
                            class="space-y-3 h-72 overflow-y-auto no-scrollbar"
                        >
                            <div
                                v-for="tip in tips"
                                :key="tip.title"
                                class="p-3 rounded-lg bg-white/70 dark:bg-gray-700/70 backdrop-blur-sm border border-white/50 dark:border-gray-600/50 hover:bg-white/90 dark:hover:bg-gray-700/90 transition-all duration-300 hover:shadow-md"
                            >
                                <div class="flex items-start gap-2 mb-2">
                                    <div
                                        class="flex-shrink-0 w-6 h-6 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center"
                                    >
                                        <component
                                            :is="tip.icon"
                                            :class="`w-4 h-4 text-blue-600 dark:text-blue-400 text-xs`"
                                        />
                                    </div>
                                    <div>
                                        <h3
                                            class="font-semibold text-gray-900 dark:text-white text-xs mb-1"
                                        >
                                            {{ tip.title }}
                                        </h3>
                                        <p
                                            class="text-xs text-gray-600 dark:text-gray-300 leading-tight"
                                        >
                                            {{ tip.description }}
                                        </p>
                                    </div>
                                </div>
                                <div
                                    class="ml-8 p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg"
                                >
                                    <p
                                        class="text-xs text-blue-700 dark:text-blue-300 italic"
                                    >
                                        {{ tip.example }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Mobile Slide Indicators -->
                    <div
                        class="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-2"
                    >
                        <button
                            v-for="(slide, index) in carouselSlides"
                            :key="slide.id"
                            @click="goToSlide(index)"
                            :class="
                                currentSlide === index
                                    ? 'bg-black dark:bg-white shadow-lg scale-110'
                                    : 'bg-gray-100 dark:bg-gray-600 hover:bg-gray-100 dark:hover:bg-gray-500'
                            "
                            class="group w-8 h-8 rounded-full transition-all duration-300 flex items-center justify-center"
                            :title="slide.title"
                        >
                            <component
                                :is="slide.icon"
                                :class="[
                                    'w-4 h-4',
                                    currentSlide === index
                                        ? `text-white dark:text-gray-900`
                                        : `text-gray-500 dark:text-gray-300`,
                                ]"
                            />
                        </button>
                    </div>
                </div>
            </div>

            <!-- Mobile Auth Section -->
            <div
                v-if="props.data.showCreateSession"
                class="w-full max-w-sm pb-4 px-1"
            >
                <div
                    class="p-2 transition-colors duration-300 max-w-full overflow-hidden"
                >
                    <!-- Progress indicator -->
                    <div class="flex justify-center mb-6">
                        <div class="flex items-center space-x-2">
                            <div
                                v-for="step in 4"
                                :key="step"
                                :class="
                                    step <= props.data.authStep
                                        ? 'bg-blue-600 dark:bg-blue-500'
                                        : 'bg-gray-300 dark:bg-gray-600'
                                "
                                class="w-2.5 h-2.5 rounded-full transition-colors duration-300"
                            ></div>
                        </div>
                    </div>

                    <!-- Mobile Multi-step form container -->
                    <div class="relative min-h-80 w-full max-w-full">
                        <!-- Mobile Step 1: Username -->
                        <div
                            :class="
                                props.data.authStep === 1
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 1
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-6">
                                <h2
                                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    Welcome!
                                </h2>
                                <p
                                    class="text-gray-600 dark:text-gray-300 text-sm"
                                >
                                    Let's start by creating your username
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-4"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                    >
                                        Choose a username
                                    </label>
                                    <input
                                        v-model="props.data.authData.username"
                                        required
                                        type="text"
                                        placeholder="johndoe"
                                        class="border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.username &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handleUsernameInput"
                                    />
                                    <p
                                        class="text-xs text-gray-500 dark:text-gray-400 mt-1"
                                    >
                                        This will be your display name
                                    </p>
                                </div>

                                <button
                                    type="submit"
                                    :disabled="
                                        !props.functions.validateCurrentStep
                                    "
                                    class="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed text-white rounded-lg px-6 py-2.5 font-semibold transition-all duration-300 transform hover:scale-[1.02] shadow-lg"
                                >
                                    Continue
                                </button>
                            </form>
                        </div>

                        <!-- Mobile Step 2: Email -->
                        <div
                            :class="
                                props.data.authStep === 2
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 2
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-6">
                                <h2
                                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    Hi {{ props.data.authData.username }}!
                                </h2>
                                <p
                                    class="text-gray-600 dark:text-gray-300 text-sm"
                                >
                                    What's your email address?
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-4"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                    >
                                        Email address
                                    </label>
                                    <input
                                        v-model="props.data.authData.email"
                                        required
                                        type="email"
                                        placeholder="johndoe@example.com"
                                        class="border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.email &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handleEmailInput"
                                    />
                                    <p
                                        class="text-xs text-gray-500 dark:text-gray-400 mt-1"
                                    >
                                        Used for session identification only
                                    </p>
                                </div>

                                <div class="flex gap-3">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-lg px-4 py-2.5 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.functions.validateCurrentStep
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-lg px-4 py-2.5 font-medium text-white transition-all duration-200"
                                    >
                                        Continue
                                    </button>
                                </div>
                            </form>
                        </div>

                        <!-- Mobile Step 3: Password -->
                        <div
                            :class="
                                props.data.authStep === 3
                                    ? 'translate-x-0 opacity-100'
                                    : props.data.authStep > 3
                                      ? '-translate-x-full opacity-0'
                                      : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-6">
                                <h2
                                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    Almost there!
                                </h2>
                                <p
                                    class="text-gray-600 dark:text-gray-300 text-sm"
                                >
                                    Create a secure password
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-4"
                            >
                                <div>
                                    <label
                                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                                    >
                                        Password
                                    </label>
                                    <input
                                        v-model="props.data.authData.password"
                                        required
                                        type="password"
                                        placeholder="Enter a secure password"
                                        minlength="8"
                                        class="border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                                        :class="
                                            props.data.authData.password &&
                                            !props.functions.validateCurrentStep
                                                ? 'border-red-300 dark:border-red-500'
                                                : ''
                                        "
                                        @input="handlePasswordInput"
                                    />
                                    <div class="mt-2">
                                        <div
                                            class="flex items-center gap-2 text-xs"
                                        >
                                            <div
                                                :class="
                                                    props.data.authData.password
                                                        .length >= 8
                                                        ? 'text-green-600 dark:text-green-400'
                                                        : 'text-gray-400 dark:text-gray-500'
                                                "
                                                class="flex items-center gap-1"
                                            >
                                                <Check
                                                    v-if="
                                                        props.data.authData
                                                            .password.length >=
                                                        8
                                                    "
                                                    class="w-4 h-4"
                                                />
                                                <Circle
                                                    class="w-4 h-4"
                                                    v-else
                                                />
                                                <span
                                                    >At least 8 characters</span
                                                >
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex gap-3">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-lg px-4 py-2.5 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.functions.validateCurrentStep
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-lg px-4 py-2.5 font-medium text-white transition-all duration-200"
                                    >
                                        Continue
                                    </button>
                                </div>
                            </form>
                        </div>

                        <!-- Mobile Step 4: Terms & Conditions -->
                        <div
                            :class="
                                props.data.authStep === 4
                                    ? 'translate-x-0 opacity-100'
                                    : 'translate-x-full opacity-0'
                            "
                            class="absolute inset-0 transition-all duration-500 ease-in-out transform"
                        >
                            <div class="text-center mb-4">
                                <h2
                                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                                >
                                    One last step
                                </h2>
                                <p
                                    class="text-gray-600 dark:text-gray-300 text-sm"
                                >
                                    Please review and accept our terms
                                </p>
                            </div>

                            <form
                                @submit.prevent="
                                    props.functions.handleStepSubmit
                                "
                                class="space-y-4"
                            >
                                <!-- Terms and Conditions Checkbox -->
                                <div class="space-y-3">
                                    <div
                                        class="border border-gray-200 dark:border-gray-600 rounded-lg p-3 bg-gray-50 dark:bg-gray-700/50"
                                    >
                                        <div class="flex items-start gap-2">
                                            <input
                                                id="mobile-agree-terms"
                                                v-model="
                                                    props.data.authData
                                                        .agreeToTerms
                                                "
                                                type="checkbox"
                                                required
                                                class="mt-0.5 h-3.5 w-3.5 text-blue-600 dark:text-blue-500 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700"
                                                @change="handleTermsToggle"
                                            />
                                            <label
                                                for="mobile-agree-terms"
                                                class="text-xs text-gray-700 dark:text-gray-300 leading-relaxed cursor-pointer"
                                            >
                                                I agree to the
                                                <router-link
                                                    to="/legal/terms"
                                                    class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline font-medium"
                                                >
                                                    Terms of Service
                                                </router-link>
                                                and
                                                <router-link
                                                    to="/legal/privacy"
                                                    class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline font-medium"
                                                >
                                                    Privacy Policy
                                                </router-link>
                                                <span class="text-red-500"
                                                    >*</span
                                                >
                                            </label>
                                        </div>
                                    </div>

                                    <!-- Key points about terms (mobile version) -->
                                    <div
                                        class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3"
                                    >
                                        <div class="flex items-start gap-2">
                                            <div
                                                class="flex-shrink-0 w-5 h-5 bg-blue-100 dark:bg-blue-800 rounded-full flex items-center justify-center"
                                            >
                                                <Info
                                                    class="w-4 h-4 text-blue-600 dark:text-blue-400"
                                                />
                                            </div>
                                            <div
                                                class="text-xs text-blue-800 dark:text-blue-200 space-y-1"
                                            >
                                                <p>
                                                    <strong
                                                        >Key highlights:</strong
                                                    >
                                                </p>
                                                <ul
                                                    class="list-disc list-inside space-y-0.5 ml-1 text-xs"
                                                >
                                                    <li>
                                                        Your data remains
                                                        private and encrypted
                                                    </li>
                                                    <li>
                                                        We don't sell your
                                                        personal information
                                                    </li>
                                                    <li>
                                                        You can delete your
                                                        account anytime
                                                    </li>
                                                    <li>
                                                        Local storage with
                                                        optional cloud sync
                                                    </li>
                                                </ul>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex gap-3">
                                    <button
                                        type="button"
                                        @click="props.functions.prevAuthStep"
                                        class="flex-1 flex gap-2 items-center justify-center bg-gray-100 dark:bg-gray-700 backdrop-blur-sm text-gray-700 dark:text-gray-300 rounded-lg px-4 py-2.5 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-all duration-200"
                                    >
                                        <ArrowLeft class="w-4 h-4" /> Back
                                    </button>
                                    <button
                                        type="submit"
                                        :disabled="
                                            !props.data.authData.agreeToTerms ||
                                            props.data.isLoading
                                        "
                                        class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 dark:disabled:from-gray-600 dark:disabled:to-gray-700 disabled:cursor-not-allowed flex-1 flex gap-2 items-center justify-center transform hover:scale-[1.02] shadow-lg rounded-lg px-4 py-2.5 font-medium text-white transition-all duration-200"
                                    >
                                        <RotateCw
                                            v-if="props.data.isLoading"
                                            class="animate-spin w-4 h-4"
                                        />
                                        <Check class="w-4 h-4" v-else />
                                        <span>{{
                                            props.data.isLoading
                                                ? "Creating..."
                                                : "Create Session"
                                        }}</span>
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>

                    <!-- Mobile Footer note -->
                    <div class="text-center mt-4">
                        <div
                            class="flex text-xs flex-wrap gap-4 justify-center leading-relaxed text-gray-600 dark:text-gray-400"
                        >
                            <router-link
                                to="/legal/privacy"
                                class="hover:text-blue-600 dark:hover:text-blue-400"
                            >
                                Privacy Policy
                            </router-link>
                            <router-link
                                to="/legal/terms"
                                class="hover:text-blue-600 dark:hover:text-blue-400"
                            >
                                Terms of Service
                            </router-link>
                            <span>© 2025 Gemmie. All rights reserved.</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Mobile Get Started Button (when not in auth mode) -->
            <div v-else class="w-full max-w-sm">
                <button
                    @click="() => props.functions.setShowCreateSession(true)"
                    class="group w-full px-6 py-3 bg-gradient-to-r from-indigo-500 to-blue-600 text-white rounded-xl hover:from-indigo-600 hover:to-blue-700 transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-xl font-medium"
                >
                    <span class="flex items-center justify-center gap-2">
                        <ArrowRight
                            class="w-4 h-4 group-hover:translate-x-1 transition-transform"
                        />
                        Get Started
                    </span>
                </button>
            </div>
        </div>

        <!-- Footer disclaimer -->
        <div v-if="isAuthenticated" class="absolute bottom-4 text-center">
            <p class="text-xs text-gray-500 dark:text-gray-400">
                © 2025 Gemmie. All rights reserved.
            </p>
        </div>
    </div>
</template>

<style scoped>
.touch-pan-y {
    touch-action: pan-y;
}

/* Optional: Add visual feedback during swipe */
.carousel-slide {
    transition: transform 0.3s ease-out;
}
</style>
