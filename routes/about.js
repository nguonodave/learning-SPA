export function aboutView() {
    const div = document.createElement('div')
    div.innerHTML = `
      <h1>About Page</h1>
      <p>This is a simple SPA built without frameworks.</p>
    `
    return div
  }
  