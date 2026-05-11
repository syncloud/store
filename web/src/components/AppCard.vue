<script setup>
import { ref } from 'vue'

const props = defineProps({
  app: { type: Object, required: true }
})

const broken = ref(false)

function initial (name) {
  return (name || '?').slice(0, 1).toUpperCase()
}
</script>

<template>
  <article class="card" data-testid="app-card" :data-name="app.name">
    <span
      v-if="app.rank"
      class="rank"
      data-testid="app-rank"
    >#{{ app.rank }}</span>
    <div class="icon-wrap">
      <img
        v-if="app.icon && !broken"
        :src="app.icon"
        :alt="app.name + ' icon'"
        loading="lazy"
        data-testid="app-icon"
        @error="broken = true"
      />
      <span v-else class="fallback" data-testid="app-icon-fallback">{{ initial(app.name) }}</span>
    </div>
    <div class="body">
      <h3 class="name" data-testid="app-name">{{ app.name }}</h3>
      <p v-if="app.summary" class="summary" data-testid="app-summary">{{ app.summary }}</p>
      <span v-if="app.version" class="version" data-testid="app-version">v{{ app.version }}</span>
    </div>
  </article>
</template>

<style scoped>
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px;
  border-radius: var(--radius);
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  height: 100%;
  transition: transform 0.15s ease, border-color 0.15s ease;
}
.rank {
  position: absolute;
  top: 10px;
  right: 10px;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 999px;
  letter-spacing: 0.02em;
  color: #fff;
  background: var(--accent);
}
.card:hover {
  transform: translateY(-2px);
  border-color: var(--accent);
}
.icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.icon-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.fallback {
  font-weight: 700;
  font-size: 22px;
  color: var(--accent);
}
.body { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
.name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.summary {
  margin: 0;
  font-size: 13.5px;
  color: var(--text-muted);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.version {
  align-self: flex-start;
  font-size: 11px;
  color: var(--text-muted);
  background: var(--accent-soft);
  padding: 2px 8px;
  border-radius: 999px;
  margin-top: auto;
}
</style>
