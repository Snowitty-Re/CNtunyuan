// 统一导出所有服务
module.exports = {
  auth: require('./auth'),
  user: require('./user'),
  missingPerson: require('./missingPerson'),
  dialect: require('./dialect'),
  task: require('./task'),
  upload: require('./upload'),
  dashboard: require('./dashboard'),
  organization: require('./organization')
}
