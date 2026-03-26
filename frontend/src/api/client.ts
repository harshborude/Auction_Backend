import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'

const BASE_URL = 'http://localhost:8080'

const client = axios.create({ baseURL: BASE_URL })

// Attach access token to every request
client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Refresh token on 401
let refreshing = false
let queue: Array<(token: string) => void> = []

client.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status !== 401 || original._retry) {
      return Promise.reject(error)
    }

    original._retry = true

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      clearSession()
      return Promise.reject(error)
    }

    if (refreshing) {
      // Queue requests while refresh is in flight
      return new Promise((resolve) => {
        queue.push((token: string) => {
          original.headers.Authorization = `Bearer ${token}`
          resolve(client(original))
        })
      })
    }

    refreshing = true
    try {
      const { data } = await axios.post(`${BASE_URL}/users/refresh`, {
        refresh_token: refreshToken,
      })
      const newToken: string = data.access_token
      localStorage.setItem('access_token', newToken)
      original.headers.Authorization = `Bearer ${newToken}`
      queue.forEach((cb) => cb(newToken))
      queue = []
      return client(original)
    } catch {
      clearSession()
      return Promise.reject(error)
    } finally {
      refreshing = false
    }
  }
)

function clearSession() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  window.location.href = '/login'
}

export default client
