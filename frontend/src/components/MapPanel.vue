<template>
  <section class="card map-card">
    <div class="section-head map-head">
      <div>
        <h2>Карта маршрута</h2>
        <p class="muted">
          {{ canEdit ? 'Выберите маршрут дня и добавляйте точки кликом по карте. Линия строится по ручному порядку точек.' : 'У вас роль viewer, карта доступна только для просмотра.' }}
        </p>
      </div>
      <div class="actions">
        <button type="button" class="button small secondary" :disabled="activeRouteLocations.length < 2" @click="openRouteMapChooser">
          Открыть маршрут на карте
        </button>
      </div>
    </div>

    <div class="map-layout">
      <aside class="route-drawer subcard">
        <div class="drawer-head">
          <h3>Маршруты</h3>
          <button type="button" class="button small secondary" @click="showRouteForm = !showRouteForm" v-if="canEdit">
            {{ showRouteForm ? 'Скрыть' : '+ Создать' }}
          </button>
        </div>
        <button type="button" :class="routeButtonClass('all')" @click="selectRoute('all')">
          <span>Все точки поездки</span>
          <small>{{ locations.length }}</small>
        </button>
        <button type="button" :class="routeButtonClass('')" @click="selectRoute('')">
          <span>Без маршрута</span>
          <small>{{ unassignedLocations.length }}</small>
        </button>
        <button v-for="route in sortedRoutes" :key="route.id" type="button" :class="routeButtonClass(route.id)" @click="selectRoute(route.id)">
          <span>{{ route.title }}</span>
          <small>{{ route.route_date ? formatDate(route.route_date) : 'Без даты' }} · {{ countLocations(route.id) }} точек</small>
        </button>

        <form v-if="showRouteForm && canEdit" class="mini-form" @submit.prevent="createRoute">
          <label>Название маршрута<input v-model="routeDraft.title" placeholder="Например, День 1 — центр" /></label>
          <label>Дата<input v-model="routeDraft.route_date" type="date" /></label>
          <p v-if="routeError" class="error">{{ routeError }}</p>
          <button class="button small" :disabled="routeSaving">Создать маршрут</button>
        </form>

        <form v-if="selectedRoute && canEdit" class="mini-form" @submit.prevent="updateSelectedRoute">
          <h4>Настройки выбранного маршрута</h4>
          <label>Название<input v-model="editRouteDraft.title" /></label>
          <label>Дата<input v-model="editRouteDraft.route_date" type="date" /></label>
          <p v-if="editRouteError" class="error">{{ editRouteError }}</p>
          <div class="actions compact-actions">
            <button class="button tiny" :disabled="editRouteSaving">Сохранить</button>
            <button type="button" class="button tiny danger" :disabled="editRouteSaving" @click="deleteSelectedRoute">Удалить</button>
          </div>
        </form>

        <div v-if="selectedRoute" class="mini-form route-order-panel">
          <div class="drawer-head">
            <h4>Порядок точек</h4>
            <span class="badge">{{ activeRouteLocations.length }}</span>
          </div>
          <p class="hint">Меняйте порядок здесь и сразу смотрите на карту. Линия и внешние карты используют этот список сверху вниз.</p>
          <div v-if="activeRouteLocations.length" class="route-order-list">
            <article
              v-for="(location, index) in activeRouteLocations"
              :key="location.id"
              class="route-order-row draggable-route-item"
              :class="{ 'is-dragging': draggingLocationId === location.id, 'is-drag-over': dragOverLocationId === location.id }"
              :data-route-id="selectedRoute?.id"
              :data-drop-location-id="location.id"
              :draggable="canEdit && !orderSaving"
              @click="focusLocation(location)"
              @dragstart="startNativeDrag(location.id, $event)"
              @dragover.prevent="markDragOver(location.id)"
              @drop.prevent="dropOnLocation(location.id)"
              @dragend="cancelDrag"
            >
              <button
                v-if="canEdit"
                type="button"
                class="drag-handle compact-drag-handle"
                :disabled="orderSaving"
                title="Перетащить точку"
                @pointerdown="startPointerDrag(location.id, $event)"
                @click.stop
              >
                ⋮⋮
              </button>
              <span class="order-number mini">{{ index + 1 }}</span>
              <div class="ordered-content">
                <strong>{{ location.name }}</strong>
                <small>{{ formatVisit(location.visit_at) }} · {{ categoryLabel(location.category) }}</small>
              </div>
              <div v-if="canEdit" class="actions compact-actions" @click.stop>
                <button type="button" class="button tiny secondary" :disabled="orderSaving || index === 0" @click="moveRouteLocation(index, -1)">↑</button>
                <button type="button" class="button tiny secondary" :disabled="orderSaving || index === activeRouteLocations.length - 1" @click="moveRouteLocation(index, 1)">↓</button>
                <button type="button" class="button tiny" @click="$emit('addExpense', location)">Расход</button>
              </div>
            </article>
          </div>
          <p v-else class="empty">В выбранном маршруте пока нет точек. Кликните по карте, чтобы добавить первую.</p>
          <p v-if="orderError" class="error">{{ orderError }}</p>
        </div>
      </aside>

      <div class="map-wrap" :class="{ editable: canEdit && selectedRouteId !== 'all' }">
        <div ref="mapEl" class="map"></div>
      </div>
    </div>

    <div v-if="panelError" class="error inline-notice">{{ panelError }}</div>

    <div class="route-legend">
      <span v-for="category in categories" :key="category.value" class="category-pill">
        <span class="category-dot">{{ category.icon }}</span>{{ category.label }}
      </span>
    </div>

    <div class="locations-strip">
      <article
        v-for="location in filteredLocations"
        :key="location.id"
        class="location-chip"
        @click="focusLocation(location)"
      >
        <strong>{{ markerNumber(location) }}{{ location.name }}</strong>
        <small>{{ categoryLabel(location.category) }}</small>
        <small>{{ location.route_title || 'Без маршрута' }}</small>
        <small>{{ formatVisit(location.visit_at) }}</small>
        <small v-if="locationExpenseTotal(location.id) > 0">Расходы: {{ money(locationExpenseTotal(location.id)) }}</small>
      </article>
      <p v-if="filteredLocations.length === 0" class="empty">Точек в выбранном режиме пока нет. {{ canEdit && selectedRouteId !== 'all' ? 'Кликните по карте, чтобы добавить первую.' : '' }}</p>
    </div>

    <p class="hint">
      Линия на карте строится только для выбранного маршрута и использует ручной порядок точек. В режиме “Все точки” показываются все маркеры без общей линии.
    </p>

    <div v-if="mapChooser" class="modal-backdrop" @click.self="closeMapChooser">
      <div class="modal card map-open-modal">
        <h2>{{ mapChooser.mode === 'route' ? 'Открыть маршрут на карте' : 'Открыть точку на карте' }}</h2>
        <p class="muted">Выберите приложение или сервис, в котором нужно открыть {{ mapChooser.mode === 'route' ? 'выбранный маршрут' : 'точку' }}.</p>
        <div class="map-provider-list">
          <button type="button" class="button secondary" @click="openChosenMap('google')">Google Maps</button>
          <button type="button" class="button secondary" @click="openChosenMap('yandex')">Яндекс Карты</button>
          <button type="button" class="button secondary" @click="openChosenMap('2gis')">2ГИС</button>
          <button v-if="mapChooser.mode === 'point'" type="button" class="button secondary" @click="openChosenMap('system')">Системная карта</button>
        </div>
        <p v-if="mapChooser.mode === 'route'" class="hint">Для маршрута передается ручной порядок точек из выбранного дня. Внешний сервис сам строит дорогу между ними.</p>
        <div class="actions right">
          <button type="button" class="button secondary" @click="closeMapChooser">Закрыть</button>
        </div>
      </div>
    </div>

    <div v-if="draft" class="modal-backdrop" @click.self="closeDraft">
      <form class="modal card form" @submit.prevent="saveDraft">
        <h2>{{ editingId ? 'Редактировать точку' : 'Добавить точку' }}</h2>
        <p v-if="draft.route_id" class="success inline-notice">Точка будет в маршруте: {{ routeTitle(draft.route_id) }}</p>
        <p v-else class="hint">Точка будет общей, без маршрута.</p>
        <label>Название<input v-model="draft.name" required /></label>
        <label>Описание<textarea v-model="draft.description" rows="3" /></label>
        <label>Маршрут
          <select v-model="draft.route_id">
            <option value="">Без маршрута</option>
            <option v-for="route in sortedRoutes" :key="route.id" :value="route.id">
              {{ route.title }}{{ route.route_date ? ` — ${formatDate(route.route_date)}` : '' }}
            </option>
          </select>
        </label>
        <label>Категория
          <select v-model="draft.category">
            <option v-for="category in categories" :key="category.value" :value="category.value">
              {{ category.icon }} {{ category.label }}
            </option>
          </select>
        </label>
        <label>Дата и время посещения
          <input v-model="draft.visit_at" type="datetime-local" />
        </label>
        <p class="hint">Время необязательно. Порядок маршрута меняется в панели слева — прямо рядом с картой.</p>
        <p v-if="draftError" class="error">{{ draftError }}</p>
        <div class="actions right">
          <button type="button" class="button secondary" @click="closeDraft">Отмена</button>
          <button class="button" :disabled="draftSaving">Сохранить</button>
        </div>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import L from 'leaflet'

