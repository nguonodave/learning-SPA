import { setHeadingText } from './js/ui.js'
import { setupAuthForms, checkAuthStatus } from './js/auth.js'

setHeadingText("PostApp")

// Initialize auth forms
setupAuthForms()

// Check auth status on page load
checkAuthStatus()