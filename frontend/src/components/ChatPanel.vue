<template>
  <section class="card panel chat-panel chat-page">
    <div class="section-head">
      <div>
        <h2>Чат</h2>
        <p class="muted">Обсуждение поездки между участниками.</p>
      </div>
      <span class="dot" :class="connected ? 'ok' : 'warn'"></span>
    </div>
    <div ref="messagesEl" class="messages chat-messages">
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
import { nextTick, ref, watch } from 'vue'
import type { Message } from '../types'

const props = defineProps<{ messages: Message[]; connected: boolean; canWrite: boolean }>()
const emit = defineEmits<{ send: [text: string] }>()
const text = ref('')
const messagesEl = ref<HTMLDivElement | null>(null)

function scrollDown() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}

watch(() => props.messages.length, scrollDown, { immediate: true })

function submit() {
  const value = text.value.trim()
  if (!value) return
  emit('send', value)
  text.value = ''
}
</script>
