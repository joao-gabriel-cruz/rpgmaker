import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { Coord, MapGrid } from '../map3d/grid'
import { useRealtime } from '../realtime/realtimeContext'
import type { SceneTokens } from '../tokens/useSceneTokens'
import { rectCells } from './bits'
import { applyEvent, loadVision, NO_FOG, visibleTokens, type FogMode, type VisionState, type Wire } from './vision'
import type { Token } from '../tokens/tokens'

export type Brush = 'reveal' | 'hide' | null

function useVisionSync(campaignId: string, sceneId: string, grid: MapGrid, isGm: boolean) {
  const { subscribe, status } = useRealtime()
  const [state, setState] = useState<VisionState>(NO_FOG)
  const gridRef = useRef(grid)
  useEffect(() => { gridRef.current = grid }, [grid])
  const reload = useCallback(() => {
    loadVision(campaignId, sceneId, gridRef.current, isGm).then(setState, () => undefined)
  }, [campaignId, sceneId, isGm])
  useEffect(reload, [reload, status])
  useEffect(() => {
    const on = (p: unknown) => {
      const e = p as Wire & { sceneId?: string }
      if (e?.sceneId === sceneId) setState((s) => applyEvent(s, e, gridRef.current, isGm))
    }
    return subscribe(isGm ? 'fog.updated' : 'vision.updated', on)
  }, [subscribe, sceneId, isGm])
  return { state, gridRef }
}

/** Fog of the scene kept live, plus the Mestre's commands. */
export function useVision(campaignId: string, sceneId: string, grid: MapGrid, isGm: boolean) {
  const { send } = useRealtime()
  const { state, gridRef } = useVisionSync(campaignId, sceneId, grid, isGm)
  const [brush, setBrush] = useState<Brush>(null)
  const [error, setError] = useState('')

  const cmd = useCallback((type: string, extra: object) => {
    send(type, { sceneId, ...extra }).catch((e) => setError(e instanceof Error ? e.message : 'Não foi possível alterar a névoa.'))
  }, [send, sceneId])

  const paint = useCallback((cells: Coord[], reveal: boolean) => cmd(reveal ? 'fog.reveal' : 'fog.hide', { cells }), [cmd])
  const configure = useCallback((mode: FogMode, visionRadius: number) => cmd('fog.configure', { mode, visionRadius }), [cmd])
  const all = useCallback((reveal: boolean) => {
    const g = gridRef.current
    paint(rectCells({ x: 0, z: 0 }, { x: g.width - 1, z: g.depth - 1 }), reveal)
  }, [paint, gridRef])

  return { state, brush, setBrush, paint, configure, all, error, clearError: () => setError('') }
}

export type Vision = ReturnType<typeof useVision>

/** Same tokens API with the ones outside the player's sight left out. */
export function useShownTokens(tokens: SceneTokens, vision: Vision, grid: MapGrid, userId: string, isGm: boolean): SceneTokens {
  return useMemo(
    () => ({ ...tokens, tokens: visibleTokens(tokens.tokens as Token[], vision.state, grid, userId, isGm) }),
    [tokens, vision.state, grid, userId, isGm],
  )
}
