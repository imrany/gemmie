<script lang="ts" setup>
import ErrorState from "@/components/ErrorState.vue";
import LoadingState from "@/components/LoadingState.vue";
import ProtectedPage from "@/layout/ProtectedPage.vue";
import type { ApiResponse, RawArcade, UserDetails } from "@/types";
import type { Ref } from "vue";
import { ref, computed } from "vue";
import { onMounted } from "vue";
import { inject } from "vue";
import { useRouter } from "vue-router";
import { ChevronLeft, Code, Calendar, User, Search, X } from "lucide-vue-next";
import { Input } from "@/components/ui/input";

const { screenWidth, parsedUserDetails, isCollapsed, apiCall } = inject(
    "globalState",
) as {
    screenWidth: Ref<number>;
    parsedUserDetails: Ref<UserDetails>;
    isCollapsed: Ref<boolean>;
    apiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
};

const router = useRouter();
const arcades = ref<RawArcade[]>([]);
const activeTab = ref<"all" | "yours">("all");
const searchQuery = ref("");

function handleBack() {
    if (window.history.state.back) {
        router.back();
        return;
    }
    router.push("/");
}

const error = ref<string | null>(null);
const isLoading = ref(true);

async function fetchArcades() {
    try {
        isLoading.value = true;
        error.value = null;
        const res = await apiCall<RawArcade[]>(`/arcades`, {
            method: "GET",
        });
        if (!res.success) {
            throw new Error(res.message || "Failed to load arcades");
        }

        if (res.data) {
            arcades.value = res.data;
            console.log(arcades);
        } else {
            throw new Error("No arcade data found");
        }
    } catch (err: any) {
        console.error("Fetch arcades error:", err);
        if (err.message?.includes("Failed to fetch")) {
            error.value =
                "Failed to connect to the server. Please check your internet connection.";
        } else {
            error.value = err.message || "Failed to load arcades";
        }
    } finally {
        isLoading.value = false;
    }
}

