<template>
  <section
    v-if="chips.length > 0"
    class="mb-5 flex flex-wrap items-center gap-2"
    aria-label="Active filters"
  >
    <button
      v-for="chip in chips"
      :key="chip.id"
      type="button"
      class="inline-flex items-center gap-2 rounded-full border border-slate-300 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
      @click="$emit('remove-chip', chip)"
    >
      <span>{{ chip.label }}</span>
      <span aria-hidden="true" class="text-xs">x</span>
      <span class="sr-only">Remove {{ chip.label }}</span>
    </button>

    <button
      type="button"
      class="inline-flex items-center rounded-full border border-transparent bg-slate-900 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-slate-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:bg-slate-100 dark:text-slate-900 dark:hover:bg-slate-300"
      @click="$emit('clear-all')"
    >
      Clear all
    </button>
  </section>
</template>

<script setup>
defineProps({
  chips: {
    type: Array,
    default: () => [],
  },
})

defineEmits(['remove-chip', 'clear-all'])
</script>
