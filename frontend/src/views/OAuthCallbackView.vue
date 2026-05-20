<template>
  <section class="card narrow">
    <h1>Завершаем вход...</h1>
    <p class="muted">Если редирект не произошёл автоматически, открой список поездок вручную.</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(async () => {
  const token = String(route.query.token || '')
  if (token) {
    auth.setSession(token)
    await auth.loadMe().catch(() => {})
    router.replace('/trips')
  } else {
    router.replace('/login')
  }
})
</script>