import { http } from '../api/http'
import type { Expense, ExpensesSummary, LocationPoint, TripRoute } from '../types'

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png'
})

const props = defineProps<{
  tripId: string
  routes: TripRoute[]
  locations: LocationPoint[]
  selectedRouteId: string
  canEdit: boolean
  expenseReloadKey: number
}>()
const emit = defineEmits<{
  selectedRouteChange: [routeId: string]
  routeCreated: [route: TripRoute]
  routeUpdated: [route: TripRoute]
  routeDeleted: [id: string]
  created: [location: LocationPoint]
  updated: [location: LocationPoint]
  deleted: [id: string]
  locationsReordered: [locations: LocationPoint[]]
  addExpense: [location: LocationPoint]
}>()

const categories = [
  { value: 'sight', label: 'Достопримечательность', icon: '🏛️' },
  { value: 'food', label: 'Еда', icon: '🍽️' },
  { value: 'hotel', label: 'Отель', icon: '🏨' },
  { value: 'transport', label: 'Транспорт', icon: '🚆' },
  { value: 'other', label: 'Другое', icon: '📍' }
]

const mapEl = ref<HTMLDivElement | null>(null)
const summary = ref<ExpensesSummary | null>(null)
const showRouteForm = ref(false)
const routeSaving = ref(false)
const routeError = ref('')
const panelError = ref('')
const draftError = ref('')
const draftSaving = ref(false)
const routeDraft = reactive({ title: '', route_date: '' })
const editRouteDraft = reactive({ title: '', route_date: '' })
const editRouteSaving = ref(false)
const editRouteError = ref('')
const orderSaving = ref(false)
const orderError = ref('')
const draggingLocationId = ref('')
const dragOverLocationId = ref('')
let map: L.Map | null = null
let routeLine: L.Polyline | null = null
const markers = new Map<string, L.Marker>()
const draft = ref<null | { name: string; description: string; category: string; lat: number; lng: number; visit_at: string; route_id: string }>(null)
type MapProvider = 'google' | 'yandex' | '2gis' | 'system'
const editingId = ref<string | null>(null)
const mapChooser = ref<null | { mode: 'route' | 'point'; location?: LocationPoint }>(null)

