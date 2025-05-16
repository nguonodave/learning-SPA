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
                    const formData = new FormData()
                    formData.append('content', content)
                    formData.append('image', imageFile)
                    
                    response = await fetch('/api/posts/create', {
                        method: 'POST',
                        body: formData,
                        credentials: 'include'
                    })
                } else {
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
                
                // Get the created post from the response
                const createdPost = await response.json()
                
                // Clear form
                document.getElementById('post-content').value = ''
                document.getElementById('post-image').value = ''
                document.getElementById('post-error').textContent = ''
                
                // Add the new post to the UI without reloading all posts
                addPostToUI(createdPost)
            } catch (err) {
                document.getElementById('post-error').textContent = 'Network error'
            }
        })
    }
}

function addPostToUI(post) {
    const postsContainer = document.getElementById('posts-container')
    if (!postsContainer) return
    
    // Create post element
    const postElement = document.createElement('div')
    postElement.className = 'post'
    postElement.dataset.id = post.id
    postElement.innerHTML = `
        <h3>${post.username}</h3>
        <p>${post.content}</p>
        ${post.image_path ? `<img src="/uploads/${post.image_path}" alt="Post image" style="max-width: 100%;">` : ''}
        <small>${new Date(post.created_at).toLocaleString()}</small>
        <div class="post-actions">
            <button class="like-btn">Like</button>
            <button class="comment-btn">Comment</button>
        </div>
        <div class="comments-container"></div>
    `
    
    // Insert at the top of the posts container
    if (postsContainer.firstChild) {
        postsContainer.insertBefore(postElement, postsContainer.firstChild)
    } else {
        postsContainer.appendChild(postElement)
    }
}

// Modify loadPosts to use addPostToUI
export async function loadPosts() {
    const postsContainer = document.getElementById('posts-container')
    if (!postsContainer) return
    
    try {
        const response = await fetch('/api/posts', {
            credentials: 'include'
        })
        
        if (!response.ok) {
            throw new Error('Failed to load posts')
        }
        
        const posts = await response.json()
        
        // Clear existing posts
        postsContainer.innerHTML = ''
        
        // Add each post to UI
        posts.forEach(post => addPostToUI(post))
        
        // Show message if no posts
        if (posts.length === 0) {
            postsContainer.innerHTML = '<p>No posts yet. Be the first to post!</p>'
        }
    } catch (err) {
        postsContainer.innerHTML = `<p class="error">Failed to load posts: ${err.message}</p>`
    }
}