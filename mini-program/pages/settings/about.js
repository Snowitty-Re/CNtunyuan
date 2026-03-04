Page({
  data: {
    version: '1.0.0',
    appInfo: {
      name: '团圆寻亲',
      slogan: '用爱点亮回家的路',
      description: '团圆寻亲志愿者系统是一个帮助寻找走失人员的公益平台，通过整合志愿者网络、方言语音数据库和任务系统，提高寻人效率。'
    },
    features: [
      { icon: '&#xe6ac;', title: '走失人员登记', desc: '快速登记走失人员信息' },
      { icon: '&#xe6ad;', title: '方言语音识别', desc: '通过方言确认身份' },
      { icon: '&#xe6ae;', title: '任务分配系统', desc: '高效协作寻人' },
      { icon: '&#xe6af;', title: '志愿者网络', desc: '全国志愿者联动' }
    ]
  },

  onLoad() {
    // 获取版本号
    const accountInfo = wx.getAccountInfoSync()
    if (accountInfo && accountInfo.miniProgram) {
      this.setData({
        version: accountInfo.miniProgram.version || '1.0.0'
      })
    }
  },

  // 检查更新
  checkUpdate() {
    const updateManager = wx.getUpdateManager()
    
    updateManager.onCheckForUpdate((res) => {
      if (res.hasUpdate) {
        wx.showLoading({ title: '更新中...' })
      } else {
        wx.showToast({
          title: '已是最新版本',
          icon: 'success'
        })
      }
    })
    
    updateManager.onUpdateReady(() => {
      wx.hideLoading()
      wx.showModal({
        title: '更新提示',
        content: '新版本已经准备好，是否重启应用？',
        success: (res) => {
          if (res.confirm) {
            updateManager.applyUpdate()
          }
        }
      })
    })
  }
})
