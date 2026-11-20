import { Vector3 } from 'three'
import type { DieKind } from './faces'

export interface DieFace {
  value: number
  normal: Vector3
  center: Vector3
  vertices: Vector3[]
  inradius: number
}

export interface DieShape {
  faces: DieFace[]
  radius: number
}

const PHI = (1 + Math.sqrt(5)) / 2

function sortAround(points: Vector3[], normal: Vector3): Vector3[] {
  const ref = Math.abs(normal.y) < 0.9 ? new Vector3(0, 1, 0) : new Vector3(1, 0, 0)
  const u = ref.sub(normal.clone().multiplyScalar(ref.dot(normal))).normalize()
  const w = new Vector3().crossVectors(normal, u)
  const center = points.reduce((s, p) => s.add(p), new Vector3()).divideScalar(points.length)
  const angle = (p: Vector3) => {
    const d = p.clone().sub(center)
    return Math.atan2(d.dot(w), d.dot(u))
  }
  return [...points].sort((a, b) => angle(a) - angle(b))
}

function makeFace(shared: Vector3[], value: number): DieFace {
  const points = shared.map((p) => p.clone())
  const center = points.reduce((s, p) => s.add(p), new Vector3()).divideScalar(points.length)
  const e1 = points[1].clone().sub(points[0])
  const e2 = points[2].clone().sub(points[0])
  const normal = new Vector3().crossVectors(e1, e2).normalize()
  if (normal.dot(center) < 0) normal.negate()
  const vertices = sortAround(points, normal)
  return { value, normal, center, vertices, inradius: normal.dot(center) }
}

function facesByNormals(verts: Vector3[], normals: Vector3[]): DieFace[] {
  return normals.map((n, i) => {
    const dir = n.clone().normalize()
    const max = Math.max(...verts.map((v) => v.dot(dir)))
    return makeFace(verts.filter((v) => v.dot(dir) > max - 1e-6), i + 1)
  })
}

function signs(): number[] {
  return [-1, 1]
}

function cube(): DieFace[] {
  const verts: Vector3[] = []
  for (const x of signs()) for (const y of signs()) for (const z of signs()) verts.push(new Vector3(x, y, z))
  const normals = [new Vector3(1, 0, 0), new Vector3(-1, 0, 0), new Vector3(0, 1, 0)]
  normals.push(new Vector3(0, -1, 0), new Vector3(0, 0, 1), new Vector3(0, 0, -1))
  return facesByNormals(verts, normals)
}

function dodecahedron(): DieFace[] {
  const verts: Vector3[] = []
  const inv = 1 / PHI
  for (const x of signs()) for (const y of signs()) for (const z of signs()) verts.push(new Vector3(x, y, z))
  for (const a of signs()) for (const b of signs()) {
    verts.push(new Vector3(0, a * inv, b * PHI), new Vector3(a * inv, b * PHI, 0), new Vector3(b * PHI, 0, a * inv))
  }
  const normals: Vector3[] = []
  for (const a of signs()) for (const b of signs()) {
    normals.push(new Vector3(0, a * PHI, b), new Vector3(b, 0, a * PHI), new Vector3(a * PHI, b, 0))
  }
  return facesByNormals(verts, normals)
}

// Pentagonal trapezohedron; apex height H makes each kite exactly planar.
function trapezohedron(): DieFace[] {
  const a = 0.12
  const c = Math.cos(Math.PI / 5)
  const h = (a * (1 + c)) / (1 - c)
  const ring = (k: number) => {
    const t = (k * Math.PI) / 5
    const kk = ((k % 10) + 10) % 10
    return new Vector3(Math.cos(t), kk % 2 === 0 ? a : -a, Math.sin(t))
  }
  const top = new Vector3(0, h, 0)
  const bottom = new Vector3(0, -h, 0)
  const faces: DieFace[] = []
  for (let i = 0; i < 5; i++) faces.push(makeFace([top, ring(2 * i), ring(2 * i + 1), ring(2 * i + 2)], i + 1))
  for (let i = 0; i < 5; i++) {
    faces.push(makeFace([bottom, ring(2 * i + 1), ring(2 * i + 2), ring(2 * i + 3)], i + 6))
  }
  return faces
}

function normalize(faces: DieFace[]): DieShape {
  const radius = Math.max(...faces.flatMap((f) => f.vertices.map((v) => v.length())))
  const k = 1 / radius
  for (const f of faces) {
    f.vertices.forEach((v) => v.multiplyScalar(k))
    f.center.multiplyScalar(k)
    f.inradius *= k
  }
  return { faces, radius: 1 }
}

const cache = new Map<DieKind, DieShape>()

export function shapeFor(kind: DieKind): DieShape {
  let shape = cache.get(kind)
  if (!shape) {
    const build = { d6: cube, d10: trapezohedron, d12: dodecahedron }[kind]
    shape = normalize(build())
    cache.set(kind, shape)
  }
  return shape
}
