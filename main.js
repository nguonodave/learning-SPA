import { homeView } from './routes/home.js'
import { aboutView } from './routes/about.js'
import { notFoundView } from './routes/not-found.js'

function render(content) {
  const container = document.getElementById('app')
  container.innerHTML = ''
  container.appendChild(content)
}

function router() {
  const routes = {
    '/': homeView,
    '/about': aboutView
  }

  const path = window.location.pathname
  const viewFunc = routes[path] || notFoundView
  const content = viewFunc()
  render(content)
}

function navigate(event) {
  const link = event.target.closest('a[data-link]')
  if (!link) return

  event.preventDefault()
  const url = link.getAttribute('href')
  history.pushState(null, '', url)
  router()
}

function start() {
  document.body.addEventListener('click', navigate)
  window.addEventListener('popstate', router)
  router()
}

start()
