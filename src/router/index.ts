import { createRouter, createWebHistory } from "vue-router";
import DeleteAccView from "../views/DeleteAccView.vue";
import UpgradeView from "@/views/UpgradeView.vue";
import SettingsView from "@/views/SettingsView.vue";
import WorkplaceView from "@/views/WorkplaceView.vue";
import ChatsView from "@/views/ChatsView.vue";
import type { UserDetails } from "@/types";
import LegalPage from "@/views/LegalPage.vue";
import CreateSessView from "@/views/CreateSessView.vue";
import ChatView from "@/views/ChatView.vue";
import ArcadeView from "@/views/ArcadeView.vue";
import SingleArcade from "@/views/SingleArcade.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/chat/:id?",
      name: "chat",
      component: ChatView,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: "/",
      name: "authentication",
      component: CreateSessView,
      meta: { requiresAuth: false },
    },
    {
      path: "/new",
      name: "new_chat",
      component: ChatView,
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
      path: "/arcade",
      name: "arcade",
      component: ArcadeView,
      meta: { requiresAuth: true },
    },
    {
      path: "/arcade/:id",
      name: "single-arcade",
      props: true,
      component: SingleArcade,
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
        return isAuthenticated ? "/" : "/";
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
      path: "/",
      query: { redirect: to.fullPath },
    });
  } else if ((to.path === "/" || to.path === "/") && isAuthenticated) {
    // Redirect authenticated users to intended destination or home
    const redirectPath = (to.query.redirect as string) || "/new";
    // Prevent redirect loops
    next(redirectPath === "/" || to.path === "/" ? "/new" : redirectPath);
  } else if (to.path === "/new" && to.query.redirect && isAuthenticated) {
    // Handle redirect on home page for authenticated users
    const redirectPath = to.query.redirect as string;
    if (redirectPath && redirectPath !== "/new" && redirectPath !== "/") {
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
