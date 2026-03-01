Component({
  properties: {
    show: {
      type: Boolean,
      value: false
    },
    text: {
      type: String,
      value: '加载中...'
    },
    type: {
      type: String,
      value: 'default' // default, circle, bounce
    },
    fullScreen: {
      type: Boolean,
      value: false
    },
    mask: {
      type: Boolean,
      value: true
    }
  }
})
