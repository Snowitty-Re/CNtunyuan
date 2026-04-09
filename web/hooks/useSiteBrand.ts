'use client'

import { useEffect, useMemo, useState } from 'react'
import { systemService } from '@/services/system'

type SiteBrand = {
  orgName: string
  title: string
  subtitle: string
  logoUrl: string
}

const defaultBrand: SiteBrand = {
  orgName: '助力团圆志愿者协会',
  title: '助力团圆管理后台',
  subtitle: '以温暖协作连接每一次线索与团圆',
  logoUrl: '',
}

let cachedBrand: SiteBrand | null = null

export function useSiteBrand() {
  const [brand, setBrand] = useState<SiteBrand>(cachedBrand || defaultBrand)

  useEffect(() => {
    if (cachedBrand) return
    systemService
      .bootstrapStatus()
      .then((data) => {
        const orgName = String(data?.site?.default_org_name || '').trim() || defaultBrand.orgName
        const logoUrl = String(data?.site?.logo_url || '').trim()
        const next = {
          orgName,
          title: `${orgName} 管理后台`,
          subtitle: '走失人员寻亲协作与线索闭环平台',
          logoUrl,
        }
        cachedBrand = next
        setBrand(next)
      })
      .catch(() => {
        cachedBrand = defaultBrand
        setBrand(defaultBrand)
      })
  }, [])

  return useMemo(() => brand, [brand])
}
