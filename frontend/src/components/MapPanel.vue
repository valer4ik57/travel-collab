<template>
  <section class="card map-card">
    <div class="section-head">
      <div>
        <h2>Карта маршрута</h2>
        <p class="muted">{{ canEdit ? 'Клик по карте добавляет новую точку интереса.' : 'У вас роль viewer, карта доступна только для просмотра.' }}</p>
      </div>
    </div>
    <div ref="mapEl" class="map"></div>

    <div class="locations-strip">
      <article v-for="location in locations" :key="location.id" class="location-chip" @click="focusLocation(location)">
        <strong>{{ location.name }}</strong>
        <small>{{ categoryLabel(location.category) }}</small>
      </article>
      <p v-if="locations.length === 0" class="empty">Точек пока нет. {{ canEdit ? 'Кликните по карте, чтобы добавить первую.' : '' }}</p>
    </div>

    <p class="hint">Следующий этап развития: к точке можно будет привязать расход, например кафе, экскурсию или транспорт.</p>

    <div v-if="draft" class="modal-backdrop" @click.self="draft = null">
      <form class="modal card form" @submit.prevent="saveDraft">
        <h2>{{ editingId ? 'Редактировать точку' : 'Добавить точку' }}</h2>
        <label>Название<input v-model="draft.name" required /></label>
        <label>Описание<textarea v-model="draft.description" rows="3" /></label>
        <label>Категория
          <select v-model="draft.category">
            <option value="sight">Достопримечательность</option>
            <option value="food">Еда</option>
            <option value="hotel">Жильё</option>
            <option value="transport">Транспорт</option>
            <option value="other">Другое</option>
          </select>
        </label>
        <div class="actions right">
          <button type="button" class="button secondary" @click="draft = null">Отмена</button>
          <button class="button">Сохранить</button>
        </div>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png'
})
import { http } from '../api/http'
import type { LocationPoint } from '../types'

const props = defineProps<{ tripId: string; locations: LocationPoint[]; canEdit: boolean }>()
const emit = defineEmits<{
  created: [location: LocationPoint]
  updated: [location: LocationPoint]
  deleted: [id: string]
}>()

const mapEl = ref<HTMLDivElement | null>(null)
let map: L.Map | null = null
const markers = new Map<string, L.Marker>()
const draft = ref<null | { name: string; description: string; category: string; lat: number; lng: number }>(null)
const editingId = ref<string | null>(null)

onMounted(() => {
  map = L.map(mapEl.value as HTMLDivElement).setView(center(), props.locations.length ? 11 : 5)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map)
  map.on('click', (event: L.LeafletMouseEvent) => {
    if (!props.canEdit) return
    editingId.value = null
    draft.value = { name: '', description: '', category: 'other', lat: event.latlng.lat, lng: event.latlng.lng }
  })
  renderMarkers()
})

onBeforeUnmount(() => {
  map?.remove()
})

watch(() => props.locations, renderMarkers, { deep: true })

function categoryLabel(category: string) {
  const labels: Record<string, string> = {
    sight: 'Достопримечательность',
    food: 'Еда',
    hotel: 'Жильё',
    transport: 'Транспорт',
    other: 'Другое'
  }
  return labels[category] || category
}

function focusLocation(location: LocationPoint) {
  map?.setView([location.lat, location.lng], Math.max(map.getZoom(), 14))
  markers.get(location.id)?.openPopup()
}

function center(): [number, number] {
  if (!props.locations.length) return [55.751244, 37.618423]
  const lat = props.locations.reduce((sum, l) => sum + l.lat, 0) / props.locations.length
  const lng = props.locations.reduce((sum, l) => sum + l.lng, 0) / props.locations.length
  return [lat, lng]
}

function renderMarkers() {
  if (!map) return
  for (const [id, marker] of markers) {
    if (!props.locations.some((l) => l.id === id)) {
      marker.remove()
      markers.delete(id)
    }
  }
  for (const location of props.locations) {
    const existing = markers.get(location.id)
    if (existing) {
      existing.setLatLng([location.lat, location.lng])
      existing.setPopupContent(popupHtml(location))
    } else {
      const marker = L.marker([location.lat, location.lng]).addTo(map)
      marker.bindPopup(popupHtml(location))
      marker.on('popupopen', () => attachPopupHandlers(location))
      markers.set(location.id, marker)
    }
  }
}

function popupHtml(location: LocationPoint) {
  const actions = props.canEdit
    ? `<div class="popup-actions">
        <button data-action="edit" data-id="${location.id}">Изменить</button>
        <button data-action="delete" data-id="${location.id}">Удалить</button>
      </div>`
    : '<p class="hint">Только просмотр</p>'
  return `
    <div class="popup">
      <strong>${escapeHtml(location.name)}</strong>
      <p>${escapeHtml(location.description || 'Без описания')}</p>
      <small>${escapeHtml(location.category)}</small>
      ${actions}
    </div>
  `
}

function attachPopupHandlers(location: LocationPoint) {
  if (!props.canEdit) return
  setTimeout(() => {
    document.querySelector(`button[data-action="edit"][data-id="${location.id}"]`)?.addEventListener('click', () => {
      editingId.value = location.id
      draft.value = {
        name: location.name,
        description: location.description || '',
        category: location.category,
        lat: location.lat,
        lng: location.lng
      }
    })
    document.querySelector(`button[data-action="delete"][data-id="${location.id}"]`)?.addEventListener('click', async () => {
      await http.delete(`/trips/${props.tripId}/locations/${location.id}`)
      emit('deleted', location.id)
    })
  })
}

async function saveDraft() {
  if (!draft.value) return
  const payload = { ...draft.value, description: draft.value.description || null }
  if (editingId.value) {
    const { data } = await http.put<LocationPoint>(`/trips/${props.tripId}/locations/${editingId.value}`, payload)
    emit('updated', data)
  } else {
    const { data } = await http.post<LocationPoint>(`/trips/${props.tripId}/locations`, payload)
    emit('created', data)
  }
  draft.value = null
  editingId.value = null
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
