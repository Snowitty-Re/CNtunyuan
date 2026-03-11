import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Spin, Space, Timeline, Tag, Modal, Input, message, Row, Col, Image } from 'antd';
import { ArrowLeftOutlined, EditOutlined, CheckCircleOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import UrgencyTag from '@/components/UrgencyTag';
import ConfirmButton from '@/components/ConfirmButton';
import { getMissingPerson, getMissingPersonTracks, markFound, markReunited, deleteMissingPerson } from '@/api/missingPerson';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { GENDER_MAP } from '@/utils/constants';
import type { MissingPersonResponse, MissingPersonTrackResponse } from '@/types';

export default function MissingPersonDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [person, setPerson] = useState<MissingPersonResponse | null>(null);
  const [tracks, setTracks] = useState<MissingPersonTrackResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [foundModal, setFoundModal] = useState(false);
  const [foundLocation, setFoundLocation] = useState('');
  const [foundNote, setFoundNote] = useState('');

  useEffect(() => {
    if (!id) return;
    Promise.all([getMissingPerson(id), getMissingPersonTracks(id)])
      .then(([p, t]) => {
        setPerson(p);
        setTracks(t || []);
      })
      .finally(() => setLoading(false));
  }, [id]);

  const handleMarkFound = async () => {
    if (!id) return;
    try {
      await markFound(id, { location: foundLocation, note: foundNote });
      message.success('已标记为找到');
      setFoundModal(false);
      const p = await getMissingPerson(id);
      setPerson(p);
    } catch { /* interceptor */ }
  };

  const handleMarkReunited = async () => {
    if (!id) return;
    try {
      await markReunited(id);
      message.success('已标记为团圆');
      const p = await getMissingPerson(id);
      setPerson(p);
    } catch { /* interceptor */ }
  };

  const handleDelete = async () => {
    if (!id) return;
    try {
      await deleteMissingPerson(id);
      message.success('删除成功');
      navigate('/missing-persons');
    } catch { /* interceptor */ }
  };

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!person) return <div>案件不存在</div>;

  return (
    <PageContainer
      title={`案件详情 - ${person.name}`}
      breadcrumb={[{ title: '首页' }, { title: '走失人员' }, { title: '案件详情' }]}
      extra={
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>
          <Button icon={<EditOutlined />} onClick={() => navigate(`/missing-persons/${id}/edit`)}>编辑</Button>
          {isManager && person.status === 'searching' && (
            <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => setFoundModal(true)}>标记找到</Button>
          )}
          {isManager && person.status === 'found' && (
            <Button type="primary" style={{ background: '#52c41a' }} onClick={handleMarkReunited}>标记团圆</Button>
          )}
          {isManager && (
            <ConfirmButton danger title="确定删除此案件?" onConfirm={handleDelete}>删除</ConfirmButton>
          )}
        </Space>
      }
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="基本信息" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="案件编号">{person.case_no}</Descriptions.Item>
              <Descriptions.Item label="姓名">{person.name}</Descriptions.Item>
              <Descriptions.Item label="性别">{GENDER_MAP[person.gender] || person.gender}</Descriptions.Item>
              <Descriptions.Item label="年龄">{person.age || '-'}</Descriptions.Item>
              <Descriptions.Item label="身高">{person.height ? `${person.height}cm` : '-'}</Descriptions.Item>
              <Descriptions.Item label="体重">{person.weight ? `${person.weight}kg` : '-'}</Descriptions.Item>
              <Descriptions.Item label="状态"><StatusTag status={person.status} /></Descriptions.Item>
              <Descriptions.Item label="紧急程度"><UrgencyTag level={person.urgency} /></Descriptions.Item>
              <Descriptions.Item label="走失时间" span={2}>{formatDateTime(person.missing_time)}</Descriptions.Item>
              <Descriptions.Item label="走失地点" span={2}>
                {[person.province, person.city, person.district, person.address].filter(Boolean).join(' ')}
              </Descriptions.Item>
              <Descriptions.Item label="衣着" span={2}>{person.clothes || '-'}</Descriptions.Item>
              <Descriptions.Item label="特征" span={2}>{person.features || '-'}</Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>{person.description || '-'}</Descriptions.Item>
            </Descriptions>
          </Card>

          <Card title="联系信息" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="联系人">{person.contact_name}</Descriptions.Item>
              <Descriptions.Item label="联系电话">{person.contact_phone}</Descriptions.Item>
              <Descriptions.Item label="关系">{person.contact_rel || '-'}</Descriptions.Item>
              <Descriptions.Item label="备用联系">{person.alt_contact || '-'}</Descriptions.Item>
            </Descriptions>
          </Card>

          {person.found_time && (
            <Card title="找到信息" style={{ marginBottom: 16 }}>
              <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
                <Descriptions.Item label="找到时间">{formatDateTime(person.found_time)}</Descriptions.Item>
                <Descriptions.Item label="找到地点">{person.found_location || '-'}</Descriptions.Item>
                <Descriptions.Item label="备注" span={2}>{person.found_note || '-'}</Descriptions.Item>
              </Descriptions>
            </Card>
          )}
        </Col>
        <Col xs={24} lg={8}>
          {person.photo_url && (
            <Card title="照片" style={{ marginBottom: 16 }}>
              <Image src={person.photo_url} style={{ width: '100%' }} />
            </Card>
          )}
          <Card title="轨迹记录">
            {tracks.length > 0 ? (
              <Timeline
                items={tracks.map((t) => ({
                  color: t.is_key_point ? 'red' : 'blue',
                  children: (
                    <div>
                      <div><strong>{t.location || t.address}</strong></div>
                      <div>{t.description}</div>
                      <div style={{ color: '#999', fontSize: 12 }}>{formatDateTime(t.time)}</div>
                      {t.reporter && <div style={{ color: '#999', fontSize: 12 }}>上报人: {t.reporter.nickname}</div>}
                    </div>
                  ),
                }))}
              />
            ) : (
              <div style={{ textAlign: 'center', color: '#999', padding: 20 }}>暂无轨迹记录</div>
            )}
          </Card>
        </Col>
      </Row>

      <Modal title="标记找到" open={foundModal} onOk={handleMarkFound} onCancel={() => setFoundModal(false)}>
        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8 }}>找到地点</div>
          <Input value={foundLocation} onChange={(e) => setFoundLocation(e.target.value)} placeholder="输入找到的地点" />
        </div>
        <div>
          <div style={{ marginBottom: 8 }}>备注</div>
          <Input.TextArea value={foundNote} onChange={(e) => setFoundNote(e.target.value)} rows={3} placeholder="备注信息" />
        </div>
      </Modal>
    </PageContainer>
  );
}
