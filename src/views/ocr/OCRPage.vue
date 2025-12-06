<script setup lang="ts">
import { inject, ref, type Ref } from "vue";
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
    Image as ImageIcon,
    Sparkles,
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

const onFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];

    if (file) {
        if (!validateImage(file)) return;

        captureImage(file);
        previewUrl.value = URL.createObjectURL(file);
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

const handleReset = () => {
    reset();
    previewUrl.value = null;
    showCropper.value = false;
    copied.value = false;
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
                    Image Text Extractor
                </h1>
                <p
                    class="text-gray-600 dark:text-gray-400 text-sm md:text-base"
                >
                    Upload an image or take a photo to extract text using OCR
                </p>
            </div>

            <!-- Main Card -->
            <Card class="overflow-hidden shadow-lg bg-gray-50 dark:bg-gray-800">
                <CardHeader
                    v-if="!previewUrl"
                    class="border-b border-gray-200 dark:border-gray-700"
                >
                    <CardTitle class="flex items-center gap-2">
                        <Camera class="w-5 h-5" />
                        Upload Image
                    </CardTitle>
                    <CardDescription>
                        Select an image file or capture from your camera
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
                                class="text-sm text-gray-500 dark:text-gray-400"
                            >
                                PNG, JPG, WebP • Max 4MB
                            </span>
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
                    <Alert v-if="error" variant="destructive" class="mb-4">
                        <AlertDescription class="text-sm">
                            {{ error }}
                        </AlertDescription>
                    </Alert>

                    <!-- Extracted Text -->
                    <div
                        v-if="extractedData.text"
                        class="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700"
                    >
                        <div class="flex items-center justify-between mb-3">
                            <h3
                                class="text-sm font-semibold text-gray-900 dark:text-white flex items-center gap-2"
                            >
                                <Sparkles
                                    class="w-4 h-4 text-purple-600 dark:text-purple-400"
                                />
                                Extracted Text
                            </h3>
                            <Button
                                @click="copyToClipboard"
                                size="sm"
                                variant="outline"
                                class="text-xs"
                            >
                                <Check v-if="copied" class="w-3 h-3 mr-1" />
                                <Copy v-else class="w-3 h-3 mr-1" />
                                {{ copied ? "Copied!" : "Copy" }}
                            </Button>
                        </div>
                        <div
                            class="bg-white dark:bg-gray-900 rounded-lg p-4 border border-gray-200 dark:border-gray-700 max-h-64 overflow-y-auto"
                        >
                            <p
                                class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap leading-relaxed"
                            >
                                {{ extractedData.text }}
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <!-- Tips -->
            <Card
                class="mt-6 border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20"
            >
                <CardContent class="p-4">
                    <h3
                        class="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-2 flex items-center gap-2"
                    >
                        <Sparkles class="w-4 h-4" />
                        Tips for better results
                    </h3>
                    <ul
                        class="text-sm text-blue-800 dark:text-blue-300 space-y-1"
                    >
                        <li>• Ensure good lighting and clear text</li>
                        <li>• Avoid blurry or tilted images</li>
                        <li>• Use crop to focus on specific text areas</li>
                        <li>• Supported formats: PNG, JPG, WebP</li>
                    </ul>
                </CardContent>
            </Card>
        </div>
    </div>
</template>
