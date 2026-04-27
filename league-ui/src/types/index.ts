export interface User {
  userId: number
  firstName: string
  lastName: string
  email: string
  isAdmin: boolean
  currentRating: number
  deviation: number
  volatility: number
}

export interface LeagueConfig {
  numberOfAdvances: number
  numberOfRecedes: number
  gamesToWin: number
  groupSize: number
}

export interface LeagueMaintainer {
  userId: number
  firstName: string
  lastName: string
}

export interface League {
  leagueId: number
  title: string
  description: string
  configuration: LeagueConfig
  created: string
  lastUpdated: string
  eventCount?: number
  latestEventDate?: string | null
  maintainers?: LeagueMaintainer[]
}

export type EventStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'
export type GroupStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'
export type MatchStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'

export interface LeagueEvent {
  eventId: number
  leagueId: number
  status: EventStatus
  title: string
  startDate: string
  endDate: string
}

export interface Group {
  groupId: number
  eventId: number
  status: GroupStatus
  division: string
  groupNo: number
  scheduled: string
}

export interface GroupPlayer {
  groupPlayerId: number
  groupId: number
  userId: number
  seed: number
  place: number
  points: number
  tiebreakPoints: number
  advances: boolean
  recedes: boolean
  isNonCalculated: boolean
  user?: User
}

export interface Match {
  matchId: number
  groupId: number
  groupPlayer1Id: number | null
  groupPlayer2Id: number | null
  score1: number | null
  score2: number | null
  withdraw1: boolean
  withdraw2: boolean
  status: MatchStatus
}

export interface RatingHistory {
  historyId: number
  userId: number
  matchId: number
  delta: number
  rating: number
  deviation: number
  volatility: number
}

export interface GroupDetail extends Group {
  players: GroupPlayer[]
  matches: Match[]
}

export interface EventDetail extends LeagueEvent {
  groups: GroupDetail[]
}

export type WSMessageType =
  | 'match_updated'
  | 'group_finished'
  | 'event_finished'
  | 'manual_placement_required'

export interface WSMessage {
  type: WSMessageType
  groupId: number
  matchId?: number
  payload: unknown
  timestamp: string
}

export interface UserRole {
  userId: number
  leagueId: number
  roleName: 'player' | 'umpire' | 'maintainer'
}

export interface PlayerMatchSummary {
  matchId: number
  opponentId: number | null
  opponentName: string
  myScore: number | null
  oppScore: number | null
  won: boolean | null
  withdraw: boolean
  oppWithdraw: boolean
  status: MatchStatus
  ratingDelta?: number
}

export interface PlayerGroupSummary {
  groupId: number
  division: string
  groupNo: number
  status: GroupStatus
  place: number
  points: number
  advances: boolean
  recedes: boolean
  matches: PlayerMatchSummary[]
}

export interface PlayerEventSummary {
  eventId: number
  leagueId: number
  title: string
  startDate: string
  endDate: string
  status: EventStatus
  ratingDelta: number
  ratingBefore?: number
  ratingAfter?: number
  groups: PlayerGroupSummary[]
}

export interface PlayerEventsPage {
  events: PlayerEventSummary[]
  total: number
  offset: number
  limit: number
}

export interface Country {
  countryId: number
  name: string
  code: string
}

export interface City {
  cityId: number
  name: string
  countryId: number
}

export interface Blade {
  bladeId: number
  name: string
}

export interface Rubber {
  rubberId: number
  name: string
}

export interface PlayerProfileDetail {
  userId: number
  firstName: string
  lastName: string
  country: Country | null
  city: City | null
  birthdate: string | null
  grip: 'penhold' | 'shakehand' | null
  gender: 'male' | 'female' | 'other' | null
  blade: Blade | null
  fhRubber: Rubber | null
  bhRubber: Rubber | null
  isComplete: boolean
}

export function isDns(groupPlayerId: number, matches: Match[]): boolean {
  const playerMatches = matches.filter(
    (m) => m.groupPlayer1Id === groupPlayerId || m.groupPlayer2Id === groupPlayerId
  )
  if (playerMatches.length === 0) return false
  return playerMatches.every((m) =>
    m.groupPlayer1Id === groupPlayerId ? m.withdraw1 : m.withdraw2
  )
}
