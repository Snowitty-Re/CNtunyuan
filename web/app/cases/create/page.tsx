'use client'

import { FormEvent, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { isMainlandPhone } from '@/lib/validators'
import { missingPersonService } from '@/services/missingPersons'
import { uploadService } from '@/services/upload'

export default function CaseCreatePage() {
  const { ready } = useAuthGuard()
  const router = useRouter()

  const [submitting, setSubmitting] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)

  const [name, setName] = useState('')
  const [gender, setGender] = useState('male')
  const [age, setAge] = useState('')
  const [height, setHeight] = useState('')
  const [weight, setWeight] = useState('')
  const [caseType, setCaseType] = useState('other')
  const [urgency, setUrgency] = useState('medium')

  const [missingTime, setMissingTime] = useState('')
  const [province, setProvince] = useState('')
  const [city, setCity] = useState('')
  const [district, setDistrict] = useState('')
  const [address, setAddress] = useState('')
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')

  const [description, setDescription] = useState('')
  const [clothes, setClothes] = useState('')
  const [features, setFeatures] = useState('')

  const [contactName, setContactName] = useState('')
  const [contactPhone, setContactPhone] = useState('')
  const [contactRel, setContactRel] = useState('')
  const [altContact, setAltContact] = useState('')

  const [photoUrl, setPhotoUrl] = useState('')

  async function uploadPhoto(files: FileList | null) {
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      const uploaded = await uploadService.uploadSingle(files[0], { entity_type: 'missing_person' })
      const url = uploaded.url || uploaded.path || ''
      if (!url) throw new Error('上传成功但未返回可用地址')
      setPhotoUrl(url)
      setNotice({ type: 'success', text: '照片上传成功，可继续提交案件' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '照片上传失败' })
    } finally {
      setUploading(false)
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()

    if (!name.trim()) {
      setNotice({ type: 'error', text: '请填写姓名' })
      return
    }
    if (!missingTime) {
      setNotice({ type: 'error', text: '请填写走失时间' })
      return
    }
    if (!contactName.trim()) {
      setNotice({ type: 'error', text: '请填写联系人' })
      return
    }
    if (!isMainlandPhone(contactPhone.trim())) {
      setNotice({ type: 'error', text: '联系人手机号格式不正确' })
      return
    }

    setSubmitting(true)
    setNotice(null)

    try {
      const created = await missingPersonService.create({
        name: name.trim(),
        gender,
        age: Number(age) || 0,
        height: Number(height) || 0,
        weight: Number(weight) || 0,
        case_type: caseType,
        urgency_level: urgency,
        missing_time: new Date(missingTime).toISOString(),
        province: province.trim(),
        city: city.trim(),
        district: district.trim(),
        address: address.trim(),
        missing_latitude: Number(lat) || 0,
        missing_longitude: Number(lng) || 0,
        description: description.trim(),
        clothes: clothes.trim(),
        features: features.trim(),
        contact_name: contactName.trim(),
        contact_phone: contactPhone.trim(),
        contact_rel: contactRel.trim(),
        alt_contact: altContact.trim(),
        photo_url: photoUrl.trim(),
      })
      router.push(created?.id ? `/cases/${created.id}` : '/cases')
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建案件失败' })
    } finally {
      setSubmitting(false)
    }
  }

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="新建案件" desc="完整录入走失人员案件信息，便于后续协作与闭环" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />

      <form className="grid" onSubmit={submit}>
        <div className="form-section">
          <h3 className="form-section-title">基本信息</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">姓名<span className="required">*</span></span>
              <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="例如：王某" />
            </label>
            <label>
              <span className="field-label">性别</span>
              <select className="select" value={gender} onChange={(e) => setGender(e.target.value)}>
                <option value="male">male</option>
                <option value="female">female</option>
                <option value="other">other</option>
              </select>
            </label>
            <label>
              <span className="field-label">年龄</span>
              <input className="input" value={age} onChange={(e) => setAge(e.target.value.replace(/[^\d]/g, ''))} placeholder="例如：11" />
            </label>
            <label>
              <span className="field-label">身高(cm)</span>
              <input className="input" value={height} onChange={(e) => setHeight(e.target.value.replace(/[^\d]/g, ''))} placeholder="例如：120" />
            </label>
            <label>
              <span className="field-label">体重(kg)</span>
              <input className="input" value={weight} onChange={(e) => setWeight(e.target.value.replace(/[^\d]/g, ''))} placeholder="例如：35" />
            </label>
            <label>
              <span className="field-label">案件类型</span>
              <select className="select" value={caseType} onChange={(e) => setCaseType(e.target.value)}>
                <option value="other">other</option>
                <option value="child">child</option>
                <option value="adult">adult</option>
                <option value="elderly">elderly</option>
                <option value="disability">disability</option>
              </select>
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">走失信息</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">走失时间<span className="required">*</span></span>
              <input className="input" type="datetime-local" value={missingTime} onChange={(e) => setMissingTime(e.target.value)} />
            </label>
            <label>
              <span className="field-label">紧急程度</span>
              <select className="select" value={urgency} onChange={(e) => setUrgency(e.target.value)}>
                <option value="low">low</option>
                <option value="medium">medium</option>
                <option value="high">high</option>
                <option value="urgent">urgent</option>
              </select>
            </label>
            <label>
              <span className="field-label">省</span>
              <input className="input" value={province} onChange={(e) => setProvince(e.target.value)} placeholder="广东省" />
            </label>
            <label>
              <span className="field-label">市</span>
              <input className="input" value={city} onChange={(e) => setCity(e.target.value)} placeholder="广州市" />
            </label>
            <label>
              <span className="field-label">区/县</span>
              <input className="input" value={district} onChange={(e) => setDistrict(e.target.value)} placeholder="越秀区" />
            </label>
            <label>
              <span className="field-label">详细地址</span>
              <input className="input" value={address} onChange={(e) => setAddress(e.target.value)} placeholder="北京街道府前路1号" />
            </label>
            <label>
              <span className="field-label">纬度</span>
              <input className="input" value={lat} onChange={(e) => setLat(e.target.value.replace(/[^\d.-]/g, ''))} placeholder="23.129" />
            </label>
            <label>
              <span className="field-label">经度</span>
              <input className="input" value={lng} onChange={(e) => setLng(e.target.value.replace(/[^\d.-]/g, ''))} placeholder="113.264" />
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">体貌特征与说明</h3>
          <div className="grid cols-2">
            <label>
              <span className="field-label">衣着描述</span>
              <input className="input" value={clothes} onChange={(e) => setClothes(e.target.value)} placeholder="例如：蓝色外套、黑色运动鞋" />
            </label>
            <label>
              <span className="field-label">外貌特征</span>
              <input className="input" value={features} onChange={(e) => setFeatures(e.target.value)} placeholder="例如：右手有疤痕" />
            </label>
            <label style={{ gridColumn: '1 / -1' }}>
              <span className="field-label">案件描述</span>
              <textarea className="textarea" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="补充走失经过、已核验信息等" />
            </label>
            <label>
              <span className="field-label">上传照片</span>
              <input className="input" type="file" accept="image/*" onChange={(e) => uploadPhoto(e.target.files)} disabled={uploading} />
            </label>
            <label>
              <span className="field-label">照片 URL</span>
              <input className="input" value={photoUrl} onChange={(e) => setPhotoUrl(e.target.value)} placeholder="可手动粘贴 URL" />
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">联系人信息</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">联系人<span className="required">*</span></span>
              <input className="input" value={contactName} onChange={(e) => setContactName(e.target.value)} placeholder="例如：李某" />
            </label>
            <label>
              <span className="field-label">联系电话<span className="required">*</span></span>
              <input className="input" value={contactPhone} onChange={(e) => setContactPhone(e.target.value.replace(/[^\d]/g, ''))} placeholder="11位手机号" />
            </label>
            <label>
              <span className="field-label">与走失者关系</span>
              <input className="input" value={contactRel} onChange={(e) => setContactRel(e.target.value)} placeholder="家属/亲友" />
            </label>
            <label style={{ gridColumn: '1 / -1' }}>
              <span className="field-label">备用联系方式</span>
              <input className="input" value={altContact} onChange={(e) => setAltContact(e.target.value)} placeholder="可选" />
            </label>
          </div>
        </div>

        <div className="panel row wrap" style={{ marginTop: 0 }}>
          <button className="btn primary" type="submit" disabled={submitting || uploading}>
            {uploading ? '文件上传中...' : submitting ? '提交中...' : '提交案件'}
          </button>
          <button className="btn" type="button" onClick={() => router.push('/cases')}>
            返回案件列表
          </button>
          <span className="hint">带 <span className="required">*</span> 为必填</span>
        </div>
      </form>
    </AppShell>
  )
}
