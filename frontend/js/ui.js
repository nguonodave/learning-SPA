export function setHeadingText(text) {
    const heading = document.getElementById("main-heading")
    if (heading) {
        heading.textContent = text
    }
}
