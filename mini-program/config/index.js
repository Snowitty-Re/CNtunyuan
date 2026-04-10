const ENV = {
  dev:  { API_BASE: 'http://127.0.0.1:8080/api/v1', WS_BASE: 'ws://127.0.0.1:8080/ws' },
  prod: { API_BASE: 'https://cntuanyuan.com/api/v1', WS_BASE: 'wss://cntuanyuan.com/ws' },
}

// 强制环境开关：'' | 'dev' | 'prod'
// - ''：按自动逻辑判断
// - 'dev'：强制走本地
// - 'prod'：强制走线上
const FORCE_ENV = 'prod'

function resolveEnv() {
  // 优先使用强制环境（便于线上联调）
  if (FORCE_ENV === 'dev' || FORCE_ENV === 'prod') {
    return FORCE_ENV
  }

  // 可通过全局变量强制指定（便于本地联调）
  if (typeof globalThis !== 'undefined' && globalThis.__CNTUANYUAN_ENV__) {
    return globalThis.__CNTUANYUAN_ENV__
  }

  // 小程序运行时根据 envVersion 自动判定环境
  try {
    if (typeof wx !== 'undefined' && wx.getAccountInfoSync) {
      const accountInfo = wx.getAccountInfoSync()
      const envVersion = accountInfo && accountInfo.miniProgram && accountInfo.miniProgram.envVersion
      if (envVersion === 'develop' || envVersion === 'trial') {
        return 'dev'
      }
      return 'prod'
    }
  } catch (e) {
    // ignore
  }

  // 兜底默认 prod，避免发布误连开发地址
  return 'prod'
}

const env = resolveEnv()
module.exports = { ...ENV[env], ENV: env }
