import { useEffect, useState } from 'react'
import { Row, Col, Statistic, Table, Tag, Carousel } from 'antd'
import {
  TeamOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  SoundOutlined,
  FileTextOutlined,
  RiseOutlined,
} from '@ant-design/icons'
import { Pie, Liquid, Gauge } from '@ant-design/charts'
import '../styles/big-screen.css'

const BigScreen = () => {
  const [currentTime, setCurrentTime] = useState(new Date())

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  const caseStatusData = [
    { type: '已团圆', value: 523, color: '#52c41a' },
    { type: '寻找中', value: 156, color: '#1890ff' },
    { type: '待核实', value: 89, color: '#faad14' },
  ]

  const pieConfig = {
    data: caseStatusData,
    angleField: 'value',
    colorField: 'type',
    radius: 0.8,
    innerRadius: 0.6,
    color: ['#52c41a', '#1890ff', '#faad14'],
    label: {
      type: 'outer',
      content: '{name}\n{percentage}',
    },
    statistic: {
      title: {
        content: '案件分布',
        style: {
          color: '#fff',
        },
      },
    },
  }

  const successRateConfig = {
    percent: 0.68,
    outline: {
      border: 4,
      distance: 8,
    },
    wave: {
      length: 128,
    },
    statistic: {
      content: {
        style: {
          fontSize: 24,
          fill: '#fff',
        },
      },
    },
  }

  const recentCases = [
    { id: 1, name: '张大爷', age: 78, status: 'found', location: '北京市朝阳区', time: '2小时前', volunteer: '李志愿者' },
    { id: 2, name: '小明', age: 6, status: 'searching', location: '上海市浦东区', time: '5小时前', volunteer: '王志愿者' },
    { id: 3, name: '李奶奶', age: 82, status: 'reunited', location: '广州市天河区', time: '1天前', volunteer: '张志愿者' },
    { id: 4, name: '王先生', age: 65, status: 'found', location: '深圳市南山区', time: '2天前', volunteer: '刘志愿者' },
    { id: 5, name: '陈阿姨', age: 71, status: 'searching', location: '杭州市西湖区', time: '3天前', volunteer: '赵志愿者' },
  ]

  const columns = [
    { title: '姓名', dataIndex: 'name', render: (text: string) => <span style={{ color: '#fff' }}>{text}</span> },
    { title: '年龄', dataIndex: 'age', render: (text: number) => <span style={{ color: '#fff' }}>{text}岁</span> },
    { title: '走失地点', dataIndex: 'location', render: (text: string) => <span style={{ color: '#fff' }}>{text}</span> },
    { title: '时间', dataIndex: 'time', render: (text: string) => <span style={{ color: '#fff' }}>{text}</span> },
    { title: '负责人', dataIndex: 'volunteer', render: (text: string) => <span style={{ color: '#fff' }}>{text}</span> },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          searching: { color: '#1890ff', text: '寻找中' },
          found: { color: '#52c41a', text: '已找到' },
          reunited: { color: '#722ed1', text: '已团圆' },
        }
        const { color, text } = statusMap[status] || { color: '#fff', text: status }
        return <Tag color={color} style={{ border: 'none' }}>{text}</Tag>
      },
    },
  ]

  const notices = [
    '🎉 恭喜！张大爷已于今日上午与家人团圆！',
    '📢 招募志愿者：上海市浦东新区需要5名志愿者参与实地寻访',
    '📹 新的方言录音已上传：四川话-成都地区',
    '⭐ 本月优秀志愿者：李志愿者已帮助3个家庭团圆',
  ]

  return (
    <div className="big-screen">
      {/* Header */}
      <div className="big-screen-header">
        <div className="big-screen-title">
          <h1>团圆寻亲志愿者系统</h1>
          <p>数据展示大屏</p>
        </div>
        <div className="big-screen-time">
          {currentTime.toLocaleString('zh-CN', { 
            year: 'numeric', 
            month: '2-digit', 
            day: '2-digit', 
            hour: '2-digit', 
            minute: '2-digit', 
            second: '2-digit',
            weekday: 'long'
          })}
        </div>
      </div>

      {/* Marquee */}
      <div className="big-screen-marquee">
        <Carousel autoplay vertical dots={false}>
          {notices.map((notice, index) => (
            <div key={index} className="marquee-item">{notice}</div>
          ))}
        </Carousel>
      </div>

      {/* Main Content */}
      <div className="big-screen-content">
        {/* Left Column */}
        <div className="big-screen-column">
          <div className="big-screen-card">
            <h3 className="card-title">志愿者统计</h3>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <div className="stat-item">
                  <TeamOutlined className="stat-icon" />
                  <div>
                    <div className="stat-value">2,856</div>
                    <div className="stat-label">注册志愿者</div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div className="stat-item">
                  <RiseOutlined className="stat-icon" />
                  <div>
                    <div className="stat-value">168</div>
                    <div className="stat-label">本月新增</div>
                  </div>
                </div>
              </Col>
            </Row>
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">案件分布</h3>
            <div style={{ height: 250 }}>
              <Pie {...pieConfig} />
            </div>
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">成功找回率</h3>
            <div style={{ height: 200, display: 'flex', justifyContent: 'center' }}>
              <Liquid {...successRateConfig} />
            </div>
          </div>
        </div>

        {/* Middle Column */}
        <div className="big-screen-column">
          <div className="big-screen-card highlight-card">
            <Row gutter={[24, 24]}>
              <Col span={12}>
                <div className="highlight-stat">
                  <SearchOutlined className="highlight-icon" />
                  <div>
                    <div className="highlight-value">768</div>
                    <div className="highlight-label">累计案件</div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div className="highlight-stat">
                  <CheckCircleOutlined className="highlight-icon" />
                  <div>
                    <div className="highlight-value">523</div>
                    <div className="highlight-label">成功团圆</div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div className="highlight-stat">
                  <SoundOutlined className="highlight-icon" />
                  <div>
                    <div className="highlight-value">1,256</div>
                    <div className="highlight-label">方言录音</div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div className="highlight-stat">
                  <FileTextOutlined className="highlight-icon" />
                  <div>
                    <div className="highlight-value">3,892</div>
                    <div className="highlight-label">完成任务</div>
                  </div>
                </div>
              </Col>
            </Row>
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">最近案件动态</h3>
            <Table 
              columns={columns} 
              dataSource={recentCases} 
              rowKey="id" 
              pagination={false}
              className="big-screen-table"
            />
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">地区案件热力图</h3>
            <div className="heatmap-placeholder">
              <div className="heatmap-item">
                <span className="region">北京市</span>
                <div className="heatmap-bar">
                  <div className="heatmap-fill" style={{ width: '85%', background: '#ff4d4f' }}></div>
                </div>
                <span className="count">128</span>
              </div>
              <div className="heatmap-item">
                <span className="region">上海市</span>
                <div className="heatmap-bar">
                  <div className="heatmap-fill" style={{ width: '72%', background: '#ff7a45' }}></div>
                </div>
                <span className="count">96</span>
              </div>
              <div className="heatmap-item">
                <span className="region">广东省</span>
                <div className="heatmap-bar">
                  <div className="heatmap-fill" style={{ width: '65%', background: '#ffa940' }}></div>
                </div>
                <span className="count">84</span>
              </div>
              <div className="heatmap-item">
                <span className="region">浙江省</span>
                <div className="heatmap-bar">
                  <div className="heatmap-fill" style={{ width: '45%', background: '#ffc53d' }}></div>
                </div>
                <span className="count">56</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column */}
        <div className="big-screen-column">
          <div className="big-screen-card">
            <h3 className="card-title">任务统计</h3>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <div className="stat-item">
                  <div className="stat-value text-blue">156</div>
                  <div className="stat-label">进行中</div>
                </div>
              </Col>
              <Col span={12}>
                <div className="stat-item">
                  <div className="stat-value text-green">3,892</div>
                  <div className="stat-label">已完成</div>
                </div>
              </Col>
              <Col span={12}>
                <div className="stat-item">
                  <div className="stat-value text-orange">23</div>
                  <div className="stat-label">待分配</div>
                </div>
              </Col>
              <Col span={12}>
                <div className="stat-item">
                  <div className="stat-value text-red">8</div>
                  <div className="stat-label">已逾期</div>
                </div>
              </Col>
            </Row>
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">方言录音分布</h3>
            <div className="dialect-list">
              <div className="dialect-item">
                <span className="dialect-name">四川话</span>
                <span className="dialect-count">286</span>
              </div>
              <div className="dialect-item">
                <span className="dialect-name">广东话</span>
                <span className="dialect-count">234</span>
              </div>
              <div className="dialect-item">
                <span className="dialect-name">湖南话</span>
                <span className="dialect-count">198</span>
              </div>
              <div className="dialect-item">
                <span className="dialect-name">河南话</span>
                <span className="dialect-count">167</span>
              </div>
              <div className="dialect-item">
                <span className="dialect-name">山东话</span>
                <span className="dialect-count">156</span>
              </div>
            </div>
          </div>

          <div className="big-screen-card">
            <h3 className="card-title">优秀志愿者</h3>
            <div className="volunteer-rank">
              <div className="rank-item rank-1">
                <span className="rank-num">1</span>
                <span className="rank-name">李志愿者</span>
                <span className="rank-score">156分</span>
              </div>
              <div className="rank-item rank-2">
                <span className="rank-num">2</span>
                <span className="rank-name">王志愿者</span>
                <span className="rank-score">142分</span>
              </div>
              <div className="rank-item rank-3">
                <span className="rank-num">3</span>
                <span className="rank-name">张志愿者</span>
                <span className="rank-score">128分</span>
              </div>
              <div className="rank-item">
                <span className="rank-num">4</span>
                <span className="rank-name">刘志愿者</span>
                <span className="rank-score">115分</span>
              </div>
              <div className="rank-item">
                <span className="rank-num">5</span>
                <span className="rank-name">陈志愿者</span>
                <span className="rank-score">98分</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default BigScreen
