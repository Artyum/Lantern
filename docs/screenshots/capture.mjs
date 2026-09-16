import { spawn } from "node:child_process";
import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const OUT_DIR = dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = join(OUT_DIR, "../..");
const CONFIG_PATH = join(OUT_DIR, "config.json");
const CACHE_DIR = join(OUT_DIR, ".cache");
const PORT = 8099;
const CDP_PORT = 9229;
const URL = `http://127.0.0.1:${PORT}/`;
const THEMES = [
  { id: "morning-mist", label: "Morning Mist", light: true },
  { id: "sandy-dawn", label: "Sandy Dawn", light: true },
  { id: "deep-ocean", label: "Deep Ocean", light: false },
  { id: "charcoal-dusk", label: "Charcoal Dusk", light: false },
];

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitHttp(url, attempts = 60) {
  for (let i = 0; i < attempts; i++) {
    try {
      const res = await fetch(url);
      if (res.ok) return;
    } catch {}
    await sleep(250);
  }
  throw new Error(`Server did not start: ${url}`);
}

async function waitDevtools() {
  for (let i = 0; i < 50; i++) {
    try {
      const res = await fetch(`http://127.0.0.1:${CDP_PORT}/json/version`);
      if (res.ok) return true;
    } catch {}
    await sleep(200);
  }
  throw new Error("Chrome DevTools did not start");
}

async function pageWsUrl() {
  for (let i = 0; i < 30; i++) {
    const pages = await fetch(`http://127.0.0.1:${CDP_PORT}/json/list`).then((res) => res.json());
    const page = pages.find((entry) => entry.type === "page" && entry.webSocketDebuggerUrl);
    if (page) return page.webSocketDebuggerUrl;
    await sleep(200);
  }
  throw new Error("No page target");
}

function attach(ws) {
  let id = 0;
  const pending = new Map();
  ws.addEventListener("message", (event) => {
    const msg = JSON.parse(event.data);
    if (msg.id != null && pending.has(msg.id)) {
      const { resolve, reject } = pending.get(msg.id);
      pending.delete(msg.id);
      if (msg.error) reject(new Error(JSON.stringify(msg.error)));
      else resolve(msg.result);
    }
  });
  return (method, params = {}) =>
    new Promise((resolve, reject) => {
      const thisId = ++id;
      pending.set(thisId, { resolve, reject });
      ws.send(JSON.stringify({ id: thisId, method, params }));
    });
}

async function openPage(wsUrl) {
  const ws = new WebSocket(wsUrl);
  await new Promise((resolve, reject) => {
    ws.addEventListener("open", resolve, { once: true });
    ws.addEventListener("error", reject, { once: true });
  });
  const send = attach(ws);
  const loaded = new Promise((resolve) => {
    ws.addEventListener("message", (event) => {
      const msg = JSON.parse(event.data);
      if (msg.method === "Page.loadEventFired") resolve();
    });
  });
  await send("Page.enable");
  await send("Page.navigate", { url: URL });
  await Promise.race([loaded, sleep(8000)]);
  await send("Emulation.setDeviceMetricsOverride", {
    width: 1440,
    height: 900,
    deviceScaleFactor: 2,
    mobile: false,
  });
  await send("Runtime.evaluate", {
    expression: `(() => {
      try { localStorage.removeItem("lantern-theme"); } catch (e) {}
      return "ok";
    })()`,
  });
  await waitForIcons(send);
  return { ws, send };
}

async function waitForIcons(send) {
  for (let i = 0; i < 20; i++) {
    const { result } = await send("Runtime.evaluate", {
      expression: `(() => {
        const imgs = [...document.querySelectorAll(".tile img")];
        if (!imgs.length) return { total: 0, loaded: 0 };
        const loaded = imgs.filter((img) => img.complete && img.naturalWidth > 0).length;
        return { total: imgs.length, loaded };
      })()`,
      returnByValue: true,
    });
    if (result.value.total > 0 && result.value.loaded === result.value.total) return;
    await sleep(300);
  }
  await sleep(1200);
}

async function applyTheme(send, theme) {
  await send("Runtime.evaluate", {
    expression: `(() => {
      const id = ${JSON.stringify(theme.id)};
      const label = ${JSON.stringify(theme.label)};
      const light = ${theme.light};
      try { localStorage.setItem("lantern-theme", id); } catch (e) {}
      document.documentElement.setAttribute("data-theme", id);
      document.documentElement.setAttribute("data-scheme", light ? "light" : "dark");
      const swatch = document.getElementById("theme-swatch");
      const themeLabel = document.getElementById("theme-label");
      if (swatch) swatch.dataset.swatch = id;
      if (themeLabel) themeLabel.textContent = label;
      document.getElementById("theme-menu")?.querySelectorAll("[role='option']").forEach((option) => {
        option.setAttribute("aria-selected", option.dataset.theme === id ? "true" : "false");
      });
      document.documentElement.style.overflow = "hidden";
      document.body.style.overflow = "hidden";
      window.scrollTo(0, 0);
      return document.documentElement.getAttribute("data-theme");
    })()`,
    returnByValue: true,
  });
  await sleep(400);
}

async function clipRect(send) {
  const { result } = await send("Runtime.evaluate", {
    expression: `(() => {
      const shell = document.querySelector(".shell");
      const r = shell.getBoundingClientRect();
      const pad = 24;
      return {
        x: Math.max(0, r.x - pad),
        y: Math.max(0, r.y),
        width: Math.min(innerWidth, r.width + pad * 2),
        height: Math.min(innerHeight - 8, r.height + pad),
      };
    })()`,
    returnByValue: true,
  });
  return result.value;
}

const lantern = spawn(
  "go",
  ["run", "./cmd/lantern"],
  {
    cwd: REPO_ROOT,
    detached: true,
    env: {
      ...process.env,
      CONFIG_PATH,
      CACHE_DIR,
      LISTEN: `:${PORT}`,
    },
    stdio: ["ignore", "pipe", "pipe"],
  }
);

const chrome = spawn(
  "chromium",
  [
    "--headless=new",
    "--disable-gpu",
    "--no-sandbox",
    "--hide-scrollbars",
    "--force-device-scale-factor=2",
    `--remote-debugging-port=${CDP_PORT}`,
    "--window-size=1440,900",
    "about:blank",
  ],
  { detached: true, stdio: ["ignore", "ignore", "pipe"] }
);

function stop(child, signal = "SIGTERM") {
  if (!child || child.killed) return;
  try {
    process.kill(-child.pid, signal);
  } catch {
    child.kill(signal);
  }
}

try {
  await mkdir(OUT_DIR, { recursive: true });
  await mkdir(CACHE_DIR, { recursive: true });
  await waitHttp(URL);
  await waitDevtools();
  const { ws, send } = await openPage(await pageWsUrl());

  for (const theme of THEMES) {
    await applyTheme(send, theme);
    const clip = await clipRect(send);
    const shot = await send("Page.captureScreenshot", {
      format: "png",
      fromSurface: true,
      captureBeyondViewport: false,
      clip: { ...clip, scale: 1 },
    });
    const dest = join(OUT_DIR, `${theme.id}.png`);
    await writeFile(dest, Buffer.from(shot.data, "base64"));
    console.log("wrote", dest, `${Math.round(clip.width)}x${Math.round(clip.height)}`);
  }

  ws.close();
} finally {
  stop(lantern);
  stop(chrome);
}
