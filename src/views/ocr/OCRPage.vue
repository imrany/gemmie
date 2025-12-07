<script setup lang="ts">
import { inject, ref, computed, type Ref } from "vue";
import { useImageOCR } from "@/composables/useImageOCR";
import { toast } from "vue-sonner";
import {
    Camera,
    Upload,
    Scissors,
    Loader2,
    Copy,
    Check,
    X,
    Sparkles,
    BookOpen,
    FileText,
    Download,
    Share2,
    ImageIcon,
    Search,
    MessageSquare,
    AlertCircle,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    CardDescription,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import type { UserDetails } from "@/types";

const {
    imageFile,
    croppedBlob,
    extractedData,
    loading,
    error,
    captureImage,
    cropImage,
    sendToBackend,
    reset,
    validateImage,
} = useImageOCR();

const { parsedUserDetails, isDarkMode } = inject("globalState") as {
    parsedUserDetails: Ref<UserDetails>;
    isDarkMode: Ref<boolean>;
};

const previewUrl = ref<string | null>(null);
const showCropper = ref(false);
const copied = ref(false);
const cropArea = ref({ x: 50, y: 50, width: 200, height: 200 });
const resultsOpen = ref(false);

// Computed
const wordCount = computed(() => {
    if (!extractedData.value.text) return 0;
    return extractedData.value.text.trim().split(/\s+/).length;
});

const characterCount = computed(() => {
    if (!extractedData.value.text) return 0;
    return extractedData.value.text_length;
});

const estimatedReadTime = computed(() => {
    const minutes = Math.ceil(wordCount.value / 200);
    return minutes || 1;
});

const onFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];

    if (file) {
        if (!validateImage(file)) return;

        captureImage(file);
        previewUrl.value = URL.createObjectURL(file);
        resultsOpen.value = false;
    }
};

const handleCrop = async () => {
    if (!imageFile.value) return;

    try {
        await cropImage(imageFile.value, cropArea.value);
        toast.success("Image cropped successfully!");
        showCropper.value = false;
    } catch (err: any) {
        toast.error(err.message || "Failed to crop image");
    }
};

const handleExtract = async () => {
    await sendToBackend(showCropper.value || croppedBlob.value !== null);
    if (extractedData.value.text) {
        resultsOpen.value = true;
    }
};

const copyToClipboard = () => {
    if (!extractedData.value.text) return;

    navigator.clipboard.writeText(extractedData.value.text);
    copied.value = true;
    toast.success("Copied to clipboard!");

    setTimeout(() => {
        copied.value = false;
    }, 2000);
};

