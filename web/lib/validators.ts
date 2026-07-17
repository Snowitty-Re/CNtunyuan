/** 与后端 mainland 手机号规则一致 */
export const MAINLAND_PHONE_RE = /^1[3-9]\d{9}$/

export function isMainlandPhone(phone?: string | null): boolean {
  return MAINLAND_PHONE_RE.test(String(phone || '').trim())
}

export function phoneRuleMessage(phone?: string | null): string | null {
  const v = String(phone || '').trim()
  if (!v) return '请输入手机号'
  if (!isMainlandPhone(v)) return '请输入正确的大陆手机号（1[3-9] 开头 11 位）'
  return null
}
