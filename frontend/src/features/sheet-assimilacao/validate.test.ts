import { expect, it } from 'vitest'
import { newItem, newSheet } from './types'
import { validateSheet } from './validate'

const codes = (s: ReturnType<typeof newSheet>, o = {}) => validateSheet(s, o).map((i) => i.code)

it('validate: accepts a new sheet', () => {
  expect(validateSheet(newSheet())).toEqual([])
})

it('flags instinct above max with field and blocking', () => {
  const s = newSheet()
  s.instincts.potency = 4
  const issue = validateSheet(s).find((i) => i.code === 'instinct_max')
  expect(issue).toMatchObject({ field: 'instincts.potency', blocking: true })
})

it('makes creation limits non-blocking when configured to warn', () => {
  const s = newSheet()
  s.instincts.potency = 3
  s.instincts.resolution = 3
  s.instincts.reaction = 3
  s.instincts.influence = 3
  const issues = validateSheet(s, { blockCreationLimits: false })
  expect(issues.map((i) => i.code)).toContain('instinct_total')
  expect(issues.every((i) => !i.blocking)).toBe(true)
})

it('flags aptitude total and max', () => {
  const s = newSheet()
  s.knowledge.biology = 3
  s.practices.weapons = 2
  s.practices.stealth = 2
  s.practices.crafting = 2
  expect(codes(s)).toEqual(expect.arrayContaining(['aptitude_max', 'aptitude_total']))
})

it('flags tug, xp and inventory problems', () => {
  const s = newSheet()
  s.tugOfWar.assimilationLevel = 3
  s.tugOfWar.determinationPoints = 12
  s.characteristics = [{ name: 'x', level: 5, description: '' }]
  s.inventory.body = [newItem(), newItem(), newItem(), newItem()]
  expect(codes(s)).toEqual(expect.arrayContaining(['tug_sum', 'determination_points', 'xp_overspent', 'body_full']))
})
