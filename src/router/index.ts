import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import DeleteAccView from "../views/DeleteAccView.vue";
import UpgradeView from "@/views/UpgradeView.vue";
import SettingsView from "@/views/SettingsView.vue";
import WorkplaceView from "@/views/WorkplaceView.vue";
import ChatsView from "@/views/ChatsView.vue";
import type { UserDetails } from "@/types";
import LegalPage from "@/views/LegalPage.vue";
import CreateSessView from "@/views/CreateSessView.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
      meta: { requiresAuth: true },
    },
    {
      path: "/auth",
      name: "authentication",
      component: CreateSessView,
      meta: { requiresAuth: false },
    },
    {
      path: "/new",
      name: "new_chat",
      component: HomeView,
      meta: { requiresAuth: true },
    },
    {
      path: "/chats",
      name: "chats",
      component: ChatsView,
      meta: { requiresAuth: true },
    },
    {
      path: "/workplace",
      name: "workplace",
      component: WorkplaceView,
      meta: { requiresAuth: true },
    },
    {
      path: "/auth/delete_account",
      name: "delete_account",
      component: DeleteAccView,
      meta: { requiresAuth: true },
    },
    {
      path: "/upgrade/:plan?",
      name: "upgrade",
      component: UpgradeView,
      props: true,
      meta: { requiresAuth: false },
    },
    {
      path: "/settings/:tab?",
      name: "settings",
      component: SettingsView,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: "/legal/:section?",
      name: "legal",
      component: LegalPage,
      meta: { title: "Legal - Gemmie", requiresAuth: false },
    },
    {
      path: "/:pathMatch(.*)*",
      name: "not-found",
      redirect: () => {
        // Smart redirect for 404s based on auth status
        const isAuthenticated = checkAuthStatus();
        return isAuthenticated ? "/" : "/auth";
      },
    },
  ],
});

// Simple navigation guard
router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth);
  const isAuthenticated = checkAuthStatus();

  if (requiresAuth && !isAuthenticated) {
    // Redirect to auth with intended destination
    next({
      path: "/auth",
      query: { redirect: to.fullPath },
    });
  } else if (to.path === "/auth" && isAuthenticated) {
    // Redirect authenticated users to intended destination or home
    const redirectPath = (to.query.redirect as string) || "/";
    // Prevent redirect loops
    next(redirectPath === "/auth" ? "/" : redirectPath);
  } else if (to.path === "/" && to.query.redirect && isAuthenticated) {
    // Handle redirect on home page for authenticated users
    const redirectPath = to.query.redirect as string;
    if (redirectPath && redirectPath !== "/" && redirectPath !== "/auth") {
      next(redirectPath);
    } else {
      next();
    }
  } else {
    // Allow navigation
    next();
  }
});

// Simple authentication check
function checkAuthStatus(): boolean {
  try {
    const userDetails = localStorage.getItem("userdetails");
    if (!userDetails) return false;

    const parsedDetails: UserDetails = JSON.parse(userDetails);
    return !!(
      parsedDetails?.username &&
      parsedDetails?.email &&
      parsedDetails?.userId
    );
  } catch (error) {
    console.error("Auth check failed:", error);
    return false;
  }
}

export default router;
