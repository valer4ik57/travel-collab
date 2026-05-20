<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink class="brand" to="/">Travel-Collab</RouterLink>
      <nav>
        <RouterLink v-if="auth.isAuthenticated" to="/trips">Поездки</RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/login">Вход</RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/register">Регистрация</RouterLink>
        <button v-if="auth.isAuthenticated" class="link-button" @click="logout">Выйти</button>
      </nav>
    </header>
    <main>
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>
