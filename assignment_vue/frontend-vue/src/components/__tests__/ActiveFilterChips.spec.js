import { mount } from '@vue/test-utils'
import ActiveFilterChips from '../ActiveFilterChips.vue'

describe('ActiveFilterChips', () => {
  it('renders active chips and emits removal events', async () => {
    const chips = [
      { id: 'search', label: 'Search: "iphone"' },
      { id: 'sort', label: 'Sort: Popularity' },
    ]

    const wrapper = mount(ActiveFilterChips, {
      props: { chips },
    })

    const chipButtons = wrapper.findAll('button')
    expect(chipButtons).toHaveLength(3)
    expect(wrapper.text()).toContain('Search: "iphone"')
    expect(wrapper.text()).toContain('Sort: Popularity')

    await chipButtons[0].trigger('click')
    expect(wrapper.emitted('remove-chip')).toBeTruthy()
    expect(wrapper.emitted('remove-chip')[0][0]).toEqual(chips[0])

    await chipButtons[2].trigger('click')
    expect(wrapper.emitted('clear-all')).toBeTruthy()
  })
})
