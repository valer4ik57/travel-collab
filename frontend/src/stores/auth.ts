import { defineStore } from 'pinia'
import { http } from '../api/http'
import type { User } from '../types'

function readStoredUser(): User | null {
  try {
    const raw = localStorage.getItem('user')
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: readStoredUser()
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token)
  },
  actions: {
    setSession(token: string, user?: User) {
      this.token = token
      localStorage.setItem('token', token)
      if (user) {
        this.user = user
        localStorage.setItem('user', JSON.stringify(user))
      }
    },
    async login(email: string, password: string) {
      const { data } = await http.post('/auth/login', { email, password })
      this.setSession(data.token, data.user)
    },
    async register(email: string, password: string, display_name: string) {
      const { data } = await http.post('/auth/register', { email, password, display_name })
      this.setSession(data.token, data.user)
    },
    async loadMe() {
      if (!this.token) return
      const { data } = await http.get<User>('/me')
      this.user = data
      localStorage.setItem('user', JSON.stringify(data))
    },
    async updateMe(display_name: string, avatar_url: string | null) {
      const { data } = await http.put<User>('/me', { display_name, avatar_url })
      this.user = data
      localStorage.setItem('user', JSON.stringify(data))
      return data
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
  }
})
