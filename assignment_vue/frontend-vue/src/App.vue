<template>
  <main class="min-h-screen bg-slate-100 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
    <div class="mx-auto max-w-[1320px] px-4 py-8 sm:px-6 lg:px-8">
      <header class="space-y-4">
        <div class="flex items-center justify-between gap-4">
          <h1 class="text-2xl font-extrabold tracking-tight sm:text-3xl">Product discovery</h1>
          <div class="flex items-center gap-3">
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">Dark mode</span>
            <button
              type="button"
              role="switch"
              :aria-checked="isDarkMode ? 'true' : 'false'"
              class="relative h-7 w-14 overflow-hidden rounded-full border border-slate-300 bg-white transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900"
              :class="isDarkMode ? 'bg-slate-800' : 'bg-white'"
              @click="toggleDarkMode"
            >
              <span
                aria-hidden="true"
                class="absolute left-0.5 top-0.5 h-6 w-6 rounded-full bg-slate-200 shadow transition-transform duration-200 dark:bg-cyan-300"
                :class="isDarkMode ? 'translate-x-0' : 'translate-x-7'"
              ></span>
            </button>
          </div>
        </div>

        <SearchBar
          :model-value="draftFilters.search"
          :disabled="isLoading || isLoadingMore"
          @update:model-value="handleSearchUpdate"
        />
      </header>

      <div class="mt-6 grid gap-6 lg:grid-cols-[260px,1fr]">
        <aside class="hidden lg:block">
          <FiltersPanel
            :selected-categories="draftFilters.categories"
            :selected-colors="draftFilters.colors"
            :selected-conditions="draftFilters.conditions"
            :bestseller="draftFilters.bestseller"
            :on-sale="draftFilters.onSale"
            :in-stock="draftFilters.inStock"
            :min-price="draftFilters.minPrice"
            :max-price="draftFilters.maxPrice"
            :price-floor="priceFloor"
            :price-ceiling="priceCeiling"
            :category-options="categoryOptions"
            :color-options="colorOptions"
            :condition-options="conditionOptions"
            :validation-message="filterValidationMessage"
            :has-pending-changes="hasPendingChanges"
            :apply-disabled="isApplyDisabled"
            :disabled="isLoading"
            @update:selected-categories="handleCategoriesUpdate"
            @update:selected-colors="handleColorsUpdate"
            @update:selected-conditions="handleConditionsUpdate"
            @update:bestseller="handleBestsellerUpdate"
            @update:onSale="handleOnSaleUpdate"
            @update:in-stock="handleInStockUpdate"
            @update:min-price="handleMinPriceUpdate"
            @update:max-price="handleMaxPriceUpdate"
            @apply="applyFilters"
            @clear="resetDraftFilters"
          />
        </aside>

        <section aria-live="polite" :aria-busy="isLoading ? 'true' : 'false'">
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-3xl font-semibold tracking-tight text-slate-600 dark:text-slate-300">
              {{ total }} products found
            </p>
            <SortSelect
              :model-value="appliedFilters.sort"
              :disabled="isLoading || isLoadingMore"
              @update:model-value="handleSortUpdate"
            />
          </div>

          <ActiveFilterChips
            :chips="activeFilterChips"
            @remove-chip="removeAppliedChip"
            @clear-all="clearFiltersAndApply"
          />

          <div
            v-if="inlineErrorMessage"
            class="mb-4 flex items-center justify-between rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-800 dark:bg-rose-950 dark:text-rose-300"
            role="alert"
          >
            <span>{{ inlineErrorMessage }}</span>
            <button
              type="button"
              class="rounded-md border border-rose-300 bg-white px-3 py-1 font-medium text-rose-700 transition hover:bg-rose-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-rose-500 dark:border-rose-700 dark:bg-rose-900 dark:text-rose-100 dark:hover:bg-rose-800"
              @click="retryFailedRequest"
            >
              Retry
            </button>
          </div>

          <section
            v-if="showFatalErrorState"
            class="grid min-h-[360px] place-items-center rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm dark:border-slate-800 dark:bg-slate-900"
            aria-label="Error state"
          >
            <div>
              <p class="text-xl font-semibold text-slate-800 dark:text-slate-100">Could not load products</p>
              <p class="mt-2 text-slate-500 dark:text-slate-400">{{ errorMessage }}</p>
              <button
                type="button"
                class="mt-5 rounded-xl border border-slate-300 bg-white px-4 py-2 font-semibold text-slate-800 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
                @click="retryFailedRequest"
              >
                Retry
              </button>
            </div>
          </section>

          <section
            v-else-if="showEmptyState"
            class="grid min-h-[360px] place-items-center rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm dark:border-slate-800 dark:bg-slate-900"
            aria-label="Empty state"
          >
            <div>
              <p class="text-2xl font-semibold text-slate-800 dark:text-slate-100">No products found</p>
              <p class="mt-2 text-slate-500 dark:text-slate-400">
                We could not find products matching your current filters.
              </p>
              <button
                type="button"
                class="mt-5 rounded-xl border border-slate-300 bg-white px-4 py-2 font-semibold text-slate-800 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
                @click="clearFiltersAndApply"
              >
                Clear all filters
              </button>
            </div>
          </section>

          <ProductGrid
            v-else
            :products="products"
            :loading="isLoading"
            :skeleton-count="pageSize"
          />

          <LoadMoreButton
            :show="products.length > 0 && hasMore && !isLoading"
            :loading="isLoadingMore"
            :disabled="isLoadingMore"
            @click="loadMore"
          />
        </section>
      </div>

      <div
        v-if="isMobileFiltersOpen"
        id="filters-modal"
        class="fixed inset-0 z-50 lg:hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="filters-modal-title"
      >
        <div
          class="absolute inset-0 bg-slate-950/45"
          @click="closeMobileFilters"
        ></div>
        <div class="absolute inset-x-0 bottom-0 z-10 max-h-[88vh] overflow-y-auto px-4 pb-4">
          <section
            ref="mobileFiltersDialogRef"
            class="rounded-2xl"
            tabindex="-1"
          >
            <div class="mb-3 flex items-center justify-between">
              <h2 id="filters-modal-title" class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                Filters
              </h2>
              <button
                ref="mobileFiltersCloseButtonRef"
                type="button"
                class="rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
                @click="closeMobileFilters"
              >
                Close
              </button>
            </div>

            <FiltersPanel
              :selected-categories="draftFilters.categories"
              :selected-colors="draftFilters.colors"
              :selected-conditions="draftFilters.conditions"
              :bestseller="draftFilters.bestseller"
              :on-sale="draftFilters.onSale"
              :in-stock="draftFilters.inStock"
              :min-price="draftFilters.minPrice"
              :max-price="draftFilters.maxPrice"
              :price-floor="priceFloor"
              :price-ceiling="priceCeiling"
              :category-options="categoryOptions"
              :color-options="colorOptions"
              :condition-options="conditionOptions"
              :validation-message="filterValidationMessage"
              :has-pending-changes="hasPendingChanges"
              :apply-disabled="isApplyDisabled"
              :disabled="isLoading"
              @update:selected-categories="handleCategoriesUpdate"
              @update:selected-colors="handleColorsUpdate"
              @update:selected-conditions="handleConditionsUpdate"
              @update:bestseller="handleBestsellerUpdate"
              @update:onSale="handleOnSaleUpdate"
              @update:in-stock="handleInStockUpdate"
              @update:min-price="handleMinPriceUpdate"
              @update:max-price="handleMaxPriceUpdate"
              @apply="applyFilters"
              @clear="resetDraftFilters"
            />
          </section>
        </div>
      </div>
    </div>

    <div class="fixed bottom-4 right-4 z-40 flex flex-col gap-2 lg:hidden">
      <button
        ref="mobileFiltersTriggerRef"
        type="button"
        class="rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 shadow-lg transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
        :aria-expanded="isMobileFiltersOpen ? 'true' : 'false'"
        aria-controls="filters-modal"
        @click="toggleMobileFilters"
      >
        {{ isMobileFiltersOpen ? 'Hide filters' : 'Filters' }}
      </button>
      <button
        v-if="showScrollTopButton"
        type="button"
        class="rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 shadow-lg transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
        @click="scrollToTop"
      >
        Top
      </button>
    </div>
  </main>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import ActiveFilterChips from './components/ActiveFilterChips.vue'
