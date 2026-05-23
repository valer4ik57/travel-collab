<template>
  <section class="card panel itinerary-page">
    <div class="section-head">
      <div>
        <h2>Маршрут по дням</h2>
        <p class="muted">Здесь собран план поездки по дням. Точки можно перетаскивать за ручку, чтобы менять порядок маршрута.</p>
      </div>
    </div>

    <div class="route-tabs compact-tabs">
      <button type="button" :class="selectedRouteId === 'all' ? 'active' : ''" @click="$emit('selectedRouteChange', 'all')">Все</button>
      <button type="button" :class="selectedRouteId === '' ? 'active' : ''" @click="$emit('selectedRouteChange', '')">Без маршрута</button>
      <button v-for="route in sortedRoutes" :key="route.id" type="button" :class="selectedRouteId === route.id ? 'active' : ''" @click="$emit('selectedRouteChange', route.id)">
        {{ route.title }}
      </button>
    </div>

    <p v-if="error" class="error inline-notice">{{ error }}</p>
    <p v-if="orderError" class="error inline-notice">{{ orderError }}</p>

    <div v-if="sortedRoutes.length === 0" class="subcard empty-state">
      <h3>Маршрутов пока нет</h3>
      <p class="muted">Создайте первый маршрут на вкладке “Карта”, затем добавьте точки и настройте порядок посещения.</p>
    </div>

    <article v-for="route in visibleRoutes" :key="route.id" class="subcard route-day-card">
      <div class="day-head">
        <div>
          <h3>{{ route.title }}</h3>
          <p class="muted">{{ route.route_date ? formatDate(route.route_date) : 'Дата не указана' }}</p>
        </div>
        <div class="status-stack small">
          <span class="badge">{{ routeLocations(route.id).length }} точек</span>
          <button type="button" class="button small secondary" :disabled="routeLocations(route.id).length < 2" @click="openRouteIn2gis(route.id)">2ГИС</button>
        </div>
      </div>

      <div v-if="routeLocations(route.id).length" class="ordered-list">
        <article
          v-for="(location, index) in routeLocations(route.id)"
          :key="location.id"
          class="ordered-item draggable-route-item"
          :class="{ 'is-dragging': draggingLocationId === location.id, 'is-drag-over': dragOverLocationId === location.id }"
          :data-route-id="route.id"
          :data-drop-location-id="location.id"
          :draggable="canEdit && !orderSaving"
          @dragstart="startNativeDrag(route.id, location.id, $event)"
          @dragover.prevent="markDragOver(route.id, location.id)"
          @drop.prevent="dropOnLocation(route.id, location.id)"
          @dragend="cancelDrag"
        >
          <button
            type="button"
            class="drag-handle"
            :disabled="!canEdit || orderSaving"
            title="Перетащить точку"
            @pointerdown="startPointerDrag(route.id, location.id, $event)"
            @click.stop
          >
            ⋮⋮
          </button>
          <span class="order-number">{{ index + 1 }}</span>
          <div class="ordered-content">
            <strong>{{ location.name }}</strong>
            <small>{{ timeOnly(location.visit_at) }} · {{ categoryLabel(location.category) }}</small>
            <small v-if="location.description" class="muted">{{ location.description }}</small>
          </div>
          <div class="actions compact-actions">
            <button type="button" class="button tiny secondary" :disabled="!canEdit || orderSaving || index === 0" @click="moveRouteLocation(route.id, index, -1)">↑</button>
            <button type="button" class="button tiny secondary" :disabled="!canEdit || orderSaving || index === routeLocations(route.id).length - 1" @click="moveRouteLocation(route.id, index, 1)">↓</button>
            <button type="button" class="button tiny" :disabled="!canEdit" @click="$emit('addExpense', location)">Расход</button>
          </div>
        </article>
      </div>
      <p v-else class="empty">В этом маршруте пока нет точек.</p>
    </article>

    <article v-if="showUnassigned" class="subcard route-day-card">
      <div class="day-head">
        <div>
          <h3>Точки без маршрута</h3>
          <p class="muted">Эти точки не входят в дневные маршруты и не участвуют в построении линии маршрута.</p>
        </div>
        <span class="badge">{{ unassignedLocations.length }} точек</span>
      </div>
      <div v-if="unassignedLocations.length" class="ordered-list">
        <div v-for="location in unassignedLocations" :key="location.id" class="ordered-item">
          <span class="order-number">•</span>
          <div class="ordered-content">
            <strong>{{ location.name }}</strong>
            <small>{{ timeOnly(location.visit_at) }} · {{ categoryLabel(location.category) }}</small>
          </div>
          <button type="button" class="button tiny" :disabled="!canEdit" @click="$emit('addExpense', location)">Расход</button>
        </div>
      </div>
      <p v-else class="empty">Нет точек без маршрута.</p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { http } from '../api/http'
