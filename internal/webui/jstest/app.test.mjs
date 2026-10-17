// Behaviour tests for internal/webui/static/theme.js and app.js.
// The real scripts run in a node:vm context against a tiny fake DOM written here, so only
// what these scripts touch exists. Run: node --test "internal/webui/jstest/*.test.mjs"
import { test } from "node:test";
import assert from "node:assert/strict";
import vm from "node:vm";
import fs from "node:fs";

const staticDir = new URL("../static/", import.meta.url);
const THEME_JS = fs.readFileSync(new URL("theme.js", staticDir), "utf8");
const APP_JS = fs.readFileSync(new URL("app.js", staticDir), "utf8");

// ---- fake DOM -------------------------------------------------------------------------

// matches supports what the scripts select with: tag, .class, [attr], [attr="v"], and commas.
function matches(el, selector) {
  return selector.split(",").some((part) => {
    part = part.trim();
    const tag = /^[a-z]+/.exec(part);
    if (tag && tag[0] !== el.tag) return false;
    const rest = tag ? part.slice(tag[0].length) : part;
    const tokens = rest.match(/\.[\w-]+|\[[\w-]+(?:="[^"]*")?\]/g) || [];
    return tokens.every((tok) => {
      if (tok[0] === ".") return el.classList.contains(tok.slice(1));
      const m = /^\[([\w-]+)(?:="([^"]*)")?\]$/.exec(tok);
      return el.hasAttribute(m[1]) && (m[2] === undefined || el.getAttribute(m[1]) === m[2]);
    });
  });
}

function camel(name) {
  return name.replace(/^data-/, "").replace(/-(\w)/g, (_, c) => c.toUpperCase());
}

class FakeEl {
  constructor(doc, tag, attrs = {}, children = []) {
    this.doc = doc;
    this.tag = tag;
    this.attrs = new Map();
    this.dataset = {};
    this.children = children;
    this.parent = null;
    children.forEach((c) => { c.parent = this; });
    this.hidden = false;
    this.textContent = "";
    const classes = new Set();
    this.classList = {
      add: (c) => classes.add(c),
      contains: (c) => classes.has(c),
    };
    for (const [k, v] of Object.entries(attrs)) {
      if (k === "class") v.split(" ").forEach((c) => classes.add(c));
      else this.setAttribute(k, v);
    }
    doc.all.push(this); // children registered themselves when they were built
  }
  setAttribute(k, v) {
    this.attrs.set(k, String(v));
    if (k.startsWith("data-")) this.dataset[camel(k)] = String(v);
  }
  getAttribute(k) { return this.attrs.has(k) ? this.attrs.get(k) : null; }
  hasAttribute(k) { return this.attrs.has(k); }
  removeAttribute(k) {
    this.attrs.delete(k);
    if (k.startsWith("data-")) delete this.dataset[camel(k)];
  }
  matches(sel) { return matches(this, sel); }
  // closest walks up through the parents, as real event delegation relies on.
  closest(sel) {
    for (let el = this; el; el = el.parent) if (matches(el, sel)) return el;
    return null;
  }
  querySelector(sel) {
    for (const c of this.children) {
      if (matches(c, sel)) return c;
      const deep = c.querySelector(sel);
      if (deep) return deep;
    }
    return null;
  }
  focus() { this.doc.activeElement = this; }
}

class FakeDocument {
  constructor() {
    this.all = [];
    this.listeners = {};
    this.readyState = "complete";
    this.currentScript = { dataset: {} };
    this.referrer = "";
    this.images = [];
    this.title = "Test page";
    this.activeElement = null;
    this.documentElement = new FakeEl(this, "html");
    this.head = { appendChild() {} };
  }
  addEventListener(type, fn) { (this.listeners[type] ||= []).push(fn); }
  dispatch(type, ev) { (this.listeners[type] || []).forEach((fn) => fn(ev)); }
  click(target) { this.dispatch("click", { target }); }
  key(key) { this.dispatch("keydown", { key }); }
  getElementById(id) { return this.all.find((e) => e.getAttribute("id") === id) || null; }
  querySelector(sel) { return this.all.find((e) => matches(e, sel)) || null; }
  querySelectorAll(sel) { return this.all.filter((e) => matches(e, sel)); }
  createElement(tag) { return new FakeEl(this, tag); }
}

function fakeStorage(initial = {}, { throws = false } = {}) {
  const m = new Map(Object.entries(initial));
  const fail = () => { throw new Error("storage blocked"); };
  return {
    getItem: (k) => (throws ? fail() : m.has(k) ? m.get(k) : null),
    setItem: (k, v) => (throws ? fail() : void m.set(k, String(v))),
    removeItem: (k) => (throws ? fail() : void m.delete(k)),
  };
}

// A page shaped like the site header: theme button, menu button, nav, star button.
function buildPage() {
  const doc = new FakeDocument();
  const themeBtn = new FakeEl(doc, "button", { "data-theme-toggle": "" });
  themeBtn.hidden = true;
  const menuBtn = new FakeEl(doc, "button", { "data-menu": "", "aria-expanded": "false" });
  const link1 = new FakeEl(doc, "a", { href: "/achievements/" });
  const link2 = new FakeEl(doc, "a", { href: "/about/" });
  const nav = new FakeEl(doc, "nav", { id: "site-nav" }, [link1, link2]);
  const starCount = new FakeEl(doc, "span", { "data-stars": "" });
  starCount.hidden = true;
  const starLong = new FakeEl(doc, "span", { class: "star-long" });
  const glyph = new FakeEl(doc, "span", { "aria-hidden": "true" });
  const starBtn = new FakeEl(doc, "a", { "data-star": "header" }, [glyph, starCount, starLong]);
  // The result page's banner: the target of a click is the inner star span.
  const ctaStar = new FakeEl(doc, "span", { "aria-hidden": "true" });
  const bannerCount = new FakeEl(doc, "span", { "data-stars": "" });
  bannerCount.hidden = true;
  const cta = new FakeEl(doc, "span", { class: "star-cta" }, [ctaStar, bannerCount]);
  new FakeEl(doc, "a", { class: "star-banner", "data-star": "result" }, [cta]);
  return { doc, themeBtn, menuBtn, nav, link1, starBtn, starCount, starLong, glyph, ctaStar, bannerCount };
}

function sandbox(doc, { local, session, fetchImpl, gtag, search = "" }) {
  const win = {
    document: doc,
    localStorage: local,
    sessionStorage: session,
    location: { origin: "https://example.test", pathname: "/", search },
    navigator: {},
    matchMedia: () => ({ matches: false, addEventListener() {} }),
    URL,
    Array,
    Number,
    String,
  };
  if (fetchImpl) win.fetch = fetchImpl;
  if (gtag) win.gtag = gtag;
  win.window = win;
  // The scripts read bare globals (localStorage, fetch, ...) as well as window.*.
  return vm.createContext(Object.assign(win, { document: doc }));
}

const settle = () => new Promise((r) => setImmediate(r));

// fakeFetch records every call and answers with the given response (or rejection).
function fakeFetch(respond) {
  const calls = [];
  const fn = (url, opts) => {
    calls.push({ url, credentials: opts && opts.credentials, referrerPolicy: opts && opts.referrerPolicy });
    return respond();
  };
  fn.calls = calls;
  return fn;
}
const json = (body, ok = true) => () => Promise.resolve({ ok, json: () => Promise.resolve(body) });

// load runs a script file in a fresh context and returns the page it ran against.
function loadApp(opts = {}) {
  const page = buildPage();
  const ctx = sandbox(page.doc, {
    local: opts.local || fakeStorage(),
    session: opts.session || fakeStorage(),
    fetchImpl: opts.fetch,
    gtag: opts.gtag,
    search: opts.search,
  });
  vm.runInContext(APP_JS, ctx, { filename: "app.js" });
  return page;
}

// ---- theme.js -------------------------------------------------------------------------

function runTheme(local) {
  const doc = new FakeDocument();
  const ctx = sandbox(doc, { local, session: fakeStorage() });
  vm.runInContext(THEME_JS, ctx, { filename: "theme.js" });
  return doc.documentElement;
}

test("theme.js: a stored dark choice sets data-theme and adds the js class", () => {
  const root = runTheme(fakeStorage({ "gha:theme": "dark" }));
  assert.equal(root.getAttribute("data-theme"), "dark");
  assert.ok(root.classList.contains("js"));
});

test("theme.js: a stored light choice sets data-theme=light", () => {
  assert.equal(runTheme(fakeStorage({ "gha:theme": "light" })).getAttribute("data-theme"), "light");
});

test("theme.js: a junk stored value sets nothing but still adds the js class", () => {
  const root = runTheme(fakeStorage({ "gha:theme": "<script>" }));
  assert.equal(root.getAttribute("data-theme"), null);
  assert.ok(root.classList.contains("js"));
});

test("theme.js: throwing storage does not break it", () => {
  const root = runTheme(fakeStorage({}, { throws: true }));
  assert.equal(root.getAttribute("data-theme"), null);
  assert.ok(root.classList.contains("js"));
});

// ---- app.js: theme button -------------------------------------------------------------

test("app.js: the theme button flips the theme, stores it and relabels itself", () => {
  const local = fakeStorage();
  const { doc, themeBtn } = loadApp({ local });
  // System theme is light, so the button offers dark.
  assert.equal(themeBtn.hidden, false);
  assert.equal(themeBtn.textContent, "Dark");
  assert.equal(themeBtn.getAttribute("aria-label"), "Switch to dark theme");

  doc.click(themeBtn);
  assert.equal(doc.documentElement.getAttribute("data-theme"), "dark");
  assert.equal(local.getItem("gha:theme"), "dark");
  assert.equal(themeBtn.textContent, "Light");
  assert.equal(themeBtn.getAttribute("aria-label"), "Switch to light theme");

  doc.click(themeBtn);
  assert.equal(doc.documentElement.getAttribute("data-theme"), "light");
  assert.equal(local.getItem("gha:theme"), "light");
  assert.equal(themeBtn.textContent, "Dark");
});

// ---- app.js: mobile menu --------------------------------------------------------------

test("app.js: the menu button opens the nav and focuses its first link; Escape closes it", () => {
  const { doc, menuBtn, nav, link1 } = loadApp();
  doc.click(menuBtn);
  assert.equal(menuBtn.getAttribute("aria-expanded"), "true");
  assert.ok(nav.hasAttribute("data-open"));
  assert.equal(doc.activeElement, link1);

  doc.key("Escape");
  assert.equal(nav.hasAttribute("data-open"), false);
  assert.equal(menuBtn.getAttribute("aria-expanded"), "false");
  assert.equal(doc.activeElement, menuBtn);
});

test("app.js: a second click on the menu button closes it; Escape on a closed menu does nothing", () => {
  const { doc, menuBtn, nav, link1 } = loadApp();
  doc.key("Escape");
  assert.notEqual(doc.activeElement, menuBtn);
  doc.click(menuBtn);
  doc.click(menuBtn);
  assert.equal(nav.hasAttribute("data-open"), false);
  assert.equal(menuBtn.getAttribute("aria-expanded"), "false");
  assert.equal(doc.activeElement, link1);
});

// ---- app.js: star count ---------------------------------------------------------------

test("app.js: a star count fills [data-stars], unhides it and hides the long label", async () => {
  const session = fakeStorage();
  const fetchImpl = fakeFetch(json({ stargazers_count: 128 }));
  const { starCount, starLong } = loadApp({ session, fetch: fetchImpl });
  await settle();
  assert.equal(fetchImpl.calls.length, 1);
  assert.equal(fetchImpl.calls[0].url, "https://api.github.com/repos/nmloc/gh-achievements");
  assert.equal(fetchImpl.calls[0].credentials, "omit");
  assert.equal(fetchImpl.calls[0].referrerPolicy, "no-referrer");
  assert.equal(starCount.textContent, "128");
  assert.equal(starCount.hidden, false);
  assert.equal(starLong.hidden, true);
  assert.equal(session.getItem("gha:stars"), "128");
});

test("app.js: a count of 0 leaves the count hidden and the long label showing", async () => {
  const { starCount, starLong } = loadApp({ fetch: fakeFetch(json({ stargazers_count: 0 })) });
  await settle();
  assert.equal(starCount.hidden, true);
  assert.equal(starCount.textContent, "");
  assert.equal(starLong.hidden, false);
});

test("app.js: a cached count is used without fetching", async () => {
  const fetchImpl = fakeFetch(json({ stargazers_count: 5 }));
  const { starCount } = loadApp({ session: fakeStorage({ "gha:stars": "42" }), fetch: fetchImpl });
  await settle();
  assert.equal(fetchImpl.calls.length, 0);
  assert.equal(starCount.textContent, "42");
});

const failures = {
  "a network error": () => Promise.reject(new Error("offline")),
  "a rate limit (403)": json({ message: "API rate limit exceeded" }, false),
  "a body without stargazers_count": json({ message: "ok but no count" }),
  "a 200 whose body is not JSON": () => Promise.resolve({ ok: true, json: () => Promise.reject(new SyntaxError("bad json")) }),
};
for (const [name, respond] of Object.entries(failures)) {
  test(`app.js: ${name} is remembered for the visit, so the next page does not fetch`, async () => {
    const session = fakeStorage();
    const first = fakeFetch(respond);
    const page1 = loadApp({ session, fetch: first });
    await settle();
    assert.equal(first.calls.length, 1);
    assert.equal(session.getItem("gha:stars"), "0");
    assert.equal(page1.starCount.hidden, true);

    // Next page of the same visit: fresh context, same sessionStorage.
    const second = fakeFetch(json({ stargazers_count: 999 }));
    const page2 = loadApp({ session, fetch: second });
    await settle();
    assert.equal(second.calls.length, 0);
    assert.equal(page2.starCount.hidden, true);
    assert.equal(page2.starLong.hidden, false);
  });
}

test("app.js: an error while filling does not overwrite the stored count with 0", async () => {
  const session = fakeStorage();
  const page = buildPage();
  Object.defineProperty(page.starCount, "hidden", { get: () => true, set() { throw new Error("DOM refused"); } });
  const ctx = sandbox(page.doc, { local: fakeStorage(), session, fetchImpl: fakeFetch(json({ stargazers_count: 128 })) });
  vm.runInContext(APP_JS, ctx, { filename: "app.js" });
  await settle();
  assert.equal(session.getItem("gha:stars"), "128");
});

// ---- app.js: star_click ---------------------------------------------------------------

test("app.js: star_click sends only the place, with consent granted", () => {
  const calls = [];
  const gtag = (...args) => calls.push(args);
  const { doc, starBtn } = loadApp({
    local: fakeStorage({ "gha:consent": "granted" }),
    gtag,
    search: "?u=octocat",
  });
  doc.click(starBtn);
  assert.equal(JSON.stringify(calls), JSON.stringify([["event", "star_click", { place: "header" }]]));
  assert.ok(!JSON.stringify(calls).includes("octocat"));
});

// Users click the children of the link, not the link itself: delegation must climb to it.
const starTargets = {
  "the star glyph in the header link": ["glyph", "header"],
  "the count inside the header link": ["starCount", "header"],
  "the long label inside the header link": ["starLong", "header"],
  "the star inside the banner's call to action": ["ctaStar", "result"],
  "the count inside the banner": ["bannerCount", "result"],
};
for (const [name, [key, place]] of Object.entries(starTargets)) {
  test(`app.js: clicking ${name} sends star_click {place: ${place}}`, () => {
    const calls = [];
    const page = loadApp({ local: fakeStorage({ "gha:consent": "granted" }), gtag: (...a) => calls.push(a) });
    page.doc.click(page[key]);
    assert.equal(JSON.stringify(calls), JSON.stringify([["event", "star_click", { place }]]));
  });
}

test("app.js: a click outside every tracked element sends nothing", () => {
  const calls = [];
  const page = loadApp({ local: fakeStorage({ "gha:consent": "granted" }), gtag: (...a) => calls.push(a) });
  page.doc.click(page.link1);
  assert.equal(calls.length, 0);
});

test("app.js: star_click sends nothing without consent", () => {
  const calls = [];
  const { doc, starBtn } = loadApp({ gtag: (...args) => calls.push(args) });
  doc.click(starBtn);
  assert.equal(calls.length, 0);
});
