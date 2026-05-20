<template>
  <section v-if="loading" class="card"><p class="muted">Загрузка поездки...</p></section>
  <section v-else-if="details" class="trip-page">
    <div class="trip-header card">
      <div>
        <p class="eyebrow">Поездка</p>
        <h1>{{ details.trip.name }}</h1>
        <p class="muted">{{ details.trip.description || 'Описание пока не заполнено' }}</p>
        <span class="badge">Моя роль: {{ roleLabel(currentRole) }}</span>
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
        :can-edit="canEdit"
        @created="addLocation"
        @updated="updateLocation"
        @deleted="removeLocation"
      />
      <aside class="side-panels">
        <MembersPanel
          :trip="details.trip"
          :members="members"
          :current-user-id="auth.user?.id || ''"
          :current-role="currentRole"
          @leave="leaveTrip"
          @remove="removeMember"
          @update-role="updateMemberRole"
        />
        <ExpensesPanel :trip-id="details.trip.id" :members="members" :reload-key="expensesReloadKey" :can-edit="canEdit" />
        <ChatPanel :messages="messages" :connected="socket.connected.value" :can-write="canEdit" @send="sendChat" />
      </aside>
    </div>
  </section>
  <section v-else class="card"><p class="error">{{ error || 'Поездка не найдена' }}</p></section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../api/http'
import type { LocationPoint, Message, TripDetails, TripMember, WSEvent } from '../types'
import { useTripSocket } from '../composables/useTripSocket'
import { useAuthStore } from '../stores/auth'
import MapPanel from '../components/MapPanel.vue'
import MembersPanel from '../components/MembersPanel.vue'
import ExpensesPanel from '../components/ExpensesPanel.vue'
import ChatPanel from '../components/ChatPanel.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const tripId = String(route.params.id)
const details = ref<TripDetails | null>(null)
const loading = ref(true)
const error = ref('')
const messages = ref<Message[]>([])
const expensesReloadKey = ref(0)

const locations = computed(() => details.value?.locations || [])
const members = computed(() => details.value?.members || [])
const currentMember = computed(() => members.value.find((m) => m.user_id === auth.user?.id) || null)
const currentRole = computed(() => currentMember.value?.role || 'viewer')
const canEdit = computed(() => currentRole.value === 'owner' || currentRole.value === 'editor')

function roleLabel(role: string) {
  if (role === 'owner') return 'владелец'
  if (role === 'editor') return 'редактор'
  return 'только просмотр'
}

function handleEvent(event: WSEvent) {
  if (!details.value) return
  if (event.type === 'LOCATION_ADDED') addLocation(event.payload as LocationPoint)
  if (event.type === 'LOCATION_UPDATED') updateLocation(event.payload as LocationPoint)
  if (event.type === 'LOCATION_DELETED') removeLocation((event.payload as { id: string }).id)
  if (event.type === 'MESSAGE_SENT') messages.value.push(event.payload as Message)
  if (event.type === 'MEMBER_JOINED') {
    const member = event.payload as TripMember
    if (!details.value.members.some((m) => m.user_id === member.user_id)) details.value.members.push(member)
  }
  if (event.type === 'MEMBER_UPDATED') {
    const member = event.payload as TripMember
    const index = details.value.members.findIndex((m) => m.user_id === member.user_id)
    if (index >= 0) details.value.members[index] = member
  }
  if (event.type === 'MEMBER_REMOVED' || event.type === 'MEMBER_LEFT') {
    const userID = (event.payload as { user_id: string }).user_id
    details.value.members = details.value.members.filter((m) => m.user_id !== userID)
    if (userID === auth.user?.id) router.push('/trips')
  }
  if (event.type === 'EXPENSE_ADDED') expensesReloadKey.value++
}

const socket = useTripSocket(tripId, handleEvent)

function addLocation(location: LocationPoint) {
  if (!details.value) return
  if (!details.value.locations.some((l) => l.id === location.id)) details.value.locations.push(location)
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
  if (!canEdit.value) return
  socket.send('SEND_MESSAGE', { text })
}

async function leaveTrip() {
  await http.post(`/trips/${tripId}/leave`)
  router.push('/trips')
}

async function removeMember(userID: string) {
  await http.delete(`/trips/${tripId}/members/${userID}`)
  if (details.value) details.value.members = details.value.members.filter((m) => m.user_id !== userID)
}

async function updateMemberRole(userID: string, role: TripMember['role']) {
  const { data } = await http.patch<TripMember>(`/trips/${tripId}/members/${userID}`, { role })
  if (!details.value) return
  const index = details.value.members.findIndex((m) => m.user_id === userID)
  if (index >= 0) details.value.members[index] = data
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    await auth.loadMe()
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
