async function loadCategories() {
    try {
        const response = await fetch('/api/categories', {
            credentials: 'include'
        });

        if (!response.ok) throw new Error('Failed to load categories');

        const categories = await response.json();
        renderCategorySelector(categories);
        setupCategoryNavigation(categories)
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

function setupCategoryNavigation(categories) {
    const categoryList = document.getElementById('category-list');
    const viewAllBtn = document.getElementById('view-all-posts');

    if (!categoryList || !viewAllBtn) return;

    categoryList.innerHTML = categories.map(cat => `
        <button class="category-filter" data-category-id="${cat.id}">
            ${escapeHtml(cat.name)}
        </button>
    `).join('');

    // Add click handlers
    document.querySelectorAll('.category-filter').forEach(btn => {
        btn.addEventListener('click', () => {
            // Highlight selected category
            document.querySelectorAll('.category-filter').forEach(el => {
                el.classList.remove('active');
            });
            btn.classList.add('active');

            // Load posts for this category
            loadPostsByCategory(btn.dataset.categoryId);
        });
    });

    // Handle "View All" button
    viewAllBtn.addEventListener('click', () => {
        loadPosts();
        document.querySelectorAll('.category-filter.active').forEach(el => {
            el.classList.remove('active');
        });
    });
}

async function loadPostsByCategory(categoryId) {
    const postsContainer = document.getElementById('posts-container');
    if (!postsContainer) return;

    try {
        postsContainer.innerHTML = '<p>Loading posts...</p>';

        const response = await fetch(`/api/categories/${categoryId}/posts`);
        if (!response.ok) throw new Error(`Server returned ${response.status}`);

        const posts = await response.json();

        postsContainer.innerHTML = '';
        if (posts.length === 0) {
            postsContainer.innerHTML = '<p>No posts in this category yet.</p>';
            return;
        }

        posts.forEach(post => addPostToUI(post));
    } catch (err) {
        console.error('Failed to load category posts:', err);
        postsContainer.innerHTML = `
            <p class="error">
                Failed to load posts: ${err.message}
                <button onclick="window.location.reload()">Retry</button>
            </p>
        `;
    }
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
        const checkedCategories = document.querySelectorAll('#category-selector input:checked')

        // Clear previous errors
        postError.textContent = '';

        if (!content) {
            postError.textContent = 'Post content cannot be empty';
            return;
        }
        if (checkedCategories.length === 0) {
            postError.textContent = 'Please select at least one category';
            return;
        }

        try {
            const formData = new FormData();
            formData.append('content', content);
            // Get selected categories
            checkedCategories.forEach(checkbox => {
                formData.append('categories', checkbox.value);
            });
            if (imageFile) {
                // Handle image upload
                formData.append('image', imageFile);
            }

            const response = await fetch('/api/posts/create', {
                method: 'POST',
                body: formData,
                credentials: 'include'
            })

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
            document.querySelectorAll('#category-selector input:checked').forEach(cb => cb.checked = false);

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

        postsContainer.innerHTML = posts.map(post => `
            <div class="post" data-id="${post.id}">
                <div class="post-body">
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

                    ${post.categories.length > 0 ? 
                        `<div class="post-categories">
                            ${post.categories.map(cat =>
                            `<span class="category-tag">${escapeHtml(cat)}</span>`
                            ).join('')}
                        </div>`:''
                    }
                </div>

                <div class="post-actions">
                    <button class="like-btn" data-post-id="${post.id}">
                        <span class="like-count">${post.likes_count}</span> Likes
                    </button>
                    <button class="dislike-btn" data-post-id="${post.id}">
                        <span class="dislike-count">${post.dislikes_count}</span> Dislikes
                    </button>
                    <button class="comment-btn" data-post-id="${post.id}">
                        <span class="comment-count">${post.comments_count}</span> Comments
                    </button>
                </div>
                <div class="comments-container" data-post-id="${post.id}">
                    <form class="comment-form">
                        <input type="text" name="comment-input" class="comment-input" placeholder="Write a comment..." required>
                        <button type="submit">Submit</button>
                    </form>
                    <div class="comments-list"></div>
                </div>
            </div>
    `).join('')
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

async function addPostToUI(post) {
    const postsContainer = document.getElementById('posts-container');
    if (!postsContainer) return;

    const postElement = document.createElement('div');
    postElement.className = 'post';
    postElement.dataset.id = post.id;

    // Add categories to post HTML
    const categoriesHtml = post.categories.length > 0
        ? `<div class="post-categories">
            ${post.categories.map(cat =>
            `<span class="category-tag">${escapeHtml(cat)}</span>`
        ).join('')}
        </div>`
        : '';

    postElement.innerHTML = `
        <div class="post-header">
            <div class="post-body">
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

            ${categoriesHtml}
        </div>

        <div class="post-actions">
            <button class="like-btn" data-post-id="${post.id}">
                <span class="like-count">0</span> Likes
            </button>
            <button class="dislike-btn" data-post-id="${post.id}">
                <span class="dislike-count">0</span> Dislikes
            </button>
            <button class="comment-btn" data-post-id="${post.id}">
                <span class="comment-count">0</span> Comments
            </button>
        </div>
        <div class="comments-container" data-post-id="${post.id}">
            <form class="comment-form">
                <input type="text" name="comment-input" class="comment-input" placeholder="Write a comment..." required>
                <button type="submit">Submit</button>
            </form>
            <div class="comments-list"></div>
        </div>
    `;

    // Insert at the top of the container
    if (postsContainer.firstChild) {
        postsContainer.insertBefore(postElement, postsContainer.firstChild);
    } else {
        postsContainer.appendChild(postElement);
    }
}

function setupPostReactions() {
    document.addEventListener('click', async (e) => {
        if (e.target.closest('.like-btn')) {
            const postId = e.target.closest('[data-post-id]').dataset.postId;
            await handleReaction(postId, 'like');
        } else if (e.target.closest('.dislike-btn')) {
            const postId = e.target.closest('[data-post-id]').dataset.postId;
            await handleReaction(postId, 'dislike');
        }
    });
}

async function handleReaction(postId, type) {
    const postElement = document.querySelector(`.post[data-id="${postId}"]`);
    // console.log(postElement)
    if (!postElement) return;

    try {
        const response = await fetch(`/api/posts/${postId}/react`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ type }),
            credentials: 'include'
        });

        if (!response.ok) throw new Error('Reaction failed');

        const { likes, dislikes, userVote } = await response.json();
        // updateReactionUI(postId, counts);

        // Update like button
        const likeBtn = postElement.querySelector('.like-btn');
        likeBtn.classList.toggle('active', userVote === 1);
        likeBtn.querySelector('.like-count').textContent = likes;

        // Update dislike button
        const dislikeBtn = postElement.querySelector('.dislike-btn');
        dislikeBtn.classList.toggle('active', userVote === -1);
        dislikeBtn.querySelector('.dislike-count').textContent = dislikes;
    } catch (err) {
        console.error('Reaction error:', err);
    }
}

// Call this in your initialization
setupPostReactions();

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