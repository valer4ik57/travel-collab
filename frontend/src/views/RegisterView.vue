<template>
  <section class="auth-page card narrow">
    <h1>Регистрация</h1>
    <p class="muted">Создай аккаунт, чтобы начать планировать поездки.</p>
    <form @submit.prevent="submit" class="form">
      <label>Имя<input v-model="displayName" required placeholder="Например, Алексей" /></label>
      <label>Email<input v-model="email" type="email" required placeholder="you@example.com" /></label>
      <label>Пароль<input v-model="password" type="password" required minlength="6" /></label>
      <button class="button" :disabled="loading">{{ loading ? 'Создаём...' : 'Создать аккаунт' }}</button>
    </form>
    <p v-if="error" class="error">{{ error }}</p>
    <p class="muted">Уже есть аккаунт? <RouterLink to="/login">Войти</RouterLink></p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const displayName = ref('')
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
    await auth.register(email.value, password.value, displayName.value)
    router.push('/trips')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Не удалось зарегистрироваться'
  } finally {
    loading.value = false
  }
}
</script>
