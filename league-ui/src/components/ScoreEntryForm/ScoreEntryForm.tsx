import type { Match } from '../../types'
import { Button } from '../Button/Button'
import { useTranslation } from 'react-i18next'

interface ScoreEntryFormProps {
  match: Match
  gamesToWin: number
  player1Name?: string
  player2Name?: string
  onSubmit: (
    score1: number,
    score2: number,
    withdraw1: boolean,
    withdraw2: boolean) => void
  onClear?: () => void
  onClose: () => void
  loading?: boolean
}

export function ScoreEntryForm({
  match,
  gamesToWin,
  player1Name = 'Player 1',
  player2Name = 'Player 2',
  onSubmit,
  onClear,
  onClose,
  loading = false,
}: ScoreEntryFormProps) {
  const { t } = useTranslation()
  const p1Wins: [number, number][] = []
  const p2Wins: [number, number][] = []
  for (let other = 0; other < gamesToWin; other++) {
    p1Wins.push([gamesToWin, other])
    p2Wins.push([other, gamesToWin])
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-3">
        <div className="flex-1">
          <p className="text-xs font-medium text-gray-500 uppercase mb-2 truncate" title={player1Name}>
            {t('scoreEntry.wins', { name: player1Name })}
          </p>
          <div className="flex flex-col gap-1.5">
            {p1Wins.map(([s1, s2]) => (
              <button
                key={`${s1}-${s2}`}
                onClick={() => onSubmit(s1, s2, false, false)}
                disabled={loading}
                className="w-full py-2 rounded-md border border-gray-200 bg-white hover:bg-green-50 hover:border-green-400 text-sm font-semibold text-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {s1} : {s2}
              </button>
            ))}
          </div>
        </div>
        <div className="flex-1">
          <p className="text-xs font-medium text-gray-500 uppercase mb-2 truncate" title={player2Name}>
            {t('scoreEntry.wins', { name: player2Name })}
          </p>
          <div className="flex flex-col gap-1.5">
            {p2Wins.map(([s1, s2]) => (
              <button
                key={`${s1}-${s2}`}
                onClick={() => onSubmit(s1, s2, false, false)}
                disabled={loading}
                className="w-full py-2 rounded-md border border-gray-200 bg-white hover:bg-green-50 hover:border-green-400 text-sm font-semibold text-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {s1} : {s2}
              </button>
            ))}
          </div>
        </div>
      </div>
        <>
          <hr className="border-gray-100" />
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase mb-2">
              {t('scoreEntry.walkover')}
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => onSubmit(3, 0, false, true)}
                disabled={loading}
                className="flex-1 py-2 rounded-md border border-amber-300 bg-amber-50 hover:bg-amber-100 text-sm font-semibold text-amber-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                W-L
              </button>
              <button
                onClick={() => onSubmit(0, 3, true, false)}
                disabled={loading}
                className="flex-1 py-2 rounded-md border border-amber-300 bg-amber-50 hover:bg-amber-100 text-sm font-semibold text-amber-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                L-W
              </button>
            </div>
          </div>
        </>
      <div className="flex justify-between">
        {onClear && match.status !== 'DRAFT' ? (
          <Button type="button" variant="danger" onClick={onClear} disabled={loading}>
            {t('scoreEntry.clear')}
          </Button>
        ) : (
          <span />
        )}
        <Button type="button" variant="secondary" onClick={onClose} disabled={loading}>
          {t('scoreEntry.cancel')}
        </Button>
      </div>
    </div>
  )
}
