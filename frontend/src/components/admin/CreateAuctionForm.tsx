import { FormEvent, useRef, useState } from 'react'
import { createAuction } from '@/api/admin'
import { uploadImage } from '@/api/upload'
import styles from './CreateAuctionForm.module.css'

const EMPTY = {
  title: '',
  description: '',
  image_url: '',
  starting_price: '',
  bid_increment: '',
  start_time: '',
  end_time: '',
}

export default function CreateAuctionForm({ onCreated }: { onCreated: () => void }) {
  const [form, setForm] = useState(EMPTY)
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [preview, setPreview] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  function update(field: keyof typeof form) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setForm((prev) => ({ ...prev, [field]: e.target.value }))
  }

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    setError('')
    try {
      const url = await uploadImage(file)
      setForm((prev) => ({ ...prev, image_url: url }))
      setPreview(url)
    } catch {
      setError('Image upload failed')
    } finally {
      setUploading(false)
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSuccess(false)
    if (!form.title || !form.starting_price || !form.bid_increment || !form.end_time) {
      setError('Title, starting price, bid increment, and end time are required')
      return
    }
    setLoading(true)
    try {
      const startTime = form.start_time
        ? new Date(form.start_time).toISOString()
        : new Date().toISOString()

      await createAuction({
        title: form.title,
        description: form.description,
        image_url: form.image_url,
        starting_price: Number(form.starting_price),
        bid_increment: Number(form.bid_increment),
        start_time: startTime,
        end_time: new Date(form.end_time).toISOString(),
      })
      setForm(EMPTY)
      setPreview(null)
      if (fileRef.current) fileRef.current.value = ''
      setSuccess(true)
      onCreated()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(msg || 'Failed to create auction')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={`form-group ${styles.full}`}>
        <label className="form-label">Title *</label>
        <input className="form-input" value={form.title} onChange={update('title')} placeholder="Item name" />
      </div>
      <div className={`form-group ${styles.full}`}>
        <label className="form-label">Description</label>
        <input className="form-input" value={form.description} onChange={update('description')} placeholder="Optional description" />
      </div>
      <div className={`form-group ${styles.full}`}>
        <label className="form-label">Image</label>
        <div className={styles.uploadArea}>
          <input
            ref={fileRef}
            className="form-input"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            onChange={handleFileChange}
            disabled={uploading}
          />
          {uploading && <span className={styles.uploadStatus}>Uploading…</span>}
          {preview && !uploading && (
            <img src={preview} alt="Preview" className={styles.preview} />
          )}
        </div>
      </div>
      <div className="form-group">
        <label className="form-label">Starting price *</label>
        <input className="form-input" type="number" min={1} value={form.starting_price} onChange={update('starting_price')} placeholder="e.g. 100" />
      </div>
      <div className="form-group">
        <label className="form-label">Bid increment *</label>
        <input className="form-input" type="number" min={1} value={form.bid_increment} onChange={update('bid_increment')} placeholder="e.g. 10" />
      </div>
      <div className="form-group">
        <label className="form-label">Start time <span style={{color:'var(--text-dim)',fontWeight:400}}>(leave blank to start immediately)</span></label>
        <input className="form-input" type="datetime-local" value={form.start_time} onChange={update('start_time')} />
      </div>
      <div className="form-group">
        <label className="form-label">End time *</label>
        <input className="form-input" type="datetime-local" value={form.end_time} onChange={update('end_time')} />
      </div>
      {error && <p className={`error-text ${styles.error}`}>{error}</p>}
      {success && <p className={`success-text ${styles.success}`}>Auction created!</p>}
      <div className={styles.submit}>
        <button type="submit" className="btn btn-primary" disabled={loading || uploading}>
          {loading ? 'Creating…' : 'Create auction'}
        </button>
      </div>
    </form>
  )
}