import FiltersPanel from './components/FiltersPanel.vue'
import LoadMoreButton from './components/LoadMoreButton.vue'
import ProductGrid from './components/ProductGrid.vue'
import SearchBar from './components/SearchBar.vue'
import SortSelect from './components/SortSelect.vue'
import { useTheme } from './composables/useTheme'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
const pageSize = 6
const searchDebounceMs = 500

const colorOptions = ref([])
const priceFloor = ref(0)
const priceCeiling = ref(0)

const categoryOptions = ['smartphones', 'tablets', 'laptops', 'desktops', 'accessories']
const conditionOptions = ['new', 'refurbished', 'used']

const mobileFiltersDialogRef = ref(null)
const mobileFiltersCloseButtonRef = ref(null)
const mobileFiltersTriggerRef = ref(null)
let previouslyFocusedElement = null
let previousBodyOverflow = ''

const { isDarkMode, initializeTheme, toggleDarkMode } = useTheme('refurbed-theme')

function createDefaultFilters() {
  return {
    search: '',
    categories: [],
    colors: [],
    conditions: [],
    bestseller: 'all',
    onSale: 'all',
    inStock: 'all',
    sort: '',
    minPrice: '',
    maxPrice: '',
  }
}

const draftFilters = reactive(createDefaultFilters())
const appliedFilters = reactive(createDefaultFilters())

