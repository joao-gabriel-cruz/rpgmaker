import { render } from '@testing-library/react'
import { expect, it } from 'vitest'
import { DEFAULT_APPEARANCE } from './appearance'
import { CharacterAvatar } from './CharacterAvatar'

const avatar = (c: HTMLElement) => c.querySelector('[data-avatar]') as HTMLElement

it('draws an SVG from the appearance colors with the token color ring', () => {
  const look = { ...DEFAULT_APPEARANCE, body: { ...DEFAULT_APPEARANCE.body, skin: '#3f2a1c' }, tokenColor: '#f59e0b', accessories: { ...DEFAULT_APPEARANCE.accessories, glasses: true } }
  const { container } = render(<CharacterAvatar name="Ana B" appearance={look} portraitUrl="http://x/p.png" />)
  expect(avatar(container)).toHaveAttribute('data-avatar', 'appearance')
  expect(container.querySelector('svg')).not.toBeNull()
  expect(container.innerHTML.toLowerCase()).toContain('#3f2a1c')
  expect(container.innerHTML).toContain(DEFAULT_APPEARANCE.outfit.topColor)
  expect(container.querySelectorAll('circle').length).toBeGreaterThan(3)
  expect(avatar(container).style.boxShadow).toContain('#f59e0b')
})

it('falls back to the portrait URL and then to initials', () => {
  const { container, rerender } = render(<CharacterAvatar name="Ana B" portraitUrl="http://x/p.png" />)
  expect(avatar(container)).toHaveAttribute('data-avatar', 'portrait')
  rerender(<CharacterAvatar name="Ana Beatriz Costa" />)
  expect(avatar(container)).toHaveAttribute('data-avatar', 'initials')
  expect(avatar(container).textContent).toBe('AC')
  rerender(<CharacterAvatar name="" />)
  expect(avatar(container).textContent).toBe('?')
})

it('survives partial appearance data', () => {
  const { container } = render(<CharacterAvatar name="X" appearance={{ hair: { style: 'mohawk' } }} />)
  expect(avatar(container)).toHaveAttribute('data-avatar', 'appearance')
})
