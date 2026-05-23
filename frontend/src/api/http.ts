import axios from 'axios'
import router from '../router'
import { useAuthStore } from '../stores/auth'

function normalizeBaseURL(value: string | undefined, fallback: string) {
  const cleaned = (value || '').trim()
  return (cleaned || fallback).replace(/\/$/, '')
}

export const API_BASE = normalizeBaseURL(import.meta.env.VITE_API_BASE_URL, '/api/v1')

export const http = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json'
  }
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)
