import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DeleteAccView from '../views/DeleteAccView.vue'
import UpgradeView from '@/views/UpgradeView.vue'
import SettingsView from '@/views/SettingsView.vue'
import EditorView from '@/views/EditorView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/editor',
      name: 'editor',
      component: EditorView
    },
    {
      path: '/auth/delete_account',
      name: 'delete_account',
      component: DeleteAccView
    },
    {
      path: '/upgrade/:plan?',
      name: 'upgrade',
      component: UpgradeView,
      props: true
    },
    {
      path: '/settings/:tab',
      name: 'settings',
      component: SettingsView
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

export default router
