import {
  BoxGeometry, CylinderGeometry, IcosahedronGeometry, MeshStandardMaterial, TorusGeometry,
  type BufferGeometry, type Material,
} from 'three'

export type V3 = [number, number, number]

// Geometries and materials are shared by every miniature in the scene.
const geos = new Map<string, BufferGeometry>()
const mats = new Map<string, Material>()

function cached<T>(map: Map<string, T>, key: string, make: () => T): T {
  let v = map.get(key)
  if (!v) {
    v = make()
    map.set(key, v)
  }
  return v
}

export const box = (w: number, h: number, d: number) => cached(geos, `b${w},${h},${d}`, () => new BoxGeometry(w, h, d))
export const cyl = (rTop: number, r: number, h: number, seg = 8) =>
  cached(geos, `c${rTop},${r},${h},${seg}`, () => new CylinderGeometry(rTop, r, h, seg))
export const ico = (r: number) => cached(geos, `i${r}`, () => new IcosahedronGeometry(r, 1))
export const arc = (r: number, tube: number) =>
  cached(geos, `t${r},${tube}`, () => new TorusGeometry(r, tube, 5, 10, Math.PI))

export function mat(color: string, emissive?: string, glow = 0): Material {
  const g = Math.round(glow * 10) / 10
  return cached(mats, `${color}|${emissive ?? ''}|${g}`, () =>
    new MeshStandardMaterial({ color, flatShading: true, roughness: 0.9, emissive: emissive ?? '#000000', emissiveIntensity: g }))
}

export const cacheSize = () => geos.size + mats.size

interface PartProps {
  geo: BufferGeometry
  color: string
  pos?: V3
  rot?: V3
  scale?: V3
  emissive?: string
  glow?: number
}

/** One low-poly mesh using the shared geometry and material caches. */
export function Part({ geo, color, pos, rot, scale, emissive, glow }: PartProps) {
  return <mesh geometry={geo} material={mat(color, emissive, glow)} position={pos} rotation={rot} scale={scale} />
}
