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

function renderPosts(posts) {
    const postsContainer = document.getElementById('posts-container')
    if (!postsContainer) return
    
    if (posts.length === 0) {
        postsContainer.innerHTML = '<p>No posts yet. Be the first to post!</p>'
        return
    }
    
    postsContainer.innerHTML = posts.map(post => `
        <div class="post" data-id="${post.id}">
            <h3>${post.username}</h3>
            <p>${post.content}</p>
            ${post.image_path ? `<img src="/uploads/${post.image_path}" alt="Post image" style="max-width: 100%;">` : ''}
            <small>${new Date(post.created_at).toLocaleString()}</small>
            <div class="post-actions">
                <button class="like-btn">Like</button>
                <button class="comment-btn">Comment</button>
            </div>
            <div class="comments-container"></div>
        </div>
    `).join('')
}
