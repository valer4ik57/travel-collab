import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { Capacitor } from '@capacitor/core'
import { StatusBar, Style } from '@capacitor/status-bar'
import 'leaflet/dist/leaflet.css'
import './style.css'
import App from './App.vue'
import router from './router'

async function prepareNativeShell() {
  if (!Capacitor.isNativePlatform()) return

  document.documentElement.classList.add('capacitor-native-root')
  document.body.classList.add('capacitor-native-shell')

  try {
    await StatusBar.setOverlaysWebView({ overlay: false })
    await StatusBar.setBackgroundColor({ color: '#EEF3F8' })
    await StatusBar.setStyle({ style: Style.Light })
  } catch (error) {
    console.warn('Unable to configure native status bar', error)
  }
}

void prepareNativeShell()

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
