<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="location-picker"
    :class="[readonly ? 'location-picker--readonly' : 'location-picker--editable', customClass]"
  >
    <div ref="mapEl" class="location-picker__map"></div>

    <div v-if="!readonly" class="location-picker__status" :aria-live="'polite'">
      <span class="location-picker__status-dot"></span>
      {{ statusLabel }}
    </div>

    <div
      ref="controlsEl"
      class="location-picker__controls"
      @click.stop
      @dblclick.stop
      @mousedown.stop
      @touchstart.stop
    >
      <button
        v-if="hasLatLng"
        type="button"
        class="location-picker__ctrl"
        :aria-label="t('extensionCard.locationOpenInMaps')"
        :title="t('extensionCard.locationOpenInMaps')"
        @click="handleOpenInGoogleMaps"
      >
        <JumpIcon />
      </button>
      <button
        type="button"
        class="location-picker__ctrl"
        :aria-label="t('editor.locationZoomIn')"
        :title="t('editor.locationZoomIn')"
        @click="handleZoomIn"
      >
        <PlusIcon />
      </button>
      <button
        type="button"
        class="location-picker__ctrl"
        :aria-label="t('editor.locationZoomOut')"
        :title="t('editor.locationZoomOut')"
        @click="handleZoomOut"
      >
        <MinusIcon />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { useThemeStore } from '@/stores/theme'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import JumpIcon from '@/components/icons/jump.vue'
import PlusIcon from '@/components/icons/plus.vue'
import MinusIcon from '@/components/icons/minus.vue'

type LatLng = { lat: number; lng: number }

const props = withDefaults(
  defineProps<{
    readonly?: boolean
    latLng?: LatLng | null
    autoLocate?: boolean
    class?: string
  }>(),
  {
    readonly: false,
    latLng: null,
    autoLocate: false,
    class: '',
  },
)

const emit = defineEmits<{
  (e: 'change', p: LatLng): void
}>()

const customClass = props.class
const { t } = useI18n()
const themeStore = useThemeStore()
const { theme } = storeToRefs(themeStore)

const mapEl = ref<HTMLDivElement | null>(null)
const controlsEl = ref<HTMLDivElement | null>(null)
let mapInstance: L.Map | null = null
let marker: L.Marker | null = null
let tileLayer: L.TileLayer | null = null
let locatedOnce = false

const hasLatLng = computed(() => !!props.latLng)

const DEFAULT_CENTER: LatLng = { lat: 48.8584, lng: 2.2945 }

const LIGHT_TILE = 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png'
const DARK_TILE = 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png'
const TILE_ATTRIBUTION =
  '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>'

const currentTileUrl = () => (theme.value === 'dark' ? DARK_TILE : LIGHT_TILE)

const pinIcon = () =>
  L.divIcon({
    className: 'location-picker__pin',
    html: `
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="28" height="28" fill="currentColor" aria-hidden="true">
        <path d="M12 2C7.58 2 4 5.58 4 10c0 5.25 7 11.5 7.3 11.77a1 1 0 0 0 1.4 0C13 21.5 20 15.25 20 10c0-4.42-3.58-8-8-8Zm0 10.5a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5Z"/>
      </svg>
    `,
    iconSize: [32, 32],
    iconAnchor: [16, 28],
  })

const statusLabel = ref<string>('')
const refreshStatusLabel = () => {
  if (props.readonly) return
  if (props.latLng) {
    statusLabel.value = String(t('editor.locationStatusSelected'))
  } else {
    statusLabel.value = String(t('editor.locationStatusChoose'))
  }
}

const ensureMarker = (latlng: LatLng) => {
  if (!mapInstance) return
  if (!marker) {
    marker = L.marker([latlng.lat, latlng.lng], { icon: pinIcon() }).addTo(mapInstance)
  } else {
    marker.setLatLng([latlng.lat, latlng.lng])
  }
}

