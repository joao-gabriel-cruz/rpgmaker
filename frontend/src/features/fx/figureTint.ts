import { Color, Mesh, type Group, type Material, type MeshStandardMaterial } from 'three'

const RED = new Color('#ef4444')

function cloneFor(clones: Map<Material, MeshStandardMaterial>, m: Material) {
  let c = clones.get(m)
  if (!c) {
    c = (m as MeshStandardMaterial).clone()
    c.transparent = true
    clones.set(m, c)
  }
  return c
}

/** Per-figure translucent/flashing copies of the shared miniature materials (the originals stay untouched). */
export function createTint(group: Group) {
  const clones = new Map<Material, MeshStandardMaterial>()
  let tinted = false
  const each = (fn: (mesh: Mesh) => void) => group.traverse((o) => { if (o instanceof Mesh) fn(o) })
  const paint = (mesh: Mesh, opacity: number, flash: number) => {
    const orig = (mesh.userData.orig ??= mesh.material) as Material
    const c = cloneFor(clones, orig)
    c.opacity = opacity
    c.emissive.copy(RED).multiplyScalar(flash)
    c.emissiveIntensity = 1
    mesh.material = c
  }
  return {
    apply(opacity: number, flash: number) {
      const on = opacity < 1 || flash > 0
      if (!on && !tinted) return
      each((mesh) => (on ? paint(mesh, opacity, flash) : void (mesh.material = (mesh.userData.orig as Material) ?? mesh.material)))
      tinted = on
    },
    dispose() {
      clones.forEach((c) => c.dispose())
      clones.clear()
    },
  }
}