const sortedRoutes = computed(() => [...props.routes].sort(compareRoutes))
const selectedRoute = computed(() => props.routes.find((route) => route.id === props.selectedRouteId) || null)
const unassignedLocations = computed(() => props.locations.filter((location) => !location.route_id).sort(compareLocations))
const filteredLocations = computed(() => {
  if (props.selectedRouteId === 'all') return [...props.locations].sort(compareLocations)
  if (props.selectedRouteId === '') return unassignedLocations.value
  return props.locations.filter((location) => location.route_id === props.selectedRouteId).sort(compareLocations)
})
const activeRouteLocations = computed(() => props.selectedRouteId && props.selectedRouteId !== 'all' ? filteredLocations.value.filter((location) => Number.isFinite(location.lat) && Number.isFinite(location.lng)) : [])
const expensesByLocation = computed(() => {
  const result = new Map<string, Expense[]>()
  for (const expense of summary.value?.expenses || []) {
    if (!expense.location_id) continue
    const list = result.get(expense.location_id) || []
    list.push(expense)
    result.set(expense.location_id, list)
  }
  return result
})

onMounted(async () => {
  map = L.map(mapEl.value as HTMLDivElement).setView(center(props.locations), props.locations.length ? 11 : 5)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map)
  map.on('click', (event: L.LeafletMouseEvent) => {
    panelError.value = ''
    if (!props.canEdit) return
    if (props.selectedRouteId === 'all') {
      panelError.value = 'Перед добавлением точки выберите конкретный маршрут или режим “Без маршрута”. Так точка не потеряется среди всех дней поездки.'
      return
    }
    editingId.value = null
    draft.value = { name: '', description: '', category: 'other', lat: event.latlng.lat, lng: event.latlng.lng, visit_at: '', route_id: props.selectedRouteId || '' }
  })
  await loadExpenses()
  renderMapLayers()
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', handlePointerMove)
  map?.remove()
})

