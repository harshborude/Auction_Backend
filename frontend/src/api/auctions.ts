import client from './client'
import { Auction, Bid, PaginatedAuctions } from '@/types'

export function fetchAuctions(page = 1, limit = 12) {
  return client.get<PaginatedAuctions>('/auctions', { params: { page, limit } })
}

export function fetchAuction(id: number) {
  return client.get<Auction>(`/auctions/${id}`)
}

export function fetchBids(id: number) {
  return client.get<Bid[]>(`/auctions/${id}/bids`)
}

export function placeBid(id: number, amount: number) {
  return client.post<{ message: string; bid: Bid }>(`/auctions/${id}/bid`, { amount })
}
