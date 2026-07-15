/** Domain org types — must match backend entity.OrganizationType */
export const ORG_TYPE_OPTIONS = [
  { value: 'root', label: '根组织' },
  { value: 'province', label: '省级' },
  { value: 'city', label: '市级' },
  { value: 'district', label: '区县级' },
  { value: 'street', label: '街道' },
  { value: 'community', label: '社区' },
  { value: 'team', label: '小队' },
] as const

export type OrgTypeValue = (typeof ORG_TYPE_OPTIONS)[number]['value']

export const ORG_STATUS_OPTIONS = [
  { value: 'active', label: '启用' },
  { value: 'inactive', label: '停用' },
] as const

export function orgTypeLabel(type?: string): string {
  const hit = ORG_TYPE_OPTIONS.find((o) => o.value === type)
  return hit?.label || type || '-'
}

export function orgStatusLabel(status?: string): string {
  const hit = ORG_STATUS_OPTIONS.find((o) => o.value === status)
  return hit?.label || status || '-'
}
