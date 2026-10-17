// internal/webui/static/app.js
// Consent-gated GA4, username memory, share and copy buttons, card events.
// No usernames or repository names ever reach GA: page_location drops the query and
// page_title is replaced on the result pages.
(function () {
  "use strict";
  var me = document.currentScript;
  var GA = me && me.dataset.ga;
  var CONSENT = "gha:consent", USER = "gha:lastUser";
  var LOGIN = /^[A-Za-z0-9][A-Za-z0-9\-]{0,38}$/;
  var THEME = "gha:theme";
  // The theme in effect: an explicit choice (set by theme.js or the button), else the system's.
  function currentTheme() {
    var t = document.documentElement.getAttribute("data-theme");
    if (t === "light" || t === "dark") return t;
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  // The button names the theme it switches to.
  function labelTheme(btn) {
    var next = currentTheme() === "dark" ? "light" : "dark";
    btn.textContent = next === "dark" ? "Dark" : "Light";
    btn.setAttribute("aria-label", "Switch to " + next + " theme");
  }
  // Star count for the repository, fetched once per visit. Anonymous and without a referrer:
  // /result URLs name a user. On any failure the buttons keep their plain label.
  var STARS = "gha:stars", REPO_API = "https://api.github.com/repos/nmloc/gh-achievements";
  var stars = null;
  function fillStars() {
    if (stars === null) return;
    var text = stars.toLocaleString("en-US");
    Array.prototype.forEach.call(document.querySelectorAll("[data-stars]"), function (el) {
      el.textContent = text;
      el.hidden = false;
    });
    Array.prototype.forEach.call(document.querySelectorAll(".star-long"), function (el) { el.hidden = true; });
  }
  function loadStars() {
    var cached = null;
    try { cached = sessionStorage.getItem(STARS); } catch (e) { /* storage blocked: fetch instead */ }
    if (cached !== null && /^\d+$/.test(cached)) {
      stars = Number(cached);
      fillStars();
      return;
    }
    if (!window.fetch || !document.querySelector("[data-stars]")) return;
    fetch(REPO_API, { credentials: "omit", referrerPolicy: "no-referrer" })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (d) {
        if (!d || typeof d.stargazers_count !== "number") return;
        stars = d.stargazers_count;
        try { sessionStorage.setItem(STARS, String(stars)); } catch (e) { /* not cached: fetched again on the next page */ }
        fillStars();
      })
      .catch(function () { /* offline or rate-limited: the label stays "Star on GitHub" */ });
  }
  function get(k) { try { return localStorage.getItem(k); } catch (e) { return null; } }
  function set(k, v) { try { localStorage.setItem(k, v); } catch (e) { /* private mode: nothing to remember */ } }
  function drop(k) { try { localStorage.removeItem(k); } catch (e) { /* nothing stored to drop */ } }

  // The consent choice lives in this variable; storage only remembers it across visits,
  // so a click is honoured for the rest of the page view even when storage is blocked.
  var consent = get(CONSENT);

  var loaded = false;
  function loadGA() {
    if (!GA) return;
    window["ga-disable-" + GA] = false;
    if (loaded) return;
    loaded = true;
    // GA reads location, referrer and document.title for every hit. /result's title names the
    // user, so it is replaced there; static pages carry no usernames.
    var path = location.pathname;
    var sensitive = path.indexOf("/result") === 0 || path.indexOf("/auth/callback") === 0;
    var sanitized = {
      page_location: location.origin + location.pathname,
      page_referrer: document.referrer ? new URL(document.referrer).origin + "/" : undefined,
      page_title: sensitive ? "GitHub achievements result" : document.title
    };
    window.dataLayer = window.dataLayer || [];
    window.gtag = function () { window.dataLayer.push(arguments); };
    window.gtag("js", new Date());
    window.gtag("set", sanitized);
    window.gtag("config", GA, sanitized);
    var s = document.createElement("script");
    s.async = true;
    s.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(GA);
    document.head.appendChild(s);
  }
  function stopGA() { if (GA) window["ga-disable-" + GA] = true; }
  function track(name, params) {
    if (consent === "granted" && window.gtag) window.gtag("event", name, params || {});
  }
  function reportView(el) {
    if (!el) return;
    track("report_view", { scope: el.dataset.scope, badges_earned: Number(el.dataset.earned) });
    if (el.dataset.scope === "private" && LOGIN.test(el.dataset.user || "")) set(USER, el.dataset.user);
  }
  window.gha = { track: track, reportView: reportView, fillStars: fillStars };

  // copy writes to the clipboard and calls done only after the write succeeded.
  // No clipboard (insecure context, old browser) or a refusal leaves the page as it was.
  function copy(text, done) {
    try {
      if (!navigator.clipboard || !navigator.clipboard.writeText) return;
      navigator.clipboard.writeText(text).then(done, function () { /* refused: label unchanged */ });
    } catch (e) { /* blocked by policy: nothing visible happens */ }
  }

  function ready() {
    var banner = document.getElementById("consent");
    var themeBtn = document.querySelector("[data-theme-toggle]");
    if (themeBtn) {
      labelTheme(themeBtn);
      themeBtn.hidden = false;
      var mq = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)");
      if (mq && mq.addEventListener) mq.addEventListener("change", function () { labelTheme(themeBtn); });
    }
    if (consent === "granted") loadGA();
    else if (consent === null && banner) banner.hidden = false;
    document.addEventListener("click", function (e) {
      var t = e.target.closest("[data-consent],[data-consent-reset],[data-signin],[data-share],[data-copy],[data-theme-toggle],[data-menu],[data-star]");
      if (!t) return;
      if (t.dataset.consent) {
        consent = t.dataset.consent;
        set(CONSENT, consent);
        if (banner) banner.hidden = true;
        if (consent === "granted") loadGA(); else stopGA();
      } else if (t.hasAttribute("data-consent-reset")) {
        consent = null;
        drop(CONSENT);
        stopGA();
        if (banner) banner.hidden = false;
      } else if (t.hasAttribute("data-signin")) {
        track("signin_click");
        track("check_start", { method: "signin" });
      } else if (t.hasAttribute("data-share")) {
        var r = document.querySelector("[data-report]");
        var url = location.origin + "/result?u=" + encodeURIComponent(r ? r.dataset.user : "");
        copy(url, function () { t.textContent = "Link copied"; track("share_click"); });
      } else if (t.hasAttribute("data-theme-toggle")) {
        var next = currentTheme() === "dark" ? "light" : "dark";
        document.documentElement.setAttribute("data-theme", next);
        set(THEME, next);
        labelTheme(t);
      } else if (t.hasAttribute("data-menu")) {
        var nav = document.getElementById("site-nav");
        var open = t.getAttribute("aria-expanded") !== "true";
        t.setAttribute("aria-expanded", String(open));
        if (nav && open) nav.setAttribute("data-open", "");
        else if (nav) nav.removeAttribute("data-open");
      } else if (t.dataset.star) {
        track("star_click", { place: t.dataset.star });
      } else if (t.dataset.copy) {
        copy(t.dataset.copy, function () { t.textContent = "Copied"; track("cli_copy"); });
      }
    });
    document.addEventListener("toggle", function (e) {
      var d = e.target;
      if (d.matches && d.matches("details[data-badge]") && d.open) track("badge_open", { badge: d.dataset.badge });
    }, true);
    var SIZES = ["badge-img", "medal-img", "badge-hero"];
    // Swap a broken badge image for a placeholder tile that keeps its size class.
    function fallback(img) {
      var size = SIZES.filter(function (c) { return img.classList.contains(c); })[0];
      if (!size) return;
      var tile = document.createElement("span");
      tile.className = size + " placeholder";
      tile.textContent = (img.alt || "?").slice(0, 2).toUpperCase();
      img.replaceWith(tile);
    }
    document.addEventListener("error", function (e) {
      if (e.target.tagName === "IMG") fallback(e.target);
    }, true);
    // Images that failed before this script ran never dispatched an event we saw.
    Array.prototype.slice.call(document.images).forEach(function (img) {
      if (img.complete && img.naturalWidth === 0) fallback(img);
    });
    var form = document.getElementById("check-form");
    if (form) {
      var input = form.querySelector('input[name="u"]');
      // Trim as the user types: the browser validates the pattern before submit fires, so a
      // pasted " octocat " would be refused if it were only trimmed on submit.
      var trim = function () {
        var v = input.value.trim();
        if (v !== input.value) input.value = v;
      };
      if (input && !input.value) input.value = get(USER) || "";
      if (input) input.addEventListener("input", trim);
      form.addEventListener("submit", function () {
        trim();
        if (LOGIN.test(input.value)) set(USER, input.value);
        track("check_start", { method: "public" });
      });
    }
    reportView(document.querySelector("[data-report]"));
    loadStars();
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", ready);
  else ready();
})();
