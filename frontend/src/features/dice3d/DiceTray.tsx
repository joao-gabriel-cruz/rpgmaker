import { Canvas, useThree } from '@react-three/fiber'
import { useEffect, useMemo, useRef } from 'react'
import { CameraRig } from './CameraRig'
import { Die } from './Die'
import type { DieResult } from './faces'
import { rngFor } from './prng'
import { TRAY_D, TRAY_W, slotFor } from './trayLayout'

export interface DiceTrayProps {
  dice: DieResult[]
  rollId: string | number
  seed: number
  selectable?: boolean
  selected?: number[]
  onToggle?: (index: number) => void
  onSettled?: () => void
  /** Size/appearance of the frame; defaults to a fixed 24rem-high card. */
  className?: string
}

function useSettleTracking(dice: DieResult[], rollId: string | number, onSettled?: () => void) {
  const settledRef = useRef<boolean[]>([])
  useEffect(() => {
    settledRef.current = dice.map(() => false)
  }, [dice, rollId])
  const markSettled = (i: number) => {
    settledRef.current[i] = true
    if (settledRef.current.every(Boolean)) onSettled?.()
  }
  return { settledRef, markSettled }
}

function Floor() {
  return (
    <>
      <ambientLight intensity={0.6} />
      <directionalLight position={[5, 10, 6]} intensity={1.6} />
      <mesh rotation={[-Math.PI / 2, 0, 0]}>
        <planeGeometry args={[TRAY_W, TRAY_D]} />
        <meshStandardMaterial color="#18181b" roughness={0.9} />
      </mesh>
    </>
  )
}

function Scene({ dice, rollId, seed, selectable, selected = [], onToggle, onSettled }: DiceTrayProps) {
  const startedAt = useThree((s) => s.clock.elapsedTime)
  const { settledRef, markSettled } = useSettleTracking(dice, rollId, onSettled)
  const slots = useMemo(
    () => dice.map((_, i) => slotFor(i, dice.length, rngFor(seed, `${rollId}:slot`, i))),
    [dice, rollId, seed],
  )
  return (
    <>
      <CameraRig total={dice.length} />
      <Floor />
      {dice.map((die, i) => (
        <Die
          key={`${rollId}:${i}`}
          die={die}
          slot={slots[i]}
          rng={rngFor(seed, rollId, i)}
          startedAt={startedAt}
          selected={selected.includes(i)}
          onSettle={() => markSettled(i)}
          onClick={() => selectable && settledRef.current[i] && onToggle?.(i)}
        />
      ))}
    </>
  )
}

export function DiceTray({ className, ...props }: DiceTrayProps) {
  return (
    <div className={className ?? 'h-96 w-full overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950'}>
      <Canvas camera={{ position: [0, 10, 8], fov: 40 }}>
        <Scene {...props} />
      </Canvas>
    </div>
  )
}
