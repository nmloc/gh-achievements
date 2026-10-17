#!/usr/bin/env node
// Capture the README screenshots of the result page, light and dark.
// Needs Google Chrome and the dev server: `go run ./cmd/devserver` (http://localhost:8787).
//   node scripts/readme-screenshots.mjs [url]
import { spawn } from "node:child_process";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const url = process.argv[2] || "http://localhost:8787/result?u=torvalds";
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const port = 9333;
const profile = mkdtempSync(join(tmpdir(), "gha-shot-"));
const proc = spawn(chrome, ["--headless=new", `--remote-debugging-port=${port}`, `--user-data-dir=${profile}`,
  "--hide-scrollbars", "--no-first-run", "about:blank"], { stdio: "ignore" });

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function target() {
  for (let i = 0; i < 50; i++) {
    try {
      const list = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
      const page = list.find((t) => t.type === "page");
      if (page) return page.webSocketDebuggerUrl;
    } catch { /* Chrome still starting */ }
    await sleep(200);
  }
  throw new Error("Chrome did not start");
}

const ws = new WebSocket(await target());
await new Promise((r) => ws.addEventListener("open", r, { once: true }));
let seq = 0;
const pending = new Map();
ws.addEventListener("message", (e) => {
  const msg = JSON.parse(e.data);
  if (msg.id && pending.has(msg.id)) {
    const { resolve, reject } = pending.get(msg.id);
    pending.delete(msg.id);
    msg.error ? reject(new Error(msg.error.message)) : resolve(msg.result);
  }
});
const send = (method, params = {}) => new Promise((resolve, reject) => {
  const id = ++seq;
  pending.set(id, { resolve, reject });
  ws.send(JSON.stringify({ id, method, params }));
});
const evaluate = (expression) => send("Runtime.evaluate", { expression, awaitPromise: true, returnByValue: true });

const width = 1280, height = 1100;
await send("Page.enable");
await send("Emulation.setDeviceMetricsOverride", { width, height, deviceScaleFactor: 1.5, mobile: false });
await send("Page.navigate", { url: new URL("/", url).href });
await sleep(1500);
for (const theme of ["light", "dark"]) {
  // No consent banner, no analytics, and the theme under test.
  await evaluate(`localStorage.setItem("gha:consent", "denied"); localStorage.setItem("gha:theme", "${theme}"); sessionStorage.removeItem("gha:stars"); true`);
  await send("Page.navigate", { url });
  await sleep(1500);
  // Wait for fonts and images, but never forever: an image that is lazy or blocked would stall the run.
  await evaluate(`Promise.race([Promise.all([document.fonts.ready, ...[...document.images].map((i) => i.complete ? 0 : new Promise((r) => { i.onload = i.onerror = r; }))]), new Promise((r) => setTimeout(r, 10000))]).then(() => true)`);
  await sleep(300);
  const { data } = await send("Page.captureScreenshot", { format: "png", clip: { x: 0, y: 0, width, height, scale: 1 } });
  const out = `docs/screenshot-${theme}.png`;
  writeFileSync(out, Buffer.from(data, "base64"));
  console.log("wrote", out);
}
ws.close();
const closed = new Promise((r) => proc.once("exit", r));
proc.kill();
await Promise.race([closed, sleep(5000)]);
rmSync(profile, { recursive: true, force: true, maxRetries: 5, retryDelay: 300 });