watch(() => props.locations, renderMapLayers, { deep: true })
watch(() => props.selectedRouteId, () => {
  renderMapLayers()
  if (filteredLocations.value.length) map?.fitBounds(filteredLocations.value.map((l) => [l.lat, l.lng] as [number, number]), { padding: [30, 30] })
})
watch(selectedRoute, (route) => {
  editRouteError.value = ''
  editRouteDraft.title = route?.title || ''
  editRouteDraft.route_date = toDateInput(route?.route_date)
}, { immediate: true })
watch(summary, renderMapLayers, { deep: true })
watch(() => props.expenseReloadKey, loadExpenses)

async function loadExpenses() {
  try {
    const { data } = await http.get<ExpensesSummary>(`/trips/${props.tripId}/expenses`)
    summary.value = data
  } catch {
    summary.value = null
  }
}

function selectRoute(routeId: string) {
  panelError.value = ''
  emit('selectedRouteChange', routeId)
}

async function createRoute() {
  routeError.value = ''
  if (!routeDraft.title.trim()) {
    routeError.value = 'Введите название маршрута'
    return
  }
  routeSaving.value = true
  try {
    const { data } = await http.post<TripRoute>(`/trips/${props.tripId}/routes`, {
      title: routeDraft.title.trim(),
      route_date: routeDraft.route_date || null
    })
    emit('routeCreated', data)
    routeDraft.title = ''
    routeDraft.route_date = ''
    showRouteForm.value = false
  } catch (e: any) {
    routeError.value = formatApiError(e, 'Не удалось создать маршрут')
  } finally {
    routeSaving.value = false
  }
}

async function updateSelectedRoute() {
  if (!selectedRoute.value) return
  editRouteError.value = ''
  if (!editRouteDraft.title.trim()) {
    editRouteError.value = 'Введите название маршрута'
    return
  }
  editRouteSaving.value = true
  try {
    const { data } = await http.put<TripRoute>(`/trips/${props.tripId}/routes/${selectedRoute.value.id}`, {
      title: editRouteDraft.title.trim(),
      route_date: editRouteDraft.route_date || null
    })
    emit('routeUpdated', data)
  } catch (e: any) {
    editRouteError.value = formatApiError(e, 'Не удалось обновить маршрут')
  } finally {
    editRouteSaving.value = false
  }
}

