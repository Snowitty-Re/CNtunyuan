import type { User } from '@/types/api'
import { isMainlandPhone } from '@/lib/validators'

export type SessionBlockReason = 'inactive' | 'no_phone' | null

export function getSessionBlockReason(user: User | null | undefined): SessionBlockReason {
  if (!user) return null
  const status = String(user.status || 'active').toLowerCase()
  if (status && status !== 'active') return 'inactive'
  if (!isMainlandPhone(user.phone)) return 'no_phone'
  return null
}

export function sessionBlockMessage(reason: SessionBlockReason): string {
  if (reason === 'inactive') return '账号未激活或已禁用，请等待管理员审批'
  if (reason === 'no_phone') return '请先绑定真实大陆手机号后再使用管理端'
  return ''
}
