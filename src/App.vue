<script setup lang="ts">
import { ConfirmDialog } from 'primevue';
import { Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
</script>

<template>
  <Toaster position="top-right" :closeButton="true" closeButtonPosition="top-left"/>
  <!-- Option 1: Using PrimeVue ConfirmDialog -->
  <ConfirmDialog />
  <RouterView/>
</template>
