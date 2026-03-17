import api from "./api";

export const registerUser = (data) => api.post("/users/register", data);
export const loginUser = (data) => api.post("/users/login", data);
export const logoutUser = (refreshToken) => api.post("/users/logout", { refresh_token: refreshToken });
export const getCurrentUser = () => api.get("/users/me");