const products = ref([])
const total = ref(0)
const hasMore = ref(false)
const currentOffset = ref(0)

const isLoading = ref(false)
const isLoadingMore = ref(false)
const errorMessage = ref('')
const lastFailedRequest = ref('reload')
const isMobileFiltersOpen = ref(false)
const showScrollTopButton = ref(false)

let activeController = null
let searchApplyTimer = null

const inlineErrorMessage = computed(() => {
  if (products.value.length === 0 || !errorMessage.value) {
    return ''
  }
  return errorMessage.value
})

const showFatalErrorState = computed(() => {
  return products.value.length === 0 && !isLoading.value && errorMessage.value.length > 0
})

const showEmptyState = computed(() => {
  return !isLoading.value && !errorMessage.value && products.value.length === 0
})

const filterValidationMessage = computed(() => buildFilterValidationMessage(draftFilters))

const hasPendingChanges = computed(() => {
  return !areFilterStatesEqual(normalizeFilterState(draftFilters), normalizeFilterState(appliedFilters))
})

const isApplyDisabled = computed(() => {
  return (
    !hasPendingChanges.value ||
    Boolean(filterValidationMessage.value) ||
    isLoading.value ||
    isLoadingMore.value
  )
})

const activeFilterChips = computed(() => {
  const normalized = normalizeFilterState(appliedFilters)
  const chips = []

  if (normalized.search) {
    chips.push({ id: 'search', field: 'search', label: `Search: "${normalized.search}"` })
  }

  for (const category of normalized.categories) {
    chips.push({
      id: `category:${category}`,
      field: 'categories',
      value: category,
      label: `Category: ${formatTokenLabel(category)}`,
    })
  }

  for (const color of normalized.colors) {
    chips.push({
      id: `color:${color}`,
      field: 'colors',
      value: color,
      label: `Color: ${formatTokenLabel(color)}`,
    })
  }

  for (const condition of normalized.conditions) {
    chips.push({
      id: `condition:${condition}`,
      field: 'conditions',
      value: condition,
      label: `Condition: ${formatTokenLabel(condition)}`,
    })
  }

  if (normalized.bestseller === 'true') {
    chips.push({ id: 'bestseller', field: 'bestseller', label: 'Bestseller only' })
  }

  if (normalized.onSale === 'true') {
    chips.push({ id: 'onSale', field: 'onSale', label: 'On sale only' })
  }

  if (normalized.inStock === 'true') {
    chips.push({ id: 'inStock', field: 'inStock', label: 'In stock only' })
  }

  if (normalized.minPrice) {
    chips.push({ id: 'minPrice', field: 'minPrice', label: `Min price: EUR ${normalized.minPrice}` })
  }

  if (normalized.maxPrice) {
    chips.push({ id: 'maxPrice', field: 'maxPrice', label: `Max price: EUR ${normalized.maxPrice}` })
  }

  if (normalized.sort === 'popularity') {
    chips.push({ id: 'sort', field: 'sort', label: 'Sort: Popularity' })
  }

  return chips
})

