import { useEffect, useLayoutEffect, useRef } from 'react'
import type { WSMessage } from '../types'

const WS_BASE_URL =
  (import.meta.env.VITE_WS_URL as string | undefined) ??
  `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`

export function useEventWebSocket(eventId: number, onMessage: (msg: WSMessage) => void) {
  const onMessageRef = useRef(onMessage)
  useLayoutEffect(() => {
    onMessageRef.current = onMessage
  })

  useEffect(() => {
    const ws = new WebSocket(`${WS_BASE_URL}/ws/events/${eventId}`)

    ws.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data as string) as WSMessage
        onMessageRef.current(msg)
      } catch {
        console.error('WebSocket message parse error', e.data)
      }
    }

    ws.onerror = () => console.error('WebSocket error')

    return () => {
      ws.close()
    }
  }, [eventId])
}
