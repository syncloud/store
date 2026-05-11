<script setup>
import { ref, onMounted, watch } from 'vue'

const STORAGE_KEY = 'syncloud-store-theme'
const theme = ref('light')

function applyTheme (value) {
  document.documentElement.setAttribute('data-theme', value)
}

function detectInitial () {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'dark' || stored === 'light') return stored
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function toggle () {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

onMounted(() => {
  theme.value = detectInitial()
  applyTheme(theme.value)
})

watch(theme, (v) => {
  applyTheme(v)
  localStorage.setItem(STORAGE_KEY, v)
})
</script>

<template>
  <button
    class="switch"
    :data-theme-state="theme"
    data-testid="theme-switcher"
    :aria-label="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'"
    @click="toggle"
  >
    <span v-if="theme === 'dark'" class="icon" aria-hidden="true">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="4" />
        <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
      </svg>
    </span>
    <span v-else class="icon" aria-hidden="true">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
      </svg>
    </span>
  </button>
</template>

<style scoped>
.switch {
  width: 38px;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  transition: border-color 0.15s ease, transform 0.15s ease;
}
.switch:hover {
  border-color: var(--accent);
  transform: rotate(15deg);
}
.icon { display: inline-flex; }
</style>
