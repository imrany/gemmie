import { createRouter, createWebHistory } from "vue-router";
import { Workplace } from "@/views/workplace";
import type { UserDetails } from "@/types";
import { Settings } from "@/views/settings";
import { Upgrade } from "@/views/upgrade";
import { Chat } from "@/views/chats/chat";
import { Chats } from "@/views/chats";
import LegalPage from "@/views/legal/LegalPage.vue";
import { Arcades } from "@/views/arcades";
import { Arcade } from "@/views/arcades/arcade";
import { CreateSession } from "@/views/auth";
import { DeleteAccount } from "@/views/DeleteAccount";
import OCRPage from "@/views/ocr/OCRPage.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/chat/:id?",
      name: "chat",
      component: Chat,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: "/",
      name: "authentication",
      component: CreateSession,
      meta: { requiresAuth: false },
    },
    {
      path: "/new",
      name: "new_chat",
      component: Chat,
      meta: { requiresAuth: true },
    },
    {
      path: "/chats",
      name: "chats",
      component: Chats,
      meta: { requiresAuth: true },
    },
    {
      path: "/workplace",
      name: "workplace",
      component: Workplace,
      meta: { requiresAuth: true },
    },
    {
      path: "/arcade",
      name: "arcade",
      component: Arcades,
      meta: { requiresAuth: true },
    },
    {
      path: "/arcade/:id",
      name: "single-arcade",
      props: true,
      component: Arcade,
    },
    {
      path: "/auth/delete_account",
      name: "delete_account",
      component: DeleteAccount,
      meta: { requiresAuth: true },
    },
    {
      path: "/upgrade/:plan?",
      name: "upgrade",
      component: Upgrade,
      props: true,
      meta: { requiresAuth: false },
    },
    {
      path: "/settings/:tab?",
      name: "settings",
      component: Settings,
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: "/ocr",
      name: "ocr",
      component: OCRPage,
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
