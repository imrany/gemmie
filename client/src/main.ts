import { createApp } from 'vue'
import router from './router'
import App from './App.vue'
import 'primeicons/primeicons.css';
import './assets/main.css'
// import 'vue-slick-carousel/dist/vue-slick-carousel.css'
import 'vue-slick-carousel/dist/vue-slick-carousel-theme.css'
import "vue-slick-ts/dist/css/slick.css";

const app = createApp(App)
app.use(router)
app.mount('#app')