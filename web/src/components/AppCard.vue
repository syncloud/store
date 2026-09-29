<script setup>
import { computed, ref } from 'vue'
import { landingUrl } from '../landing'

const props = defineProps({
  app: { type: Object, required: true }
})

const broken = ref(false)
const landing = computed(() => landingUrl(props.app.id))

function initial (name) {
  return (name || '?').slice(0, 1).toUpperCase()
}
</script>

<template>
  <component
    :is="landing ? 'a' : 'article'"
    class="card"
    :class="{ 'card-link': landing }"
    :href="landing || undefined"
    data-testid="app-card"
    :data-name="app.name"
  >
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
      <div class="foot">
        <span v-if="app.version" class="version" data-testid="app-version">v{{ app.version }}</span>
        <span v-if="landing" class="more" data-testid="app-more">More…</span>
      </div>
    </div>
  </component>
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
.card-link {
  text-decoration: none;
  color: inherit;
}
.card-link:hover {
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
.foot {
  margin-top: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.version {
  font-size: 11px;
  color: var(--text-muted);
  background: var(--accent-soft);
  padding: 2px 8px;
  border-radius: 999px;
}
.more {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--accent);
  white-space: nowrap;
}
.card-link:hover .more {
  text-decoration: underline;
}
</style>
