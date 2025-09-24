import { createApp } from 'vue'
import router from './router'
import App from './App.vue'
import 'primeicons/primeicons.css';
import './assets/main.css'
import { ConfirmationService } from 'primevue';          
import SideNav from './components/SideNav.vue';

const app = createApp(App)
app.use(ConfirmationService)
app.use(router)
app.component('SideNav', SideNav)
app.mount('#app')


if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/serviceworker.js') 
      .then(registration => {
        console.log('ServiceWorker registration successful with scope: ', registration.scope);
      })
      .catch(error => {
        console.log('ServiceWorker registration failed: ', error);
      });
  });
}