watch(
  [total, isLoading],
  () => {
    if (isLoading.value) {
      document.title = 'Loading products...'
      return
    }
    document.title = `${total.value} products found | Refurbed`
  },
  { immediate: true },
)

watch(
  () => isMobileFiltersOpen.value,
  async (isOpen) => {
    if (typeof window === 'undefined' || typeof document === 'undefined') {
      return
    }

    if (isOpen) {
      previouslyFocusedElement =
        document.activeElement instanceof HTMLElement ? document.activeElement : null
      previousBodyOverflow = document.body.style.overflow
      document.body.style.overflow = 'hidden'
      window.addEventListener('keydown', handleMobileDialogKeydown)
      await nextTick()
      focusMobileDialog()
      return
    }

    document.body.style.overflow = previousBodyOverflow
    window.removeEventListener('keydown', handleMobileDialogKeydown)
    await nextTick()

    if (mobileFiltersTriggerRef.value instanceof HTMLElement) {
      mobileFiltersTriggerRef.value.focus()
    } else if (previouslyFocusedElement instanceof HTMLElement) {
      previouslyFocusedElement.focus()
    }
  },
)

function parsePrice(rawValue) {
  const normalized = String(rawValue ?? '').trim()
  if (!normalized) {
    return null
  }

  const parsed = Number.parseFloat(normalized)
  if (Number.isNaN(parsed)) {
    return Number.NaN
  }

  return parsed
}

function normalizeBestsellerValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

function normalizeOnSaleValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

function normalizeInStockValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

function normalizeSortValue(rawValue) {
  return rawValue === 'popularity' ? 'popularity' : ''
}

function normalizeTokenList(values) {
  return Array.from(
    new Set(
      values
        .map((value) => value.toLowerCase().trim())
        .filter(Boolean),
    ),
  ).sort()
}

function normalizeFilterState(source) {
  return {
    search: String(source.search ?? '').trim(),
    categories: normalizeTokenList(source.categories ?? []),
    colors: normalizeTokenList(source.colors ?? []),
    conditions: normalizeTokenList(source.conditions ?? []),
    bestseller: normalizeBestsellerValue(source.bestseller),
    onSale: normalizeOnSaleValue(source.onSale),
    inStock: normalizeInStockValue(source.inStock),
    sort: normalizeSortValue(source.sort),
    minPrice: String(source.minPrice ?? '').trim(),
    maxPrice: String(source.maxPrice ?? '').trim(),
  }
}

function assignFilterState(target, source) {
  target.search = source.search
  target.categories = [...source.categories]
  target.colors = [...source.colors]
  target.conditions = [...source.conditions]
  target.bestseller = source.bestseller
  target.onSale = source.onSale
  target.inStock = source.inStock
  target.sort = source.sort
  target.minPrice = source.minPrice
  target.maxPrice = source.maxPrice
}

function areFilterStatesEqual(left, right) {
  return JSON.stringify(left) === JSON.stringify(right)
}

