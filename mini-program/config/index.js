const ENV = {
  dev:  { API_BASE: 'http://localhost:8080/api/v1', WS_BASE: 'ws://localhost:8080/ws' },
  prod: { API_BASE: 'https://cntuanyuan.com/api/v1', WS_BASE: 'wss://cntuanyuan.com/ws' },
}
const env = 'dev' // 切换环境只改这一行
module.exports = { ...ENV[env], ENV: env }
