async function loadCategories() {
    try {
        const response = await fetch('/api/categories', {
            credentials: 'include'
        });
        
        if (!response.ok) throw new Error('Failed to load categories');
        
        const categories = await response.json();
        renderCategorySelector(categories);
    } catch (err) {
        const container = document.getElementById('category-selector');
        if (container) {
            container.innerHTML = 
                `<p class="error">Failed to load categories: ${err.message}</p>`;
        }
    }
}

function renderCategorySelector(categories) {
    const container = document.getElementById('category-selector');
    if (!container) return;
    
    container.innerHTML = categories.map(cat => `
        <label class="category-option">
            <input type="checkbox" name="categories" value="${cat.id}">
            ${escapeHtml(cat.name)}
        </label>
    `).join('');
}

loadCategories()

export function setupPostForm() {
    const postForm = document.getElementById('create-post');
    const postError = document.getElementById('post-error');
    
    if (!postForm) return;

    postForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const content = document.getElementById('post-content').value.trim();
        const imageFile = document.getElementById('post-image').files[0];
        
        // Clear previous errors
        postError.textContent = '';
        
        try {
            let response;
            
            if (imageFile) {
                // Handle image upload
                const formData = new FormData();
                formData.append('content', content);
                formData.append('image', imageFile);
                
                response = await fetch('/api/posts/create', {
                    method: 'POST',
                    body: formData,
                    credentials: 'include'
                });
            } else {
                // Handle text-only post
                if (!content) {
                    postError.textContent = 'Post content cannot be empty';
                    return;
                }
                
                response = await fetch('/api/posts/create', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ content }),
                    credentials: 'include'
                });
            }

            if (!response.ok) {
                const error = await response.json().catch(() => ({ message: 'Unknown error' }));
                postError.textContent = error.message || 'Failed to create post';
                return;
            }

            // Success - handle response
            const createdPost = await response.json();
            
            // Clear form
            document.getElementById('post-content').value = '';
            document.getElementById('post-image').value = '';
            
            // Add new post to UI
            addPostToUI(createdPost);
            
        } catch (err) {
            console.error('Post creation error:', err);
            postError.textContent = 'Network error - please try again';
        }
    });
}

export async function loadPosts() {
    const postsContainer = document.getElementById('posts-container');
    if (!postsContainer) return;
    
    try {
        postsContainer.innerHTML = '<p>Loading posts...</p>';
        
        const response = await fetch('/api/posts', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error(`Server returned ${response.status}`);
        }
        
        const posts = await response.json();
        
        // Clear container
        postsContainer.innerHTML = '';
        
        if (posts.length === 0) {
            postsContainer.innerHTML = '<p class="no-posts">No posts yet. Be the first to post!</p>';
            return;
        }
        
        // Add each post to UI
        posts.forEach(post => addPostToUI(post));
        
    } catch (err) {
        console.error('Failed to load posts:', err);
        postsContainer.innerHTML = `
            <p class="error">
                Failed to load posts: ${err.message}
                <button onclick="window.location.reload()">Retry</button>
            </p>
        `;
    }
}

function addPostToUI(post) {
    const postsContainer = document.getElementById('posts-container');
    if (!postsContainer) return;
    
    const postElement = document.createElement('div');
    postElement.className = 'post';
    postElement.dataset.id = post.id;
    
    postElement.innerHTML = `
        <div class="post-header">
            <h3 class="post-username">${post.username}</h3>
            <small class="post-time">${formatDate(post.created_at)}</small>
        </div>
        <div class="post-content">
            <p>${escapeHtml(post.content)}</p>
            ${post.image_path ? `
                <div class="post-image">
                    <img src="/uploads/${post.image_path}" alt="Post image">
                </div>
            ` : ''}
        </div>
        <div class="post-actions">
            <button class="like-btn" data-post-id="${post.id}">
                <span class="like-count">0</span> Likes
            </button>
            <button class="comment-btn" data-post-id="${post.id}">
                <span class="comment-count">0</span> Comments
            </button>
        </div>
        <div class="comments-container" data-post-id="${post.id}"></div>
    `;
    
    // Insert at the top of the container
    if (postsContainer.firstChild) {
        postsContainer.insertBefore(postElement, postsContainer.firstChild);
    } else {
        postsContainer.appendChild(postElement);
    }
    
    // TODO: Add event listeners for like/comment buttons
}

// Helper functions
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString();
}

function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}