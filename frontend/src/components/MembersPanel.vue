<template>
  <section class="card panel">
    <div class="section-head">
      <h2>Участники</h2>
      <button class="button small secondary" @click="copyInvite">Скопировать код</button>
    </div>
    <p class="muted">Invite-код: <code>{{ trip.invite_code }}</code></p>
    <ul class="members">
      <li v-for="member in members" :key="member.user_id">
        <img v-if="member.avatar_url" :src="member.avatar_url" alt="" />
        <span v-else class="avatar">{{ member.display_name.slice(0, 1).toUpperCase() }}</span>
        <div>
          <strong>{{ member.display_name }}</strong>
          <small>{{ roleLabel(member.role) }}</small>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import type { Trip, TripMember } from '../types'

const props = defineProps<{ trip: Trip; members: TripMember[] }>()

function roleLabel(role: string) {
  if (role === 'owner') return 'владелец'
  if (role === 'viewer') return 'просмотр'
  return 'редактор'
}

async function copyInvite() {
  await navigator.clipboard.writeText(props.trip.invite_code)
}
</script>
