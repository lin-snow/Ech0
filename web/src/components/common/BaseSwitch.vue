<template>
  <button
    type="button"
    class="toggle-switch-btn"
    :class="[customClass, { 'is-on': modelValue, 'is-disabled': disabled }]"
    :disabled="disabled"
    @click="onToggle"
  >
    <span class="toggle-track">
      <span class="toggle-thumb"></span>
    </span>
    <span class="toggle-label">
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
  disabled?: boolean
  class?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'click', event: MouseEvent): void
}>()

const customClass = props.class

function onToggle(event: MouseEvent) {
  if (!props.disabled) {
    emit('update:modelValue', !props.modelValue)
    emit('click', event)
  }
}
</script>

<style scoped>
.toggle-switch-btn {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  background: none;
  border: none;
  outline: none;
  padding: 0.25rem 0.5rem;
  user-select: none;
  transition: opacity 0.2s;
}
.toggle-switch-btn.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.toggle-track {
  width: 40px;
  height: 20px;
  background: #ccc;
  border-radius: 5px;
  position: relative;
  transition: background 0.3s;
  margin-right: 0.5rem;
  flex-shrink: 0;
}
.toggle-switch-btn.is-on .toggle-track {
  background: #ebc78c;
}
.toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background: #fff;
  border-radius: 25%;
  transition: transform 0.3s;
}
.toggle-switch-btn.is-on .toggle-thumb {
  transform: translateX(20px);
}
.toggle-label {
  font-size: 0.95em;
  color: #333;
}
</style>
