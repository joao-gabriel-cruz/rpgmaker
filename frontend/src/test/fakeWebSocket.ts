/** Minimal controllable WebSocket double; instances are tracked for tests to drive. */
export class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onmessage: ((e: { data: string }) => void) | null = null
  sent: string[] = []

  url: string

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    this.onclose?.()
  }

  // Test drivers
  open() {
    this.onopen?.()
  }

  receive(msg: unknown) {
    this.onmessage?.({ data: JSON.stringify(msg) })
  }

  drop() {
    this.onclose?.()
  }

  static get last() {
    return FakeWebSocket.instances[FakeWebSocket.instances.length - 1]
  }
}
