import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Spin, Button, Descriptions } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatCard from '@/components/StatCard';
import { getFileStats } from '@/api/upload';
import { formatFileSize } from '@/utils/format';
import type { FileStatsResponse } from '@/types';

export default function FileStatsPage() {
  const navigate = useNavigate();
  const [stats, setStats] = useState<FileStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getFileStats().then(setStats).finally(() => setLoading(false));
  }, []);

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!stats) return null;

  return (
    <PageContainer
      title="存储统计"
      breadcrumb={[{ title: '首页' }, { title: '文件管理' }, { title: '存储统计' }]}
      extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>}
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}><StatCard title="文件总数" value={stats.total_count} /></Col>
        <Col xs={24} sm={12} lg={6}><StatCard title="总存储" value={0} suffix={formatFileSize(stats.total_size)} /></Col>
        <Col xs={24} sm={12} lg={6}><StatCard title="图片数" value={stats.image_count} color="#fa541c" /></Col>
        <Col xs={24} sm={12} lg={6}><StatCard title="音频数" value={stats.audio_count} color="#13c2c2" /></Col>
      </Row>
      <Card title="分类详情" style={{ marginTop: 16 }}>
        <Descriptions bordered column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="图片">{stats.image_count} 个 / {formatFileSize(stats.image_size)}</Descriptions.Item>
          <Descriptions.Item label="音频">{stats.audio_count} 个 / {formatFileSize(stats.audio_size)}</Descriptions.Item>
          <Descriptions.Item label="视频">{stats.video_count} 个 / {formatFileSize(stats.video_size)}</Descriptions.Item>
          <Descriptions.Item label="文档">{stats.doc_count} 个 / {formatFileSize(stats.doc_size)}</Descriptions.Item>
        </Descriptions>
      </Card>
    </PageContainer>
  );
}
