import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import type { WSMessage } from '../types'

const WS_BASE_URL =
  (import.meta.env.VITE_WS_URL as string | undefined) ??
  `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`

const BACKOFF_INITIAL_MS = 1_000
const BACKOFF_MAX_MS = 30_000

export function useEventWebSocket(
  eventId: number,
  onMessage: (msg: WSMessage) => void,
  { enabled = true }: { enabled?: boolean } = {}
): { connected: boolean } {
  const onMessageRef = useRef(onMessage)
  useLayoutEffect(() => {
    onMessageRef.current = onMessage
  })

  const [connected, setConnected] = useState(false)

  // Persists last received timestamp across reconnects.
  const lastTimestampRef = useRef<string | null>(null)
  // Unmount flag — stops the retry loop.
  const destroyedRef = useRef(false)
  // Current delay for exponential backoff.
  const backoffRef = useRef(BACKOFF_INITIAL_MS)
  // Holds the active WebSocket so cleanup can close it.
  const wsRef = useRef<WebSocket | null>(null)
  // Holds the pending retry timer handle.
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!enabled) {
      setConnected(false)
      return
    }

    destroyedRef.current = false
    backoffRef.current = BACKOFF_INITIAL_MS

    function buildURL(): string {
      const base = `${WS_BASE_URL}/ws/events/${eventId}`
      if (lastTimestampRef.current) {
        return `${base}?since=${encodeURIComponent(lastTimestampRef.current)}`
      }
      return base
    }

    function connect() {
      if (destroyedRef.current) return

      const ws = new WebSocket(buildURL())
      wsRef.current = ws

      ws.onopen = () => {
        if (destroyedRef.current) {
          ws.close()
          return
        }
        backoffRef.current = BACKOFF_INITIAL_MS
        setConnected(true)
      }

      ws.onmessage = (e: MessageEvent) => {
        try {
          const msg = JSON.parse(e.data as string) as WSMessage
          if (msg.timestamp) {
            lastTimestampRef.current = msg.timestamp
          }
          onMessageRef.current(msg)
        } catch {
          console.error('WebSocket message parse error', e.data)
        }
      }

      ws.onerror = () => {
        // onerror is always followed by onclose; let onclose drive reconnect.
        console.error('WebSocket error')
      }

      ws.onclose = () => {
        setConnected(false)
        wsRef.current = null
        if (destroyedRef.current) return

        const delay = backoffRef.current
        backoffRef.current = Math.min(backoffRef.current * 2, BACKOFF_MAX_MS)
        retryTimerRef.current = setTimeout(connect, delay)
      }
    }

    connect()

    return () => {
      destroyedRef.current = true
      if (retryTimerRef.current !== null) {
        clearTimeout(retryTimerRef.current)
        retryTimerRef.current = null
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
      setConnected(false)
    }
  }, [eventId, enabled])

  return { connected }
}
