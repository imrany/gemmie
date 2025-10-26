<script lang="ts" setup>
import type { LinkPreview } from "@/types";
import { ExternalLink, LoaderCircle, PlayCircle, Video } from "lucide-vue-next";
import { ref, onUnmounted, computed } from "vue";

interface Props {
    preview: LinkPreview;
}

const props = defineProps<Props>();

const videoId = ref(`video-${Math.random().toString(36).substr(2, 9)}`);
const showControls = ref(false);
const isPlaying = ref(false);

const hasVideo = computed(() => {
    return props.preview.video || props.preview.embedHtml;
});

const isEmbeddableVideo = computed(() => {
    return (
        props.preview.embedHtml &&
        (props.preview.videoType === "youtube" ||
            props.preview.videoType === "vimeo")
    );
});

const isDirectVideo = computed(() => {
    return props.preview.videoType === "direct" && props.preview.video;
});

const isSocialVideo = computed(() => {
    return (
        (props.preview.videoType === "twitter" ||
            props.preview.videoType === "tiktok") &&
        props.preview.previewImage
    );
});

// Video control functions
const playEmbeddedVideo = (element: HTMLElement) => {
    const container = element.closest(".video-embed-container") as HTMLElement;
    if (container && container.dataset.embed) {
        container.innerHTML = container.dataset.embed;
        showControls.value = true;
        isPlaying.value = true;
    }
};

const pauseVideo = () => {
    const video = document.getElementById(
        `${videoId.value}-video`,
    ) as HTMLVideoElement;
    if (video) {
        video.pause();
        isPlaying.value = false;
    }
};

const resumeVideo = () => {
    const video = document.getElementById(
        `${videoId.value}-video`,
    ) as HTMLVideoElement;
    if (video) {
        video.play();
        isPlaying.value = true;
    }
};

const stopVideo = () => {
    const video = document.getElementById(
        `${videoId.value}-video`,
    ) as HTMLVideoElement;
    if (video) {
        video.pause();
        video.currentTime = 0;
        isPlaying.value = false;
    }
};

const toggleDirectVideo = () => {
    const video = document.getElementById(
        `${videoId.value}-video`,
    ) as HTMLVideoElement;
    if (video) {
        if (video.paused) {
            video.play();
            isPlaying.value = true;
        } else {
            video.pause();
            isPlaying.value = false;
        }
    }
};

const stopDirectVideo = () => {
    const video = document.getElementById(
        `${videoId.value}-video`,
    ) as HTMLVideoElement;
    if (video) {
        video.pause();
        video.currentTime = 0;
        isPlaying.value = false;
    }
};

const showVideoControls = () => {
    showControls.value = true;
};

const updateVideoControls = (state: "playing" | "paused" | "ended") => {
    isPlaying.value = state === "playing";
    if (state === "ended") {
        showControls.value = false;
    }
};

const playSocialVideo = (url: string) => {
    window.open(url, "_blank");
};

// Handle image errors
const handleImageError = (event: Event) => {
    const img = event.target as HTMLImageElement;
    img.style.display = "none";
    const errorDiv = img.nextElementSibling as HTMLElement;
    if (errorDiv) {
        errorDiv.style.display = "block";
    }
};

// Cleanup on unmount
onUnmounted(() => {
    // Clean up any video resources if needed
});
</script>

