import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Spin, Space, Tag, List, Input, message, Row, Col } from 'antd';
import { ArrowLeftOutlined, EditOutlined, HeartOutlined, HeartFilled, StarOutlined, StarFilled } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import AudioPlayer from '@/components/AudioPlayer';
import ConfirmButton from '@/components/ConfirmButton';
import { getDialect, getDialectComments, addComment, likeDialect, unlikeDialect, featureDialect, unfeatureDialect, updateDialectStatus, deleteDialect, recordPlay } from '@/api/dialect';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime, formatFileSize, formatDuration } from '@/utils/format';
import type { DialectResponse, DialectCommentResponse } from '@/types';

export default function DialectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isManager, isAdmin } = usePermission();
  const [dialect, setDialect] = useState<DialectResponse | null>(null);
  const [comments, setComments] = useState<DialectCommentResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [liked, setLiked] = useState(false);
  const [commentText, setCommentText] = useState('');

  const reload = async () => {
    if (!id) return;
    const [d, c] = await Promise.all([getDialect(id), getDialectComments(id)]);
    setDialect(d);
    setComments(c?.list || []);
  };

  useEffect(() => {
    reload().finally(() => setLoading(false));
    if (id) recordPlay(id).catch(() => {});
  }, [id]);

  const handleLike = async () => {
    if (!id) return;
    if (liked) {
      await unlikeDialect(id);
      setLiked(false);
    } else {
      await likeDialect(id);
      setLiked(true);
    }
    reload();
  };

  const handleComment = async () => {
    if (!id || !commentText.trim()) return;
    await addComment(id, { content: commentText });
    setCommentText('');
    message.success('评论成功');
    reload();
  };

  const handleFeature = async () => {
    if (!id || !dialect) return;
    if (dialect.is_featured) {
      await unfeatureDialect(id);
    } else {
      await featureDialect(id);
    }
    reload();
  };

  const handleApprove = async () => {
    if (!id) return;
    await updateDialectStatus(id, 'active');
    message.success('已通过');
    reload();
  };

  const handleReject = async () => {
    if (!id) return;
    await updateDialectStatus(id, 'rejected');
    message.success('已拒绝');
    reload();
  };

  const handleDelete = async () => {
    if (!id) return;
    await deleteDialect(id);
    message.success('删除成功');
    navigate('/dialects');
  };

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!dialect) return <div>方言不存在</div>;

  return (
    <PageContainer
      title={`方言详情 - ${dialect.title}`}
      breadcrumb={[{ title: '首页' }, { title: '方言库' }, { title: '详情' }]}
      extra={
        <Space wrap>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>
          <Button icon={<EditOutlined />} onClick={() => navigate(`/dialects/${id}/edit`)}>编辑</Button>
          {isAdmin && (
            <Button onClick={handleFeature} icon={dialect.is_featured ? <StarFilled /> : <StarOutlined />}>
              {dialect.is_featured ? '取消精选' : '设为精选'}
            </Button>
          )}
          {isManager && dialect.status === 'pending' && (
            <>
              <Button type="primary" onClick={handleApprove}>通过</Button>
              <Button danger onClick={handleReject}>拒绝</Button>
            </>
          )}
          {isManager && <ConfirmButton danger title="确定删除?" onConfirm={handleDelete}>删除</ConfirmButton>}
        </Space>
      }
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="音频播放" style={{ marginBottom: 16 }}>
            <AudioPlayer src={dialect.audio_url} duration={dialect.duration} />
          </Card>
          <Card title="基本信息" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="标题">{dialect.title}</Descriptions.Item>
              <Descriptions.Item label="地区">{dialect.region}</Descriptions.Item>
              <Descriptions.Item label="方言类型">{dialect.dialect_type || '-'}</Descriptions.Item>
              <Descriptions.Item label="状态"><StatusTag status={dialect.status} /></Descriptions.Item>
              <Descriptions.Item label="时长">{formatDuration(dialect.duration)}</Descriptions.Item>
              <Descriptions.Item label="文件大小">{formatFileSize(dialect.file_size)}</Descriptions.Item>
              <Descriptions.Item label="播放次数">{dialect.play_count}</Descriptions.Item>
              <Descriptions.Item label="点赞数">{dialect.like_count}</Descriptions.Item>
              <Descriptions.Item label="精选">{dialect.is_featured ? <Tag color="gold">精选</Tag> : '否'}</Descriptions.Item>
              <Descriptions.Item label="上传者">{dialect.uploader?.nickname || '-'}</Descriptions.Item>
              <Descriptions.Item label="标签">{dialect.tags || '-'}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(dialect.created_at)}</Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>{dialect.description || '-'}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card
            title={`评论 (${dialect.comment_count})`}
            extra={
              <Button type="text" onClick={handleLike} icon={liked ? <HeartFilled style={{ color: '#ff4d4f' }} /> : <HeartOutlined />}>
                {dialect.like_count}
              </Button>
            }
          >
            <div style={{ marginBottom: 16 }}>
              <Input.TextArea
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
                rows={2}
                placeholder="写下你的评论..."
              />
              <Button type="primary" size="small" style={{ marginTop: 8 }} onClick={handleComment} disabled={!commentText.trim()}>
                发表评论
              </Button>
            </div>
            <List
              dataSource={comments}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={item.user?.nickname || '匿名'}
                    description={
                      <div>
                        <div>{item.content}</div>
                        <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                          {formatDateTime(item.created_at)}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
              locale={{ emptyText: '暂无评论' }}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
}
