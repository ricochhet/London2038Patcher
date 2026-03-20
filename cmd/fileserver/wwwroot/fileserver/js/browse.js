import {
    t, detectLang, loadLocale, setStrings,
    applyI18n, buildLangSelector,
} from "./i18n.js";
import { renderTable, initSortHeaders } from "./table.js";
import { initSearch, parseQueryTags, updateTagHint } from "./search.js";
import { initPreview } from "./preview.js";

const CFG = window.FS_CONFIG ?? {};
const ROUTE = CFG.route ?? "";
const ENTRIES = CFG.entries ?? [];
const IMAGE_EXTS = CFG.imageExts ?? {};
const TEXT_EXTS = CFG.textExts ?? {};

const showPreview = initPreview(IMAGE_EXTS, TEXT_EXTS);

const render = hlName => renderTable(ENTRIES, IMAGE_EXTS, TEXT_EXTS, hlName, showPreview);

initSortHeaders(ENTRIES, IMAGE_EXTS, TEXT_EXTS, showPreview);
initSearch(ROUTE);

const params = new URLSearchParams(window.location.search);
const hlName = params.get("highlight") ?? null;
const savedQ = sessionStorage.getItem(`fs_q_${location.pathname}`);

render(hlName);
applyI18n(CFG);
buildLangSelector(async lang => {
    localStorage.setItem("fs_lang", lang);
    const data = await loadLocale(lang);
    setStrings(data);
    applyI18n(CFG);
    render(null);
});

if (savedQ) updateTagHint(parseQueryTags(savedQ));

const initialLang = detectLang();
if (initialLang !== "en") {
    loadLocale(initialLang).then(data => {
        if (Object.keys(data).length) {
            setStrings(data);
            applyI18n(CFG);
            render(hlName);
        }
    });
}