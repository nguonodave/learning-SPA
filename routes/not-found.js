export function notFoundView() {
    const div = document.createElement('div')
    div.innerHTML = `
      <h1>404 - Not Found</h1>
      <p>That page does not exist.</p>
    `
    return div
  }
  