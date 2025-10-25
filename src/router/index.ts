import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import DeleteAccView from "../views/DeleteAccView.vue";
import UpgradeView from "@/views/UpgradeView.vue";
import SettingsView from "@/views/SettingsView.vue";
import WorkplaceView from "@/views/WorkplaceView.vue";
import ChatsView from "@/views/ChatsView.vue";
import type { UserDetails } from "@/types";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
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
      path: "/:pathMatch(.*)*",
      redirect: "/",
    },
  ],
});

// Navigation guard to check authentication
router.beforeEach((to, from, next) => {
  // Check if route requires authentication
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth);

  // Get authentication status from localStorage
  const isAuthenticated = checkAuthStatus();

  if (requiresAuth && !isAuthenticated) {
    // User is not authenticated but trying to access protected route
    console.log("Access denied: Authentication required");

    // Redirect to home with return URL
    next({
      path: "/",
      query: {
        redirect: to.fullPath,
        from: (to.name as string) || "protected",
      },
    });
  } else if (to.path === "/" && isAuthenticated && to.query.redirect) {
    // User is authenticated and home has a redirect query
    // Allow them to see home (they might want to create new session)
    next();
  } else {
    // Allow navigation
    next();
  }
});

// Helper function to check authentication status
function checkAuthStatus(): boolean {
  try {
    // Check if user details exist in localStorage
    const userDetails = localStorage.getItem("userdetails");

    if (!userDetails) {
      return false;
    }

    // Parse and validate user details
    const parsedDetails: UserDetails = JSON.parse(userDetails);

    // Check if user has required authentication fields
    const hasValidAuth = !!(
      parsedDetails &&
      parsedDetails.username &&
      parsedDetails.email &&
      parsedDetails.userId
    );

    return hasValidAuth;
  } catch (error) {
    console.error("Error checking authentication status:", error);
    return false;
  }
}

// Optional: Export helper for use in components
export { checkAuthStatus };

export default router;