const filteredArcades = computed(() => {
    let filtered = arcades.value;

    // Filter by tab
    if (activeTab.value === "yours") {
        filtered = filtered.filter(
            (arcade) => arcade.user_id === parsedUserDetails.value?.userId,
        );
    }

    // Filter by search query
    if (searchQuery.value.trim()) {
        const query = searchQuery.value.toLowerCase();
        filtered = filtered.filter(
            (arcade) =>
                arcade.label.toLowerCase().includes(query) ||
                arcade.description.toLowerCase().includes(query) ||
                arcade.code_type.toLowerCase().includes(query),
        );
    }

    // Sort by created date (newest first)
    return filtered.sort(
        (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
});

const yourArcadesCount = computed(() => {
    return arcades.value.filter(
        (arcade) => arcade.user_id === parsedUserDetails.value?.userId,
    ).length;
});

function clearSearch() {
    searchQuery.value = "";
}

function viewArcade(arcadeId: string) {
    router.push(`/arcade/${arcadeId}`);
}

onMounted(async () => {
    await fetchArcades();
});
</script>

<template>
    <ProtectedPage>
        <!-- Main Content - Centered -->
        <div
            :class="[
                'flex-grow flex flex-col transition-all duration-300 ease-in-out h-screen overflow-hidden min-w-0',
                screenWidth > 720 && parsedUserDetails?.username
                    ? isCollapsed
                        ? 'ml-[60px]'
                        : 'ml-[270px]'
                    : '',
            ]"
        >
            <!-- Loading State -->
            <LoadingState
                description="Please wait while we fetch your content"
                label="Loading Arcade"
                v-if="isLoading"
            />

            <!-- Error State -->
            <ErrorState
                v-else-if="error"
                @retry="fetchArcades"
                back-button-text="Back"
                :error="error"
            />

            <!-- Main Content -->
            <div
                v-else
                class="flex-1 flex overflow-hidden h-full justify-center p-3 sm:p-4 md:p-6 max-w-full"
            >
                <div
                    class="w-full px-2 flex flex-col h-full overflow-hidden max-w-7xl"
                >
                    <!-- Header Section -->
                    <div class="flex-shrink-0 w-full mb-6 space-y-4">
                        <!-- Top Bar -->
                        <div class="flex items-center justify-between gap-4">
                            <!-- Back Button (Mobile Only) -->
                            <button
                                @click="handleBack"
                                class="md:hidden flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors flex-shrink-0"
                                aria-label="Go back"
                            >
                                <ChevronLeft
                                    class="w-5 h-5 text-gray-700 dark:text-gray-300"
                                />
                            </button>

                            <h1
                                class="text-2xl md:text-3xl font-bold text-gray-900 dark:text-white"
                            >
                                Arcade
                            </h1>

                            <div class="flex-1"></div>
                        </div>

                        <!-- Tabs and Search -->
                        <div
                            class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between"
                        >
                            <!-- Tabs -->
                            <div
                                class="inline-flex items-center gap-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg"
                            >
                                <button
                                    @click="activeTab = 'all'"
                                    :class="[
                                        'px-4 py-2 text-sm font-medium rounded-md transition-all duration-200',
                                        activeTab === 'all'
                                            ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                    ]"
                                >
                                    All
                                    <span
                                        class="ml-1.5 text-xs"
                                        :class="
                                            activeTab === 'all'
                                                ? 'text-gray-500 dark:text-gray-400'
                                                : 'text-gray-400 dark:text-gray-500'
                                        "
                                    >
                                        {{ arcades.length }}
                                    </span>
                                </button>
                                <button
                                    @click="activeTab = 'yours'"
                                    :class="[
                                        'px-4 py-2 text-sm font-medium rounded-md transition-all duration-200',
                                        activeTab === 'yours'
                                            ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                            : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                    ]"
                                >
                                    Your Arcade
                                    <span
                                        class="ml-1.5 text-xs"
                                        :class="
                                            activeTab === 'yours'
                                                ? 'text-gray-500 dark:text-gray-400'
                                                : 'text-gray-400 dark:text-gray-500'
                                        "
                                    >
                                        {{ yourArcadesCount }}
                                    </span>
                                </button>
                            </div>

                            <!-- Search -->
                            <div class="relative w-full sm:w-64">
                                <Search
                                    class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
                                />
                                <Input
                                    v-model="searchQuery"
                                    placeholder="Search arcades..."
                                    class="pl-9 pr-9 h-10"
                                />
                                <button
                                    v-if="searchQuery"
                                    @click="clearSearch"
                                    class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                                >
                                    <X class="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Arcades Grid -->
                    <div
                        class="flex-1 overflow-y-auto custom-scrollbar"
                        v-if="filteredArcades.length > 0"
                    >
                        <div
                            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 pb-6"
                        >
                            <div
                                v-for="arcade in filteredArcades"
                                :key="arcade.id"
                                @click="viewArcade(arcade.id!)"
                                class="group relative bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden hover:shadow-lg hover:border-gray-300 dark:hover:border-gray-600 transition-all duration-200 cursor-pointer"
                            >
                                <!-- Preview Container -->
                                <div
                                    class="relative w-full aspect-video bg-gray-100 dark:bg-gray-900 overflow-hidden no-scrollbar"
                                >
                                    <!-- Iframe Preview -->
                                    <iframe
                                        :srcdoc="arcade.code"
                                        class="w-full h-full border-0 no-scrollbar overflow-hidden pointer-events-none scale-[0.5] origin-top-left"
                                        :style="{
                                            width: '200%',
                                            height: '200%',
                                        }"
                                        sandbox="allow-scripts"
                                        title="Preview"
                                    />

                                    <!-- Overlay on Hover -->
                                    <div
                                        class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center"
                                    >
                                        <div
                                            class="text-white text-sm font-medium"
                                        >
                                            View Full Preview
                                        </div>
                                    </div>

                                    <!-- Code Type Badge -->
                                    <div
                                        class="absolute top-2 right-2 px-2 py-1 bg-gray-900/80 backdrop-blur-sm text-white text-xs font-mono rounded"
                                    >
                                        {{ arcade.code_type }}
                                    </div>
                                </div>

                                <!-- Content -->
                                <div class="p-4 space-y-2">
                                    <!-- Title -->
                                    <h3
                                        class="font-semibold text-gray-900 dark:text-white line-clamp-1 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors"
                                    >
                                        {{ arcade.label }}
                                    </h3>

                                    <!-- Description -->
                                    <p
                                        class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2"
                                    >
                                        {{ arcade.description }}
                                    </p>

                                    <!-- Footer Info -->
                                    <div
                                        class="flex items-center gap-3 pt-2 text-xs text-gray-500 dark:text-gray-500"
                                    >
                                        <div class="flex items-center gap-1">
                                            <Calendar class="w-3.5 h-3.5" />
                                            {{ arcade.created_at }}
                                        </div>
                                        <div
                                            v-if="
                                                arcade.user_id ===
                                                parsedUserDetails?.userId
                                            "
                                            class="flex items-center gap-1 text-blue-600 dark:text-blue-400"
                                        >
                                            <User class="w-3.5 h-3.5" />
                                            You
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Empty State -->
                    <div v-else class="flex-1 flex items-center justify-center">
                        <div class="text-center max-w-md px-4">
                            <div
                                class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center"
                            >
                                <Code
                                    class="w-8 h-8 text-gray-400 dark:text-gray-600"
                                />
                            </div>
                            <h3
                                class="text-lg font-semibold text-gray-900 dark:text-white mb-2"
                            >
                                {{
                                    searchQuery
                                        ? "No results found"
                                        : activeTab === "yours"
                                          ? "No arcades yet"
                                          : "No arcades available"
                                }}
                            </h3>
                            <p class="text-sm text-gray-600 dark:text-gray-400">
                                {{
                                    searchQuery
                                        ? "Try adjusting your search terms"
                                        : activeTab === "yours"
                                          ? "Create your first arcade by publishing code from the editor"
                                          : "Be the first to publish an arcade!"
                                }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </ProtectedPage>
</template>