<template>
    <!-- Loading State -->
    <div
        v-if="preview.loading"
        class="link-preview loading border border-gray-200 dark:border-gray-700 rounded-lg p-3 my-2 bg-gray-50 dark:bg-gray-800 max-w-full transition-colors duration-200"
    >
        <div class="flex items-center gap-2">
            <LoaderCircle
                class="w-4 h-4 animate-spin text-gray-400 dark:text-gray-500 flex-shrink-0"
            />
            <span class="text-sm text-gray-500 dark:text-gray-400 truncate"
                >Loading preview...</span
            >
        </div>
    </div>

    <!-- Error State -->
    <div
        v-else-if="preview.error"
        class="link-preview error border border-gray-200 dark:border-gray-700 rounded-lg p-3 my-2 bg-gray-50 dark:bg-gray-800 max-w-full transition-colors duration-200"
    >
        <div class="flex items-center gap-2 min-w-0">
            <ExternalLink
                class="w-4 h-4 text-gray-400 dark:text-gray-500 flex-shrink-0"
            />
            <a
                :href="preview.url"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium truncate min-w-0 flex-1 transition-colors duration-200"
            >
                {{ preview.domain }}
            </a>
        </div>
    </div>

    <div
        v-else
        class="link-preview border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden my-2 bg-white dark:bg-gray-800 hover:shadow-md dark:hover:shadow-gray-900/30 transition-all duration-300 w-fit max-w-full"
    >
        <!-- Video Preview -->
        <div v-if="hasVideo" class="w-full max-w-[500px]">
            <!-- Embeddable Videos (YouTube, Vimeo) -->
            <div
                v-if="isEmbeddableVideo"
                class="aspect-video w-full bg-black dark:bg-gray-900 relative group overflow-hidden"
            >
                <div
                    class="video-embed-container object-cover w-full h-full"
                    :data-embed="preview.embedHtml"
                    :data-video-type="preview.videoType"
                    :data-video-id="videoId"
                >
                    <!-- Initial thumbnail state -->
                    <div
                        class="video-thumbnail w-full h-full bg-gray-900 dark:bg-gray-800 flex items-center justify-center cursor-pointer overflow-hidden"
                        @click="
                            playEmbeddedVideo(
                                $event.currentTarget as HTMLElement,
                            )
                        "
                    >
                        <img
                            v-if="
                                preview.videoThumbnail || preview.previewImage
                            "
                            :src="
                                preview.videoThumbnail || preview.previewImage
                            "
                            :alt="preview.title"
                            class="w-full h-full object-cover"
                        />
                        <div
                            class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 dark:bg-opacity-50 group-hover:bg-opacity-20 dark:group-hover:bg-opacity-40 transition-colors duration-200"
                        >
                            <div
                                class="w-12 h-12 sm:w-16 sm:h-16 bg-red-600 hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200"
                            >
                                <svg
                                    class="w-4 h-4 sm:w-6 sm:h-6 text-white ml-0.5 sm:ml-1"
                                    fill="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path d="M8 5v14l11-7z" />
                                </svg>
                            </div>
                        </div>
                        <div
                            v-if="preview.videoDuration"
                            class="absolute bottom-2 right-2 bg-black bg-opacity-80 dark:bg-opacity-90 text-white text-xs px-2 py-1 rounded max-w-[calc(100%-1rem)] truncate"
                        >
                            {{ preview.videoDuration }}
                        </div>
                    </div>
                </div>

                <!-- Video controls overlay -->
                <div
                    v-if="showControls"
                    class="video-controls absolute top-2 right-2 flex gap-2 transition-opacity duration-200"
                >
                    <button
                        v-if="isPlaying"
                        @click="pauseVideo"
                        class="pause-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                        title="Pause"
                    >
                        <svg
                            class="w-4 h-4"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
                        </svg>
                    </button>
                    <button
                        v-else
                        @click="resumeVideo"
                        class="play-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                        title="Resume"
                    >
                        <svg
                            class="w-4 h-4 ml-0.5"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M8 5v14l11-7z" />
                        </svg>
                    </button>
                    <button
                        @click="stopVideo"
                        class="stop-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                        title="Stop"
                    >
                        <svg
                            class="w-4 h-4"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M6 6h12v12H6z" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Direct Video Files -->
            <div
                v-else-if="isDirectVideo"
                class="aspect-video w-full bg-black dark:bg-gray-900 overflow-hidden relative group"
            >
                <video
                    :id="`${videoId}-video`"
                    controls
                    preload="metadata"
                    class="w-full h-full object-contain"
                    :poster="preview.previewImage || ''"
                    @play="showVideoControls"
                    @pause="updateVideoControls('paused')"
                    @ended="updateVideoControls('ended')"
                >
                    <source :src="preview.video" type="video/mp4" />
                    <source :src="preview.video" type="video/webm" />
                    Your browser does not support the video tag.
                </video>

                <!-- Custom controls overlay for direct videos -->
                <div
                    class="video-controls absolute top-2 right-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                >
                    <button
                        @click="toggleDirectVideo"
                        class="toggle-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                        title="Play/Pause"
                    >
                        <svg
                            v-if="!isPlaying"
                            class="play-icon w-4 h-4 ml-0.5"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M8 5v14l11-7z" />
                        </svg>
                        <svg
                            v-else
                            class="pause-icon w-4 h-4"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
                        </svg>
                    </button>
                    <button
                        @click="stopDirectVideo"
                        class="stop-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                        title="Stop"
                    >
                        <svg
                            class="w-4 h-4"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M6 6h12v12H6z" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Social Media Videos -->
            <div
                v-else-if="isSocialVideo"
                class="aspect-video w-full bg-gray-100 dark:bg-gray-800 relative group overflow-hidden cursor-pointer"
                @click="playSocialVideo(preview.url)"
            >
                <img
                    :src="preview.previewImage"
                    :alt="preview.title"
                    class="w-full h-full object-cover"
                />
                <div
                    class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 dark:bg-opacity-50 group-hover:bg-opacity-20 dark:group-hover:bg-opacity-40 transition-colors duration-200"
                >
                    <div
                        class="w-10 h-10 sm:w-12 sm:h-12 bg-white dark:bg-gray-300 bg-opacity-90 dark:bg-opacity-90 hover:bg-opacity-100 dark:hover:bg-opacity-100 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200"
                    >
                        <svg
                            class="w-3 h-3 sm:w-4 sm:h-4 text-gray-800 dark:text-gray-900 ml-0.5"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M8 5v14l11-7z" />
                        </svg>
                    </div>
                </div>
                <div
                    class="absolute top-2 left-2 bg-black bg-opacity-80 dark:bg-opacity-90 text-white text-xs px-2 py-1 rounded capitalize"
                >
                    {{ preview.videoType }}
                </div>
            </div>

            <!-- Video Preview Content -->
            <div class="p-3 sm:p-4 min-w-0">
                <div class="flex items-start justify-between gap-2 min-w-0">
                    <div class="flex-1 min-w-0">
                        <h4
                            class="font-medium text-gray-900 dark:text-white text-sm sm:text-base line-clamp-2 mb-1 break-words"
                        >
                            <PlayCircle
                                class="w-4 h-4 text-red-600 dark:text-red-500 mr-1 flex-shrink-0"
                            />
                            <a
                                :href="preview.url"
                                target="_blank"
                                rel="noopener noreferrer"
                                class="hover:text-blue-600 dark:hover:text-blue-400 break-words transition-colors duration-200"
                            >
                                {{ preview.title }}
                            </a>
                        </h4>
                        <p
                            v-if="preview.description"
                            class="text-gray-600 dark:text-gray-400 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed transition-colors duration-200"
                        >
                            {{ preview.description }}
                        </p>
                        <div
                            class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 dark:text-gray-400 min-w-0 transition-colors duration-200"
                        >
                            <Video
                                class="w-4 h-4 text-red-600 dark:text-red-500 flex-shrink-0"
                            />
                            <span class="truncate min-w-0 flex-1">{{
                                preview.domain
                            }}</span>
                            <span
                                v-if="preview.videoDuration"
                                class="ml-2 flex-shrink-0 hidden xs:inline"
                            >
                                • {{ preview.videoDuration }}
                            </span>
                        </div>
                        <div
                            v-if="preview.videoDuration"
                            class="text-xs text-gray-500 dark:text-gray-500 mt-1 xs:hidden transition-colors duration-200"
                        >
                            Duration: {{ preview.videoDuration }}
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Regular Link Preview -->
        <div v-else>
            <a
                :href="preview.url"
                class="block w-full max-w-[400px]"
                target="_blank"
                rel="noopener noreferrer"
            >
                <div
                    v-if="preview.previewImage"
                    class="aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 transition-colors duration-200"
                >
                    <img
                        :src="preview.previewImage"
                        :alt="preview.title"
                        class="w-full h-full object-cover"
                        @error="handleImageError"
                    />
                </div>
                <div class="p-3 sm:p-4 min-w-0">
                    <div class="flex items-start justify-between gap-2 min-w-0">
                        <div class="flex-1 min-w-0">
                            <h4
                                class="font-medium text-gray-900 dark:text-white text-sm sm:text-base line-clamp-2 mb-1 break-words transition-colors duration-200"
                            >
                                <span class="break-words">{{
                                    preview.title
                                }}</span>
                            </h4>
                            <p
                                v-if="preview.description"
                                class="text-gray-600 dark:text-gray-400 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed transition-colors duration-200"
                            >
                                {{ preview.description }}
                            </p>
                            <div
                                class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 dark:text-gray-400 min-w-0 transition-colors duration-200"
                            >
                                <ExternalLink class="w-4 h-4 flex-shrink-0" />
                                <span class="truncate min-w-0 flex-1">{{
                                    preview.domain
                                }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </a>
        </div>
    </div>
</template>
