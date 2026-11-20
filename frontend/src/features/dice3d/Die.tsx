import { useFrame } from '@react-three/fiber'
import { useMemo, useRef } from 'react'
import { BufferGeometry, Float32BufferAttribute, Group, Matrix4, Quaternion, Vector3 } from 'three'
import { faceTexture } from './faceTexture'
import type { DieResult } from './faces'
import { shapeFor, type DieFace } from './polyhedra'
import { planRoll, type RollPlan } from './rollPlan'
import type { Rng } from './prng'

const COLORS = { normal: '#e4e4e7', difficulty: '#7f1d1d' }

function bodyGeometry(faces: DieFace[]): BufferGeometry {
  const pos: number[] = []
  for (const f of faces) {
    for (let i = 1; i < f.vertices.length - 1; i++) {
      for (const v of [f.vertices[0], f.vertices[i], f.vertices[i + 1]]) pos.push(v.x, v.y, v.z)
    }
  }
  const g = new BufferGeometry()
  g.setAttribute('position', new Float32BufferAttribute(pos, 3))
  g.computeVertexNormals()
  return g
}

function decalQuaternion(f: DieFace): Quaternion {
  const y = f.vertices[0].clone().sub(f.center)
  y.sub(f.normal.clone().multiplyScalar(y.dot(f.normal))).normalize()
  const x = new Vector3().crossVectors(y, f.normal)
  return new Quaternion().setFromRotationMatrix(new Matrix4().makeBasis(x, y, f.normal))
}

function decalSize(f: DieFace): number {
  const edge = f.vertices[0].distanceTo(f.vertices[1])
  return Math.min(edge * 1.5, f.inradius * 1.6)
}

interface Props {
  die: DieResult
  slot: Vector3
  rng: Rng
  startedAt: number
  selected: boolean
  onSettle: () => void
  onClick: () => void
}

function useRollAnimation(plan: RollPlan, startedAt: number, onSettle: () => void) {
  const group = useRef<Group>(null)
  const settled = useRef(false)
  useFrame(({ clock }) => {
    const g = group.current
    if (!g) return
    const t = clock.elapsedTime - startedAt
    const pose = plan.sample(t)
    g.position.copy(pose.position)
    g.quaternion.copy(pose.quaternion)
    if (!settled.current && t >= plan.duration) {
      settled.current = true
      onSettle()
    }
  })
  return group
}

export function Die({ die, slot, rng, startedAt, selected, onSettle, onClick }: Props) {
  const shape = shapeFor(die.kind)
  const geometry = useMemo(() => bodyGeometry(shape.faces), [shape])
  const plan = useMemo(() => planRoll(shape, die.face, slot, rng), [shape, die.face, slot, rng])
  const group = useRollAnimation(plan, startedAt, onSettle)
  return (
    <group ref={group} onClick={(e) => (e.stopPropagation(), onClick())}>
      <mesh geometry={geometry}>
        <meshStandardMaterial
          color={COLORS[die.variant ?? 'normal']}
          flatShading
          roughness={0.35}
          emissive={selected ? '#facc15' : '#000000'}
          emissiveIntensity={selected ? 0.55 : 0}
        />
      </mesh>
      {shape.faces.map((f) => (
        <FaceDecal key={f.value} face={f} kind={die.kind} />
      ))}
    </group>
  )
}

function FaceDecal({ face, kind }: { face: DieFace; kind: DieResult['kind'] }) {
  const texture = faceTexture(kind, face.value)
  const quaternion = useMemo(() => decalQuaternion(face), [face])
  if (!texture) return null
  const size = decalSize(face)
  const pos = face.center.clone().add(face.normal.clone().multiplyScalar(0.004))
  return (
    <mesh position={pos} quaternion={quaternion}>
      <planeGeometry args={[size, size]} />
      <meshBasicMaterial map={texture} transparent depthWrite={false} polygonOffset polygonOffsetFactor={-2} />
    </mesh>
  )
}
