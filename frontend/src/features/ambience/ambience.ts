import { CloudFog, CloudLightning, CloudRain, CloudSnow, Flame, Moon, Sun, Sunrise, Sunset, type LucideIcon } from 'lucide-react'

export type TimeOfDay = 'dawn' | 'day' | 'dusk' | 'night'
export type Weather = 'clear' | 'rain' | 'storm' | 'fog' | 'snow' | 'ash'

export interface Ambience {
  sceneId: string
  timeOfDay: TimeOfDay
  weather: Weather
}

export const DEFAULT_AMBIENCE = { timeOfDay: 'day', weather: 'clear' } as const

export const TIMES: { id: TimeOfDay; label: string; icon: LucideIcon }[] = [
  { id: 'dawn', label: 'Amanhecer', icon: Sunrise },
  { id: 'day', label: 'Dia', icon: Sun },
  { id: 'dusk', label: 'Entardecer', icon: Sunset },
  { id: 'night', label: 'Noite', icon: Moon },
]

export const WEATHERS: { id: Weather; label: string; icon: LucideIcon }[] = [
  { id: 'clear', label: 'Limpo', icon: Sun },
  { id: 'rain', label: 'Chuva', icon: CloudRain },
  { id: 'storm', label: 'Tempestade', icon: CloudLightning },
  { id: 'fog', label: 'Neblina', icon: CloudFog },
  { id: 'snow', label: 'Neve', icon: CloudSnow },
  { id: 'ash', label: 'Cinzas', icon: Flame },
]

export interface AtmosphereParams {
  sky: string
  ambientColor: string
  ambientIntensity: number
  sunColor: string
  sunIntensity: number
  fogColor: string
  fogNear: number
  fogFar: number
  /** 0 (full light) to 1 (night): how much token and campfire lights show. */
  darkness: number
}

// "day" keeps the original look of the table, so scenes nobody configured do not change.
const BY_TIME: Record<TimeOfDay, Omit<AtmosphereParams, 'fogColor' | 'fogNear' | 'fogFar'>> = {
  day: { sky: '#18181b', ambientColor: '#ffffff', ambientIntensity: 0.8, sunColor: '#ffffff', sunIntensity: 1.6, darkness: 0 },
  dawn: { sky: '#3a2a33', ambientColor: '#ffd7b8', ambientIntensity: 0.7, sunColor: '#ffb98a', sunIntensity: 1.2, darkness: 0.25 },
  dusk: { sky: '#2a1f33', ambientColor: '#ffb48f', ambientIntensity: 0.55, sunColor: '#ff8a5c', sunIntensity: 0.9, darkness: 0.55 },
  night: { sky: '#0b1020', ambientColor: '#7d93d6', ambientIntensity: 0.38, sunColor: '#9db4ff', sunIntensity: 0.5, darkness: 1 },
}

// Fog distance in cells; clear weather pushes it out of reach. `dim` scales the light under heavy skies.
const BY_WEATHER: Record<Weather, { color: string | null; near: number; far: number; dim: number }> = {
  clear: { color: null, near: 200, far: 1000, dim: 1 },
  rain: { color: '#3b4452', near: 14, far: 70, dim: 0.88 },
  storm: { color: '#262c38', near: 10, far: 52, dim: 0.72 },
  fog: { color: '#7a828c', near: 4, far: 24, dim: 0.9 },
  snow: { color: '#a9b8c8', near: 14, far: 76, dim: 1 },
  ash: { color: '#4a3f3a', near: 10, far: 55, dim: 0.8 },
}

/** Lighting, sky and fog for a time of day and weather (pure; the scene eases toward these values). */
export function atmosphereFor(time: TimeOfDay, weather: Weather): AtmosphereParams {
  const t = BY_TIME[time]
  const w = BY_WEATHER[weather]
  return {
    ...t,
    ambientIntensity: t.ambientIntensity * w.dim,
    sunIntensity: t.sunIntensity * w.dim,
    fogColor: w.color ?? t.sky,
    fogNear: w.near,
    fogFar: w.far,
  }
}

/** Weather that makes a rain-like sound bed. */
export const isWet = (w: Weather) => w === 'rain' || w === 'storm'

/** Lights shine from dusk on: below this darkness, torches add nothing the eye would notice. */
export const LIGHT_MIN_DARKNESS = 0.2

/** Whether a server payload is a well-formed ambience (guards against stale or unexpected bodies). */
export function isAmbience(a: unknown): a is Ambience {
  const v = a as Partial<Ambience> | null
  return !!v && TIMES.some((t) => t.id === v.timeOfDay) && WEATHERS.some((w) => w.id === v.weather)
}
