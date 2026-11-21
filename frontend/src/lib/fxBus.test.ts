import { expect, it, vi } from 'vitest'
import { fxBus } from './fxBus'

it('delivers payloads to subscribers of that name only and stops after unsubscribe', () => {
  const a = vi.fn()
  const b = vi.fn()
  const off = fxBus.on('fx:test_a', a)
  fxBus.on('fx:test_b', b)
  fxBus.emit('fx:test_a', { x: 1 })
  expect(a).toHaveBeenCalledWith({ x: 1 })
  expect(b).not.toHaveBeenCalled()
  off()
  fxBus.emit('fx:test_a', {})
  expect(a).toHaveBeenCalledTimes(1)
})
