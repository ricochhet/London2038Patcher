const src = document.getElementById("readme-src");
const render = document.getElementById("readme-render");

if (src && render && typeof marked !== "undefined") {
    render.innerHTML = marked.parse(src.textContent);
}