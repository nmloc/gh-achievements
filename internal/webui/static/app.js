// internal/webui/static/app.js
// Consent-gated GA4, username memory, share and copy buttons, card events.
// No usernames or repository names ever reach GA: page_location drops the query.
(function () {
  "use strict";
  var me = document.currentScript;
  var GA = me && me.dataset.ga;
  var CONSENT = "gha:consent", USER = "gha:lastUser";
  var LOGIN = /^[A-Za-z0-9][A-Za-z0-9-]{0,38}$/;
  function get(k) { try { return localStorage.getItem(k); } catch (e) { return null; } }
  function set(k, v) { try { localStorage.setItem(k, v); } catch (e) { /* private mode: nothing to remember */ } }

  var loaded = false;
  function loadGA() {
    if (!GA || loaded) return;
    loaded = true;
    window.dataLayer = window.dataLayer || [];
    window.gtag = function () { window.dataLayer.push(arguments); };
    var ref = document.referrer ? new URL(document.referrer).origin + "/" : undefined;
    window.gtag("js", new Date());
    window.gtag("config", GA, { page_location: location.origin + location.pathname, page_referrer: ref });
    var s = document.createElement("script");
    s.async = true;
    s.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(GA);
    document.head.appendChild(s);
  }
  function track(name, params) {
    if (get(CONSENT) === "granted" && window.gtag) window.gtag("event", name, params || {});
  }
  function reportView(el) {
    if (!el) return;
    track("report_view", { scope: el.dataset.scope, badges_earned: Number(el.dataset.earned) });
    if (el.dataset.scope === "private" && LOGIN.test(el.dataset.user || "")) set(USER, el.dataset.user);
  }
  window.gha = { track: track, reportView: reportView };

  function ready() {
    var banner = document.getElementById("consent");
    var choice = get(CONSENT);
    if (choice === "granted") loadGA();
    else if (choice === null && banner) banner.hidden = false;
    document.addEventListener("click", function (e) {
      var t = e.target.closest("[data-consent],[data-signin],[data-share],[data-copy]");
      if (!t) return;
      if (t.dataset.consent) {
        set(CONSENT, t.dataset.consent);
        if (banner) banner.hidden = true;
        if (t.dataset.consent === "granted") loadGA();
      } else if (t.hasAttribute("data-signin")) {
        track("signin_click");
        track("check_start", { method: "signin" });
      } else if (t.hasAttribute("data-share")) {
        var r = document.querySelector("[data-report]");
        var url = location.origin + "/result?u=" + encodeURIComponent(r ? r.dataset.user : "");
        navigator.clipboard.writeText(url).then(function () { t.textContent = "Link copied"; });
        track("share_click");
      } else if (t.dataset.copy) {
        navigator.clipboard.writeText(t.dataset.copy).then(function () { t.textContent = "Copied"; });
        track("cli_copy");
      }
    });
    document.addEventListener("toggle", function (e) {
      var d = e.target;
      if (d.matches && d.matches("details[data-badge]") && d.open) track("badge_open", { badge: d.dataset.badge });
    }, true);
    document.addEventListener("error", function (e) {
      var img = e.target;
      if (img.tagName === "IMG" && img.classList.contains("badge-img")) {
        var tile = document.createElement("span");
        tile.className = "badge-img placeholder";
        tile.textContent = (img.alt || "?").slice(0, 2).toUpperCase();
        img.replaceWith(tile);
      }
    }, true);
    var form = document.getElementById("check-form");
    if (form) {
      var input = form.querySelector('input[name="u"]');
      if (input && !input.value) input.value = get(USER) || "";
      form.addEventListener("submit", function () {
        var v = input.value.trim();
        input.value = v;
        if (LOGIN.test(v)) set(USER, v);
        track("check_start", { method: "public" });
      });
    }
    reportView(document.querySelector("[data-report]"));
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", ready);
  else ready();
})();
