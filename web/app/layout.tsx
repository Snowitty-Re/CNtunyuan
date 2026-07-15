import './globals.css'
import type { Metadata } from 'next'
import { AppProviders } from '@/components/providers/AppProviders'

export const metadata: Metadata = {
  title: '助力团圆 Web',
  description: '团圆寻亲志愿者系统 Web 管理端',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  )
}