async function deleteSelectedRoute() {
  if (!selectedRoute.value) return
  editRouteError.value = ''
  if (!window.confirm(`Удалить маршрут «${selectedRoute.value.title}»? Это возможно только если в нём нет точек и расходов.`)) return
  editRouteSaving.value = true
  try {
    await http.delete(`/trips/${props.tripId}/routes/${selectedRoute.value.id}`)
    emit('routeDeleted', selectedRoute.value.id)
  } catch (e: any) {
    editRouteError.value = formatApiError(e, 'Не удалось удалить маршрут')
  } finally {
    editRouteSaving.value = false
  }
}

async function moveRouteLocation(index: number, direction: -1 | 1) {
  if (!selectedRoute.value) return
  const targetIndex = index + direction
  const items = [...activeRouteLocations.value]
  if (targetIndex < 0 || targetIndex >= items.length) return
  const [item] = items.splice(index, 1)
  items.splice(targetIndex, 0, item)
  await saveRouteOrder(items)
}

function startNativeDrag(locationID: string, event: DragEvent) {
  if (!props.canEdit || orderSaving.value || !selectedRoute.value) {
    event.preventDefault()
    return
  }
  draggingLocationId.value = locationID
  dragOverLocationId.value = locationID
  event.dataTransfer?.setData('text/plain', locationID)
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function markDragOver(locationID: string) {
  if (!draggingLocationId.value) return
  dragOverLocationId.value = locationID
}

async function dropOnLocation(targetID: string) {
  if (!draggingLocationId.value || draggingLocationId.value === targetID) {
    cancelDrag()
    return
  }
  await reorderByLocationIDs(draggingLocationId.value, targetID)
  cancelDrag()
}

function startPointerDrag(locationID: string, event: PointerEvent) {
  if (!props.canEdit || orderSaving.value || !selectedRoute.value || event.pointerType === 'mouse') return
  draggingLocationId.value = locationID
  dragOverLocationId.value = locationID
  event.preventDefault()
  window.addEventListener('pointermove', handlePointerMove, { passive: false })
  window.addEventListener('pointerup', handlePointerUp, { once: true })
  window.addEventListener('pointercancel', handlePointerCancel, { once: true })
}

function handlePointerMove(event: PointerEvent) {
  if (!draggingLocationId.value || !selectedRoute.value) return
  event.preventDefault()
  const element = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null
  const target = element?.closest('[data-drop-location-id]') as HTMLElement | null
  if (!target || target.dataset.routeId !== selectedRoute.value.id) return
  dragOverLocationId.value = target.dataset.dropLocationId || ''
}

async function handlePointerUp() {
  window.removeEventListener('pointermove', handlePointerMove)
  const sourceID = draggingLocationId.value
  const targetID = dragOverLocationId.value
  if (sourceID && targetID && sourceID !== targetID) await reorderByLocationIDs(sourceID, targetID)
  cancelDrag()
}

function handlePointerCancel() {
  window.removeEventListener('pointermove', handlePointerMove)
  cancelDrag()
}

async function reorderByLocationIDs(sourceID: string, targetID: string) {
  const items = [...activeRouteLocations.value]
  const sourceIndex = items.findIndex((location) => location.id === sourceID)
  const targetIndex = items.findIndex((location) => location.id === targetID)
  if (sourceIndex < 0 || targetIndex < 0) return
  const [item] = items.splice(sourceIndex, 1)
  items.splice(targetIndex, 0, item)
  await saveRouteOrder(items)
}

async function saveRouteOrder(items: LocationPoint[]) {
  if (!selectedRoute.value) return
  orderError.value = ''
  orderSaving.value = true
  try {
    const { data } = await http.post<LocationPoint[]>(`/trips/${props.tripId}/routes/${selectedRoute.value.id}/locations/reorder`, {
      location_ids: items.map((location) => location.id)
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
  dragOverLocationId.value = ''
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

function routeButtonClass(routeId: string) {
  return ['route-picker-item', props.selectedRouteId === routeId ? 'active' : '']
}

function countLocations(routeId: string) {
  return props.locations.filter((location) => location.route_id === routeId).length
}

function routeTitle(routeId: string) {
  return props.routes.find((route) => route.id === routeId)?.title || 'Маршрут'
}

function categoryLabel(category: string) {
  return categories.find((item) => item.value === category)?.label || 'Другое'
}

function categoryIcon(category: string) {
  return categories.find((item) => item.value === category)?.icon || '📍'
}

function focusLocation(location: LocationPoint) {
  map?.setView([location.lat, location.lng], Math.max(map.getZoom(), 14))
  markers.get(location.id)?.openPopup()
}

function center(locations: LocationPoint[]): [number, number] {
  if (!locations.length) return [55.751244, 37.618423]
  const lat = locations.reduce((sum, l) => sum + l.lat, 0) / locations.length
  const lng = locations.reduce((sum, l) => sum + l.lng, 0) / locations.length
  return [lat, lng]
}

function renderMapLayers() {
  if (!map) return
  for (const [id, marker] of markers) {
    if (!filteredLocations.value.some((l) => l.id === id)) {
      marker.remove()
      markers.delete(id)
    }
  }
  for (const location of filteredLocations.value) {
    const existing = markers.get(location.id)
    if (existing) {
      existing.setLatLng([location.lat, location.lng])
      existing.setIcon(makeIcon(location))
      existing.setPopupContent(popupHtml(location))
    } else {
      const marker = L.marker([location.lat, location.lng], { icon: makeIcon(location) }).addTo(map)
      marker.bindPopup(popupHtml(location))
      marker.on('popupopen', () => attachPopupHandlers(location.id))
      markers.set(location.id, marker)
    }
  }
  renderRouteLine()
}

function renderRouteLine() {
  if (!map) return
  routeLine?.remove()
  routeLine = null
  if (activeRouteLocations.value.length < 2) return
  const coordinates = activeRouteLocations.value.map((location) => [location.lat, location.lng] as [number, number])
  routeLine = L.polyline(coordinates, { weight: 4, opacity: 0.78, dashArray: '8 8' }).addTo(map)
}

function markerNumber(location: LocationPoint) {
  if (props.selectedRouteId === 'all' || !location.route_id) return ''
  const index = activeRouteLocations.value.findIndex((item) => item.id === location.id)
  return index >= 0 ? `${index + 1}. ` : ''
}

function makeIcon(location: LocationPoint) {
  const index = props.selectedRouteId !== 'all' && location.route_id ? activeRouteLocations.value.findIndex((item) => item.id === location.id) : -1
  const number = index >= 0 ? `<b>${index + 1}</b>` : ''
  return L.divIcon({
    className: 'tc-marker-wrap',
    html: `<span class="tc-marker">${number}<em>${escapeHtml(categoryIcon(location.category))}</em></span>`,
    iconSize: [38, 38],
    iconAnchor: [19, 19],
    popupAnchor: [0, -20]
  })
}

function popupHtml(location: LocationPoint) {
  const linkedExpenses = expensesByLocation.value.get(location.id) || []
  const expensesHtml = linkedExpenses.length
    ? `<div class="popup-expenses">
        <small>Связанные расходы</small>
        ${linkedExpenses.slice(0, 3).map((expense) => `<p>${escapeHtml(expense.description)} — <strong>${money(expense.amount, expense.currency)}</strong></p>`).join('')}
        ${linkedExpenses.length > 3 ? `<p>+ ещё ${linkedExpenses.length - 3}</p>` : ''}
        <p class="popup-total">Итого: <strong>${money(locationExpenseTotal(location.id), linkedExpenses[0]?.currency || 'RUB')}</strong></p>
      </div>`
    : '<p class="hint">Связанных расходов пока нет</p>'
  const editActions = props.canEdit
    ? `<button data-action="expense" data-id="${location.id}">Добавить расход</button>
       <button data-action="edit" data-id="${location.id}">Изменить</button>
       <button data-action="delete" data-id="${location.id}">Удалить</button>`
    : '<span class="hint">Только просмотр</span>'
  return `
    <div class="popup">
      <strong>${escapeHtml(markerNumber(location))}${escapeHtml(location.name)}</strong>
      <p>${escapeHtml(location.description || 'Без описания')}</p>
      <small>${escapeHtml(categoryIcon(location.category))} ${escapeHtml(categoryLabel(location.category))}</small><br />
      <small>${escapeHtml(location.route_title || 'Без маршрута')}</small><br />
      <small>${escapeHtml(formatVisit(location.visit_at))}</small>
      ${expensesHtml}
      <div class="popup-actions">
        <button data-action="map" data-id="${location.id}">Открыть на карте</button>
        ${editActions}
      </div>
    </div>
  `
}

function attachPopupHandlers(locationID: string) {
  setTimeout(() => {
    const location = props.locations.find((item) => item.id === locationID)
    if (!location) return
    document.querySelector(`button[data-action="map"][data-id="${location.id}"]`)?.addEventListener('click', () => {
      openPointMapChooser(location)
    })
    if (!props.canEdit) return
    document.querySelector(`button[data-action="expense"][data-id="${location.id}"]`)?.addEventListener('click', () => {
      emit('addExpense', location)
    })
    document.querySelector(`button[data-action="edit"][data-id="${location.id}"]`)?.addEventListener('click', () => {
      editingId.value = location.id
      draftError.value = ''
      draft.value = {
        name: location.name,
        description: location.description || '',
        category: location.category,
        lat: location.lat,
        lng: location.lng,
        visit_at: toDateTimeLocal(location.visit_at),
        route_id: location.route_id || ''
      }
    })
    document.querySelector(`button[data-action="delete"][data-id="${location.id}"]`)?.addEventListener('click', async () => {
      if (!window.confirm(`Удалить точку «${location.name}»? Связанные расходы останутся, но отвяжутся от точки.`)) return
      await http.delete(`/trips/${props.tripId}/locations/${location.id}`)
      emit('deleted', location.id)
    })
  })
}

async function saveDraft() {
  if (!draft.value) return
  draftError.value = ''
  if (!draft.value.name.trim()) {
    draftError.value = 'Введите название точки'
    return
  }
  draftSaving.value = true
  const payload = {
    ...draft.value,
    name: draft.value.name.trim(),
    description: draft.value.description || null,
    route_id: draft.value.route_id || null,
    visit_at: draft.value.visit_at ? new Date(draft.value.visit_at).toISOString() : null
  }
  try {
    if (editingId.value) {
      const { data } = await http.put<LocationPoint>(`/trips/${props.tripId}/locations/${editingId.value}`, payload)
      emit('updated', data)
    } else {
      const { data } = await http.post<LocationPoint>(`/trips/${props.tripId}/locations`, payload)
      emit('created', data)
    }
    closeDraft()
  } catch (e: any) {
    draftError.value = formatApiError(e, 'Не удалось сохранить точку')
  } finally {
    draftSaving.value = false
  }
}

function closeDraft() {
  draft.value = null
  editingId.value = null
  draftError.value = ''
}

function locationExpenseTotal(locationID: string) {
  return (expensesByLocation.value.get(locationID) || []).reduce((sum, expense) => sum + Number(expense.amount || 0), 0)
}

function money(value: number, currency = 'RUB') {
  return `${Number(value || 0).toFixed(2)} ${currency}`
}

function formatVisit(value?: string | null) {
  if (!value) return 'Без времени посещения'
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function formatDate(value?: string | null) {
  if (!value) return 'Без даты'
  return new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' }).format(new Date(value))
}

function toDateInput(value?: string | null) {
  if (!value) return ''
  return new Date(value).toISOString().slice(0, 10)
}

function toDateTimeLocal(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  const offsetMs = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function openPointMapChooser(location: LocationPoint) {
  panelError.value = ''
  mapChooser.value = { mode: 'point', location }
}

function openRouteMapChooser() {
  panelError.value = ''
  if (activeRouteLocations.value.length < 2) {
    panelError.value = 'Для построения маршрута на внешней карте добавьте минимум две точки в выбранный маршрут.'
    return
  }
  mapChooser.value = { mode: 'route' }
}

function closeMapChooser() {
  mapChooser.value = null
}

function openChosenMap(provider: MapProvider) {
  const chooser = mapChooser.value
  if (!chooser) return
  const url = chooser.mode === 'point' && chooser.location
    ? pointMapUrl(chooser.location, provider)
    : routeMapUrl(provider)
  if (!url) {
    panelError.value = 'Не удалось сформировать ссылку для выбранной карты.'
    return
  }
  closeMapChooser()
  if (url.startsWith('geo:')) {
    window.location.href = url
    return
  }
  window.open(url, '_blank', 'noopener,noreferrer')
}

function pointMapUrl(location: LocationPoint, provider: MapProvider) {
  const lat = location.lat
  const lng = location.lng
  const name = encodeURIComponent(location.name)
  if (provider === 'google') return `https://www.google.com/maps/search/?api=1&query=${lat},${lng}`
  if (provider === 'yandex') return `https://yandex.ru/maps/?ll=${lng}%2C${lat}&z=16&pt=${lng}%2C${lat}%2Cpm2rdm`
  if (provider === '2gis') return `https://2gis.ru/directions/points/|${lng},${lat}`
  return `geo:${lat},${lng}?q=${lat},${lng}(${name})`
}

function routeMapUrl(provider: Exclude<MapProvider, 'system'> | MapProvider) {
  const points = activeRouteLocations.value.filter((location) => Number.isFinite(location.lat) && Number.isFinite(location.lng))
  if (points.length < 2) return ''
  const first = points[0]
  const last = points[points.length - 1]
  const middle = points.slice(1, -1)
  if (provider === 'google' || provider === 'system') {
    const origin = `${first.lat},${first.lng}`
    const destination = `${last.lat},${last.lng}`
    const waypoints = middle.map((location) => `${location.lat},${location.lng}`).join('|')
    return `https://www.google.com/maps/dir/?api=1&origin=${encodeURIComponent(origin)}&destination=${encodeURIComponent(destination)}${waypoints ? `&waypoints=${encodeURIComponent(waypoints)}` : ''}&travelmode=walking`
  }
  if (provider === 'yandex') {
    const rtext = points.map((location) => `${location.lat},${location.lng}`).join('~')
    return `https://yandex.ru/maps/?rtext=${encodeURIComponent(rtext)}&rtt=pd`
  }
  const dgisPoints = points.slice(0, 10).map((location) => `${location.lng},${location.lat}`).join('|')
  return `https://2gis.ru/directions/tab/pedestrian/points/${dgisPoints}`
}

function formatApiError(e: any, fallback: string) {
  const data = e.response?.data
  if (!data?.error) return fallback
  if (!data.details) return data.error
  const details = Object.entries(data.details).map(([key, value]) => `${key}: ${value}`).join(', ')
  return `${data.error}. ${details}`
}

function escapeHtml(value: string) {
  const replacements: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    "'": '&#039;',
    '"': '&quot;'
  }
  return value.replace(/[&<>'"]/g, (char) => replacements[char] ?? char)
}
</script>

<style scoped>
.map-provider-list {
  display: grid;
  gap: 10px;
  margin: 16px 0;
}

.map-open-modal {
  max-width: 520px;
}
</style>
