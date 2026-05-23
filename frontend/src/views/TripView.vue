<template>
  <section v-if="loading" class="card"><p class="muted">Загрузка поездки...</p></section>
  <section v-else-if="details" class="trip-page mobile-app-layout">
    <div class="mobile-trip-appbar">
      <button type="button" class="mobile-menu-button" aria-label="Открыть меню" @click="mobileMenuOpen = true">☰</button>
      <div class="mobile-appbar-title">
        <strong>{{ activeTabLabel }}</strong>
        <small>{{ details.trip.name }}</small>
      </div>
      <span class="mobile-online-dot" :class="socket.connected.value ? 'ok' : 'warn'"></span>
    </div>

    <div v-if="mobileMenuOpen" class="mobile-menu-backdrop" @click.self="mobileMenuOpen = false">
      <aside class="mobile-side-menu">
        <div class="mobile-menu-head">
          <div>
            <strong>Travel-Collab</strong>
            <small>{{ details.trip.name }}</small>
          </div>
          <button type="button" class="mobile-menu-close" aria-label="Закрыть меню" @click="mobileMenuOpen = false">×</button>
        </div>

        <button type="button" class="mobile-invite-card" @click="copyInvite">
          <span>Invite-код</span>
          <code>{{ details.trip.invite_code }}</code>
        </button>

        <nav class="mobile-menu-nav">
          <button type="button" :class="mobileNavClass('map')" @click="setActiveTab('map')"><span>🗺️</span>Карта</button>
          <button type="button" :class="mobileNavClass('expenses')" @click="setActiveTab('expenses')"><span>💸</span>Расходы</button>
          <button type="button" :class="mobileNavClass('itinerary')" @click="setActiveTab('itinerary')"><span>🧭</span>Маршрут</button>
          <button type="button" :class="mobileNavClass('chat')" @click="setActiveTab('chat')"><span>💬</span>Чат</button>
          <button type="button" :class="mobileNavClass('members')" @click="setActiveTab('members')"><span>👥</span>Участники</button>
        </nav>

        <div class="mobile-menu-footer">
          <router-link to="/trips" @click="mobileMenuOpen = false">Мои поездки</router-link>
          <button type="button" @click="logoutFromTrip">Выйти</button>
        </div>
      </aside>
    </div>

    <div class="trip-header compact-trip-header card">
      <div class="trip-title-block">
        <div class="trip-title-row">
          <router-link to="/trips" class="back-link" title="Вернуться к списку поездок">←</router-link>
          <div class="trip-heading-text">
            <p class="eyebrow compact-eyebrow">Поездка</p>
            <h1>{{ details.trip.name }}</h1>
          </div>
        </div>
        <p class="trip-description">{{ details.trip.description || 'Описание пока не заполнено' }}</p>
      </div>

      <div class="trip-quick-meta">
        <span class="badge role-badge">{{ roleLabel(currentRole) }}</span>
        <span class="badge connection-badge" :class="socket.connected.value ? 'ok' : 'warn'">
          <span class="status-dot-mini"></span>
          {{ socket.connected.value ? 'online' : 'offline' }}
        </span>
        <button type="button" class="invite-chip" title="Скопировать invite-код" @click="copyInvite">
          <span>Invite</span>
          <code>{{ details.trip.invite_code }}</code>
        </button>
      </div>
    </div>

    <div v-if="notice" class="success inline-notice">{{ notice }}</div>
    <div v-if="error" class="error inline-notice">{{ error }}</div>

    <div class="tabs card desktop-tabs">
      <button :class="tabClass('map')" @click="setActiveTab('map')">Карта</button>
      <button :class="tabClass('expenses')" @click="setActiveTab('expenses')">Расходы</button>
      <button :class="tabClass('itinerary')" @click="setActiveTab('itinerary')">Маршрут</button>
      <button :class="tabClass('chat')" @click="setActiveTab('chat')">Чат</button>
      <button :class="tabClass('members')" @click="setActiveTab('members')">Участники</button>
    </div>

    <MapPanel
      v-if="activeTab === 'map'"
      :trip-id="details.trip.id"
      :routes="routes"
      :locations="locations"
      :selected-route-id="selectedRouteId"
      :can-edit="canEdit"
      :expense-reload-key="expensesReloadKey"
      @selected-route-change="selectedRouteId = $event"
      @route-created="addRoute"
      @route-updated="updateRoute"
      @route-deleted="removeRoute"
      @created="addLocation"
      @updated="updateLocation"
      @deleted="removeLocation"
      @locations-reordered="replaceLocations"
      @add-expense="startExpenseFromLocation"
    />

    <ExpensesPanel
      v-if="activeTab === 'expenses'"
      :trip-id="details.trip.id"
      :members="members"
      :routes="routes"
      :locations="locations"
      :selected-route-id="selectedRouteId"
      :reload-key="expensesReloadKey"
      :can-edit="canEdit"
      :preselected-location-id="expenseLocationId"
      @created="handleExpenseCreated"
    />

    <ItineraryPanel
      v-if="activeTab === 'itinerary'"
      :trip-id="details.trip.id"
      :routes="routes"
      :locations="locations"
      :selected-route-id="selectedRouteId"
      :can-edit="canEdit"
      @selected-route-change="selectedRouteId = $event"
      @locations-reordered="replaceLocations"
      @add-expense="startExpenseFromLocation"
    />

    <ChatPanel
      v-if="activeTab === 'chat'"
      :messages="messages"
      :connected="socket.connected.value"
      :can-write="canEdit"
      @send="sendChat"
    />

    <MembersPanel
      v-if="activeTab === 'members'"
      :trip="details.trip"
      :members="members"
      :current-user-id="auth.user?.id || ''"
      :current-role="currentRole"
      @leave="leaveTrip"
      @remove="removeMember"
      @update-role="updateMemberRole"
    />
  </section>
  <section v-else class="card"><p class="error">{{ error || 'Поездка не найдена' }}</p></section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { http } from '../api/http'
