<template>
  <div class="flex items-center gap-3">
    <label
      for="product-sort"
      class="text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400"
    >
      Sort
    </label>
    <select
      id="product-sort"
      :value="modelValue"
      :disabled="disabled"
      class="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 disabled:cursor-not-allowed disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
      @change="emitUpdate"
    >
      <option value="">Default</option>
      <option value="popularity">Popularity</option>
    </select>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])

function emitUpdate(event) {
  emit('update:modelValue', String(event.target.value ?? ''))
}
</script>
