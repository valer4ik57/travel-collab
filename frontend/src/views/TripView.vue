<template>
  <section v-if="loading" class="card"><p class="muted">Загрузка поездки...</p></section>
  <section v-else-if="details" class="trip-page">
    <div class="trip-header card">
      <div>
        <p class="eyebrow">Поездка</p>
        <h1>{{ details.trip.name }}</h1>
        <p class="muted">{{ details.trip.description || 'Описание пока не заполнено' }}</p>
      </div>
      <div class="status-stack">
        <span class="badge" :class="socket.connected.value ? 'ok' : 'warn'">
          {{ socket.connected.value ? 'WebSocket подключен' : 'WebSocket отключён' }}
        </span>
        <code class="invite">{{ details.trip.invite_code }}</code>
      </div>
    </div>

    <div class="workspace">
      <MapPanel
        :trip-id="details.trip.id"
        :locations="locations"
        @created="addLocation"
        @updated="updateLocation"
        @deleted="removeLocation"
      />
      <aside class="side-panels">
        <MembersPanel :trip="details.trip" :members="members" />
        <ExpensesPanel :trip-id="details.trip.id" :members="members" :reload-key="expensesReloadKey" />
        <ChatPanel :messages="messages" :connected="socket.connected.value" @send="sendChat" />
      </aside>
    </div>
  </section>
  <section v-else class="card"><p class="error">{{ error || 'Поездка не найдена' }}</p></section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { http } from '../api/http'
import type { Expense, LocationPoint, Message, TripDetails, TripMember, WSEvent } from '../types'
import { useTripSocket } from '../composables/useTripSocket'
import MapPanel from '../components/MapPanel.vue'
import MembersPanel from '../components/MembersPanel.vue'
import ExpensesPanel from '../components/ExpensesPanel.vue'
import ChatPanel from '../components/ChatPanel.vue'

const route = useRoute()
const tripId = String(route.params.id)
const details = ref<TripDetails | null>(null)
const loading = ref(true)
const error = ref('')
const messages = ref<Message[]>([])
const expensesReloadKey = ref(0)

const locations = computed(() => details.value?.locations || [])
const members = computed(() => details.value?.members || [])

function handleEvent(event: WSEvent) {
  if (!details.value) return
  if (event.type === 'LOCATION_ADDED') {
    addLocation(event.payload as LocationPoint)
  }
  if (event.type === 'LOCATION_UPDATED') {
    updateLocation(event.payload as LocationPoint)
  }
  if (event.type === 'LOCATION_DELETED') {
    const payload = event.payload as { id: string }
    removeLocation(payload.id)
  }
  if (event.type === 'MESSAGE_SENT') {
    messages.value.push(event.payload as Message)
  }
  if (event.type === 'MEMBER_JOINED') {
    const member = event.payload as TripMember
    if (!details.value.members.some((m) => m.user_id === member.user_id)) {
      details.value.members.push(member)
    }
  }
  if (event.type === 'EXPENSE_ADDED') {
    expensesReloadKey.value++
  }
}

const socket = useTripSocket(tripId, handleEvent)

function addLocation(location: LocationPoint) {
  if (!details.value) return
  if (!details.value.locations.some((l) => l.id === location.id)) {
    details.value.locations.push(location)
  }
}

function updateLocation(location: LocationPoint) {
  if (!details.value) return
  const index = details.value.locations.findIndex((l) => l.id === location.id)
  if (index >= 0) details.value.locations[index] = location
}

function removeLocation(id: string) {
  if (!details.value) return
  details.value.locations = details.value.locations.filter((l) => l.id !== id)
}

function sendChat(text: string) {
  socket.send('SEND_MESSAGE', { text })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [tripResponse, messagesResponse] = await Promise.all([
      http.get<TripDetails>(`/trips/${tripId}`),
      http.get<Message[]>(`/trips/${tripId}/messages`)
    ])
    details.value = tripResponse.data
    messages.value = messagesResponse.data
    socket.connect()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Не удалось загрузить поездку'
  } finally {
    loading.value = false
  }
}

onMounted(load)
onBeforeUnmount(() => socket.close())
</script>
