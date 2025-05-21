import { setHeadingText } from './js/ui.js'
import { setupAuthForms, checkAuthStatus } from './js/auth.js'
import { setupPostForm, loadPosts } from './js/posts.js'
import { setupCommentForm } from './js/comments.js'

setHeadingText("PostApp")

// Initialize auth forms
setupAuthForms()

// Check auth status on page load and setup post functionality if authenticated
async function initializeApp() {
    try {
        console.log("Checking authentication status...") // This will now log
        const authenticated = await checkAuthStatus()
        console.log("Authenticated:", authenticated) // This will show the status

        if (authenticated) {
            // Setup post form and load posts if authenticated
            setupPostForm()
            loadPosts()
            setupCommentForm()
        }
    } catch (err) {
        console.error("Initialization error:", err)
    }
}

// Start the app
initializeApp()