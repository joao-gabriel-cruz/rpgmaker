import { Vector3 } from 'three'
import { describe, expect, it } from 'vitest'
import { countSymbols, facesFor, type DieKind } from './faces'
import { shapeFor } from './polyhedra'
import { createRng, rngFor } from './prng'
import { planRoll } from './rollPlan'
import { layoutSymbols } from './symbolIcons'

const KINDS: DieKind[] = ['d6', 'd10', 'd12']

describe('faces model', () => {
  it('has the spec table for d12', () => {
    const f = facesFor('d12')
    expect(f).toHaveLength(12)
    expect(f[0]).toEqual([])
    expect(f[1]).toEqual([])
    expect(f[2]).toEqual(['pressure'])
    expect(f[4]).toEqual(['adaptation', 'pressure'])
    expect(f[6]).toEqual(['success', 'success'])
    expect(f[10]).toEqual(['success', 'adaptation', 'adaptation', 'pressure'])
    expect(f[11]).toEqual(['pressure', 'pressure'])
  })
  it('limits face count per kind', () => {
    expect(facesFor('d6')).toHaveLength(6)
    expect(facesFor('d10')).toHaveLength(10)
  })
  it('counts symbols', () => {
    const c = countSymbols([
      { kind: 'd6', face: 6 },
      { kind: 'd10', face: 10 },
      { kind: 'd12', face: 12 },
      { kind: 'd12', face: 1 },
    ])
    expect(c).toEqual({ success: 3, adaptation: 0, pressure: 3 })
  })
  it('lays out up to 4 symbols', () => {
    for (let n = 1; n <= 4; n++) expect(layoutSymbols(n)).toHaveLength(n)
  })
})

describe('geometry', () => {
  it.each([['d6', 4], ['d10', 4], ['d12', 5]] as const)('%s faces are planar polygons', (kind, verts) => {
    const shape = shapeFor(kind)
    expect(shape.faces).toHaveLength(facesFor(kind).length)
    for (const f of shape.faces) {
      expect(f.vertices).toHaveLength(verts)
      for (const v of f.vertices) expect(Math.abs(v.clone().sub(f.center).dot(f.normal))).toBeLessThan(1e-6)
      expect(f.inradius).toBeGreaterThan(0)
    }
  })
})

describe('prng', () => {
  it('is deterministic per seed', () => {
    const a = createRng(42)
    const b = createRng(42)
    expect([a(), a(), a()]).toEqual([b(), b(), b()])
    expect(createRng(1)()).not.toEqual(createRng(2)())
  })
  it('derives the same stream for same seed/rollId/index', () => {
    expect(rngFor(7, 'r1', 2)()).toEqual(rngFor(7, 'r1', 2)())
    expect(rngFor(7, 'r1', 2)()).not.toEqual(rngFor(7, 'r2', 2)())
  })
})

describe('roll plan', () => {
  it.each(KINDS)('%s ends with the target face up for every face', (kind) => {
    const shape = shapeFor(kind)
    for (const face of shape.faces) {
      const plan = planRoll(shape, face.value, new Vector3(1, 0, 1), rngFor(3, 'x', face.value))
      for (const q of [plan.finalQuaternion, plan.sample(plan.duration).quaternion, plan.sample(99).quaternion]) {
        const up = face.normal.clone().applyQuaternion(q)
        expect(Math.abs(up.dot(new Vector3(0, 1, 0)) - 1)).toBeLessThan(1e-6)
      }
      expect(plan.sample(plan.duration).position.y).toBeCloseTo(face.inradius, 6)
    }
  })
  it('is reproducible', () => {
    const shape = shapeFor('d10')
    const mk = () => planRoll(shape, 4, new Vector3(), rngFor(5, 9, 0)).sample(0.7)
    expect(mk().position.toArray()).toEqual(mk().position.toArray())
    expect(mk().quaternion.toArray()).toEqual(mk().quaternion.toArray())
  })
})
