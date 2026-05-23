<template>
  <section class="card panel members-panel">
    <div class="section-head">
      <div>
        <h2>Участники</h2>
        <p class="muted">Роли определяют, кто может менять маршрут, расходы и писать в чат.</p>
      </div>
      <button class="button small secondary" @click="copyInvite">Скопировать код</button>
    </div>
    <p class="muted">Invite-код: <code>{{ trip.invite_code }}</code></p>

    <ul class="members large">
      <li v-for="member in members" :key="member.user_id" class="member-row">
        <img
          v-if="member.avatar_url && !brokenAvatars.has(member.user_id)"
          :src="member.avatar_url"
          alt=""
          @error="brokenAvatars.add(member.user_id)"
        />
        <span v-else class="avatar">{{ member.display_name.slice(0, 1).toUpperCase() }}</span>
        <div class="member-main">
          <strong>{{ member.display_name }} <span v-if="member.user_id === currentUserId" class="muted">(это вы)</span></strong>
          <small>{{ roleLabel(member.role) }}</small>
          <small v-if="member.email">{{ member.email }}</small>
        </div>

        <div v-if="currentRole === 'owner' && member.user_id !== currentUserId" class="member-actions">
          <select :value="member.role" @change="emitRole(member.user_id, ($event.target as HTMLSelectElement).value)">
            <option value="owner">owner</option>
            <option value="editor">editor</option>
            <option value="viewer">viewer</option>
          </select>
          <button class="button tiny danger" type="button" @click="emit('remove', member.user_id)">Выгнать</button>
        </div>
      </li>
    </ul>

    <div class="roles-help subcard">
      <p><strong>owner</strong> — управляет поездкой и участниками.</p>
      <p><strong>editor</strong> — редактирует маршрут, расходы и чат.</p>
      <p><strong>viewer</strong> — только просматривает данные.</p>
    </div>

    <p v-if="copied" class="success">Код скопирован</p>
    <button v-if="currentRole !== 'owner'" class="button small secondary full" type="button" @click="emit('leave')">Выйти из поездки</button>
    <p v-else class="hint">Владелец не может выйти, пока он единственный owner. Сначала назначьте другого owner.</p>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { Trip, TripMember } from '../types'

const props = defineProps<{ trip: Trip; members: TripMember[]; currentUserId: string; currentRole: string }>()
const emit = defineEmits<{
  leave: []
  remove: [userId: string]
  updateRole: [userId: string, role: TripMember['role']]
}>()
const copied = ref(false)
const brokenAvatars = reactive(new Set<string>())

function roleLabel(role: string) {
  if (role === 'owner') return 'владелец'
  if (role === 'viewer') return 'только просмотр'
  return 'редактор'
}

function emitRole(userId: string, role: string) {
  emit('updateRole', userId, role as TripMember['role'])
}

async function copyInvite() {
  await navigator.clipboard.writeText(props.trip.invite_code)
  copied.value = true
  setTimeout(() => (copied.value = false), 1600)
}
</script>
