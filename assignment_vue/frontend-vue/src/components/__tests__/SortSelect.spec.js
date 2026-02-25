import { mount } from '@vue/test-utils'
import SortSelect from '../SortSelect.vue'

describe('SortSelect', () => {
  it('emits selected sort value', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: '',
      },
    })

    const select = wrapper.get('select')
    await select.setValue('popularity')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')[0][0]).toBe('popularity')
  })
})