function buildFilterValidationMessage(source) {
  const min = parsePrice(source.minPrice)
  const max = parsePrice(source.maxPrice)

  if (Number.isNaN(min)) {
    return 'Minimum price must be a valid number.'
  }
  if (Number.isNaN(max)) {
    return 'Maximum price must be a valid number.'
  }
  if (min !== null && min < 0) {
    return 'Minimum price must be greater than or equal to 0.'
  }
  if (max !== null && max < 0) {
    return 'Maximum price must be greater than or equal to 0.'
  }
  if (min !== null && max !== null && min > max) {
    return 'Minimum price cannot be greater than maximum price.'
  }

  return ''
}

function parseParamList(params, key) {
  return normalizeTokenList(
    params
      .getAll(key)
      .flatMap((value) => value.split(','))
      .map((value) => value.toLowerCase().trim())
      .filter(Boolean),
  )
}

function parseOffset(rawValue) {
  const parsed = Number.parseInt(rawValue ?? '', 10)
  if (Number.isNaN(parsed) || parsed < 0) {
    return 0
  }

  return Math.floor(parsed / pageSize) * pageSize
}

function parseURLState() {
  const params = new URLSearchParams(window.location.search)
  const nextFilters = createDefaultFilters()

  nextFilters.search = params.get('search')?.trim() ?? ''
  nextFilters.categories = parseParamList(params, 'category')
  nextFilters.colors = parseParamList(params, 'color')
  nextFilters.conditions = parseParamList(params, 'condition')
  nextFilters.bestseller = normalizeBestsellerValue(params.get('bestseller'))
  nextFilters.onSale = normalizeOnSaleValue(params.get('onSale'))
  nextFilters.inStock = normalizeInStockValue(params.get('inStock'))
  nextFilters.sort = normalizeSortValue(params.get('sort'))
  nextFilters.minPrice = params.get('minPrice')?.trim() ?? ''
  nextFilters.maxPrice = params.get('maxPrice')?.trim() ?? ''

  return {
    filters: normalizeFilterState(nextFilters),
    offset: parseOffset(params.get('offset')),
  }
}

function syncURLState() {
  const params = new URLSearchParams()

  if (appliedFilters.search) {
    params.set('search', appliedFilters.search)
  }

  for (const category of appliedFilters.categories) {
    params.append('category', category)
  }

  for (const color of appliedFilters.colors) {
    params.append('color', color)
  }

  for (const condition of appliedFilters.conditions) {
    params.append('condition', condition)
  }

  if (appliedFilters.bestseller !== 'all') {
    params.set('bestseller', appliedFilters.bestseller)
  }

  if (appliedFilters.onSale !== 'all') {
    params.set('onSale', appliedFilters.onSale)
  }

  if (appliedFilters.inStock !== 'all') {
    params.set('inStock', appliedFilters.inStock)
  }

  if (appliedFilters.sort) {
    params.set('sort', appliedFilters.sort)
  }

  if (appliedFilters.minPrice) {
    params.set('minPrice', appliedFilters.minPrice)
  }

  if (appliedFilters.maxPrice) {
    params.set('maxPrice', appliedFilters.maxPrice)
  }

  if (currentOffset.value > 0) {
    params.set('offset', String(currentOffset.value))
  }

  const query = params.toString()
  const nextURL = query ? `${window.location.pathname}?${query}` : window.location.pathname
  window.history.replaceState(null, '', nextURL)
}

