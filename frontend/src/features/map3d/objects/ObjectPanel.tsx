import type { MapGrid } from '../grid'
import { PREFAB_META_BY_ID } from './catalogData'
import { removeObject, rotateObject, setObjectProps, type MapObject } from './model'
import type { ObjectPatchHandler } from './useObjectTools'

interface Props {
  obj: MapObject
  objects: readonly MapObject[]
  grid: MapGrid
  emit: ObjectPatchHandler
  onDone: () => void
}

const btn = 'rounded bg-zinc-800 px-2 py-1 text-sm hover:bg-zinc-700'

export function ObjectPanel({ obj, objects, grid, emit, onDone }: Props) {
  const meta = PREFAB_META_BY_ID[obj.prefab]
  const open = obj.props.open === true
  return (
    <div className="absolute right-3 top-3 flex w-48 flex-col gap-2 rounded bg-zinc-900/95 p-3 text-zinc-100 shadow">
      <h3 className="text-sm font-semibold">{meta?.name ?? obj.prefab}</h3>
      <p className="text-xs text-zinc-400">
        Posição {obj.x}, {obj.z} · {obj.rotation}°
      </p>
      <button className={btn} onClick={() => emit(rotateObject(objects, grid, obj.id).patch)}>
        Girar 90°
      </button>
      {obj.prefab === 'door' && (
        <button className={btn} onClick={() => emit(setObjectProps(objects, obj.id, { open: !open }).patch)}>
          {open ? 'Fechar porta' : 'Abrir porta'}
        </button>
      )}
      <button
        className={`${btn} text-red-300`}
        onClick={() => {
          emit(removeObject(objects, obj.id).patch)
          onDone()
        }}
      >
        Remover
      </button>
    </div>
  )
}
