import { Crosshair, LocateFixed, Minus, Plus, RotateCcw, RotateCw, ScanEye, type LucideIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { SURFACE } from '../surface'
import { Kbd, Tip } from '../Tip'
import type { CameraHandle } from './cameraActions'

export interface CameraControlsProps {
  handle: CameraHandle
  onCenterMap: () => void
  /** Absent when the user has no token on this scene. */
  onCenterToken?: () => void
}

const BTN =
  'group relative flex h-8 w-8 items-center justify-center rounded-md text-zinc-300 transition-[background-color,color,transform] ' +
  'hover:bg-zinc-800 hover:text-white active:scale-[0.96] disabled:opacity-40 disabled:hover:bg-transparent ' +
  'focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400'

interface BtnProps {
  icon: LucideIcon
  label: string
  keys?: string
  onClick: () => void
  pressed?: boolean
  disabled?: boolean
}

function CamButton({ icon: Icon, label, keys, onClick, pressed, disabled }: BtnProps) {
  return (
    <button type="button" aria-label={label} aria-pressed={pressed} disabled={disabled} onClick={onClick}
      className={`${BTN} ${pressed ? 'bg-emerald-600 text-white hover:bg-emerald-500' : ''}`}>
      <Icon size={16} strokeWidth={pressed ? 2.25 : 1.5} aria-hidden />
      <Tip side="left">{label}{keys && <Kbd>{keys}</Kbd>}</Tip>
    </button>
  )
}

const isTyping = (t: EventTarget | null) =>
  t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))

interface Item extends Omit<BtnProps, 'onClick'> {
  run: () => void
}

function useCameraKeys(items: Item[]) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e.target) || e.metaKey || e.ctrlKey || e.altKey || e.shiftKey) return
      const item = items.find((i) => i.keys?.toLowerCase() === e.key.toLowerCase())
      if (item && !item.disabled) item.run()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [items])
}

/** Bottom-right zone buttons (and their keys) for everyone: recenter, zoom, rotate and top-down view. */
export function CameraControls({ handle, onCenterMap, onCenterToken }: CameraControlsProps) {
  const [topDown, setTopDown] = useState(false)
  const act = (f: (a: NonNullable<CameraHandle['current']>) => void) => () => handle.current && f(handle.current)
  const items: Item[] = [
    { icon: Crosshair, label: 'Centralizar no mapa', keys: 'C', run: onCenterMap },
    { icon: LocateFixed, label: 'Centralizar no meu token', keys: 'H', run: () => onCenterToken?.(), disabled: !onCenterToken },
    { icon: Plus, label: 'Aproximar', keys: 'Z', run: act((a) => a.zoomIn()) },
    { icon: Minus, label: 'Afastar', keys: 'X', run: act((a) => a.zoomOut()) },
    { icon: RotateCcw, label: 'Girar 90° para a esquerda', keys: 'Q', run: act((a) => a.rotate(-90)) },
    { icon: RotateCw, label: 'Girar 90° para a direita', keys: 'E', run: act((a) => a.rotate(90)) },
    { icon: ScanEye, label: 'Visão de cima', keys: 'U', pressed: topDown, run: act((a) => setTopDown(a.toggleTopDown())) },
  ]
  useCameraKeys(items)
  return (
    <div role="toolbar" aria-label="Câmera" className={`pointer-events-auto flex flex-col gap-0.5 rounded-xl p-1 ${SURFACE}`}>
      {items.map(({ run, ...b }) => <CamButton key={b.label} {...b} onClick={run} />)}
    </div>
  )
}