import type { LocationPoint, Message, TripDetails, TripMember, TripRoute, WSEvent } from '../types'
import { useTripSocket } from '../composables/useTripSocket'
import { useAuthStore } from '../stores/auth'
import MapPanel from '../components/MapPanel.vue'
import MembersPanel from '../components/MembersPanel.vue'
import ExpensesPanel from '../components/ExpensesPanel.vue'
import ItineraryPanel from '../components/ItineraryPanel.vue'
import ChatPanel from '../components/ChatPanel.vue'

type TripTab = 'map' | 'expenses' | 'itinerary' | 'chat' | 'members'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const tripId = String(route.params.id)
const details = ref<TripDetails | null>(null)
const loading = ref(true)
const error = ref('')
const notice = ref('')
const messages = ref<Message[]>([])
const expensesReloadKey = ref(0)
const expenseLocationId = ref<string | null>(null)
const selectedRouteId = ref<string>('all')
const activeTab = ref<TripTab>('map')
const mobileMenuOpen = ref(false)

const locations = computed(() => details.value?.locations || [])
const routes = computed(() => details.value?.routes || [])
const members = computed(() => details.value?.members || [])
const currentMember = computed(() => members.value.find((m) => m.user_id === auth.user?.id) || null)
const currentRole = computed(() => currentMember.value?.role || 'viewer')
const canEdit = computed(() => currentRole.value === 'owner' || currentRole.value === 'editor')

const tabLabels: Record<TripTab, string> = {
  map: 'Карта',
  expenses: 'Расходы',
  itinerary: 'Маршрут',
  chat: 'Чат',
  members: 'Участники'
}
const activeTabLabel = computed(() => tabLabels[activeTab.value])

function roleLabel(role: string) {
  if (role === 'owner') return 'владелец'
  if (role === 'editor') return 'редактор'
  return 'только просмотр'
}

async function copyInvite() {
  if (!details.value?.trip.invite_code) return
  await navigator.clipboard.writeText(details.value.trip.invite_code)
  notice.value = 'Invite-код скопирован'
  setTimeout(() => {
    if (notice.value === 'Invite-код скопирован') notice.value = ''
  }, 1600)
}

function tabClass(tab: TripTab) {
  return ['tab-button', activeTab.value === tab ? 'active' : '']
}

function mobileNavClass(tab: TripTab) {
  return ['mobile-menu-link', activeTab.value === tab ? 'active' : '']
}

function setActiveTab(tab: TripTab) {
  activeTab.value = tab
  mobileMenuOpen.value = false
}

function logoutFromTrip() {
  mobileMenuOpen.value = false
  auth.logout()
  router.push('/login')
}

function handleEvent(event: WSEvent) {
  if (!details.value) return
  if (event.type === 'LOCATION_ADDED') addLocation(event.payload as LocationPoint)
  if (event.type === 'LOCATION_UPDATED') updateLocation(event.payload as LocationPoint)
  if (event.type === 'LOCATION_DELETED') removeLocation((event.payload as { id: string }).id)
  if (event.type === 'LOCATIONS_REORDERED') replaceLocations(event.payload as LocationPoint[])
  if (event.type === 'ROUTE_ADDED') addRoute(event.payload as TripRoute)
  if (event.type === 'ROUTE_UPDATED') updateRoute(event.payload as TripRoute)
  if (event.type === 'ROUTE_DELETED') removeRoute((event.payload as { id: string }).id)
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
  if (expenseLocationId.value === id) expenseLocationId.value = null
}

function replaceLocations(newLocations: LocationPoint[]) {
  if (!details.value) return
  details.value.locations = newLocations
}

function addRoute(route: TripRoute) {
  if (!details.value) return
  if (!details.value.routes.some((r) => r.id === route.id)) details.value.routes.push(route)
  selectedRouteId.value = route.id
}

function updateRoute(route: TripRoute) {
  if (!details.value) return
  const index = details.value.routes.findIndex((r) => r.id === route.id)
  if (index >= 0) details.value.routes[index] = route
}

function removeRoute(routeID: string) {
  if (!details.value) return
  details.value.routes = details.value.routes.filter((r) => r.id !== routeID)
  if (selectedRouteId.value === routeID) selectedRouteId.value = 'all'
}

function startExpenseFromLocation(location: LocationPoint) {
  expenseLocationId.value = location.id
  selectedRouteId.value = location.route_id || ''
  activeTab.value = 'expenses'
}

function handleExpenseCreated() {
  expensesReloadKey.value++
  expenseLocationId.value = null
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
    details.value = {
      ...tripResponse.data,
      routes: tripResponse.data.routes || []
    }
    messages.value = messagesResponse.data
    socket.connect()
  } catch (e: any) {
    error.value = formatApiError(e, 'Не удалось загрузить поездку')
  } finally {
    loading.value = false
  }
}

function formatApiError(e: any, fallback: string) {
  const data = e.response?.data
  if (!data?.error) return fallback
  if (!data.details) return data.error
  const details = Object.entries(data.details).map(([key, value]) => `${key}: ${value}`).join(', ')
  return `${data.error}. ${details}`
}

onMounted(() => {
  document.body.classList.add('trip-mobile-shell')
  load()
})

onBeforeUnmount(() => {
  document.body.classList.remove('trip-mobile-shell')
  socket.close()
})

watch(activeTab, () => {
  mobileMenuOpen.value = false
})
</script>
