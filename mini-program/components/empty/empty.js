Component({
  properties: {
    show: {
      type: Boolean,
      value: false
    },
    text: {
      type: String,
      value: '暂无数据'
    },
    subText: {
      type: String,
      value: ''
    },
    icon: {
      type: String,
      value: 'empty'
    },
    image: {
      type: String,
      value: ''
    },
    buttonText: {
      type: String,
      value: ''
    }
  },

  methods: {
    onButtonTap() {
      this.triggerEvent('buttontap')
    }
  }
})
