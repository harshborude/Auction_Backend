import api from "./api"

export const getUsers = () =>
    api.get("/admin/users")

export const assignCredits = (userId, amount) =>
    api.patch(`/admin/users/${userId}/credits`, { amount })

export const createAuction = (data) =>
    api.post("/admin/auctions", data)

export const endAuction = (id) =>
    api.post(`/admin/auctions/${id}/end`)

export const cancelAuction = (id) =>
    api.post(`/admin/auctions/${id}/cancel`)