import type { LocationPoint, TripRoute } from '../types'

const props = defineProps<{
  tripId: string
  routes: TripRoute[]
  locations: LocationPoint[]
  selectedRouteId: string
  canEdit: boolean
}>()
const emit = defineEmits<{
  selectedRouteChange: [routeId: string]
  locationsReordered: [locations: LocationPoint[]]
  addExpense: [location: LocationPoint]
}>()

const error = ref('')
const orderError = ref('')
const orderSaving = ref(false)
const draggingLocationId = ref('')
const draggingRouteId = ref('')
const dragOverLocationId = ref('')

const categories = [
  { value: 'sight', label: 'Достопримечательность' },
  { value: 'food', label: 'Еда' },
  { value: 'hotel', label: 'Отель' },
  { value: 'transport', label: 'Транспорт' },
  { value: 'other', label: 'Другое' }
]

const sortedRoutes = computed(() => [...props.routes].sort(compareRoutes))
const unassignedLocations = computed(() => props.locations.filter((location) => !location.route_id).sort(compareLocations))
const visibleRoutes = computed(() => {
  if (props.selectedRouteId && props.selectedRouteId !== 'all') return sortedRoutes.value.filter((route) => route.id === props.selectedRouteId)
  return sortedRoutes.value
})
const showUnassigned = computed(() => props.selectedRouteId === 'all' || props.selectedRouteId === '')

function routeLocations(routeID: string) {
  return props.locations.filter((location) => location.route_id === routeID).sort(compareLocations)
}



async function moveRouteLocation(routeID: string, index: number, direction: -1 | 1) {
  const items = routeLocations(routeID)
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= items.length) return
  const reordered = [...items]
  const [item] = reordered.splice(index, 1)
  reordered.splice(targetIndex, 0, item)
  await saveRouteOrder(routeID, reordered)
}

