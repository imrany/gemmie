import type { Ref } from "vue";
import { toast } from "vue-sonner/src/packages/state.js";

export function useVoiceRecord({
  voiceRecognition,
  isRecording,
  isTranscribing,
  transcribedText,
  microphonePermission,
  autoGrow,
  transcriptionDuration,
  updateTimeout,
  transcriptionTimer,
}: {
  voiceRecognition: Ref<any | null>;
  isRecording: Ref<boolean>;
  isTranscribing: Ref<boolean>;
  transcribedText: Ref<string>;
  microphonePermission: Ref<string | boolean>;
  autoGrow: (data: any) => void;
  transcriptionDuration: Ref<number>;
  updateTimeout: number | null;
  transcriptionTimer: number | null;
}) {
  // Initialize Speech Recognition (add this in the onMounted hook)
  function initializeSpeechRecognition() {
    const SpeechRecognition =
      (window as any).SpeechRecognition ||
      (window as any).webkitSpeechRecognition;
    if (!SpeechRecognition) {
      console.warn("Speech recognition not supported");
      return;
    }

    const recognition = new SpeechRecognition();
    recognition.continuous = true;
    recognition.interimResults = true;
    recognition.lang = "en-US";
    recognition.maxAlternatives = 1;

    recognition.onresult = (event: any) => {
      // Only process if we're still recording (FIX 6)
      if (!isRecording.value) return;

      let interimTranscript = "";
      for (let i = event.resultIndex; i < event.results.length; i++) {
        const transcript = event.results[i][0].transcript;
        if (event.results[i].isFinal) {
          transcribedText.value += transcript + " ";
        } else {
          interimTranscript += transcript;
        }
      }
      updateTextarea(interimTranscript);
    };

    recognition.onerror = (event: any) => {
      console.error("Speech recognition error:", event.error);
      isRecording.value = false;
      isTranscribing.value = false;
      microphonePermission.value =
        event.error === "not-allowed" ? "denied" : microphonePermission.value;

      if (event.error !== "aborted") {
        // Don't show toast for manual stops
        toast.error("Voice Input Error", {
          duration: 4000,
          description:
            event.error === "not-allowed"
              ? "Microphone access denied"
              : event.error,
        });
      }
    };

    recognition.onend = () => {
      // Only restart if we're still supposed to be recording (FIX 5)
      if (isRecording.value && !isTranscribing.value) {
        setTimeout(() => {
          if (isRecording.value) {
            // Double check we're still recording
            recognition.start();
          }
        }, 500);
      } else {
        isTranscribing.value = false;
      }
    };

    recognition.onstart = () => {
      isTranscribing.value = true;
    };

    voiceRecognition.value = recognition;
  }
  // Toggle voice recording
  async function toggleVoiceRecording() {
    if (!voiceRecognition.value) {
      toast.error("Speech recognition not available", {
        duration: 3000,
        description: "Your browser may not support speech recognition.",
      });
      return;
    }

    if (isRecording.value) {
      stopVoiceRecording(false); // Don't clear text - let user decide
    } else {
      await startVoiceRecording();
    }
  }

  // Start voice recording
  async function startVoiceRecording() {
    try {
      await navigator.mediaDevices.getUserMedia({ audio: true });
      microphonePermission.value = "granted";
      isRecording.value = true;
      transcribedText.value = "";
      transcriptionDuration.value = 0;
      startTimer();

      const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
      if (textarea) {
        textarea.value = "";
        autoGrow({ target: textarea } as any);
      }

      voiceRecognition.value?.start();
    } catch (error) {
      microphonePermission.value = "denied";
      isRecording.value = false;
      toast.error("Microphone Access Denied", {
        duration: 5000,
        description: "Please allow microphone access.",
      });
    }
  }
  function updateTextarea(interim: string) {
    if (updateTimeout) clearTimeout(updateTimeout);
    updateTimeout = window.setTimeout(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
      if (textarea && (isRecording.value || transcribedText.value)) {
        // Only update if we're actively recording or have transcribed text
        textarea.value = transcribedText.value + interim;
        autoGrow({ target: textarea } as any);
      }
    }, 100);
  }

  // Stop voice recording
  function stopVoiceRecording(clearText: boolean = true) {
    isRecording.value = false;
    isTranscribing.value = false;
    stopTimer();
    voiceRecognition.value?.stop();

    // Clear transcribed text if requested (FIX 2 & 5)
    if (clearText) {
      clearVoiceTranscription();
    }
  }

  function startTimer() {
    transcriptionTimer = window.setInterval(() => {
      transcriptionDuration.value += 1;
    }, 1000);
  }

  function stopTimer() {
    if (transcriptionTimer) clearInterval(transcriptionTimer);
  }

  // Clear voice transcription
  function clearVoiceTranscription() {
    transcribedText.value = "";
    transcriptionDuration.value = 0; // Reset duration
    const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
    if (textarea) {
      textarea.value = "";
      autoGrow({ target: textarea } as any);
      textarea.focus();
    }
  }

  return {
    startVoiceRecording,
    stopVoiceRecording,
    clearVoiceTranscription,
    startTimer,
    stopTimer,
    toggleVoiceRecording,
    initializeSpeechRecognition,
  };
}
