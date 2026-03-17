import api from "./api"

export const fetchAuctions = async () => {
    const res = await api.get("/auctions?page=1&limit=20")
    return res.data
}

export const fetchAuction = (id) =>
    api.get(`/auctions/${id}`)

export const fetchBids = (id) =>
    api.get(`/auctions/${id}/bids`)

export const placeBid = (id, amount) =>
    api.post(`/auctions/${id}/bid`, { amount })