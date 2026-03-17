import axios from "axios";

const instance = axios.create({
    baseURL: "http://localhost:8080", // your Go backend
});

// 🔐 Attach access token automatically
instance.interceptors.request.use((config) => {
    const token = localStorage.getItem("access_token");

    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
});

// 🔄 Handle token refresh automatically
instance.interceptors.response.use(
    (res) => res,
    async (err) => {

        const originalRequest = err.config;

        // If access token expired → try refresh
        if (err.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;

            try {
                const refreshToken = localStorage.getItem("refresh_token");

                const res = await axios.post(
                    "http://localhost:8080/users/refresh",
                    {},
                    {
                        headers: {
                            Authorization: `Bearer ${refreshToken}`
                        }
                    }
                );

                localStorage.setItem("access_token", res.data.access_token);

                // retry original request
                originalRequest.headers.Authorization = `Bearer ${res.data.access_token}`;

                return instance(originalRequest);

            } catch (refreshErr) {
                console.log("Refresh failed");

                localStorage.removeItem("access_token");
                localStorage.removeItem("refresh_token");

                window.location.href = "/login";
            }
        }

        return Promise.reject(err);
    }
);

export default instance;