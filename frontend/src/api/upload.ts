import client from './client'

/**
 * Upload an image. In production (Cloudinary env vars set), uploads directly
 * to Cloudinary CDN. In development, falls back to the local backend endpoint.
 */
export async function uploadImage(file: File): Promise<string> {
  const cloudName = import.meta.env.VITE_CLOUDINARY_CLOUD_NAME
  const uploadPreset = import.meta.env.VITE_CLOUDINARY_UPLOAD_PRESET

  if (cloudName && uploadPreset) {
    // Direct Cloudinary upload — no backend needed
    const formData = new FormData()
    formData.append('file', file)
    formData.append('upload_preset', uploadPreset)

    const res = await fetch(
      `https://api.cloudinary.com/v1_1/${cloudName}/image/upload`,
      { method: 'POST', body: formData }
    )
    if (!res.ok) throw new Error('Cloudinary upload failed')
    const data = await res.json()
    return data.secure_url as string
  }

  // Local dev fallback — backend saves to ./uploads/
  const formData = new FormData()
  formData.append('image', file)
  const { data } = await client.post<{ url: string }>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data.url
}
