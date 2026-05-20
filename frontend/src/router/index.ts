import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import TripsView from '../views/TripsView.vue'
import TripView from '../views/TripView.vue'
import OAuthCallbackView from '../views/OAuthCallbackView.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/login', component: LoginView },
    { path: '/register', component: RegisterView },
    { path: '/oauth-callback', component: OAuthCallbackView },
    { path: '/trips', component: TripsView, meta: { requiresAuth: true } },
    { path: '/trips/:id', component: TripView, meta: { requiresAuth: true } }
  ]
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return '/login'
  }
})

export default router
