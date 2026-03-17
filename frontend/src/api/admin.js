import axios from "./axiosInstance";

export const fetchUsers = async () => {
    const res = await axios.get("/admin/users");
    return res.data;
};

export const assignCredits = async (userId, amount) => {
    return axios.patch(`/admin/users/${userId}/credits`, {
        amount
    });
};

export const createAuction = async (data) => {
    return axios.post("/admin/auctions", data);
};

export const endAuction = async (id) => {
    return axios.post(`/admin/auctions/${id}/end`);
};

export const cancelAuction = async (id) => {
    return axios.post(`/admin/auctions/${id}/cancel`);
};