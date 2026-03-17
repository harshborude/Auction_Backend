import { useState, useContext } from "react"
import { AuthContext } from "../context/AuthContext"
import styles from "./Login.module.css"

function Login() {

    const { login } = useContext(AuthContext)

    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")

    const handleSubmit = async (e) => {
        e.preventDefault()

        if (!email || !password) {
            alert("Please fill all fields")
            return
        }

        try {
            await login({ email, password })
        } catch (err) {
            alert("Invalid credentials")
        }
    }

    return (

        <div className={styles.container}>

            <h2>Login</h2>

            <form onSubmit={handleSubmit} className={styles.form}>

                <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />

                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />

                <button type="submit">Login</button>

            </form>

        </div>

    )
}

export default Login