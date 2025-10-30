import { ref } from "vue";

export function useSync() {
  const syncStatus = ref({
    lastSync: null as Date | null,
    syncing: false,
    hasUnsyncedChanges: false,
    lastError: null as string | null,
    retryCount: 0,
    maxRetries: 3,
    showSyncIndicator: false,
    syncMessage: "",
    syncProgress: 0,
  });

  function showSyncIndicator(message: string, progress: number = 0) {
    syncStatus.value.showSyncIndicator = true;
    syncStatus.value.syncMessage = message;
    syncStatus.value.syncProgress = progress;
  }

  function hideSyncIndicator() {
    syncStatus.value.showSyncIndicator = false;
    syncStatus.value.syncMessage = "";
    syncStatus.value.syncProgress = 0;
  }

  function updateSyncProgress(message: string, progress: number) {
    syncStatus.value.syncMessage = message;
    syncStatus.value.syncProgress = progress;
  }

  return {
    syncStatus,
    showSyncIndicator,
    hideSyncIndicator,
    updateSyncProgress,
  };
}
