Page({
  data: {
    faqList: [
      {
        question: '如何发布走失人员信息？',
        answer: '点击首页的"发布案件"按钮或进入"案件"页面点击右上角+号，填写走失人员的基本信息、失踪时间、地点和照片后提交即可。'
      },
      {
        question: '如何录制方言语音？',
        answer: '进入"工作台"页面，点击"录制方言"按钮，按住录音按钮录制15-20秒的方言语音，选择所在地区并添加相关标签后上传。'
      },
      {
        question: '如何接受任务？',
        answer: '在"工作台"页面查看待分配任务列表，点击任务卡片进入详情页，点击"接受任务"按钮即可。'
      },
      {
        question: '志愿者积分如何获得？',
        answer: '完成任务可获得10积分，成功帮助找到走失人员可获得100积分，上传方言录音可获得5积分。'
      },
      {
        question: '如何修改个人信息？',
        answer: '进入"我的"页面，点击"编辑资料"即可修改头像、昵称、手机号等信息。'
      }
    ]
  },

  onLoad() {},

  // 展开/收起问题
  toggleFaq(e) {
    const index = e.currentTarget.dataset.index
    const faqList = this.data.faqList
    faqList[index].expanded = !faqList[index].expanded
    this.setData({ faqList })
  },

  // 联系客服
  contactService() {
    wx.showModal({
      title: '联系客服',
      content: '客服电话：400-123-4567\n工作时间：9:00-18:00',
      showCancel: false
    })
  }
})
