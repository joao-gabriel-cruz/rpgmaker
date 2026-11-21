import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'
import { createCameraActions, type CameraHandle } from './cameraActions'
import { CameraPortal } from '../CameraPortal'
import { TableToolProvider, useTableTool } from '../useTableTool'
import { CameraControls } from './CameraControls'
import { isTopDown, rotateOffset, topDownOffset, zoomOffset, MAX_DISTANCE, MIN_DISTANCE } from './cameraMath'

const near = (a: number, b: number) => expect(a).toBeCloseTo(b, 5)

it('zooms within limits and rotates around the vertical axis keeping the distance', () => {
  const o = { x: 10, y: 10, z: 0 }
  expect(Math.hypot(...Object.values(zoomOffset(o, 0.5)))).toBeCloseTo(Math.hypot(10, 10) / 2, 5)
  expect(Math.hypot(...Object.values(zoomOffset(o, 0.0001)))).toBeCloseTo(MIN_DISTANCE, 5)
  expect(Math.hypot(...Object.values(zoomOffset(o, 1000)))).toBeCloseTo(MAX_DISTANCE, 5)
  const r = rotateOffset(o, 90)
  near(r.x, 0); near(r.z, -10); near(r.y, 10)
  const back = rotateOffset(r, -90)
  near(back.x, 10); near(back.z, 0)
})

it('puts the camera above the target for the top-down view', () => {
  const t = topDownOffset({ x: 10, y: 10, z: 10 })
  expect(isTopDown(t)).toBe(true)
  expect(isTopDown({ x: 10, y: 10, z: 10 })).toBe(false)
  near(Math.hypot(t.x, t.y, t.z), Math.hypot(10, 10, 10))
})

function rig() {
  const camera = { position: { x: 10, y: 10, z: 10, set(x: number, y: number, z: number) { Object.assign(this, { x, y, z }) } } }
  const controls = { target: { x: 0, y: 0, z: 0 }, update: vi.fn() }
  return { camera, controls, actions: createCameraActions({ camera, controls }) }
}

it('drives the orbit rig: zoom, rotate and a reversible top-down toggle', () => {
  const { camera, controls, actions } = rig()
  actions.zoomIn()
  expect(camera.position.y).toBeCloseTo(8)
  actions.rotate(90)
  expect(camera.position.x).toBeCloseTo(8)
  expect(camera.position.z).toBeCloseTo(-8)
  expect(controls.update).toHaveBeenCalledTimes(2)
  const before = { ...camera.position }
  expect(actions.toggleTopDown()).toBe(true)
  expect(camera.position.y).toBeGreaterThan(13.8)
  expect(Math.hypot(camera.position.x, camera.position.z)).toBeLessThan(1)
  expect(actions.toggleTopDown()).toBe(false)
  near(camera.position.x, before.x); near(camera.position.y, before.y); near(camera.position.z, before.z)
})

function mount(token = true) {
  const actions = { zoomIn: vi.fn(), zoomOut: vi.fn(), rotate: vi.fn(), toggleTopDown: vi.fn(() => true) }
  const handle: CameraHandle = { current: actions }
  const onCenterMap = vi.fn()
  const onCenterToken = vi.fn()
  render(<CameraControls handle={handle} onCenterMap={onCenterMap} onCenterToken={token ? onCenterToken : undefined} />)
  return { actions, onCenterMap, onCenterToken }
}

it('calls the right action from each camera button', async () => {
  const { actions, onCenterMap, onCenterToken } = mount()
  const click = (n: string) => userEvent.click(screen.getByRole('button', { name: n }))
  await click('Centralizar no mapa')
  await click('Centralizar no meu token')
  await click('Aproximar')
  await click('Afastar')
  await click('Girar 90° para a esquerda')
  await click('Girar 90° para a direita')
  await click('Visão de cima')
  expect(onCenterMap).toHaveBeenCalled()
  expect(onCenterToken).toHaveBeenCalled()
  expect(actions.zoomIn).toHaveBeenCalled()
  expect(actions.zoomOut).toHaveBeenCalled()
  expect(actions.rotate.mock.calls).toEqual([[-90], [90]])
  expect(actions.toggleTopDown).toHaveBeenCalled()
  expect(screen.getByRole('button', { name: 'Visão de cima' })).toHaveAttribute('aria-pressed', 'true')
})

it('disables "my token" without a token and ignores keys while typing', () => {
  const { actions } = mount(false)
  expect(screen.getByRole('button', { name: 'Centralizar no meu token' })).toBeDisabled()
  const input = document.body.appendChild(document.createElement('input'))
  fireEvent.keyDown(input, { key: 'q' })
  expect(actions.rotate).not.toHaveBeenCalled()
  fireEvent.keyDown(window, { key: 'q' })
  fireEvent.keyDown(window, { key: 'E' })
  fireEvent.keyDown(window, { key: 'z' })
  fireEvent.keyDown(window, { key: 'z', ctrlKey: true })
  expect(actions.rotate.mock.calls).toEqual([[-90], [90]])
  expect(actions.zoomIn).toHaveBeenCalledTimes(1)
})

function Host() {
  const { setCameraHost } = useTableTool()
  return <div data-testid="host" ref={setCameraHost} />
}

it('portals the camera controls into the bottom-right zone of the table', () => {
  render(<TableToolProvider isGm={false}><Host /><CameraPortal><button type="button">cam</button></CameraPortal></TableToolProvider>)
  expect(screen.getByTestId('host')).toContainElement(screen.getByRole('button', { name: 'cam' }))
})
