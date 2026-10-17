// internal/webui/static/theme.js
// Runs in <head> before the stylesheet, so the first paint already uses the visitor's
// chosen theme. app.js owns the toggle button; this only applies a stored choice.
(function () {
  "use strict";
  var root = document.documentElement;
  root.classList.add("js");
  try {
    var t = localStorage.getItem("gha:theme");
    if (t === "light" || t === "dark") root.setAttribute("data-theme", t);
  } catch (e) { /* storage blocked: follow the system setting */ }
})();