const downloadAsText = () => {
    if (!extractedData.value.text) return;

    const blob = new Blob([extractedData.value.text], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `extracted-text-${Date.now()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast.success("Downloaded as text file!");
};

const handleReset = () => {
    reset();
    previewUrl.value = null;
    showCropper.value = false;
    copied.value = false;
    resultsOpen.value = false;
};

const shareText = async () => {
    if (!extractedData.value.text) return;

    if (navigator.share) {
        try {
            await navigator.share({
                title: "Extracted Text",
                text: extractedData.value.text,
            });
            toast.success("Shared successfully!");
        } catch (err) {
            console.log("Share cancelled");
        }
    } else {
        copyToClipboard();
    }
};

const analyzeWithGemmie = () => {
    toast.info("Analyzing with Gemmie... (Coming soon)");
};

const closeResults = () => {
    resultsOpen.value = false;
};
</script>

<template>
    <div class="min-h-screen bg-gray-50 dark:bg-gray-900 py-8 px-4">
        <div class="max-w-4xl mx-auto">
            <!-- Header -->
            <div class="text-center mb-8">
                <img
                    :src="
                        parsedUserDetails?.theme === 'dark' ||
                        (parsedUserDetails?.theme === 'system' && isDarkMode)
                            ? '/favicon-light.svg'
                            : '/favicon.svg'
                    "
                    alt="Gemmie Logo"
                    class="w-16 h-16 items-center justify-center inline-flex rounded-2xl mb-4 shadow-lg"
                />

                <h1
                    class="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-2"
                >
                    Gemmie Extractor
                </h1>
                <p
                    class="text-gray-600 dark:text-gray-400 text-sm md:text-base max-w-2xl mx-auto"
                >
                    Extract text from research papers, textbooks, lecture
                    slides, and documents. Perfect for students, scholars, and
                    researchers.
                </p>
            </div>

            <!-- Quick Stats -->
            <div v-if="extractedData.text" class="grid grid-cols-3 gap-3 mb-6">
                <Card class="text-center">
                    <CardContent class="p-4">
                        <FileText
                            class="w-6 h-6 mx-auto mb-2 text-purple-600 dark:text-purple-400"
                        />
                        <p
                            class="text-2xl font-bold text-gray-900 dark:text-white"
                        >
                            {{ wordCount }}
                        </p>
                        <p class="text-xs text-gray-600 dark:text-gray-400">
                            Words
                        </p>
                    </CardContent>
                </Card>
                <Card class="text-center">
                    <CardContent class="p-4">
                        <BookOpen
                            class="w-6 h-6 mx-auto mb-2 text-blue-600 dark:text-blue-400"
                        />
                        <p
                            class="text-2xl font-bold text-gray-900 dark:text-white"
                        >
                            {{ estimatedReadTime }}
                        </p>
                        <p class="text-xs text-gray-600 dark:text-gray-400">
                            Min Read
                        </p>
                    </CardContent>
                </Card>
                <Card class="text-center">
                    <CardContent class="p-4">
                        <Sparkles
                            class="w-6 h-6 mx-auto mb-2 text-green-600 dark:text-green-400"
                        />
                        <p
                            class="text-2xl font-bold text-gray-900 dark:text-white"
                        >
                            {{ characterCount }}
                        </p>
                        <p class="text-xs text-gray-600 dark:text-gray-400">
                            Characters
                        </p>
                    </CardContent>
                </Card>
            </div>

            <!-- Main Card -->
            <Card class="overflow-hidden shadow-lg bg-gray-50 dark:bg-gray-800">
                <CardHeader
                    v-if="!previewUrl"
                    class="border-b border-gray-200 dark:border-gray-700"
                >
                    <CardTitle class="flex items-center gap-2">
                        <Camera class="w-5 h-5" />
                        Upload Research Material
                    </CardTitle>
                    <CardDescription>
                        Upload images of textbooks, papers, slides, or
                        handwritten notes
                    </CardDescription>
                </CardHeader>

                <CardContent class="p-6">
                    <!-- Upload Section -->
                    <div v-if="!previewUrl">
                        <label
                            for="imageInput"
                            class="flex flex-col items-center justify-center border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl cursor-pointer hover:border-purple-500 dark:hover:border-purple-400 transition-all duration-200 p-8 md:p-12 bg-gray-50 dark:bg-gray-800/50 group"
                        >
                            <div
                                class="p-4 bg-purple-100 dark:bg-purple-900/30 rounded-full mb-4 group-hover:scale-110 transition-transform"
                            >
                                <Upload
                                    class="w-8 h-8 text-purple-600 dark:text-purple-400"
                                />
                            </div>
                            <span
                                class="text-lg font-semibold text-gray-700 dark:text-gray-200 mb-2"
                            >
                                Click to upload or capture
                            </span>
                            <span
                                class="text-sm text-gray-500 dark:text-gray-400 mb-3"
                            >
                                PNG, JPG, WebP • Max 4MB
                            </span>
                            <div class="flex gap-2 flex-wrap justify-center">
                                <Badge
                                    class="bg-gray-200 dark:bg-gray-700"
                                    variant="secondary"
                                >
                                    Research Papers
                                </Badge>
                                <Badge
                                    class="bg-gray-200 dark:bg-gray-700"
                                    variant="secondary"
                                    >Textbooks</Badge
                                >
                                <Badge
                                    class="bg-gray-200 dark:bg-gray-700"
                                    variant="secondary"
                                    >Lecture Slides</Badge
                                >
                                <Badge
                                    class="bg-gray-200 dark:bg-gray-700"
                                    variant="secondary"
                                    >Notes</Badge
                                >
                            </div>
                            <input
                                id="imageInput"
                                type="file"
                                accept="image/*"
                                capture="environment"
                                class="hidden"
                                @change="onFileChange"
                            />
                        </label>
                    </div>

                    <!-- Preview & Controls -->
                    <div v-else>
                        <div
                            class="relative rounded-xl overflow-hidden bg-gray-100 dark:bg-gray-800 mb-4"
                        >
                            <img
                                :src="previewUrl"
                                alt="Preview"
                                class="w-full max-h-96 object-contain"
                            />

                            <button
                                @click="handleReset"
                                class="absolute top-3 right-3 p-2 bg-red-500 hover:bg-red-600 text-white rounded-full shadow-lg transition-colors"
                            >
                                <X class="w-5 h-5" />
                            </button>

                            <Badge
                                v-if="croppedBlob"
                                class="absolute border-none top-3 left-3 text-white dark:text-white bg-purple-600 hover:bg-purple-700 dark:bg-purple-600 dark:hover:bg-purple-700"
                            >
                                <Scissors class="w-3 h-3 mr-1" />
                                Cropped
                            </Badge>
                        </div>

                        <!-- Action Buttons -->
                        <div class="flex flex-wrap gap-3 mb-4">
                            <Button
                                @click="showCropper = !showCropper"
                                variant="outline"
                                class="flex-1 min-w-[140px] bg-gray-800 hover:bg-gray-800/80 dark:bg-gray-100 hover:dark:bg-gray-100/80 text-gray-100 hover:text-gray-100 dark:text-gray-900 hover:dark:text-gray-900"
                            >
                                <Scissors class="w-4 h-4 mr-2" />
                                {{ showCropper ? "Cancel Crop" : "Crop Image" }}
                            </Button>

                            <Button
                                @click="handleExtract"
                                :disabled="loading"
                                class="flex-1 min-w-[140px] disabled:cursor-not-allowed disabled:opacity-50 text-gray-100 dark:text-white bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700"
                            >
                                <Loader2
                                    v-if="loading"
                                    class="w-4 h-4 mr-2 animate-spin"
                                />
                                <ImageIcon v-else class="w-4 h-4 mr-2" />
                                {{ loading ? "Processing..." : "Extract Text" }}
                            </Button>
                        </div>

                        <!-- Crop Controls -->
                        <div
                            v-if="showCropper"
                            class="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg border border-purple-200 dark:border-purple-800 mb-4"
                        >
                            <div class="flex items-center justify-between mb-3">
                                <h3
                                    class="text-sm font-semibold text-purple-900 dark:text-purple-200"
                                >
                                    Crop Settings
                                </h3>
                                <Button
                                    @click="handleCrop"
                                    size="sm"
                                    class="bg-purple-600 hover:bg-purple-700"
                                >
                                    Apply Crop
                                </Button>
                            </div>
                            <div class="grid grid-cols-2 gap-3 text-sm">
                                <div>
                                    <label
                                        class="block text-gray-700 dark:text-gray-300 mb-1 text-xs"
                                        >X Position</label
                                    >
                                    <input
                                        v-model.number="cropArea.x"
                                        type="number"
                                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                                    />
                                </div>
                                <div>
                                    <label
                                        class="block text-gray-700 dark:text-gray-300 mb-1 text-xs"
                                        >Y Position</label
                                    >
                                    <input
                                        v-model.number="cropArea.y"
                                        type="number"
                                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                                    />
                                </div>
                                <div>
                                    <label
                                        class="block text-gray-700 dark:text-gray-300 mb-1 text-xs"
                                        >Width</label
                                    >
                                    <input
                                        v-model.number="cropArea.width"
                                        type="number"
                                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                                    />
                                </div>
                                <div>
                                    <label
                                        class="block text-gray-700 dark:text-gray-300 mb-1 text-xs"
                                        >Height</label
                                    >
                                    <input
                                        v-model.number="cropArea.height"
                                        type="number"
                                        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Error Alert -->
                    <Alert
                        v-if="error"
                        variant="destructive"
                        class="bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 mb-4"
                    >
                        <AlertCircle class="h-4 w-4" />
                        <AlertDescription class="text-sm ml-2">
                            <strong class="font-semibold">Error:</strong>
                            {{ error }}
                        </AlertDescription>
                    </Alert>
                </CardContent>
            </Card>

            <!-- Use Cases -->
            <Card
                class="mt-6 border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20"
            >
                <CardContent class="p-4">
                    <h3
                        class="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-3 flex items-center gap-2"
                    >
                        <Sparkles class="w-4 h-4" />
                        Perfect for Research & Study
                    </h3>
                    <div class="grid md:grid-cols-2 gap-3 text-sm">
                        <div
                            class="flex items-start gap-2 text-blue-800 dark:text-blue-300"
                        >
                            <BookOpen class="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>Extract quotes from research papers</span>
                        </div>
                        <div
                            class="flex items-start gap-2 text-blue-800 dark:text-blue-300"
                        >
                            <FileText class="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>Digitize handwritten notes</span>
                        </div>
                        <div
                            class="flex items-start gap-2 text-blue-800 dark:text-blue-300"
                        >
                            <Search class="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>Copy text from lecture slides</span>
                        </div>
                        <div
                            class="flex items-start gap-2 text-blue-800 dark:text-blue-300"
                        >
                            <MessageSquare
                                class="w-4 h-4 mt-0.5 flex-shrink-0"
                            />
                            <span>Reference textbook passages</span>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>

        <!-- Slide Panel Overlay -->
        <Transition
            enter-active-class="transition-opacity duration-300"
            leave-active-class="transition-opacity duration-300"
            enter-from-class="opacity-0"
            leave-to-class="opacity-0"
        >
            <div
                v-if="resultsOpen"
                @click="closeResults"
                class="fixed inset-0 bg-black/50 backdrop-blur-sm z-40"
            />
        </Transition>

        <!-- Slide Panel - Half Width Desktop, Full Width Mobile -->
        <Transition
            enter-active-class="transition-transform duration-300 ease-out"
            leave-active-class="transition-transform duration-300 ease-in"
            enter-from-class="translate-x-full"
            leave-to-class="translate-x-full"
        >
            <div
                v-if="resultsOpen"
                class="fixed inset-y-0 right-0 w-full md:w-1/2 bg-white dark:bg-gray-900 shadow-2xl z-50 overflow-y-auto"
            >
                <div
                    class="sticky top-0 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 z-10"
                >
                    <div class="flex items-center justify-between p-4">
                        <div class="flex-1 min-w-0">
                            <h2
                                class="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2"
                            >
                                <Sparkles
                                    class="w-5 h-5 text-purple-600 flex-shrink-0"
                                />
                                <span class="truncate">Extracted Text</span>
                            </h2>
                            <p
                                class="text-sm text-gray-600 dark:text-gray-400 mt-1"
                            >
                                {{ wordCount }} words •
                                {{ characterCount }} characters • ~{{
                                    estimatedReadTime
                                }}
                                min read
                            </p>
                        </div>
                        <button
                            @click="closeResults"
                            class="ml-4 p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors flex-shrink-0"
                        >
                            <X
                                class="w-5 h-5 text-gray-600 dark:text-gray-400"
                            />
                        </button>
                    </div>

                    <!-- Action Buttons -->
                    <div class="flex flex-wrap gap-2 px-4 pb-4">
                        <Button
                            @click="copyToClipboard"
                            variant="outline"
                            size="sm"
                        >
                            <Check v-if="copied" class="w-4 h-4 mr-2" />
                            <Copy v-else class="w-4 h-4 mr-2" />
                            {{ copied ? "Copied!" : "Copy" }}
                        </Button>
                        <Button
                            @click="downloadAsText"
                            variant="outline"
                            size="sm"
                        >
                            <Download class="w-4 h-4 mr-2" />
                            Download
                        </Button>
                        <Button @click="shareText" variant="outline" size="sm">
                            <Share2 class="w-4 h-4 mr-2" />
                            Share
                        </Button>
                        <Button
                            @click="analyzeWithGemmie"
                            variant="default"
                            size="sm"
                            class="ml-auto bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700"
                        >
                            <MessageSquare class="w-4 h-4 mr-2" />
                            Analyze
                        </Button>
                    </div>
                </div>

                <!-- Extracted Text Content -->
                <div class="p-4">
                    <Card>
                        <CardContent class="p-6">
                            <div
                                class="prose dark:prose-invert max-w-none text-sm leading-relaxed"
                            >
                                <p
                                    class="whitespace-pre-wrap text-gray-800 dark:text-gray-200"
                                >
                                    {{ extractedData.text }}
                                </p>
                            </div>
                        </CardContent>
                    </Card>

                    <!-- Processing Info -->
                    <div
                        v-if="extractedData.processing_ms"
                        class="mt-4 text-xs text-gray-500 dark:text-gray-400 text-center"
                    >
                        Processed in {{ extractedData.processing_ms }}ms
                    </div>
                </div>
            </div>
        </Transition>
    </div>
</template>
