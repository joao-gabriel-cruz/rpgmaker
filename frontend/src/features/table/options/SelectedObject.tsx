import { DoorClosed, DoorOpen, RotateCw, Trash2 } from 'lucide-react'
import type { MapGrid } from '../../map3d/grid'
import { PREFAB_META_BY_ID } from '../../map3d/objects/catalogData'
import { removeObject, rotateObject, setObjectProps, type MapObject, type ObjectPatch } from '../../map3d/objects/model'
import { Hint, BTN, BTN_DANGER } from './ui'

interface Props {
  obj: MapObject | null
  objects: readonly MapObject[]
  grid: MapGrid
  emit: (patch: ObjectPatch) => void
  onDone: () => void
}

/** Rotate / open-close / remove the object picked with the Select mode. */
export function SelectedObject({ obj, objects, grid, emit, onDone }: Props) {
  if (!obj) return <Hint>Clique em um objeto do mapa para selecioná-lo.</Hint>
  const open = obj.props.open === true
  return (
    <div role="group" aria-label="Objeto selecionado" className="flex flex-col gap-2 rounded-lg bg-zinc-950/50 p-2">
      <div>
        <h4 className="text-sm font-medium text-zinc-100">{PREFAB_META_BY_ID[obj.prefab]?.name ?? obj.prefab}</h4>
        <p className="text-xs text-zinc-400">Posição {obj.x}, {obj.z} · {obj.rotation}°</p>
      </div>
      <div className="flex flex-wrap gap-1">
        <button type="button" className={`${BTN} flex items-center gap-1`} onClick={() => emit(rotateObject(objects, grid, obj.id).patch)}>
          <RotateCw size={14} strokeWidth={1.5} aria-hidden />Girar 90°
        </button>
        {obj.prefab === 'door' && (
          <button type="button" className={`${BTN} flex items-center gap-1`} onClick={() => emit(setObjectProps(objects, obj.id, { open: !open }).patch)}>
            {open ? <DoorClosed size={14} strokeWidth={1.5} aria-hidden /> : <DoorOpen size={14} strokeWidth={1.5} aria-hidden />}
            {open ? 'Fechar porta' : 'Abrir porta'}
          </button>
        )}
        <button type="button" className={`${BTN_DANGER} flex items-center gap-1`}
          onClick={() => { emit(removeObject(objects, obj.id).patch); onDone() }}>
          <Trash2 size={14} strokeWidth={1.5} aria-hidden />Remover
        </button>
      </div>
    </div>
  )
}