function buildApiQuery(offset) {
  const params = new URLSearchParams()
  params.set('limit', String(pageSize))
  params.set('offset', String(offset))

  if (appliedFilters.search) {
    params.set('search', appliedFilters.search)
  }

  for (const category of appliedFilters.categories) {
    params.append('category', category)
  }

  for (const color of appliedFilters.colors) {
    params.append('color', color)
  }

  for (const condition of appliedFilters.conditions) {
    params.append('condition', condition)
  }

  if (appliedFilters.bestseller !== 'all') {
    params.set('bestseller', appliedFilters.bestseller)
  }

  if (appliedFilters.onSale !== 'all') {
    params.set('onSale', appliedFilters.onSale)
  }

  if (appliedFilters.inStock !== 'all') {
    params.set('inStock', appliedFilters.inStock)
  }

  if (appliedFilters.sort) {
    params.set('sort', appliedFilters.sort)
  }

  if (appliedFilters.minPrice) {
    params.set('minPrice', appliedFilters.minPrice)
  }

  if (appliedFilters.maxPrice) {
    params.set('maxPrice', appliedFilters.maxPrice)
  }

  return params
}

function extractColorOptionsFromPayload(payload) {
  if (Array.isArray(payload?.available_colors)) {
    return normalizeTokenList(payload.available_colors)
  }

  if (Array.isArray(payload?.items)) {
    return normalizeTokenList(
      payload.items.flatMap((item) => {
        if (Array.isArray(item?.colors)) {
          return item.colors
        }
        return []
      }),
    )
  }

  return []
}

function syncColorOptionsFromPayload(payload) {
  const available = extractColorOptionsFromPayload(payload)
  const selected = normalizeTokenList([...draftFilters.colors, ...appliedFilters.colors])
  const merged = normalizeTokenList([...available, ...selected])
  colorOptions.value = merged
}

function syncPriceBoundsFromPayload(payload) {
  const parsedMin = Number(payload?.price_min)
  const parsedMax = Number(payload?.price_max)

  if (!Number.isFinite(parsedMin) || !Number.isFinite(parsedMax) || parsedMax < parsedMin) {
    return
  }

  priceFloor.value = Math.max(0, Math.floor(parsedMin))
  priceCeiling.value = Math.max(priceFloor.value, Math.ceil(parsedMax))
}

