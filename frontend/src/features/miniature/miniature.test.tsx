import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, it, vi } from 'vitest'
import { AppearanceEditor } from './AppearanceEditor'
import { DEFAULT_APPEARANCE, normalizeAppearance, randomAppearance, type Appearance } from './appearance'
import { effectiveLevel, marksFor, MAX_MARKS } from './assimilationMarks'
import { markSlots } from './Marks'
import { cacheSize, mat } from './shared'

function Harness({ onChange }: { onChange: (a: Appearance) => void }) {
  const [a, setA] = useState<Appearance>(DEFAULT_APPEARANCE)
  return <AppearanceEditor appearance={a} onChange={(n) => { setA(n); onChange(n) }} />
}

function last(fn: ReturnType<typeof vi.fn>): Appearance {
  return fn.mock.calls[fn.mock.calls.length - 1][0]
}

it('changes styles, toggles and colors in the JSON', async () => {
  const onChange = vi.fn()
  render(<Harness onChange={onChange} />)
  await userEvent.selectOptions(screen.getByLabelText('Cabelo'), 'mohawk')
  expect(last(onChange).hair.style).toBe('mohawk')
  await userEvent.selectOptions(screen.getByLabelText('Arma'), 'axe')
  await userEvent.click(screen.getByLabelText('Mochila'))
  expect(last(onChange).accessories).toMatchObject({ weapon: 'axe', backpack: true })
  await userEvent.click(screen.getByRole('button', { name: 'Pele #6b4429' }))
  expect(last(onChange).body.skin).toBe('#6b4429')
  await userEvent.click(screen.getByRole('button', { name: 'Cor do token #f59e0b' }))
  expect(last(onChange).tokenColor).toBe('#f59e0b')
})

it('randomizes keeping the token color and restores the default', async () => {
  const onChange = vi.fn()
  render(<Harness onChange={onChange} />)
  await userEvent.click(screen.getByRole('button', { name: 'Cor do token #60a5fa' }))
  await userEvent.click(screen.getByRole('button', { name: 'Aleatório' }))
  const random = last(onChange)
  expect(random.tokenColor).toBe('#60a5fa')
  expect(normalizeAppearance(random)).toEqual(random)
  await userEvent.click(screen.getByRole('button', { name: 'Restaurar padrão' }))
  expect(last(onChange)).toEqual({ ...DEFAULT_APPEARANCE, tokenColor: '#60a5fa' })
})

it('is read only for viewers', () => {
  render(<AppearanceEditor appearance={null} onChange={vi.fn()} readOnly />)
  expect(screen.getByLabelText('Cabelo')).toBeDisabled()
  expect(screen.queryByRole('button', { name: 'Aleatório' })).toBeNull()
})

it('normalizes partial or invalid appearances', () => {
  const n = normalizeAppearance({ hair: { style: 'afro', color: 'red' }, body: { height: 5 }, infected: 42 })
  expect(n.hair).toEqual(DEFAULT_APPEARANCE.hair)
  expect(n.body.height).toBe(1.1)
  expect(n.infected).toBe(10)
  expect(normalizeAppearance(undefined)).toEqual(DEFAULT_APPEARANCE)
  const rng = (() => { let i = 0; return () => ((i += 0.137) % 1) })()
  expect(normalizeAppearance(randomAppearance(rng))).toEqual(randomAppearance((() => { let i = 0; return () => ((i += 0.137) % 1) })()))
})

it('grows the Assimilation marks with the level', () => {
  expect(marksFor(undefined)).toEqual({ count: 0, intensity: 0 })
  expect(marksFor(1).count).toBe(0)
  const levels = [2, 3, 4, 5, 6, 7, 8, 9, 10].map(marksFor)
  levels.slice(1).forEach((m, i) => {
    expect(m.count).toBeGreaterThanOrEqual(levels[i].count)
    expect(m.intensity).toBeGreaterThan(levels[i].intensity)
  })
  expect(marksFor(10)).toEqual({ count: 14, intensity: 1 })
  expect(marksFor(99).count).toBeLessThanOrEqual(MAX_MARKS)
  expect(markSlots(1)).toHaveLength(MAX_MARKS)
})

it('uses the sheet level first and the NPC infection as fallback', () => {
  expect(effectiveLevel(5, 9)).toBe(5)
  expect(effectiveLevel(undefined, 7)).toBe(7)
  expect(effectiveLevel(null, undefined)).toBe(0)
})

it('reuses materials across miniatures', () => {
  const before = cacheSize()
  expect(mat('#123456')).toBe(mat('#123456'))
  expect(cacheSize()).toBe(before + 1)
})
