(function () {
    "use strict";

    var LOCALES = window.FS_LOCALES || {
        en: "English",
    };
    var LOCALE_BASE = "/js/locales/";

    var FALLBACK = {
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
        lang_label: "Language"
    };

    var strings = FALLBACK;

    function t(key, vars) {
        var s = strings[key] || FALLBACK[key] || key;
        if (vars) {
            Object.keys(vars).forEach(function (k) {
                s = s.split("{" + k + "}").join(vars[k]);
            });
        }
        return s;
    }

    function detectLang() {
        var saved = localStorage.getItem("fs_lang");
        if (saved && LOCALES[saved]) { return saved; }
        var nav = (navigator.language || "en").toLowerCase().split("-")[0];
        return LOCALES[nav] ? nav : "en";
    }

    function loadLocale(lang) {
        return fetch(LOCALE_BASE + lang + ".json")
            .then(function (r) {
                if (!r.ok) { throw new Error("not found"); }
                return r.json();
            })
            .catch(function () {
                return {};
            });
    }

    function applyI18n() {
        document.querySelectorAll("[data-i18n]").forEach(function (el) {
            el.textContent = t(el.getAttribute("data-i18n"));
        });
        document.querySelectorAll("[data-i18n-placeholder]").forEach(function (el) {
            el.placeholder = t(el.getAttribute("data-i18n-placeholder"));
        });
        document.querySelectorAll("[data-i18n-title]").forEach(function (el) {
            el.title = t(el.getAttribute("data-i18n-title"));
        });

        document.querySelectorAll("th[data-sort]").forEach(function (th) {
            th.title = t("sort_click_title");
        });

        var metaEl = document.getElementById("fs-dir-meta");
        if (metaEl) {
            var n = CFG.fileCount;
            if (n > 0) {
                var label = t(n === 1 ? "files_count_one" : "files_count_many", { n: n });
                if (CFG.totalSize) { label += ", " + CFG.totalSize; }
                metaEl.textContent = label;
            } else {
                metaEl.textContent = "";
            }
        }
    }

    function buildLangSelector() {
        var sel = document.getElementById("fs-lang-select");
        if (!sel) { return; }

        var current = detectLang();
        sel.innerHTML = "";

        Object.keys(LOCALES).forEach(function (code) {
            var opt = document.createElement("option");
            opt.value = code;
            opt.textContent = LOCALES[code];
            if (code === current) { opt.selected = true; }
            sel.appendChild(opt);
        });

        sel.addEventListener("change", function () {
            var lang = sel.value;
            localStorage.setItem("fs_lang", lang);
            loadLocale(lang).then(function (data) {
                strings = Object.keys(data).length ? data : FALLBACK;
                applyI18n();
                renderTable(null);
            });
        });
    }

    var CFG = window.FS_CONFIG || {};
    var ROUTE = CFG.route || "";
    var ENTRIES = CFG.entries || [];
    var IMAGE_EXTS = CFG.imageExts || {};
    var TEXT_EXTS = CFG.textExts || {};

    var input = document.getElementById("fs-search-input");
    var results = document.getElementById("fs-search-results");
    var contentToggle = document.getElementById("fs-content-toggle");
    var tagHint = document.getElementById("fs-tag-hint");
    var tbody = document.getElementById("fs-tbody");
    var previewModal = document.getElementById("fs-preview-modal");
    var previewClose = document.getElementById("fs-preview-close");
    var previewBackdrop = document.getElementById("fs-preview-backdrop");
    var previewTitle = document.getElementById("fs-preview-title");
    var previewDL = document.getElementById("fs-preview-dl");
    var previewBody = document.getElementById("fs-preview-body");

    var SK = location.pathname;
    var sortCol = sessionStorage.getItem("fs_sc_" + SK) || "name";
    var sortDir = sessionStorage.getItem("fs_sd_" + SK) || "asc";
    var contentSearch = false;

    function isPreviewable(ext) {
        return IMAGE_EXTS[ext] || TEXT_EXTS[ext] || ext === ".pdf";
    }

    function escHtml(s) {
        return String(s)
            .replace(/&/g, "&amp;").replace(/</g, "&lt;")
            .replace(/>/g, "&gt;").replace(/"/g, "&quot;");
    }

    function sortedEntries() {
        var arr = ENTRIES.slice();
        arr.sort(function (a, b) {
            if (a.isDir !== b.isDir) { return a.isDir ? -1 : 1; }
            var va, vb;
            if (sortCol === "size") { va = a.sizeBytes; vb = b.sizeBytes; }
            else if (sortCol === "modified") { va = a.modUnix; vb = b.modUnix; }
            else { va = a.name.toLowerCase(); vb = b.name.toLowerCase(); }
            if (va < vb) { return sortDir === "asc" ? -1 : 1; }
            if (va > vb) { return sortDir === "asc" ? 1 : -1; }
            return 0;
        });
        return arr;
    }

    function makeRow(e) {
        var tr = document.createElement("tr");

        var td1 = document.createElement("td");
        var nameDiv = document.createElement("div");
        nameDiv.className = "fs-name";

        var icon = document.createElement("span");
        icon.className = "fs-icon";
        icon.textContent = e.isDir ? "\uD83D\uDCC1" : "\uD83D\uDCC4";

        var link = document.createElement("a");
        link.textContent = e.name;

        if (e.isDir) {
            link.href = e.browseURL;
        } else {
            link.href = e.downloadURL;
            if (isPreviewable(e.ext)) {
                link.addEventListener("click", function (ev) {
                    ev.preventDefault();
                    showPreview(e);
                });
            }
        }

        nameDiv.appendChild(icon);
        nameDiv.appendChild(link);
        td1.appendChild(nameDiv);

        var td2 = document.createElement("td");
        td2.className = "fs-meta";
        td2.textContent = e.sizeStr;

        var td3 = document.createElement("td");
        td3.className = "fs-meta";
        td3.textContent = e.modStr;

        var td4 = document.createElement("td");
        var actDiv = document.createElement("div");
        actDiv.className = "fs-actions";

        var dlBtn = document.createElement("a");
        dlBtn.className = "fs-btn";
        dlBtn.href = e.downloadURL;
        dlBtn.textContent = e.isDir ? t("btn_zip") : t("btn_download");
        actDiv.appendChild(dlBtn);

        if (!e.isDir) {
            var infoBtn = document.createElement("a");
            infoBtn.className = "fs-btn secondary";
            infoBtn.href = e.infoURL;
            infoBtn.textContent = t("btn_info");
            actDiv.appendChild(infoBtn);

            var copyBtn = document.createElement("button");
            copyBtn.className = "fs-btn secondary fs-copy-btn";
            copyBtn.textContent = t("btn_copy_url");
            (function (btn, url) {
                btn.addEventListener("click", function () {
                    var full = location.origin + url;
                    var reset = function () { btn.textContent = t("btn_copy_url"); };
                    if (navigator.clipboard) {
                        navigator.clipboard.writeText(full).then(function () {
                            btn.textContent = t("btn_copied");
                            setTimeout(reset, 1500);
                        });
                    } else {
                        var ta = document.createElement("textarea");
                        ta.value = full;
                        ta.style.position = "fixed";
                        ta.style.opacity = "0";
                        document.body.appendChild(ta);
                        ta.select();
                        document.execCommand("copy");
                        document.body.removeChild(ta);
                        btn.textContent = t("btn_copied");
                        setTimeout(reset, 1500);
                    }
                });
            })(copyBtn, e.downloadURL);
            actDiv.appendChild(copyBtn);
        }

        td4.appendChild(actDiv);
        tr.appendChild(td1);
        tr.appendChild(td2);
        tr.appendChild(td3);
        tr.appendChild(td4);
        return tr;
    }

    function updateSortIndicators() {
        document.querySelectorAll("th[data-sort]").forEach(function (th) {
            th.classList.remove("fs-sort-asc", "fs-sort-desc");
            if (th.dataset.sort === sortCol) {
                th.classList.add(sortDir === "asc" ? "fs-sort-asc" : "fs-sort-desc");
            }
        });
    }

    function renderTable(hlName) {
        var parentRow = tbody.querySelector(".fs-parent");
        tbody.innerHTML = "";
        if (parentRow) { tbody.appendChild(parentRow); }

        var sorted = sortedEntries();
        sorted.forEach(function (e) {
            var tr = makeRow(e);
            if (hlName && e.name === hlName) {
                tr.classList.add("fs-row-highlight");
                setTimeout(function () {
                    tr.classList.remove("fs-row-highlight");
                    tr.classList.add("fs-row-highlight-fade");
                }, 1500);
                setTimeout(function () {
                    tr.scrollIntoView({ behavior: "smooth", block: "center" });
                }, 50);
            }
            tbody.appendChild(tr);
        });

        if (sorted.length === 0 && !parentRow) {
            var emptyRow = document.createElement("tr");
            var emptyTd = document.createElement("td");
            emptyTd.className = "fs-empty";
            emptyTd.setAttribute("colspan", "4");
            emptyTd.textContent = t("empty_dir");
            emptyRow.appendChild(emptyTd);
            tbody.appendChild(emptyRow);
        }

        updateSortIndicators();
    }

    document.querySelectorAll("th[data-sort]").forEach(function (th) {
        th.style.cursor = "pointer";
        th.addEventListener("click", function () {
            var col = th.dataset.sort;
            if (col === sortCol) {
                sortDir = sortDir === "asc" ? "desc" : "asc";
            } else {
                sortCol = col;
                sortDir = "asc";
            }
            sessionStorage.setItem("fs_sc_" + SK, sortCol);
            sessionStorage.setItem("fs_sd_" + SK, sortDir);
            renderTable(null);
        });
    });

    var searchTimer = null;
    var activeIdx = -1;

    function parseQueryTags(raw) {
        var tokens = raw.trim().split(/\s+/);
        var rest = [];
        var ext = null;
        tokens.forEach(function (tok) {
            var lower = tok.toLowerCase();
            var val = null;
            if (lower.indexOf("extension:") === 0) { val = lower.slice("extension:".length); }
            else if (lower.indexOf("ext:") === 0) { val = lower.slice("ext:".length); }
            if (val !== null && val !== "") {
                if (val.charAt(0) !== ".") { val = "." + val; }
                ext = val;
            } else {
                rest.push(tok);
            }
        });
        return { base: rest.join(" "), ext: ext };
    }

    function updateTagHint(parsed) {
        if (!parsed.ext) {
            tagHint.innerHTML = "";
            tagHint.style.display = "none";
            return;
        }
        tagHint.style.display = "flex";
        tagHint.innerHTML =
            '<span class="fs-tag-chip">'
            + escHtml(t("ext_filter_label")) + " <strong>" + escHtml(parsed.ext) + "</strong>"
            + '<button class="fs-tag-remove" title="' + escHtml(t("ext_filter_remove_title")) + '">\xd7</button>'
            + "</span>";
        tagHint.querySelector(".fs-tag-remove").addEventListener("click", function () {
            input.value = input.value
                .replace(/\b(?:ext|extension):\S*/gi, "")
                .replace(/\s+/g, " ").trim();
            input.dispatchEvent(new Event("input"));
            input.focus();
        });
    }

    function closeResults() {
        results.classList.remove("open");
        results.innerHTML = "";
        activeIdx = -1;
    }

    function doSearch(raw) {
        activeIdx = -1;
        var parsed = parseQueryTags(raw);
        updateTagHint(parsed);
        if (!parsed.base && !parsed.ext) { closeResults(); return; }

        var fetchURL = ROUTE + "?search=" + encodeURIComponent(raw);
        if (contentSearch) { fetchURL += "&content=1"; }

        fetch(fetchURL)
            .then(function (r) { return r.json(); })
            .then(function (items) {
                if (!items || items.length === 0) {
                    var label = parsed.base || ("*" + (parsed.ext || ""));
                    results.innerHTML =
                        '<div class="fs-search-empty">'
                        + escHtml(t("search_no_results", { q: label }))
                        + "</div>";
                } else {
                    results.innerHTML = items.map(function (item) {
                        var badge = item.matchType === "content"
                            ? '<span class="fs-match-badge">' + escHtml(t("match_in_file")) + "</span>"
                            : "";
                        var snippet = item.snippet
                            ? '<span class="fs-search-snippet">' + escHtml(item.snippet) + "</span>"
                            : "";
                        return '<a class="fs-search-item" href="' + escHtml(item.highlightURL) + '">'
                            + '<div class="fs-search-item-top">'
                            + '<span class="fs-search-item-name">' + escHtml(item.name) + "</span>"
                            + badge
                            + "</div>"
                            + '<span class="fs-search-item-path">' + escHtml(item.relPath) + "</span>"
                            + snippet
                            + "</a>";
                    }).join("");
                }
                results.classList.add("open");
            })
            .catch(function () { closeResults(); });
    }

    var savedQ = sessionStorage.getItem("fs_q_" + SK);
    if (savedQ) { input.value = savedQ; }

    input.addEventListener("input", function () {
        clearTimeout(searchTimer);
        var q = input.value.trim();
        sessionStorage.setItem("fs_q_" + SK, q);
        if (!q) {
            closeResults();
            updateTagHint({ ext: null });
            return;
        }
        searchTimer = setTimeout(function () { doSearch(q); }, 220);
    });

    contentToggle.addEventListener("click", function () {
        contentSearch = !contentSearch;
        contentToggle.classList.toggle("active", contentSearch);
        var q = input.value.trim();
        if (q) { doSearch(q); }
    });

    function setActive(idx) {
        var items = results.querySelectorAll(".fs-search-item");
        if (activeIdx >= 0 && activeIdx < items.length) {
            items[activeIdx].classList.remove("fs-search-item-active");
        }
        activeIdx = idx;
        if (activeIdx >= 0 && activeIdx < items.length) {
            items[activeIdx].classList.add("fs-search-item-active");
            items[activeIdx].scrollIntoView({ block: "nearest" });
        }
    }

    input.addEventListener("keydown", function (e) {
        var items = results.querySelectorAll(".fs-search-item");
        if (e.key === "Escape") {
            closeResults();
            input.value = "";
            updateTagHint({ ext: null });
            sessionStorage.removeItem("fs_q_" + SK);
        } else if (e.key === "ArrowDown") {
            e.preventDefault();
            if (results.classList.contains("open")) {
                setActive(Math.min(activeIdx + 1, items.length - 1));
            }
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            if (results.classList.contains("open")) {
                setActive(Math.max(activeIdx - 1, 0));
            }
        } else if (e.key === "Enter" && activeIdx >= 0 && activeIdx < items.length) {
            e.preventDefault();
            items[activeIdx].click();
        }
    });

    document.addEventListener("click", function (e) {
        if (!input.contains(e.target) && !results.contains(e.target) &&
            !contentToggle.contains(e.target) && !tagHint.contains(e.target)) {
            closeResults();
        }
    });

    document.addEventListener("keydown", function (e) {
        if (e.key === "/" &&
            document.activeElement !== input &&
            document.activeElement.tagName !== "INPUT" &&
            document.activeElement.tagName !== "TEXTAREA" &&
            document.activeElement.tagName !== "SELECT") {
            e.preventDefault();
            input.focus();
        }
    });

    function showPreview(entry) {
        previewTitle.textContent = entry.name;
        previewDL.href = entry.downloadURL;
        previewBody.innerHTML = "";
        previewModal.classList.add("open");
        document.body.style.overflow = "hidden";

        if (IMAGE_EXTS[entry.ext]) {
            var img = document.createElement("img");
            img.src = entry.previewURL;
            img.className = "fs-preview-img";
            previewBody.appendChild(img);
        } else if (entry.ext === ".pdf") {
            var iframe = document.createElement("iframe");
            iframe.src = entry.previewURL;
            iframe.className = "fs-preview-iframe";
            previewBody.appendChild(iframe);
        } else if (TEXT_EXTS[entry.ext]) {
            previewBody.innerHTML =
                '<div class="fs-preview-loading">' + escHtml(t("preview_loading")) + "</div>";
            fetch(entry.previewURL)
                .then(function (r) { return r.text(); })
                .then(function (text) {
                    previewBody.innerHTML = "";
                    var pre = document.createElement("pre");
                    pre.className = "fs-preview-text";
                    pre.textContent = text;
                    previewBody.appendChild(pre);
                })
                .catch(function () {
                    previewBody.innerHTML =
                        '<div class="fs-preview-error">' + escHtml(t("preview_failed")) + "</div>";
                });
        }
    }

    function closePreview() {
        previewModal.classList.remove("open");
        previewBody.innerHTML = "";
        document.body.style.overflow = "";
    }

    previewClose.addEventListener("click", closePreview);
    previewBackdrop.addEventListener("click", closePreview);
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape" && previewModal.classList.contains("open")) {
            closePreview();
        }
    });

    var params = new URLSearchParams(window.location.search);
    renderTable(params.get("highlight") || null);
    applyI18n();
    buildLangSelector();
    if (savedQ) { updateTagHint(parseQueryTags(savedQ)); }

    var initialLang = detectLang();
    if (initialLang !== "en") {
        loadLocale(initialLang).then(function (data) {
            if (Object.keys(data).length) {
                strings = data;
                applyI18n();
                renderTable(params.get("highlight") || null);
            }
        });
    }

})();