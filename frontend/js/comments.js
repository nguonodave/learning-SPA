export function setupCommentForm() {
    document.addEventListener('submit', async (e) => {
        if (e.target.matches('.comment-form')) {
            e.preventDefault()
            const form = e.target
            const postId = e.target.closest('[data-post-id]').dataset.postId
            const content = form.querySelector('.comment-input').value
            console.log(postId, content)
            
            createComment(postId, content)
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

        console.log(commentsCount)
    } catch (error) {
        console.log("comment creation error", error)
    }
}