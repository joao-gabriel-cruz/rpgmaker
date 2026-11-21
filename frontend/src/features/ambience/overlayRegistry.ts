/** DOM nodes of the token overlays by token id; the in-canvas projector moves them every frame. */
export type OverlayRegistry = Map<string, HTMLElement>

export const createOverlayRegistry = (): OverlayRegistry => new Map()
