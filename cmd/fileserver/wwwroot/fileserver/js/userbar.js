/**
 * userbar.js — side-effect module.
 *
 * Fetches /chat/api/me and, when the user is authenticated via form auth,
 * populates #slv-user-bar with a "signed in as <DisplayName>" label and a
 * logout button, reusing existing CSS classes from the rest of the UI.
 *
 * Safe to load on pages where form auth is not active (or the visitor is not
 * logged in) — it simply renders nothing in that case.
 *
 * Usage:
 *   As a standalone script tag (404, index, …):
 *     <script type="module" src="/js/userbar.js"></script>
 *
 *   From another entry-point module (browse.js, …):
 *     import "/js/userbar.js";
 *
 * Because ES modules are cached, this file is only ever evaluated once per
 * page load regardless of how many times it is imported.
 */

import { t, whenReady } from "./i18n.js";

function escHtml(s) {
    return String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;");
}

// Fire immediately so the request runs in parallel with the i18n locale fetch.
const mePromise = fetch(`${window.FS_CHAT_ROUTE ?? "/chat"}/api/me`)
    .then(r => (r.ok ? r.json() : null))
    .catch(() => null);

await whenReady(async () => {
    const me = await mePromise;

    // Empty or missing username → form auth not active, or no valid session.
    if (!me?.username) return;

    const container = document.getElementById("slv-user-bar");
    if (!container) return;

    const name = me.displayName || me.username;

    container.innerHTML =
        `<span class="slv-user-bar-identity">` +
            `<span class="slv-chat-identity-label">${escHtml(t("chat_signed_in_as"))}</span>` +
            `<span class="slv-chat-display-name">${escHtml(name)}</span>` +
        `</span>` +
        `<a href="/auth/logout" class="slv-btn secondary">${escHtml(t("chat_logout_btn"))}</a>`;
});