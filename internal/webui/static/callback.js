// internal/webui/static/callback.js
// GitHub sends the browser here with ?code&state. Drop them from the address bar,
// POST them to /api/private (the state cookie rides along), show the result or one message.
(function () {
  "use strict";
  var p = new URLSearchParams(location.search);
  var code = p.get("code"), state = p.get("state"), denied = p.get("error") === "access_denied";
  history.replaceState(null, "", location.pathname);
  function show(id) {
    document.getElementById("status").hidden = true;
    var el = document.getElementById("error-" + id) || document.getElementById("error-internal");
    el.hidden = false;
    if (window.gha) window.gha.track("check_error", { reason: id });
  }
  function run() {
    if (denied) return show("denied");
    if (!code || !state) return show("state");
    fetch("/api/private", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ code: code, state: state })
    }).then(function (r) {
      if (!r.ok) return show(r.headers.get("X-Error-Reason") || "internal");
      return r.text().then(function (html) {
        document.getElementById("status").hidden = true;
        var out = document.getElementById("result");
        out.innerHTML = html; // server-rendered by html/template on our own origin
        if (window.gha) {
          window.gha.track("signin_complete");
          window.gha.reportView(out.querySelector("[data-report]"));
        }
      });
    }).catch(function () { show("network"); });
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", run);
  else run();
})();
