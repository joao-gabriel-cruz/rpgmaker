import { render } from '@testing-library/react'
import { expect, it } from 'vitest'
import type { CharacterView } from '../characters/characters'
import { DEFAULT_APPEARANCE } from '../miniature/appearance'
import { Portrait } from './HotbarStatus'

const char = (sheet: object) => ({ id: 'c', name: 'Ana B', ownerId: 'u1', sheet } as unknown as CharacterView)

it('draws the avatar from the appearance, else the portrait URL, else initials', () => {
  const kind = (c: CharacterView) => render(<Portrait character={c} />).container.querySelector('[data-avatar]')?.getAttribute('data-avatar')
  expect(kind(char({ appearance: DEFAULT_APPEARANCE, identity: { portraitUrl: 'http://x/p.png' } }))).toBe('appearance')
  expect(kind(char({ identity: { portraitUrl: 'http://x/p.png' } }))).toBe('portrait')
  expect(kind(char({}))).toBe('initials')
})
