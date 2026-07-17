import { caseStatusLabel, taskStatusLabel, userStatusLabel } from '@/lib/status'

const styleMap: Record<string, string> = {
  missing: 'status-tag warning',
  searching: 'status-tag processing',
  found: 'status-tag success',
  reunited: 'status-tag success',
  closed: 'status-tag neutral',
  draft: 'status-tag neutral',
  pending: 'status-tag warning',
  assigned: 'status-tag processing',
  processing: 'status-tag processing',
  in_progress: 'status-tag processing',
  completed: 'status-tag success',
  cancelled: 'status-tag danger',
  overdue: 'status-tag danger',
  active: 'status-tag success',
  inactive: 'status-tag warning',
  banned: 'status-tag danger',
  approved: 'status-tag success',
  rejected: 'status-tag danger',
}

function labelOf(status?: string | null): string {
  const key = String(status || '').toLowerCase()
  if (CASE_KEYS.has(key)) return caseStatusLabel(key)
  if (TASK_KEYS.has(key)) return taskStatusLabel(key)
  if (USER_KEYS.has(key)) return userStatusLabel(key)
  return status || '-'
}

const CASE_KEYS = new Set(['missing', 'searching', 'found', 'reunited', 'closed'])
const TASK_KEYS = new Set(['draft', 'pending', 'assigned', 'processing', 'in_progress', 'completed', 'cancelled', 'overdue'])
const USER_KEYS = new Set(['active', 'inactive', 'banned'])

export function StatusTag({ status }: { status?: string | null }) {
  const value = (status || '-').toLowerCase()
  return <span className={styleMap[value] || 'status-tag neutral'}>{labelOf(status)}</span>
}
