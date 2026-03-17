export const mapAuction = (a = {}) => ({
    id: a.ID ?? null,
    title: a.Title ?? "",
    description: a.Description ?? "",
    image_url: a.ImageURL ?? "",
    starting_price: a.StartingPrice ?? 0,
    bid_increment: a.BidIncrement ?? 1,
    current_highest_bid: a.CurrentHighestBid ?? 0,
    current_highest_bidder_id: a.CurrentHighestBidderID ?? null,
    bid_count: a.BidCount ?? 0,
    status: a.Status ?? "UNKNOWN",
    start_time: a.StartTime ?? null,
    end_time: a.EndTime ?? null,
});

export const mapBid = (b) => ({
    id: b.ID,
    user_id: b.UserID,
    amount: b.Amount,
    created_at: b.CreatedAt,
})

export const mapWallet = (w) => ({
    available_credits: (w.Balance || 0) - (w.ReservedBalance || 0),
    reserved_credits: w.ReservedBalance || 0,
    total_credits: w.Balance || 0,
})
export const mapTransaction = (t) => ({
    id: t.ID,
    type: t.Type,
    amount: t.Amount,
    description: t.Reference || "",
})