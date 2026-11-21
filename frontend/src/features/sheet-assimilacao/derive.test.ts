import { expect, it } from 'vitest'
import { boxesPerLevel, currentLevel, currentLevelName, isDebilitated, isIncapacitated, successPenalty, xpAvailable, xpSpent } from './derive'
import { newSheet } from './types'

function sheetWith(potency: number, resolution: number, damage: number) {
  const s = newSheet()
  s.instincts.potency = potency
  s.instincts.resolution = resolution
  s.health.damage = damage
  return s
}

it('computes boxes per level', () => {
  expect(boxesPerLevel(sheetWith(2, 1, 0))).toBe(4)
  expect(boxesPerLevel(newSheet())).toBe(3)
})

it('maps damage 9 with 4 boxes to Laceração and -1 success', () => {
  const s = sheetWith(2, 1, 9)
  expect(currentLevel(s)).toBe(2)
  expect(currentLevelName(s)).toBe('Laceração')
  expect(successPenalty(s)).toBe(1)
})

it('has no penalty when healthy and caps at level 5', () => {
  expect(successPenalty(sheetWith(2, 1, 0))).toBe(0)
  const dead = sheetWith(2, 1, 100)
  expect(currentLevel(dead)).toBe(5)
  expect(isIncapacitated(dead)).toBe(true)
  expect(successPenalty(dead)).toBe(2)
})

it('flags debilitated from level 4', () => {
  expect(isDebilitated(sheetWith(2, 1, 16))).toBe(true)
  expect(isIncapacitated(sheetWith(2, 1, 16))).toBe(false)
  expect(isDebilitated(sheetWith(2, 1, 15))).toBe(false)
})

it('computes xp spent and available', () => {
  const s = newSheet()
  s.characteristics = [
    { name: 'a', level: 1, description: '' },
    { name: 'b', level: 2, description: '' },
  ]
  expect(xpSpent(s)).toBe(6)
  expect(xpAvailable(s)).toBe(1)
})
