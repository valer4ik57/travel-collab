<template>
  <section class="auth-page card narrow">
    <h1>Вход</h1>
    <p class="muted">Войди по email и паролю, чтобы открыть свои поездки.</p>
    <form @submit.prevent="submit" class="form">
      <label>Email<input v-model="email" type="email" required placeholder="you@example.com" /></label>
      <label>Пароль<input v-model="password" type="password" required minlength="6" /></label>
      <button class="button" :disabled="loading">{{ loading ? 'Входим...' : 'Войти' }}</button>
    </form>
    <p v-if="error" class="error">{{ error }}</p>
    <p class="muted">Нет аккаунта? <RouterLink to="/register">Зарегистрироваться</RouterLink></p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    router.push('/trips')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Не удалось войти'
  } finally {
    loading.value = false
  }
}

</script>
