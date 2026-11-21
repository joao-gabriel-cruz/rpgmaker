export type ConnectionStatus = 'connecting' | 'open' | 'reconnecting' | 'closed'

export type Handler = (payload: unknown) => void

export interface SocketMessage {
  type: string
  payload?: unknown
  reqId?: string
}

type Pending = { resolve: () => void; reject: (e: Error) => void; timer: ReturnType<typeof setTimeout> }

export const REQUEST_TIMEOUT_MS = 10_000
const BASE_DELAY_MS = 500
const MAX_DELAY_MS = 10_000
// Events replayed to late subscribers so they never miss the state sent on connect.
const STICKY_TYPES = new Set(['presence.updated'])

export interface SocketOptions {
  url: string
  createSocket?: (url: string) => WebSocket
}

/** Framework-free WebSocket client with auto-reconnect, request/ack correlation and subscriptions. */
export class CampaignSocket {
  status: ConnectionStatus = 'connecting'
  private ws: WebSocket | null = null
  private attempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | undefined
  private nextReq = 0
  private disposed = false
  private pending = new Map<string, Pending>()
  private handlers = new Map<string, Set<Handler>>()
  private sticky = new Map<string, unknown>()
  private statusListeners = new Set<() => void>()

  private opts: SocketOptions

  constructor(opts: SocketOptions) {
    this.opts = opts
  }

  /** Starts (or restarts after close) the connection; idempotent while running. */
  open() {
    if (this.ws || this.reconnectTimer !== undefined) return
    this.disposed = false
    this.attempts = 0
    this.setStatus('connecting')
    this.connect()
  }

  private connect() {
    const make = this.opts.createSocket ?? ((u: string) => new WebSocket(u))
    const ws = make(this.opts.url)
    this.ws = ws
    ws.onopen = () => {
      this.attempts = 0
      this.setStatus('open')
    }
    ws.onmessage = (e) => this.onMessage(e.data as string)
    ws.onclose = () => this.onClose(ws)
  }

  private onClose(ws: WebSocket) {
    if (this.ws !== ws) return
    this.ws = null
    this.failPending('Conexão perdida')
    this.sticky.clear()
    if (this.disposed) return
    this.setStatus('reconnecting')
    const delay = Math.min(MAX_DELAY_MS, BASE_DELAY_MS * 2 ** this.attempts)
    this.attempts += 1
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = undefined
      this.connect()
    }, delay * (0.75 + Math.random() * 0.25))
  }

  private onMessage(raw: string) {
    let msg: SocketMessage
    try {
      msg = JSON.parse(raw)
    } catch {
      return
    }
    const reqId = msg.reqId ?? (msg.payload as { reqId?: string } | undefined)?.reqId
    if ((msg.type === 'ack' || msg.type === 'error') && reqId && this.pending.has(reqId)) {
      this.settle(reqId, msg)
      return
    }
    if (STICKY_TYPES.has(msg.type)) this.sticky.set(msg.type, msg.payload)
    this.handlers.get(msg.type)?.forEach((h) => h(msg.payload))
  }

  private settle(reqId: string, msg: SocketMessage) {
    const p = this.pending.get(reqId)
    if (!p) return
    clearTimeout(p.timer)
    this.pending.delete(reqId)
    if (msg.type === 'ack') p.resolve()
    else p.reject(new Error((msg.payload as { message?: string }).message ?? 'Algo deu errado. Tente novamente'))
  }

  private failPending(message: string) {
    for (const [id, p] of this.pending) {
      clearTimeout(p.timer)
      p.reject(new Error(message))
      this.pending.delete(id)
    }
  }

  private setStatus(s: ConnectionStatus) {
    this.status = s
    this.statusListeners.forEach((l) => l())
  }

  /** Sends a command; resolves on the server's ack and rejects with the server's PT-BR error message. */
  send(type: string, payload?: unknown): Promise<void> {
    if (this.status !== 'open' || !this.ws) return Promise.reject(new Error('Sem conexão com a mesa'))
    const reqId = `r${++this.nextReq}`
    const ws = this.ws
    return new Promise<void>((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pending.delete(reqId)
        reject(new Error('A mesa demorou para responder'))
      }, REQUEST_TIMEOUT_MS)
      this.pending.set(reqId, { resolve, reject, timer })
      ws.send(JSON.stringify({ type, payload, reqId }))
    })
  }

  /** Subscribes to a server event type; returns the unsubscribe function. */
  subscribe(type: string, handler: Handler): () => void {
    const set = this.handlers.get(type) ?? new Set<Handler>()
    set.add(handler)
    this.handlers.set(type, set)
    if (this.sticky.has(type)) handler(this.sticky.get(type))
    return () => set.delete(handler)
  }

  /** Notifies on connection status changes; returns the unsubscribe function. */
  onStatusChange(listener: () => void): () => void {
    this.statusListeners.add(listener)
    return () => this.statusListeners.delete(listener)
  }

  close() {
    this.disposed = true
    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = undefined
    this.failPending('Conexão encerrada')
    this.ws?.close()
    this.ws = null
    this.setStatus('closed')
  }
}
