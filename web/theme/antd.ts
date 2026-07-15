import type { ThemeConfig } from 'antd'

export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: '#d97706',
    colorInfo: '#d97706',
    colorSuccess: '#15803d',
    colorWarning: '#b45309',
    colorError: '#be123c',
    colorLink: '#b45309',
    borderRadius: 10,
    fontFamily:
      "'PingFang SC', 'Hiragino Sans GB', 'Source Han Sans SC', 'Microsoft YaHei', sans-serif",
    colorBgLayout: '#fff9f3',
    colorBgContainer: '#fffdf9',
  },
  components: {
    Layout: {
      siderBg: '#8f4f1f',
      triggerBg: '#7a4219',
      headerBg: '#fffdf9',
      bodyBg: '#fff9f3',
    },
    Menu: {
      darkItemBg: 'transparent',
      darkSubMenuItemBg: 'transparent',
      darkItemSelectedBg: 'rgba(255,255,255,0.18)',
      darkItemHoverBg: 'rgba(255,255,255,0.12)',
    },
  },
}
