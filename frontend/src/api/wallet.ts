import client from './client'
import { Wallet, PaginatedTransactions } from '@/types'

export function getWallet() {
  return client.get<Wallet>('/users/wallet')
}

export function getTransactions(page = 1, limit = 20) {
  return client.get<PaginatedTransactions>('/users/wallet/transactions', {
    params: { page, limit },
  })
}
