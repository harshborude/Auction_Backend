import client from './client'
import { User } from '@/types'

export function registerUser(data: { username: string; email: string; password: string }) {
  return client.post('/users/register', data)
}

export function loginUser(data: { email: string; password: string }) {
  return client.post<{ access_token: string; refresh_token: string }>('/users/login', data)
}

export function logoutUser() {
  return client.post('/users/logout')
}

export function refreshToken(refreshToken: string) {
  return client.post<{ access_token: string }>('/users/refresh', { refresh_token: refreshToken })
}

export function getCurrentUser() {
  return client.get<User>('/users/me')
}

export function changePassword(data: { old_password: string; new_password: string }) {
  return client.patch('/users/change-password', data)
}
