import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import { DEFAULT_APPEARANCE } from '../miniature/appearance'
import { applyCharacter, type Token } from './tokens'
import { TokenFigure } from './TokenFigure'

vi.mock('./MiniatureToken', () => ({ MiniatureToken: () => <div data-testid="miniature" /> }))
vi.mock('./Standee', () => ({ Standee: () => <div data-testid="standee" /> }))

const tok = (o: Partial<Token> = {}): Token => ({
  id: 't1', sceneId: 's1', characterId: 'c1', x: 1, z: 1, name: 'Ana', portraitUrl: null, ownerId: 'u1', ...o,
})

it('draws the miniature only when the character has an appearance', () => {
  const { rerender } = render(<TokenFigure token={tok()} x={1} y={1} z={1} />)
  expect(screen.getByTestId('standee')).toBeInTheDocument()
  rerender(<TokenFigure token={tok({ appearance: null })} x={1} y={1} z={1} />)
  expect(screen.getByTestId('standee')).toBeInTheDocument()
  rerender(<TokenFigure token={tok({ appearance: DEFAULT_APPEARANCE })} x={1} y={1} z={1} />)
  expect(screen.getByTestId('miniature')).toBeInTheDocument()
  expect(screen.queryByTestId('standee')).toBeNull()
})

it('mirrors character.updated into the matching tokens', () => {
  const list = [tok(), tok({ id: 't2', characterId: 'c2' })]
  const out = applyCharacter(list, { id: 'c1', name: 'Ana B', sheet: { appearance: DEFAULT_APPEARANCE, tugOfWar: { assimilationLevel: 6 } } })
  expect(out[0]).toMatchObject({ name: 'Ana B', appearance: DEFAULT_APPEARANCE, assimilationLevel: 6 })
  expect(out[1]).toBe(list[1])
  expect(applyCharacter(out, { id: 'c1', sheet: {} })[0]).toMatchObject({ appearance: null, assimilationLevel: null })
})