onMounted(() => {
  if (!mapEl.value) return
  const initialCenter = props.latLng ?? DEFAULT_CENTER
  const initialZoom = props.latLng ? 14 : 11

  mapInstance = L.map(mapEl.value, {
    center: [initialCenter.lat, initialCenter.lng],
    zoom: initialZoom,
    zoomControl: false,
    attributionControl: true,
    scrollWheelZoom: !props.readonly,
    dragging: true,
    doubleClickZoom: !props.readonly,
  })

  if (controlsEl.value) {
    L.DomEvent.disableClickPropagation(controlsEl.value)
    L.DomEvent.disableScrollPropagation(controlsEl.value)
  }

  tileLayer = L.tileLayer(currentTileUrl(), {
    maxZoom: 19,
    attribution: TILE_ATTRIBUTION,
  }).addTo(mapInstance)

  if (props.latLng) {
    ensureMarker(props.latLng)
  }

  if (!props.readonly) {
    mapInstance.on('click', (e: L.LeafletMouseEvent) => {
      const p: LatLng = { lat: e.latlng.lat, lng: e.latlng.lng }
      ensureMarker(p)
      emit('change', p)
    })

    if (props.autoLocate && !props.latLng && !locatedOnce) {
      locatedOnce = true
      mapInstance.locate({ setView: true, maxZoom: 14, timeout: 8000 })
      mapInstance.once('locationfound', (e: L.LocationEvent) => {
        const p: LatLng = { lat: e.latlng.lat, lng: e.latlng.lng }
        ensureMarker(p)
        emit('change', p)
      })
    }
  }

  refreshStatusLabel()
})

watch(
  () => props.latLng,
  (next) => {
    if (!mapInstance) return
    if (next) {
      ensureMarker(next)
      mapInstance.setView([next.lat, next.lng], Math.max(mapInstance.getZoom(), 13), {
        animate: true,
      })
    } else if (marker) {
      marker.remove()
      marker = null
    }
    refreshStatusLabel()
  },
  { deep: true },
)

watch(theme, () => {
  if (!mapInstance || !tileLayer) return
  tileLayer.setUrl(currentTileUrl())
})

function handleZoomIn() {
  mapInstance?.zoomIn()
}

function handleZoomOut() {
  mapInstance?.zoomOut()
}

function handleOpenInGoogleMaps() {
  if (!props.latLng) return
  const url = `https://www.google.com/maps?q=${props.latLng.lat},${props.latLng.lng}`
  window.open(url, '_blank', 'noopener,noreferrer')
}

onBeforeUnmount(() => {
  marker?.remove()
  marker = null
  tileLayer?.remove()
  tileLayer = null
  mapInstance?.remove()
  mapInstance = null
})
</script>

<style scoped>
.location-picker {
  position: relative;
  width: 100%;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-muted);
  isolation: isolate;
}

.location-picker--editable {
  height: 18rem;
}

.location-picker--readonly {
  height: 14rem;
}

.location-picker__map {
  width: 100%;
  height: 100%;
}

.location-picker__map :deep(.leaflet-container) {
  width: 100%;
  height: 100%;
  background: var(--color-bg-muted);
  font-family: inherit;
}

.location-picker__map :deep(.leaflet-control-attribution) {
  font-size: 0.65rem;
  background: var(--color-bg-surface);
  color: var(--color-text-muted);
  border-radius: var(--radius-sm);
  padding: 0 0.35rem;
}

.location-picker__map :deep(.leaflet-control-attribution a) {
  color: var(--color-accent);
}

.location-picker__map :deep(.location-picker__pin) {
  color: var(--color-accent);
  filter: drop-shadow(0 4px 6px rgb(0 0 0 / 18%));
}

.location-picker__status {
  position: absolute;
  left: 0.6rem;
  top: 0.6rem;
  z-index: 500;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem;
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 9999px;
  pointer-events: none;
  box-shadow: var(--shadow-sm);
}

.location-picker__status-dot {
  display: inline-block;
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 9999px;
  background: var(--color-accent);
}

.location-picker__controls {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 500;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  pointer-events: auto;
}

.location-picker__ctrl {
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  box-shadow: var(--shadow-sm);
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease,
    transform 0.12s ease;
}

.location-picker__ctrl:hover {
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
  transform: scale(1.04);
}

.location-picker__ctrl:active {
  transform: scale(0.96);
}

.location-picker__ctrl:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-focus-ring),
    var(--shadow-sm);
}

.location-picker__ctrl :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
}
</style>
