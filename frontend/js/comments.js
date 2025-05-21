export function setupCommentForm() {
    document.addEventListener('submit', async (e) => {
        if (e.target.matches('.comment-form')) {
            e.preventDefault()
            const form = e.target
            const postId = e.target.closest('[data-post-id]').dataset.postId
            const content = form.querySelector('.comment-input').value
            console.log(postId, content)

            createComment(postId, content)

            form.querySelector('.comment-input').value = ''
        }
    });
}

async function createComment(postId, content) {
    try {
        const response = await fetch(`/api/posts/${postId}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ content }),
            credentials: 'include'
        })

        if (!response.ok) throw new Error('Reaction failed');

        const commentsCount = await response.json()

        const commentBtn = document.querySelector(`.comment-btn[data-post-id="${postId}"]`);
        commentBtn.querySelector('.comment-count').textContent = commentsCount

        console.log(commentsCount)
    } catch (error) {
        console.log("comment creation error", error)
    }
}

async function revealComments(postId) {
    const commentsContainer = document.querySelector(`.comments-container[data-post-id="${postId}"]`);
    if (!commentsContainer) return
    const commentsList = commentsContainer.querySelector('.comments-list');
    commentsList.innerHTML = '<p>Loading comments...</p>'

    try {
        const response = await fetch(`/api/posts/${postId}/comments`)

        if (!response.ok) throw new Error('Failed to load comments');

        const comments = await response.json();
        commentsList.innerHTML = ''

        if (comments.length === 0) {
            container.innerHTML = '<p>No comments yet</p>';
        } else {
            // comments.forEach(comment => addCommentToUI(comment, container));
            commentsList.innerHTML = comments.map(comment => `
                <div class="user-comment-container">
                    <p class="commenter">By <span>${comment.username}</span></p>
                    <p class="user-comment-content">${comment.content}</p>
                    <p class="comment-created-time">${comment.createdAt}</p>
                </div>
            `).join('')
        }
    } catch (error) {
        console.error('Failed to load comments:', error)
    }
}

document.addEventListener('click', (e) => {
    // Handle post body click
    const postBody = e.target.closest('.post-body');
    if (postBody) {
        const postId = postBody.closest('.post').getAttribute('data-id');
        if (postId) {
            revealComments(postId);
        }
    }

    // Handle comment button click
    const commentBtn = e.target.closest('.comment-btn');
    if (commentBtn) {
        e.stopPropagation(); // prevent bubbling to post body (Prevent the .post-body handler from also firing when the "Comment" button is clicked.)
        const postId = commentBtn.getAttribute('data-post-id');
        if (postId) {
            revealComments(postId);
        }
    }
});
