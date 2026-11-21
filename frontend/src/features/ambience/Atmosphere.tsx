import { useFrame, useThree } from '@react-three/fiber'
import { useEffect, useMemo, useRef } from 'react'
import { AmbientLight, Color, DirectionalLight, Fog } from 'three'
import type { MapGrid } from '../map3d/grid'
import type { MapObject } from '../map3d/objects/model'
import type { Token } from '../tokens/tokens'
import { atmosphereFor, LIGHT_MIN_DARKNESS, type Ambience } from './ambience'
import { lightPoints } from './lightPoints'
import { LightSources } from './LightSources'
import { Precipitation } from './Precipitation'
import { useLightning } from './useLightning'
import { useReducedMotion } from './useReducedMotion'

interface Props {
  ambience: Pick<Ambience, 'timeOfDay' | 'weather'>
  grid: MapGrid
  tokens: readonly Token[]
  objects: readonly MapObject[]
  /** Full daylight regardless of the ambience (Mestre editing the terrain needs to see everything). */
  neutral?: boolean
}

const EASE = 2.2

// Eases sky, fog and both lights toward the target values so a time or weather change fades in.
function useEasing(target: ReturnType<typeof atmosphereFor>, flash: { current: number }) {
  const scene = useThree((s) => s.scene)
  const ambient = useRef<AmbientLight>(null)
  const sun = useRef<DirectionalLight>(null)
  const cur = useRef({ ...target, sky: new Color(target.sky), fogColor: new Color(target.fogColor), ambientColor: new Color(target.ambientColor), sunColor: new Color(target.sunColor) })
  const level = useRef(target.darkness)
  useEffect(() => {
    scene.background = cur.current.sky
    scene.fog = new Fog(cur.current.fogColor, cur.current.fogNear, cur.current.fogFar)
    return () => { scene.fog = null }
  }, [scene])
  useFrame((_, dt) => {
    const c = cur.current
    const k = 1 - Math.exp(-Math.min(dt, 0.1) * EASE)
    c.sky.lerp(new Color(target.sky), k)
    c.fogColor.lerp(new Color(target.fogColor), k)
    c.ambientColor.lerp(new Color(target.ambientColor), k)
    c.sunColor.lerp(new Color(target.sunColor), k)
    for (const key of ['ambientIntensity', 'sunIntensity', 'fogNear', 'fogFar', 'darkness'] as const) c[key] += (target[key] - c[key]) * k
    level.current = c.darkness
    const fog = scene.fog as Fog | null
    if (fog) { fog.color.copy(c.fogColor); fog.near = c.fogNear; fog.far = c.fogFar }
    if (ambient.current) { ambient.current.color.copy(c.ambientColor); ambient.current.intensity = c.ambientIntensity + flash.current * 2.2 }
    if (sun.current) { sun.current.color.copy(c.sunColor); sun.current.intensity = c.sunIntensity + flash.current * 3 }
    flash.current *= Math.exp(-dt * 9)
  })
  return { ambient, sun, level }
}

/** Sky, lights, fog, weather particles and torch/campfire light for the scene; lives inside the Canvas. */
export function Atmosphere({ ambience, grid, tokens, objects, neutral }: Props) {
  const reduced = useReducedMotion()
  const time = neutral ? 'day' : ambience.timeOfDay
  const weather = neutral ? 'clear' : ambience.weather
  const target = useMemo(() => atmosphereFor(time, weather), [time, weather])
  const flash = useLightning(weather, reduced)
  const { ambient, sun, level } = useEasing(target, flash)
  const points = useMemo(() => (target.darkness >= LIGHT_MIN_DARKNESS ? lightPoints(grid, tokens, objects) : []), [target.darkness, grid, tokens, objects])
  const cx = grid.width / 2
  const cz = grid.depth / 2
  return (
    <>
      <ambientLight ref={ambient} />
      <directionalLight ref={sun} position={[cx + 30, 60, cz + 20]} />
      <LightSources points={points} level={level} still={reduced} />
      <Precipitation weather={weather} width={grid.width} depth={grid.depth} still={reduced} />
    </>
  )
}
