import axios from "axios"

const api = axios.create({
    baseURL: "http://localhost:8080",
})

api.interceptors.request.use((config) => {

    const token = localStorage.getItem("access_token")

    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }

    return config
})

api.interceptors.response.use(

    (response) => response,

    async (error) => {

        const originalRequest = error.config

        if (error.response?.status === 401 && !originalRequest._retry) {

            originalRequest._retry = true

            try {

                const refreshToken = localStorage.getItem("refresh_token")

                const res = await axios.post(
                    "http://localhost:8080/users/refresh",
                    { refresh_token: refreshToken }
                )

                const newToken = res.data.access_token

                localStorage.setItem("access_token", newToken)

                originalRequest.headers.Authorization = `Bearer ${newToken}`

                return api(originalRequest)

            } catch (err) {

                localStorage.clear()
                window.location.href = "/login"

            }

        }

        return Promise.reject(error)

    }

)

export default api