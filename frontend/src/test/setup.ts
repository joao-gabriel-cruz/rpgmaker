import '@testing-library/jest-dom/vitest'
import { cleanup, configure } from '@testing-library/react'
import { afterEach } from 'vitest'
import { FakeWebSocket } from './fakeWebSocket'

afterEach(cleanup)

// Parallel test files make the first render slow; the default 1s is flaky on loaded machines.
configure({ asyncUtilTimeout: 10000 })

// Never open real sockets in tests (not unstubbed by vi.unstubAllGlobals).
globalThis.WebSocket = FakeWebSocket as unknown as typeof WebSocket
