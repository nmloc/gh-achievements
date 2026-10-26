// Checks the Markdown route in vercel.json with a lookahead-capable regex engine.
// Vercel matches route sources as PCRE; Go's regexp has no lookahead, so the Go tests
// cannot run this pattern. JavaScript's RegExp agrees with PCRE on everything it uses.
import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const cfg = JSON.parse(fs.readFileSync(new URL("../../../vercel.json", import.meta.url), "utf8"));
const route = cfg.routes.find((r) => r.dest && r.dest.startsWith("/api/markdown"));

test("the Markdown route sends page paths to the function, keeping the path", () => {
  const src = new RegExp(route.src);
  for (const [path, captured] of [
    ["/", ""],
    ["/achievements/", "achievements/"],
    ["/achievements/pull-shark/", "achievements/pull-shark/"],
    ["/about", "about"],
    ["/__ora-404-probe-79rn61ru", "__ora-404-probe-79rn61ru"],
    ["/results", "results"],
  ]) {
    const m = src.exec(path);
    assert.ok(m, `${path} should match`);
    assert.equal(route.dest.replace("$1", m[1]), "/api/markdown?path=" + captured);
  }
});

test("the Markdown route leaves files and the functions alone", () => {
  const src = new RegExp(route.src);
  for (const path of ["/llms.txt", "/achievements.json", "/sitemap.xml", "/style.css", "/index.html",
    "/result", "/result/", "/api/markdown", "/api/result", "/auth/login", "/auth/callback/"]) {
    assert.equal(src.test(path), false, `${path} should not match`);
  }
});

test("the Markdown route only fires for Accept: text/markdown", () => {
  const [has] = route.has;
  assert.equal(has.type, "header");
  assert.equal(has.key.toLowerCase(), "accept");
  const re = new RegExp("^" + has.value + "$");
  assert.ok(re.test("text/markdown"));
  assert.ok(re.test("text/markdown, text/html;q=0.9"));
  assert.ok(re.test("text/html, text/markdown;q=0.5"));
  assert.equal(re.test("text/html,application/xhtml+xml,*/*;q=0.8"), false);
});
