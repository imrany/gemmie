import { createApp, reactive } from 'vue'
import router from './router'
import App from './App.vue'
import 'primeicons/primeicons.css';
import './assets/main.css'
import { ConfirmationService } from 'primevue';
import SideNav from './components/SideNav.vue';
import UpdateModal from './components/Modals/UpdateModal.vue';

// reactive state for update modal
export const updateState = reactive({
  showModal: false,
  registration: null as ServiceWorkerRegistration | null
})

export function showUpdateModal(registration: ServiceWorkerRegistration) {
  updateState.showModal = true
  updateState.registration = registration
}

export function hideUpdateModal() {
  updateState.showModal = false
}

export function installUpdate() {
  console.log('⬇️ installUpdate called', updateState.registration);
  if (updateState.registration?.waiting) {
    // Send message to service worker to skip waiting
    updateState.registration.waiting.postMessage({ type: 'SKIP_WAITING' })
    console.log('📤 Sent SKIP_WAITING message to service worker');
    
    // Reload the page once the new service worker takes control
    let refreshing = false
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      console.log('🔄 Controller changed, reloading...');
      if (!refreshing) {
        refreshing = true
        window.location.reload()
      }
    })
  } else {
    console.warn('⚠️ No waiting service worker found');
  }
  hideUpdateModal()
}

const app = createApp(App)
app.use(ConfirmationService)
app.use(router)
app.component('SideNav', SideNav)
app.component('UpdateModal', UpdateModal)
app.mount('#app')

console.log('✅ Vue app mounted');

// Service Worker Registration
if ('serviceWorker' in navigator) {
  console.log('🔧 Service Worker supported, registering...');
  
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/serviceworker.js')
      .then(registration => {
        console.log('✅ ServiceWorker registration successful with scope:', registration.scope);
        console.log('📋 Registration state:', {
          installing: registration.installing,
          waiting: registration.waiting,
          active: registration.active,
          controller: navigator.serviceWorker.controller
        });
        
        // Check for updates periodically (every 30 minutes)
        setInterval(() => {
          console.log('🔍 Checking for updates...');
          registration.update()
        }, 30 * 60 * 1000)
        
        // Listen for updatefound event
        registration.addEventListener('updatefound', () => {
          console.log('🆕 Update found!');
          const newWorker = registration.installing;
          
          if (newWorker) {
            console.log('👷 New worker installing...');
            newWorker.addEventListener('statechange', () => {
              console.log('📊 New worker state:', newWorker.state);
              if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                console.log('✨ New service worker is installed and ready!');
                showUpdateModal(registration)
              }
            })
          }
        })
        
        // Check if there's already a waiting worker
        if (registration.waiting && navigator.serviceWorker.controller) {
          console.log('⏳ Found waiting service worker on page load');
          showUpdateModal(registration)
        }
      })
      .catch(error => {
        console.error('❌ ServiceWorker registration failed:', error);
      });
  });
} else {
  console.warn('⚠️ Service Worker not supported in this browser');
}