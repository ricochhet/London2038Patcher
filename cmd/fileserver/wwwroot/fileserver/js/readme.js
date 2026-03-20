(function () {
    "use strict";

    var src = document.getElementById("readme-src");
    var render = document.getElementById("readme-render");

    if (src && render && typeof marked !== "undefined") {
        render.innerHTML = marked.parse(src.textContent);
    }
})();