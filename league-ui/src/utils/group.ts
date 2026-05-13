const DIVISION_ORDER = ['S', 'A', 'B', 'C', 'D', 'E', 'F']

export function groupTitle(division: string, groupNo: number): string {
  if (division === 'S' || division === 'Superleague') return 'Superleague'
  return `${division}${groupNo}`
}

export function groupSortKey(division: string, groupNo: number): number {
  const idx = DIVISION_ORDER.indexOf(division.toUpperCase())
  return (idx === -1 ? 99 : idx) * 1000 + groupNo
}
