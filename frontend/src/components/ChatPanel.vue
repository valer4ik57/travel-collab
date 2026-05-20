<template>
  <section class="card panel chat-panel">
    <div class="section-head">
      <h2>Чат</h2>
      <span class="dot" :class="connected ? 'ok' : 'warn'"></span>
    </div>
    <div class="messages">
      <p v-if="messages.length === 0" class="empty">Сообщений пока нет.</p>
      <article v-for="message in messages" :key="message.id" class="message">
        <strong>{{ message.display_name }}</strong>
        <p>{{ message.text }}</p>
        <small>{{ new Date(message.sent_at).toLocaleString() }}</small>
      </article>
    </div>
    <form v-if="canWrite" class="chat-form" @submit.prevent="submit">
      <input v-model="text" :disabled="!connected" placeholder="Написать сообщение..." />
      <button class="button small" :disabled="!connected || !text.trim()">Отправить</button>
    </form>
    <p v-else class="empty">У вас роль viewer, чат доступен только для чтения.</p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Message } from '../types'

defineProps<{ messages: Message[]; connected: boolean; canWrite: boolean }>()
const emit = defineEmits<{ send: [text: string] }>()
const text = ref('')

function submit() {
  const value = text.value.trim()
  if (!value) return
  emit('send', value)
  text.value = ''
}
</script>
