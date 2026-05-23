<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink class="brand" to="/">Travel-Collab</RouterLink>
      <nav>
        <RouterLink v-if="auth.isAuthenticated" to="/trips">Мои поездки</RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/login">Вход</RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/register">Создать аккаунт</RouterLink>

        <button v-if="auth.isAuthenticated" class="user-chip" type="button" @click="openProfile">
          <img
            v-if="avatarVisible"
            class="avatar-img small"
            :src="auth.user?.avatar_url || ''"
            alt=""
            @error="avatarBroken = true"
          />
          <span v-else class="avatar small">{{ userInitial }}</span>
          <span>
            <strong>{{ auth.user?.display_name || 'Пользователь' }}</strong>
            <small>{{ auth.user?.email || 'локальный аккаунт' }}</small>
          </span>
        </button>
        <button v-if="auth.isAuthenticated" class="link-button" @click="logout">Выйти</button>
      </nav>
    </header>

    <main>
      <RouterView />
    </main>

    <div v-if="profileOpen" class="modal-backdrop" @click.self="profileOpen = false">
      <form class="modal card form" @submit.prevent="saveProfile">
        <h2>Профиль</h2>
        <label>Отображаемое имя<input v-model="profileForm.display_name" required maxlength="100" /></label>
        <label>Ссылка на аватар<input v-model="profileForm.avatar_url" placeholder="https://example.com/avatar.png" /></label>
        <p class="hint">Пока используется ссылка на картинку. Если она не открывается напрямую, вместо аватарки будет показана первая буква имени.</p>
        <p v-if="profileError" class="error">{{ profileError }}</p>
        <p v-if="profileSaved" class="success">Профиль обновлён</p>
        <div class="actions right">
          <button type="button" class="button secondary" @click="profileOpen = false">Закрыть</button>
          <button class="button" :disabled="profileSaving">Сохранить</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const router = useRouter()
const profileOpen = ref(false)
const profileSaving = ref(false)
const profileError = ref('')
const profileSaved = ref(false)
const avatarBroken = ref(false)
const profileForm = reactive({ display_name: '', avatar_url: '' })

const userInitial = computed(() => (auth.user?.display_name || '?').slice(0, 1).toUpperCase())
const avatarVisible = computed(() => Boolean(auth.user?.avatar_url) && !avatarBroken.value)

watch(() => auth.user?.avatar_url, () => {
  avatarBroken.value = false
})

function openProfile() {
  profileError.value = ''
  profileSaved.value = false
  profileForm.display_name = auth.user?.display_name || ''
  profileForm.avatar_url = auth.user?.avatar_url || ''
  profileOpen.value = true
}

async function saveProfile() {
  profileSaving.value = true
  profileError.value = ''
  profileSaved.value = false
  try {
    await auth.updateMe(profileForm.display_name, profileForm.avatar_url || null)
    profileSaved.value = true
  } catch (e: any) {
    profileError.value = e.response?.data?.error || 'Не удалось обновить профиль'
  } finally {
    profileSaving.value = false
  }
}

function logout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  if (auth.isAuthenticated && !auth.user) {
    auth.loadMe().catch(() => undefined)
  }
})
</script>