function startNativeDrag(routeID: string, locationID: string, event: DragEvent) {
  if (!props.canEdit || orderSaving.value) {
    event.preventDefault()
    return
  }
  draggingRouteId.value = routeID
  draggingLocationId.value = locationID
  dragOverLocationId.value = locationID
  event.dataTransfer?.setData('text/plain', locationID)
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function markDragOver(routeID: string, locationID: string) {
  if (draggingRouteId.value !== routeID || !draggingLocationId.value) return
  dragOverLocationId.value = locationID
}

async function dropOnLocation(routeID: string, targetID: string) {
  if (draggingRouteId.value !== routeID || !draggingLocationId.value || draggingLocationId.value === targetID) {
    cancelDrag()
    return
  }
  await reorderByLocationIDs(routeID, draggingLocationId.value, targetID)
  cancelDrag()
}

function startPointerDrag(routeID: string, locationID: string, event: PointerEvent) {
  if (!props.canEdit || orderSaving.value || event.pointerType === 'mouse') return
  draggingRouteId.value = routeID
  draggingLocationId.value = locationID
  dragOverLocationId.value = locationID
  event.preventDefault()
  window.addEventListener('pointermove', handlePointerMove, { passive: false })
  window.addEventListener('pointerup', handlePointerUp, { once: true })
  window.addEventListener('pointercancel', handlePointerCancel, { once: true })
}

function handlePointerMove(event: PointerEvent) {
  if (!draggingLocationId.value || !draggingRouteId.value) return
  event.preventDefault()
  const element = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null
  const target = element?.closest('[data-drop-location-id]') as HTMLElement | null
  if (!target || target.dataset.routeId !== draggingRouteId.value) return
  dragOverLocationId.value = target.dataset.dropLocationId || ''
}

async function handlePointerUp() {
  window.removeEventListener('pointermove', handlePointerMove)
  const routeID = draggingRouteId.value
  const sourceID = draggingLocationId.value
  const targetID = dragOverLocationId.value
  if (routeID && sourceID && targetID && sourceID !== targetID) await reorderByLocationIDs(routeID, sourceID, targetID)
  cancelDrag()
}

function handlePointerCancel() {
  window.removeEventListener('pointermove', handlePointerMove)
  cancelDrag()
}

async function reorderByLocationIDs(routeID: string, sourceID: string, targetID: string) {
  const reordered = routeLocations(routeID)
  const sourceIndex = reordered.findIndex((location) => location.id === sourceID)
  const targetIndex = reordered.findIndex((location) => location.id === targetID)
  if (sourceIndex < 0 || targetIndex < 0) return
  const [item] = reordered.splice(sourceIndex, 1)
  reordered.splice(targetIndex, 0, item)
  await saveRouteOrder(routeID, reordered)
}

async function saveRouteOrder(routeID: string, reordered: LocationPoint[]) {
  orderError.value = ''
  orderSaving.value = true
  try {
    const { data } = await http.post<LocationPoint[]>(`/trips/${props.tripId}/routes/${routeID}/locations/reorder`, {
      location_ids: reordered.map((location) => location.id)
    })
    emit('locationsReordered', data)
  } catch (e: any) {
    orderError.value = formatApiError(e, 'Не удалось изменить порядок точек')
  } finally {
    orderSaving.value = false
  }
}

function cancelDrag() {
  draggingLocationId.value = ''
  draggingRouteId.value = ''
  dragOverLocationId.value = ''
}

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', handlePointerMove)
})

function openRouteIn2gis(routeID: string) {
  error.value = ''
  const locations = routeLocations(routeID)
  if (locations.length < 2) {
    error.value = 'Для построения маршрута в 2ГИС добавьте минимум две точки.'
    return
  }
  const points = locations.slice(0, 10).map((location) => `${location.lng},${location.lat}`).join('|')
  window.open(`https://2gis.ru/directions/tab/pedestrian/points/${points}`, '_blank', 'noopener,noreferrer')
}

function compareRoutes(a: TripRoute, b: TripRoute) {
  const aDate = a.route_date ? new Date(a.route_date).getTime() : Number.POSITIVE_INFINITY
  const bDate = b.route_date ? new Date(b.route_date).getTime() : Number.POSITIVE_INFINITY
  if (aDate !== bDate) return aDate - bDate
  if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
  return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
}

function compareLocations(a: LocationPoint, b: LocationPoint) {
  const aOrder = a.route_order ?? Number.POSITIVE_INFINITY
  const bOrder = b.route_order ?? Number.POSITIVE_INFINITY
  if (a.route_id && b.route_id && a.route_id === b.route_id && aOrder !== bOrder) return aOrder - bOrder
  const aTime = a.visit_at ? new Date(a.visit_at).getTime() : Number.POSITIVE_INFINITY
  const bTime = b.visit_at ? new Date(b.visit_at).getTime() : Number.POSITIVE_INFINITY
  if (aTime !== bTime) return aTime - bTime
  return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
}

function categoryLabel(category: string) {
  return categories.find((item) => item.value === category)?.label || 'Другое'
}

function formatDate(value?: string | null) {
  if (!value) return 'Без даты'
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value))
}

function timeOnly(value?: string | null) {
  if (!value) return 'Без времени'
  return new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

function formatApiError(e: any, fallback: string) {
  const data = e.response?.data
  if (!data?.error) return fallback
  if (!data.details) return data.error
  const details = Object.entries(data.details).map(([key, value]) => `${key}: ${value}`).join(', ')
  return `${data.error}. ${details}`
}
</script>
