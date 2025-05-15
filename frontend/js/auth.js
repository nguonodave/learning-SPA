export function setupAuthUI(setHeadingText) {
    const loginForm = document.getElementById("login-form")
    const registerForm = document.getElementById("register-form")
    const showLoginBtn = document.getElementById("show-login")
    const showRegisterBtn = document.getElementById("show-register")
    const logoutBtn = document.getElementById("logout-btn")

    showLoginBtn.addEventListener("click", function () {
        loginForm.style.display = "block"
        registerForm.style.display = "none"
        setHeadingText("Login")
    })

    showRegisterBtn.addEventListener("click", function () {
        registerForm.style.display = "block"
        loginForm.style.display = "none"
        setHeadingText("Register")
    })

    loginForm.addEventListener("submit", async function (e) {
        e.preventDefault()
        const username = document.getElementById("login-username").value
        const password = document.getElementById("login-password").value

        const res = await fetch("/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        })

        if (res.ok) {
            localStorage.setItem("username", username)
            setHeadingText(`Welcome, ${username}`)
            loginForm.style.display = "none"
            document.getElementById("auth-toggle").style.display = "none"
            logoutBtn.style.display = "inline"
        } else {
            alert("Login failed")
        }
    })

    registerForm.addEventListener("submit", async function (e) {
        e.preventDefault()
        const username = document.getElementById("register-username").value
        const password = document.getElementById("register-password").value

        const res = await fetch("/api/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        })

        if (res.ok) {
            alert("Registered successfully. You can now log in.")
            showLoginBtn.click()
        } else {
            alert("Registration failed: Username may already be taken.")
        }
    })

    logoutBtn.addEventListener("click", function () {
        localStorage.removeItem("username")
        setHeadingText("PostApp - Logged out view")
        logoutBtn.style.display = "none"
        document.getElementById("auth-toggle").style.display = "block"
    })

    // On load, check session
    const user = localStorage.getItem("username")
    if (user) {
        setHeadingText(`Welcome, ${user}`)
        loginForm.style.display = "none"
        registerForm.style.display = "none"
        document.getElementById("auth-toggle").style.display = "none"
        logoutBtn.style.display = "inline"
    } else {
        setHeadingText("PostApp - Logged out view")
    }
}
