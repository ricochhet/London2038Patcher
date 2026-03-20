import { t } from "./i18n.js";
import { escHtml } from "./table.js";

export function initPreview(imageExts, textExts) {
    const modal = document.getElementById("fs-preview-modal");
    const close = document.getElementById("fs-preview-close");
    const backdrop = document.getElementById("fs-preview-backdrop");
    const title = document.getElementById("fs-preview-title");
    const dl = document.getElementById("fs-preview-dl");
    const body = document.getElementById("fs-preview-body");

    const closePreview = () => {
        modal.classList.remove("open");
        body.innerHTML = "";
        document.body.style.overflow = "";
    };

    close.addEventListener("click", closePreview);
    backdrop.addEventListener("click", closePreview);
    document.addEventListener("keydown", e => {
        if (e.key === "Escape" && modal.classList.contains("open")) closePreview();
    });

    return async function showPreview(entry) {
        title.textContent = entry.name;
        dl.href = entry.downloadURL;
        body.innerHTML = "";
        modal.classList.add("open");
        document.body.style.overflow = "hidden";

        if (imageExts[entry.ext]) {
            const img = Object.assign(document.createElement("img"), {
                src: entry.previewURL,
                className: "fs-preview-img",
            });
            body.appendChild(img);
        } else if (entry.ext === ".pdf") {
            const iframe = Object.assign(document.createElement("iframe"), {
                src: entry.previewURL,
                className: "fs-preview-iframe",
            });
            body.appendChild(iframe);
        } else if (textExts[entry.ext]) {
            body.innerHTML = `<div class="fs-preview-loading">${escHtml(t("preview_loading"))}</div>`;
            try {
                const r = await fetch(entry.previewURL);
                const text = await r.text();
                body.innerHTML = "";
                const pre = Object.assign(document.createElement("pre"), {
                    className: "fs-preview-text",
                    textContent: text,
                });
                body.appendChild(pre);
            } catch {
                body.innerHTML = `<div class="fs-preview-error">${escHtml(t("preview_failed"))}</div>`;
            }
        }
    };
}