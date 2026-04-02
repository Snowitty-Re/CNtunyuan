const styleMap: Record<string, string> = {
  missing: 'status-tag warning',
  searching: 'status-tag processing',
  found: 'status-tag success',
  reunited: 'status-tag success',
  closed: 'status-tag neutral',
  pending: 'status-tag warning',
  assigned: 'status-tag processing',
  processing: 'status-tag processing',
  completed: 'status-tag success',
  cancelled: 'status-tag danger',
  active: 'status-tag success',
  inactive: 'status-tag danger',
  approved: 'status-tag success',
  rejected: 'status-tag danger',
}

export function StatusTag({ status }: { status?: string | null }) {
  const value = (status || '-').toLowerCase()
  return <span className={styleMap[value] || 'status-tag neutral'}>{status || '-'}</span>
}

