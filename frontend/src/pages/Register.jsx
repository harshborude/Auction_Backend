import { useState } from "react"
import { registerUser } from "../api/auth"
import { useNavigate } from "react-router-dom"
import styles from "./Register.module.css"

function Register() {

    const navigate = useNavigate()

    const [username, setUsername] = useState("")
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")

    const handleSubmit = async (e) => {
        e.preventDefault()

        if (!username || !email || !password) {
            alert("Please fill all fields")
            return
        }

        try {

            await registerUser({
                username,
                email,
                password
            })

            alert("Registration successful!")

            navigate("/login")

        } catch (err) {
            alert("Registration failed")
        }
    }

    return (

        <div className={styles.container}>

            <h2>Register</h2>

            <form onSubmit={handleSubmit} className={styles.form}>

                <input
                    type="text"
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />

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

                <button type="submit">Register</button>

            </form>

        </div>

    )
}

export default Register