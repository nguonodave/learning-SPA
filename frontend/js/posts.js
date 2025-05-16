export function setupPostForm() {
    const postForm = document.getElementById('create-post')
    
    if (postForm) {
        postForm.addEventListener('submit', async (e) => {
            e.preventDefault()
            const content = document.getElementById('post-content').value
            const imageFile = document.getElementById('post-image').files[0]
            
            try {
                let response
                if (imageFile) {
                    // We'll implement image upload later
                    alert("Image upload will be implemented in the next step")
                    return
                } else {
                    // Text-only post
                    response = await fetch('/api/posts/create', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({ content }),
                        credentials: 'include'
                    })
                }
                
                if (!response.ok) {
                    const error = await response.json()
                    document.getElementById('post-error').textContent = error.message || 'Post failed'
                    return
                }
                
                // Clear form and reload posts
                document.getElementById('post-content').value = ''
                document.getElementById('post-image').value = ''
                document.getElementById('post-error').textContent = ''
            } catch (err) {
                document.getElementById('post-error').textContent = 'Network error'
            }
        })
    }
}
