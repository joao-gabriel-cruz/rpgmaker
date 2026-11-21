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

it('carries action payloads with targets to fx:<animation> subscribers', () => {
  const strike = vi.fn()
  const off = fxBus.on('fx:melee_strike', strike)
  fxBus.emit('fx:melee_strike', { x: 2, z: 3, target: { kind: 'token', tokenId: 't1' } })
  expect(strike).toHaveBeenCalledWith({ x: 2, z: 3, target: { kind: 'token', tokenId: 't1' } })
  off()
  fxBus.emit('fx:melee_strike', { x: 0, z: 0 })
  expect(strike).toHaveBeenCalledTimes(1)
})
