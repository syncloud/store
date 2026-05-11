<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import AppCard from './components/AppCard.vue'
import ThemeSwitcher from './components/ThemeSwitcher.vue'

const apps = ref([])
const loading = ref(true)
const error = ref(null)
const query = ref('')
const version = ref(null)
const sortBy = ref('rank')
const sortDir = ref('desc')

function toggleSort (col) {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = col === 'rank' ? 'desc' : 'asc'
  }
}

async function loadVersion () {
  try {
    const res = await fetch('/api/ui/v1/version')
    if (res.ok) version.value = await res.json()
  } catch (_) { /* version is best-effort */ }
}

async function load () {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/ui/v1/apps')
    if (!res.ok) throw new Error('http ' + res.status)
    const data = await res.json()
    let rank = 0
    apps.value = (data || []).map(a => {
      const popularity = a.popularity || 0
      return {
        id: a.snapId,
        name: a.name,
        summary: a.summary || '',
        description: a.description || '',
        version: a.version || '',
        icon: a.iconUrl || '',
        popularity,
        rank: popularity > 0 ? ++rank : 0
      }
    })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

const buildDate = computed(() => {
  const t = version.value?.buildTime
  if (!t) return null
  const d = new Date(t)
  if (isNaN(d.getTime())) return null
  return d.toISOString().slice(0, 10)
})

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  const arr = q
    ? apps.value.filter(a =>
        a.name.toLowerCase().includes(q) ||
        (a.summary || '').toLowerCase().includes(q))
    : apps.value.slice()
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    let cmp
    if (sortBy.value === 'rank') {
      cmp = (b.popularity || 0) - (a.popularity || 0)
      cmp = cmp * (sortDir.value === 'desc' ? 1 : -1)
    } else {
      cmp = a.name.localeCompare(b.name) * dir
    }
    if (cmp === 0) cmp = a.name.localeCompare(b.name)
    return cmp
  })
  return arr
})

onMounted(() => {
  load()
  loadVersion()
})
</script>

<template>
  <div class="page">
    <header class="header">
      <div class="header-inner">
        <div class="brand" data-testid="brand">
          <img src="/syncloud-logo.svg" alt="Syncloud" class="logo" />
          <span class="brand-name">Syncloud Store</span>
        </div>
        <ThemeSwitcher />
      </div>
    </header>

    <main class="main">
      <section class="hero">
        <h1 class="title">Apps for your device</h1>
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
          <span class="count" data-testid="results-count">{{ filtered.length }} of {{ apps.length }} apps</span>
          <div class="sort-controls" data-testid="sort-controls">
            <button
              type="button"
              class="sort-btn"
              :class="{ active: sortBy === 'rank' }"
              data-testid="sort-rank"
              :title="sortBy === 'rank' ? 'Toggle direction' : 'Sort by popularity'"
              @click="toggleSort('rank')"
            >
              Rank
              <span v-if="sortBy === 'rank'" class="arrow">{{ sortDir === 'desc' ? '↓' : '↑' }}</span>
            </button>
            <button
              type="button"
              class="sort-btn"
              :class="{ active: sortBy === 'name' }"
              data-testid="sort-name"
              :title="sortBy === 'name' ? 'Toggle direction' : 'Sort alphabetically'"
              @click="toggleSort('name')"
            >
              Name
              <span v-if="sortBy === 'name'" class="arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </button>
          </div>
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
      <span v-if="version" class="version-pill" data-testid="version" :title="`build ${version.buildNumber} · ${version.buildTime}`">
        <span class="version-label">build</span>
        <span class="version-num">{{ version.buildNumber }}</span>
        <span class="version-sha">{{ version.gitSha.slice(0, 7) }}</span>
        <span v-if="buildDate" class="version-date" data-testid="version-date">{{ buildDate }}</span>
      </span>
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
.logo {
  width: 28px;
  height: 28px;
  display: block;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.sort-controls {
  display: inline-flex;
  gap: 6px;
}
.sort-btn {
  font: inherit;
  font-size: 12px;
  color: var(--text-muted);
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  padding: 4px 10px;
  border-radius: 999px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: border-color 0.15s ease, color 0.15s ease;
}
.sort-btn:hover { border-color: var(--accent); }
.sort-btn.active {
  color: var(--accent);
  border-color: var(--accent);
}
.sort-btn .arrow { font-weight: 700; }
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
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
.version-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  font-size: 11px;
  letter-spacing: 0.02em;
  font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  cursor: default;
}
.version-label { color: var(--text-muted); }
.version-num { font-weight: 600; color: var(--text); }
.version-sha {
  color: var(--accent);
  padding-left: 6px;
  border-left: 1px solid var(--border);
  margin-left: 2px;
}
.version-date {
  color: var(--text-muted);
  padding-left: 6px;
  border-left: 1px solid var(--border);
  margin-left: 2px;
}
@media (max-width: 480px) {
  .main { padding: 20px 14px 48px; }
  .grid { grid-template-columns: 1fr 1fr; gap: 12px; }
}
</style>
