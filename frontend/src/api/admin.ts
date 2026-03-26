import client from './client'
import { User, Auction } from '@/types'

export function fetchUsers() {
  return client.get<User[]>('/admin/users')
}

export function assignCredits(userId: number, amount: number) {
  return client.patch(`/admin/users/${userId}/credits`, { amount })
}

export function promoteUser(userId: number) {
  return client.patch(`/admin/promote/${userId}`)
}

export function banUser(userId: number) {
  return client.patch(`/admin/users/${userId}/ban`)
}

export function getAdminAuctions() {
  return client.get<{ auctions: Auction[] }>('/admin/auctions')
}

export function createAuction(data: {
  title: string
  description: string
  image_url: string
  starting_price: number
  bid_increment: number
  start_time: string   // always required by backend — send "now" if not specified
  end_time: string
}) {
  return client.post<Auction>('/admin/auctions', data)
}

export function endAuction(id: number) {
  return client.post(`/admin/auctions/${id}/end`)
}

export function cancelAuction(id: number) {
  return client.post(`/admin/auctions/${id}/cancel`)
}
