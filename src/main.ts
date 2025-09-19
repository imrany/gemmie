import { createApp } from 'vue'
import router from './router'
import App from './App.vue'
import 'primeicons/primeicons.css';
import './assets/main.css'
import { ConfirmationService } from 'primevue';          

const app = createApp(App)
app.use(ConfirmationService)
app.use(router)
app.mount('#app')
