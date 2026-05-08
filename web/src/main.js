import { createApp } from 'vue'
import App from './App.vue'
import './style/global.css'

if (import.meta.env.VITE_STUB) {
  const { mock } = await import('./stub/api.js')
  mock()
}

createApp(App).mount('#app')
