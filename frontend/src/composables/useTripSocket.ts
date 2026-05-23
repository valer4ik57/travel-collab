import { ref } from 'vue'
import type { WSEvent } from '../types'

function buildWebSocketURL(tripId: string, token: string) {
  const configuredBase = (import.meta.env.VITE_WS_BASE_URL || '').trim().replace(/\/$/, '')
  const encodedToken = encodeURIComponent(token)

  if (configuredBase) {
    return `${configuredBase}/${tripId}?token=${encodedToken}`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = window.location.host
  return `${protocol}://${host}/api/v1/ws/${tripId}?token=${encodedToken}`
}

export function useTripSocket(tripId: string, onEvent: (event: WSEvent) => void) {
  const connected = ref(false)
  const reconnecting = ref(false)
  let socket: WebSocket | null = null
  let closedByUser = false
  let reconnectTimer: number | undefined

  function connect() {
    const token = localStorage.getItem('token')
    if (!token) return
    socket = new WebSocket(buildWebSocketURL(tripId, token))

    socket.onopen = () => {
      connected.value = true
      reconnecting.value = false
    }
    socket.onmessage = (message) => {
      try {
        onEvent(JSON.parse(message.data))
      } catch (error) {
        console.warn('WS parse error', error)
      }
    }
    socket.onclose = () => {
      connected.value = false
      if (!closedByUser) {
        reconnecting.value = true
        reconnectTimer = window.setTimeout(connect, 3000)
      }
    }
  }

  function send(type: string, payload: unknown) {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type, payload }))
    }
  }

  function close() {
    closedByUser = true
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    socket?.close()
  }

  return { connected, reconnecting, connect, send, close }
}
