import api from "./api"

// ✅ Get all auctions
export const fetchAuctions = async (page = 1, limit = 20) => {
    const res = await api.get(`/auctions?page=${page}&limit=${limit}`);
    return res.data?.auctions || res.data || [];
};

// ✅ Get single auction
export const fetchAuction = async (id) => {
    const res = await api.get(`/auctions/${id}`)
    return res.data
}

// ✅ Get bids
export const fetchBids = async (id) => {
    const res = await api.get(`/auctions/${id}/bids`)
    return res.data || []
}

// ✅ Place bid
export const placeBid = async (id, amount) => {
    const res = await api.post(`/auctions/${id}/bid`, { amount })
    return res.data
}