import api from "./api"

export const getWallet = () => api.get("/users/wallet")

export const getTransactions = () =>
    api.get("/users/wallet/transactions")