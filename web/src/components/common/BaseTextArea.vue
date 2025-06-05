<template>
  <div class="base-textarea w-full">
    <!-- Label -->
    <label v-if="label" :for="id" class="block text-sm font-medium text-gray-700 mb-1">
      {{ label }}
    </label>

    <!-- Textarea Wrapper -->
    <div class="relative">
      <textarea
        :id="id"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :rows="rows"
        :class="[
          'block w-full px-3 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-300 transition duration-150 ease-in-out shadow-sm sm:text-sm',
          disabled
            ? 'bg-gray-100 cursor-not-allowed opacity-70'
            : 'bg-white hover:border-orange-400',
          customClass,
        ]"
        :maxlength="maxLength"
        @input="
          $emit('update:modelValue', $event.target && ($event.target as HTMLTextAreaElement).value)
        "
        v-bind="$attrs"
      ></textarea>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: string
  id?: string
  label?: string
  placeholder?: string
  rows?: number // 默认行数
  disabled?: boolean
  readonly?: boolean
  customClass?: string
  maxLength?: number // 最大长度
}>()

const customClass = props.customClass
const rows = props.rows || 3 // 默认行数为 3
</script>

<style scoped>
.base-textarea {
  display: flex;
  flex-direction: column;
}

textarea {
  resize: vertical; /* 允许用户垂直调整大小 */
  outline: none;
}
</style>
