const LOCALE_BASE = "/js/locales/";

const FALLBACK = {
    col_name: "Name",
    col_size: "Size",
    col_modified: "Modified",
    col_actions: "Actions",
    empty_dir: "This directory is empty.",
    btn_zip: "ZIP",
    btn_download: "Download",
    btn_info: "Info",
    btn_copy_url: "Copy URL",
    btn_copied: "Copied!",
    search_placeholder: "Search files\u2026 (press / to focus, e.g. config ext:json)",
    search_in_files: "\u229e In files",
    search_in_files_title: "Also search inside file contents",
    search_no_results: "No results for \u201c{q}\u201d",
    ext_filter_label: "ext filter:",
    ext_filter_remove_title: "Remove filter",
    sort_click_title: "Click to sort",
    match_in_file: "in file",
    preview_loading: "Loading\u2026",
    preview_failed: "Failed to load file.",
    readme_link_text: "README \u2193",
    readme_link_title: "Jump to README",
    files_count_one: "{n} file",
    files_count_many: "{n} files",
    lang_label: "Language",
};

let strings = { ...FALLBACK };

export function t(key, vars) {
    let s = strings[key] ?? FALLBACK[key] ?? key;
    if (vars) {
        for (const [k, v] of Object.entries(vars)) {
            s = s.replaceAll(`{${k}}`, v);
        }
    }
    return s;
}

export function detectLang() {
    const saved = localStorage.getItem("fs_lang");
    const locales = window.FS_LOCALES ?? { en: "English" };
    if (saved && locales[saved]) return saved;
    const nav = (navigator.language ?? "en").toLowerCase().split("-")[0];
    return locales[nav] ? nav : "en";
}

export async function loadLocale(lang) {
    try {
        const r = await fetch(`${LOCALE_BASE}${lang}.json`);
        if (!r.ok) throw new Error("not found");
        return await r.json();
    } catch {
        return {};
    }
}

export function setStrings(data) {
    strings = Object.keys(data).length ? data : { ...FALLBACK };
}

export function applyI18n(cfg) {
    for (const el of document.querySelectorAll("[data-i18n]")) {
        el.textContent = t(el.dataset.i18n);
    }
    for (const el of document.querySelectorAll("[data-i18n-placeholder]")) {
        el.placeholder = t(el.dataset.i18nPlaceholder);
    }
    for (const el of document.querySelectorAll("[data-i18n-title]")) {
        el.title = t(el.dataset.i18nTitle);
    }
    for (const th of document.querySelectorAll("th[data-sort]")) {
        th.title = t("sort_click_title");
    }

    const metaEl = document.getElementById("fs-dir-meta");
    if (metaEl) {
        const n = cfg.fileCount;
        if (n > 0) {
            let label = t(n === 1 ? "files_count_one" : "files_count_many", { n });
            if (cfg.totalSize) label += `, ${cfg.totalSize}`;
            metaEl.textContent = label;
        } else {
            metaEl.textContent = "";
        }
    }
}

export function buildLangSelector(onchange) {
    const sel = document.getElementById("fs-lang-select");
    if (!sel) return;

    const locales = window.FS_LOCALES ?? { en: "English" };
    const current = detectLang();
    sel.innerHTML = "";

    for (const [code, label] of Object.entries(locales)) {
        const opt = document.createElement("option");
        opt.value = code;
        opt.textContent = label;
        opt.selected = code === current;
        sel.appendChild(opt);
    }

    sel.addEventListener("change", () => onchange(sel.value));
}