async function fetchProductsPage(offset, signal) {
  const query = buildApiQuery(offset)
  const response = await fetch(`${apiBaseUrl}/products?${query.toString()}`, {
    method: 'GET',
    signal,
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    let message = 'Failed to load products.'
    try {
      const payload = await response.json()
      if (payload?.error) {
        message = payload.error
      }
    } catch {
      // Ignore malformed error payload.
    }
    throw new Error(message)
  }

  return response.json()
}

async function hydrateToOffset(targetOffset = 0) {
  const validationMessage = buildFilterValidationMessage(appliedFilters)
  if (validationMessage) {
    errorMessage.value = validationMessage
    products.value = []
    total.value = 0
    hasMore.value = false
    currentOffset.value = 0
    syncURLState()
    return
  }

  if (activeController) {
    activeController.abort()
  }

  const controller = new AbortController()
  activeController = controller

  lastFailedRequest.value = 'reload'
  isLoading.value = true
  errorMessage.value = ''
  products.value = []
  total.value = 0
  hasMore.value = false
  currentOffset.value = 0

  let pageOffset = 0
  const normalizedTarget = parseOffset(targetOffset)

  try {
    while (true) {
      const payload = await fetchProductsPage(pageOffset, controller.signal)
      syncColorOptionsFromPayload(payload)
      syncPriceBoundsFromPayload(payload)

      if (pageOffset === 0) {
        products.value = payload.items
      } else {
        products.value = [...products.value, ...payload.items]
      }

      total.value = payload.total
      hasMore.value = payload.has_more
      currentOffset.value = pageOffset

      if (pageOffset >= normalizedTarget || !payload.has_more) {
        break
      }

      pageOffset += pageSize
    }

    syncURLState()
  } catch (error) {
    if (controller.signal.aborted) {
      return
    }

    errorMessage.value = error instanceof Error ? error.message : 'Failed to load products.'
    products.value = []
    total.value = 0
    hasMore.value = false
    currentOffset.value = 0
    syncURLState()
  } finally {
    if (activeController === controller) {
      activeController = null
    }
    if (!controller.signal.aborted) {
      isLoading.value = false
    }
  }
}

async function loadMore() {
  if (isLoading.value || isLoadingMore.value || !hasMore.value) {
    return
  }

  if (activeController) {
    activeController.abort()
  }

  const controller = new AbortController()
  activeController = controller

  lastFailedRequest.value = 'load-more'
  isLoadingMore.value = true
  errorMessage.value = ''

  const nextOffset = currentOffset.value + pageSize

  try {
    const payload = await fetchProductsPage(nextOffset, controller.signal)
    syncColorOptionsFromPayload(payload)
    syncPriceBoundsFromPayload(payload)

    products.value = [...products.value, ...payload.items]
    total.value = payload.total
    hasMore.value = payload.has_more
    currentOffset.value = nextOffset

    syncURLState()
  } catch (error) {
    if (controller.signal.aborted) {
      return
    }
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load products.'
  } finally {
    if (activeController === controller) {
      activeController = null
    }
    if (!controller.signal.aborted) {
      isLoadingMore.value = false
    }
  }
}

function applyNormalizedFilters(nextFilters) {
  const normalized = normalizeFilterState(nextFilters)
  assignFilterState(draftFilters, normalized)
  assignFilterState(appliedFilters, normalized)
  currentOffset.value = 0
  syncURLState()
  void hydrateToOffset(0)
}

function applyFilters() {
  if (isApplyDisabled.value) {
    return
  }

  clearSearchApplyTimer()
  closeMobileFilters()
  applyNormalizedFilters(draftFilters)
}

function handleSearchUpdate(value) {
  draftFilters.search = value
  scheduleSearchApply()
}

function handleCategoriesUpdate(value) {
  draftFilters.categories = normalizeTokenList(value)
}

function handleColorsUpdate(value) {
  draftFilters.colors = normalizeTokenList(value)
}

function handleConditionsUpdate(value) {
  draftFilters.conditions = normalizeTokenList(value)
}

function handleBestsellerUpdate(value) {
  draftFilters.bestseller = normalizeBestsellerValue(value)
}

function handleOnSaleUpdate(value) {
  draftFilters.onSale = normalizeOnSaleValue(value)
}

function handleInStockUpdate(value) {
  draftFilters.inStock = normalizeInStockValue(value)
}

function handleSortUpdate(value) {
  const normalizedSort = normalizeSortValue(value)
  if (normalizedSort === appliedFilters.sort) {
    return
  }

  draftFilters.sort = normalizedSort
  appliedFilters.sort = normalizedSort
  currentOffset.value = 0
  syncURLState()
  void hydrateToOffset(0)
}

function handleMinPriceUpdate(value) {
  draftFilters.minPrice = value
}

function handleMaxPriceUpdate(value) {
  draftFilters.maxPrice = value
}

function removeAppliedChip(chip) {
  const nextFilters = normalizeFilterState(appliedFilters)

  switch (chip.field) {
    case 'search':
      nextFilters.search = ''
      break
    case 'categories':
      nextFilters.categories = nextFilters.categories.filter((value) => value !== chip.value)
      break
    case 'colors':
      nextFilters.colors = nextFilters.colors.filter((value) => value !== chip.value)
      break
    case 'conditions':
      nextFilters.conditions = nextFilters.conditions.filter((value) => value !== chip.value)
      break
    case 'bestseller':
      nextFilters.bestseller = 'all'
      break
    case 'onSale':
      nextFilters.onSale = 'all'
      break
    case 'inStock':
      nextFilters.inStock = 'all'
      break
    case 'sort':
      nextFilters.sort = ''
      break
    case 'minPrice':
      nextFilters.minPrice = ''
      break
    case 'maxPrice':
      nextFilters.maxPrice = ''
      break
    default:
      return
  }

  clearSearchApplyTimer()
  applyNormalizedFilters(nextFilters)
}

function resetDraftFilters() {
  assignFilterState(draftFilters, createDefaultFilters())
}

function clearFiltersAndApply() {
  clearSearchApplyTimer()
  closeMobileFilters()
  applyNormalizedFilters(createDefaultFilters())
}

function toggleMobileFilters() {
  if (isMobileFiltersOpen.value) {
    closeMobileFilters()
    return
  }
  openMobileFilters()
}

function openMobileFilters() {
  isMobileFiltersOpen.value = true
}

function closeMobileFilters() {
  isMobileFiltersOpen.value = false
}

function clearSearchApplyTimer() {
  if (searchApplyTimer !== null) {
    clearTimeout(searchApplyTimer)
    searchApplyTimer = null
  }
}

function scheduleSearchApply() {
  clearSearchApplyTimer()
  const normalizedSearch = String(draftFilters.search ?? '').trim()

  if (normalizedSearch === appliedFilters.search) {
    return
  }

  searchApplyTimer = setTimeout(() => {
    searchApplyTimer = null
    draftFilters.search = normalizedSearch
    appliedFilters.search = normalizedSearch
    currentOffset.value = 0
    syncURLState()
    void hydrateToOffset(0)
  }, searchDebounceMs)
}

function retryFailedRequest() {
  if (lastFailedRequest.value === 'load-more' && products.value.length > 0) {
    void loadMore()
    return
  }

  void hydrateToOffset(currentOffset.value)
}

function handleWindowScroll() {
  showScrollTopButton.value = window.scrollY > 420
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handlePopState() {
  clearSearchApplyTimer()
  const { filters, offset } = parseURLState()
  assignFilterState(draftFilters, filters)
  assignFilterState(appliedFilters, filters)
  void hydrateToOffset(offset)
}

function formatTokenLabel(token) {
  return token
    .split('-')
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' ')
}

function focusMobileDialog() {
  if (!(mobileFiltersDialogRef.value instanceof HTMLElement)) {
    return
  }

  const focusable = getFocusableElements(mobileFiltersDialogRef.value)
  if (mobileFiltersCloseButtonRef.value instanceof HTMLElement) {
    mobileFiltersCloseButtonRef.value.focus()
    return
  }

  if (focusable.length > 0) {
    focusable[0].focus()
    return
  }

  mobileFiltersDialogRef.value.focus()
}

function getFocusableElements(container) {
  const selector = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(',')

  return [...container.querySelectorAll(selector)].filter((element) => {
    return element instanceof HTMLElement && element.offsetParent !== null
  })
}

function handleMobileDialogKeydown(event) {
  if (!isMobileFiltersOpen.value) {
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    closeMobileFilters()
    return
  }

  if (event.key !== 'Tab') {
    return
  }

  if (!(mobileFiltersDialogRef.value instanceof HTMLElement)) {
    return
  }

  const focusable = getFocusableElements(mobileFiltersDialogRef.value)
  if (focusable.length === 0) {
    event.preventDefault()
    mobileFiltersDialogRef.value.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement

  if (event.shiftKey) {
    if (active === first || !mobileFiltersDialogRef.value.contains(active)) {
      event.preventDefault()
      last.focus()
    }
    return
  }

  if (active === last || !mobileFiltersDialogRef.value.contains(active)) {
    event.preventDefault()
    first.focus()
  }
}

onMounted(() => {
  initializeTheme()
  const { filters, offset } = parseURLState()
  assignFilterState(draftFilters, filters)
  assignFilterState(appliedFilters, filters)
  void hydrateToOffset(offset)

  handleWindowScroll()
  window.addEventListener('scroll', handleWindowScroll, { passive: true })
  window.addEventListener('popstate', handlePopState)
})

onBeforeUnmount(() => {
  clearSearchApplyTimer()
  window.removeEventListener('scroll', handleWindowScroll)
  window.removeEventListener('popstate', handlePopState)
  window.removeEventListener('keydown', handleMobileDialogKeydown)

  if (typeof document !== 'undefined') {
    document.body.style.overflow = previousBodyOverflow
  }

  if (activeController) {
    activeController.abort()
  }
})
</script>
