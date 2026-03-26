// All field names match Go struct field names (PascalCase, no json tags on models)
// Exception: Bid.user has json:"user,omitempty" so it stays lowercase

export interface User {
  ID: number
  Username: string
  Email: string
  Role: 'USER' | 'ADMIN'
  IsActive: boolean
  Rating: number
  TotalAuctionsWon: number
  TotalAuctionsCreated: number
  Wallet?: Wallet
  CreatedAt: string
  UpdatedAt: string
}

export interface Wallet {
  ID: number
  UserID: number
  Balance: number
  ReservedBalance: number
  CreatedAt: string
  UpdatedAt: string
}

export interface Auction {
  ID: number
  Title: string
  Description: string
  ImageURL: string
  StartingPrice: number
  BidIncrement: number
  CurrentHighestBid: number
  CurrentHighestBidderID: number | null
  BidCount: number
  ExtensionCount: number
  Status: 'ACTIVE' | 'SCHEDULED' | 'ENDED' | 'CANCELLED'
  StartTime: string
  EndTime: string
  CreatedBy: number
  CreatedAt: string
  UpdatedAt: string
}

export interface Bid {
  ID: number
  UserID: number
  Amount: number
  AuctionID: number
  user?: { ID: number; Username: string }  // json:"user,omitempty" — lowercase key, PascalCase fields inside
  CreatedAt: string
}

export interface Transaction {
  ID: number
  UserID: number
  Amount: number
  Type: 'BID_RESERVE' | 'BID_RELEASE' | 'AUCTION_WIN' | 'ADMIN_ASSIGN'
  Reference: string
  CreatedAt: string
}

// WebSocket Message struct has explicit json tags — all snake_case
export interface WsMessage {
  type: 'BID_UPDATE' | 'OUTBID' | 'AUCTION_EXTENDED' | 'AUCTION_STARTED' | 'AUCTION_END'
  auction_id: number
  amount: number
  bidder_id: number
  message?: string
  end_time?: string
}

// These come from gin.H{} which uses explicit string keys
export interface PaginatedAuctions {
  page: number
  limit: number
  auctions: Auction[]
}

export interface PaginatedTransactions {
  page: number
  limit: number
  transactions: Transaction[]
}
