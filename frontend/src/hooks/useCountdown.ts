import { useEffect, useState } from 'react'

function formatCountdown(ms: number): string {
  if (ms <= 0) return 'Ended'
  const totalSeconds = Math.floor(ms / 1000)
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s.toString().padStart(2, '0')}s`
  return `${s}s`
}

export type CountdownState = 'normal' | 'warning' | 'critical' | 'ended'

export function useCountdown(endTime: string | null): {
  label: string
  state: CountdownState
} {
  const [ms, setMs] = useState<number>(() => {
    if (!endTime) return 0
    return new Date(endTime).getTime() - Date.now()
  })

  useEffect(() => {
    if (!endTime) return
    const tick = () => setMs(new Date(endTime).getTime() - Date.now())
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [endTime])

  const label = formatCountdown(ms)
  const state: CountdownState =
    ms <= 0 ? 'ended' : ms < 5 * 60 * 1000 ? 'critical' : ms < 60 * 60 * 1000 ? 'warning' : 'normal'

  return { label, state }
}
