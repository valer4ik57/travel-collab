<template>
  <section class="page-grid">
    <div class="card">
      <h1>Мои поездки</h1>
      <p class="muted">Создай новую поездку или вступи в поездку по invite-коду.</p>

      <div class="split">
        <form @submit.prevent="createTrip" class="form subcard">
          <h2>Новая поездка</h2>
          <label>Название<input v-model="newTrip.name" required placeholder="Например, Санкт-Петербург" /></label>
          <label>Описание<textarea v-model="newTrip.description" rows="3" placeholder="Коротко о маршруте" /></label>
          <div class="two-cols">
            <label>Начало<input v-model="newTrip.start_date" type="date" /></label>
            <label>Конец<input v-model="newTrip.end_date" type="date" /></label>
          </div>
          <p v-if="createError" class="error">{{ createError }}</p>
          <button class="button" :disabled="creating">Создать</button>
        </form>

        <form @submit.prevent="joinTrip" class="form subcard">
          <h2>Вступить по коду</h2>
          <label>Invite-код<input v-model="inviteCode" required placeholder="ABCD1234" /></label>
          <p v-if="joinError" class="error">{{ joinError }}</p>
          <button class="button secondary" :disabled="joining">Вступить</button>
        </form>
      </div>
    </div>

    <div class="card">
      <div class="section-head">
        <h2>Список поездок</h2>
        <button class="button small secondary" @click="loadTrips">Обновить</button>
      </div>
      <p v-if="loading" class="muted">Загрузка...</p>
      <p v-else-if="trips.length === 0" class="empty">Пока поездок нет. Создай первую поездку слева.</p>
      <div v-else class="trip-list">
        <RouterLink v-for="trip in trips" :key="trip.id" class="trip-card" :to="`/trips/${trip.id}`">
          <div>
            <h3>{{ trip.name }}</h3>
            <p>{{ trip.description || 'Описание не заполнено' }}</p>
          </div>
          <div class="trip-meta">
            <span>{{ trip.members_count || 1 }} участников</span>
            <span>{{ trip.locations_count || 0 }} точек</span>
            <code>{{ trip.invite_code }}</code>
          </div>
        </RouterLink>
      </div>
      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { http } from '../api/http'
import type { Trip } from '../types'

const trips = ref<Trip[]>([])
const loading = ref(false)
const creating = ref(false)
const joining = ref(false)
const error = ref('')
const createError = ref('')
const joinError = ref('')
const inviteCode = ref('')
const router = useRouter()
const newTrip = reactive({ name: '', description: '', start_date: '', end_date: '' })

async function loadTrips() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await http.get<Trip[]>('/trips')
    trips.value = data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Не удалось загрузить поездки'
  } finally {
    loading.value = false
  }
}

async function createTrip() {
  creating.value = true
  createError.value = ''
  try {
    const payload = {
      name: newTrip.name,
      description: newTrip.description || null,
      start_date: newTrip.start_date || null,
      end_date: newTrip.end_date || null
    }
    const { data } = await http.post<Trip>('/trips', payload)
    router.push(`/trips/${data.id}`)
  } catch (e: any) {
    createError.value = e.response?.data?.error || 'Не удалось создать поездку'
  } finally {
    creating.value = false
  }
}

async function joinTrip() {
  joining.value = true
  joinError.value = ''
  try {
    const { data } = await http.post('/trips/join', { invite_code: inviteCode.value })
    router.push(`/trips/${data.trip.id}`)
  } catch (e: any) {
    if (e.response?.status === 404) joinError.value = 'Код приглашения не найден'
    else if (e.response?.status === 403) joinError.value = 'Владелец удалил вас из этой поездки. Повторный вход по коду недоступен.'
    else joinError.value = e.response?.data?.error || 'Не удалось вступить в поездку'
  } finally {
    joining.value = false
  }
}

onMounted(loadTrips)
</script>
