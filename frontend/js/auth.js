export function setupAuthForms() {
    const registerForm = document.getElementById('register')
    const loginForm = document.getElementById('login')
    const logoutBtn = document.getElementById('logout-btn')
    
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault()
            const username = document.getElementById('reg-username').value
            const password = document.getElementById('reg-password').value
            
            try {
                const response = await fetch('/api/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ username, password })
                })
                
                if (!response.ok) {
                    const error = await response.json()
                    document.getElementById('register-error').textContent = error.message || 'Registration failed'
                    return
                }
                
                // On successful registration, show login form
                document.getElementById('register-error').textContent = ''
                document.getElementById('reg-username').value = ''
                document.getElementById('reg-password').value = ''
                alert('Registration successful! Please login.')
            } catch (err) {
                document.getElementById('register-error').textContent = 'Network error'
            }
        })
    }
    
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault()
            const username = document.getElementById('login-username').value
            const password = document.getElementById('login-password').value
            
            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ username, password })
                })
                
                if (!response.ok) {
                    const error = await response.json()
                    document.getElementById('login-error').textContent = error.message || 'Login failed'
                    return
                }
                
                // On successful login, show app content
                document.getElementById('login-error').textContent = ''
                document.getElementById('auth-forms').classList.add('hidden')
                document.getElementById('app-content').classList.remove('hidden')
                checkAuthStatus()
            } catch (err) {
                document.getElementById('login-error').textContent = 'Network error'
            }
        })
    }
    
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            try {
                const response = await fetch('/api/logout', {
                    method: 'POST'
                })
                
                if (!response.ok) {
                    throw new Error('Logout failed')
                }
                
                // On successful logout, show auth forms
                document.getElementById('auth-forms').classList.remove('hidden')
                document.getElementById('app-content').classList.add('hidden')
            } catch (err) {
                alert('Failed to logout')
            }
        })
    }
}

export async function checkAuthStatus() {
    try {
        const response = await fetch('/api/check-auth', {
            method: 'GET',
            credentials: 'include'
        })
        
        if (response.ok) {
            document.getElementById('auth-forms').classList.add('hidden')
            document.getElementById('app-content').classList.remove('hidden')
        } else {
            document.getElementById('auth-forms').classList.remove('hidden')
            document.getElementById('app-content').classList.add('hidden')
        }
    } catch (err) {
        console.error('Auth check failed:', err)
    }
}