import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import FiltersPanel from '../FiltersPanel.vue'

function baseProps(overrides = {}) {
  return {
    selectedCategories: [],
    selectedColors: [],
    selectedConditions: [],
    bestseller: 'all',
    onSale: 'all',
    inStock: 'all',
    minPrice: '',
    maxPrice: '',
    priceFloor: 0,
    priceCeiling: 100,
    categoryOptions: ['smartphones'],
    colorOptions: ['blue', 'red'],
    conditionOptions: ['refurbished'],
    validationMessage: '',
    hasPendingChanges: true,
    applyDisabled: false,
    disabled: false,
    ...overrides,
  }
}

describe('FiltersPanel', () => {
  it('clamps minimum slider value to current maximum slider value', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '10',
        maxPrice: '50',
      }),
    })

    const sliders = wrapper.findAll('input[type="range"]')
    expect(sliders).toHaveLength(2)

    await sliders[0].setValue('90')

    const emitted = wrapper.emitted('update:minPrice')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toBe('50')
  })

  it('clamps maximum slider value to current minimum slider value', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '40',
        maxPrice: '60',
      }),
    })

    const sliders = wrapper.findAll('input[type="range"]')
    expect(sliders).toHaveLength(2)

    await sliders[1].setValue('20')

    const emitted = wrapper.emitted('update:maxPrice')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toBe('40')
  })

  it('shows bottom sticky actions when top controls scroll out of view', async () => {
    let observerCallback = null
    const observe = vi.fn()
    const disconnect = vi.fn()

    globalThis.IntersectionObserver = vi.fn((cb) => {
      observerCallback = cb
      return { observe, disconnect }
    })

    const wrapper = mount(FiltersPanel, {
      props: baseProps(),
    })

    expect(observe).toHaveBeenCalled()
    expect(wrapper.find('div.sticky.bottom-3').exists()).toBe(false)
    expect(wrapper.text()).toContain('You have unapplied changes.')

    observerCallback?.([{ isIntersecting: false }])
    await nextTick()

    expect(wrapper.find('div.sticky.bottom-3').exists()).toBe(true)
    expect(wrapper.text()).toContain('You have unapplied changes.')

    wrapper.unmount()
    expect(disconnect).toHaveBeenCalled()
  })
})
