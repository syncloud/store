<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import AppCard from './components/AppCard.vue'
import ThemeSwitcher from './components/ThemeSwitcher.vue'

const apps = ref([])
const loading = ref(true)
const error = ref(null)
const query = ref('')

async function load () {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/ui/v1/apps')
    if (!res.ok) throw new Error('http ' + res.status)
    const data = await res.json()
    apps.value = (data || []).map(a => ({
      id: a.snapId,
      name: a.name,
      summary: a.summary || '',
      description: a.description || '',
      version: a.version || '',
      icon: a.iconUrl || ''
    }))
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return apps.value
  return apps.value.filter(a =>
    a.name.toLowerCase().includes(q) ||
    (a.summary || '').toLowerCase().includes(q)
  )
})

onMounted(load)
</script>

<template>
  <div class="page">
    <header class="header">
      <div class="header-inner">
        <div class="brand" data-testid="brand">
          <span class="logo-dot" />
          <span class="brand-name">Syncloud Store</span>
        </div>
        <ThemeSwitcher />
      </div>
    </header>

    <main class="main">
      <section class="hero">
        <h1 class="title">Apps for your Syncloud</h1>
        <p class="subtitle">Self-hosted, private, one click away.</p>
        <div class="search-wrap">
          <input
            v-model="query"
            type="search"
            class="search"
            placeholder="Search apps…"
            data-testid="search"
            autocomplete="off"
          />
        </div>
      </section>

      <div v-if="loading" class="state" data-testid="loading">Loading…</div>
      <div v-else-if="error" class="state state-error" data-testid="error">
        Could not load apps: {{ error }}
      </div>
      <div v-else>
        <div class="meta-row">
          <span data-testid="results-count">{{ filtered.length }} of {{ apps.length }} apps</span>
        </div>
        <div v-if="filtered.length === 0" class="state" data-testid="empty">
          No apps match "{{ query }}".
        </div>
        <ul v-else class="grid" data-testid="app-list">
          <li v-for="app in filtered" :key="app.id" data-testid="app-card-item">
            <AppCard :app="app" />
          </li>
        </ul>
      </div>
    </main>

    <footer class="footer">
      <span>syncloud.org</span>
    </footer>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: saturate(180%) blur(10px);
  background: color-mix(in srgb, var(--bg) 85%, transparent);
  border-bottom: 1px solid var(--border);
}
.header-inner {
  max-width: 1100px;
  margin: 0 auto;
  padding: 14px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 17px;
  letter-spacing: -0.01em;
}
.logo-dot {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  background: linear-gradient(135deg, var(--accent), color-mix(in srgb, var(--accent) 60%, white));
  box-shadow: var(--shadow);
}
.main {
  flex: 1;
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  padding: 32px 20px 64px;
}
.hero {
  text-align: center;
  margin: 8px 0 28px;
}
.title {
  font-size: clamp(28px, 5vw, 44px);
  margin: 0 0 8px;
  letter-spacing: -0.02em;
}
.subtitle {
  margin: 0 0 24px;
  color: var(--text-muted);
  font-size: clamp(15px, 2vw, 17px);
}
.search-wrap {
  display: flex;
  justify-content: center;
}
.search {
  width: 100%;
  max-width: 520px;
  padding: 14px 18px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font-size: 16px;
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.search:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-soft);
}
.meta-row {
  color: var(--text-muted);
  font-size: 14px;
  margin: 0 0 14px;
  padding: 0 4px;
}
.grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
}
.state {
  text-align: center;
  padding: 48px 20px;
  color: var(--text-muted);
  border: 1px dashed var(--border);
  border-radius: var(--radius);
  background: var(--bg-elevated);
}
.state-error { color: #ef4444; border-color: #ef4444; }
.footer {
  padding: 24px 20px 36px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}
@media (max-width: 480px) {
  .main { padding: 20px 14px 48px; }
  .grid { grid-template-columns: 1fr 1fr; gap: 12px; }
}
</style>
