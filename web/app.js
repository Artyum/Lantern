const board = document.getElementById("board");
const empty = document.getElementById("empty");
const titleEl = document.getElementById("title");
const q = document.getElementById("q");
const themePicker = document.getElementById("theme-picker");
const themeToggle = document.getElementById("theme-toggle");
const themeMenu = document.getElementById("theme-menu");
const themeLabel = document.getElementById("theme-label");
const themeSwatch = document.getElementById("theme-swatch");
const addSectionBtn = document.getElementById("add-section");
const sectionNav = document.getElementById("section-nav");
const sectionNavList = document.getElementById("section-nav-list");
const sectionNavToggle = document.getElementById("section-nav-toggle");
const widthPicker = document.getElementById("width-picker");
const widthToggle = document.getElementById("width-toggle");
const widthPanel = document.getElementById("width-panel");
const widthClose = document.getElementById("width-close");
const contentWidthInput = document.getElementById("content-width");
const contentWidthLabel = document.getElementById("content-width-label");

const THEME_KEY = "lantern-theme";
const COLLAPSED_KEY = "lantern-collapsed";
const SIDEBAR_KEY = "lantern-sidebar-collapsed";
const CONTENT_WIDTH_KEY = "lantern-content-width";
const THEME_LIST = [
  { id: "morning-mist", label: "Morning Mist", light: true },
  { id: "sandy-dawn", label: "Sandy Dawn", light: true },
  { id: "green-meadow", label: "Green Meadow", light: true },
  { id: "graphite-day", label: "Graphite Day", light: true },
  { id: "berry-dawn", label: "Berry Dawn", light: true },
  { id: "rose-dawn", label: "Rose Dawn", light: true },
  { id: "deep-ocean", label: "Deep Ocean", light: false },
  { id: "amber-night", label: "Amber Night", light: false },
  { id: "northern-spruce", label: "Northern Spruce", light: false },
  { id: "charcoal-dusk", label: "Charcoal Dusk", light: false },
  { id: "midnight-berry", label: "Midnight Berry", light: false },
  { id: "rose-night", label: "Rose Night", light: false },
];
const THEMES = new Set(THEME_LIST.map((entry) => entry.id));
const THEME_ALIASES = {
  dark: "deep-ocean",
  light: "morning-mist",
  lantern: "amber-night",
  parchment: "sandy-dawn",
  polar: "morning-mist",
  sakura: "rose-dawn",
  amethyst: "midnight-berry",
  graphite: "charcoal-dusk",
  moss: "green-meadow",
  copper: "sandy-dawn",
  ocean: "deep-ocean",
  wine: "rose-night",
  latarnia: "amber-night",
  pergamin: "sandy-dawn",
  ametyst: "midnight-berry",
  grafit: "charcoal-dusk",
  mech: "green-meadow",
  miedz: "sandy-dawn",
  wino: "rose-night",
};
const ICON_UP = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 14l6-6 6 6"/></svg>';
const ICON_DOWN = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 10l6 6 6-6"/></svg>';
const ICON_MORE = '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><circle cx="12" cy="5" r="1.7"/><circle cx="12" cy="12" r="1.7"/><circle cx="12" cy="19" r="1.7"/></svg>';
const ICON_GRIP = '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><circle cx="9" cy="7" r="1.4"/><circle cx="15" cy="7" r="1.4"/><circle cx="9" cy="12" r="1.4"/><circle cx="15" cy="12" r="1.4"/><circle cx="9" cy="17" r="1.4"/><circle cx="15" cy="17" r="1.4"/></svg>';
const ICON_TRASH = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 7h14M10 7V5h4v2M8 7l1 12h6l1-12"/></svg>';
const ICON_PLUS = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14M5 12h14"/></svg>';
const ICON_CHEVRON = '<svg class="section-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 10l6 6 6-6"/></svg>';

const editor = document.getElementById("item-editor");
const itemForm = document.getElementById("item-form");
const itemName = document.getElementById("item-name");
const itemUrl = document.getElementById("item-url");
const itemIcon = document.getElementById("item-icon");
const itemIconFile = document.getElementById("item-icon-file");
const itemIconPreview = document.getElementById("item-icon-preview");
const itemError = document.getElementById("item-error");
const itemTitle = document.getElementById("item-title");
const itemCancel = document.getElementById("item-cancel");
const itemSave = document.getElementById("item-save");
const itemDelete = document.getElementById("item-delete");
const itemConfirm = document.getElementById("item-confirm");
const itemConfirmForm = document.getElementById("item-confirm-form");
const itemConfirmText = document.getElementById("item-confirm-text");
const itemConfirmError = document.getElementById("item-confirm-error");
const itemConfirmCancel = document.getElementById("item-confirm-cancel");
const itemConfirmOk = document.getElementById("item-confirm-ok");
const sectionEditor = document.getElementById("section-editor");
const sectionForm = document.getElementById("section-form");
const sectionNameInput = document.getElementById("section-name");
const sectionError = document.getElementById("section-error");
const sectionCancel = document.getElementById("section-cancel");
const sectionSave = document.getElementById("section-save");
const sectionConfirm = document.getElementById("section-confirm");
const sectionConfirmForm = document.getElementById("section-confirm-form");
const sectionConfirmText = document.getElementById("section-confirm-text");
const sectionConfirmError = document.getElementById("section-confirm-error");
const sectionConfirmCancel = document.getElementById("section-confirm-cancel");
const sectionConfirmOk = document.getElementById("section-confirm-ok");

let editing = null;
let previewObjectUrl = "";
let deletingSection = null;
let deletingItem = null;

function themeMeta(id) {
  return THEME_LIST.find((entry) => entry.id === id) || THEME_LIST[0];
}

function applyTheme(id) {
  const meta = themeMeta(id);
  document.documentElement.setAttribute("data-theme", id);
  document.documentElement.setAttribute("data-scheme", meta.light ? "light" : "dark");
  themeSwatch.dataset.swatch = id;
  themeLabel.textContent = meta.label;
  for (const option of themeMenu.querySelectorAll("[role='option']")) {
    const selected = option.dataset.theme === id;
    option.setAttribute("aria-selected", selected ? "true" : "false");
  }
}

function preferredTheme() {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "deep-ocean" : "morning-mist";
}

function resolveTheme(value) {
  const mapped = THEME_ALIASES[value] || value;
  return THEMES.has(mapped) ? mapped : preferredTheme();
}

function setTheme(id) {
  const mode = resolveTheme(id);
  localStorage.setItem(THEME_KEY, mode);
  applyTheme(mode);
  closeThemeMenu();
}

function themeMenuOpen() {
  return themeToggle.getAttribute("aria-expanded") === "true";
}

function openThemeMenu() {
  closeWidthPanel();
  themeMenu.hidden = false;
  themeToggle.setAttribute("aria-expanded", "true");
  themePicker.classList.add("open");
  const selected = themeMenu.querySelector('[aria-selected="true"]');
  (selected || themeMenu.querySelector("[role='option']"))?.focus();
}

function closeThemeMenu() {
  themeMenu.hidden = true;
  themeToggle.setAttribute("aria-expanded", "false");
  themePicker.classList.remove("open");
}

function initThemeMenu() {
  for (const entry of THEME_LIST) {
    const item = document.createElement("li");
    const option = document.createElement("button");
    option.type = "button";
    option.setAttribute("role", "option");
    option.className = "theme-option";
    option.dataset.theme = entry.id;
    option.setAttribute("aria-selected", "false");
    const swatch = document.createElement("span");
    swatch.className = "theme-swatch";
    swatch.dataset.swatch = entry.id;
    swatch.setAttribute("aria-hidden", "true");
    const name = document.createElement("span");
    name.textContent = entry.label;
    option.append(swatch, name);
    option.addEventListener("click", () => setTheme(entry.id));
    item.append(option);
    themeMenu.append(item);
  }
}

function initTheme() {
  const mode = resolveTheme(localStorage.getItem(THEME_KEY));
  initThemeMenu();
  applyTheme(mode);
  return mode;
}

initTheme();
themeToggle.addEventListener("click", () => {
  if (themeMenuOpen()) closeThemeMenu();
  else openThemeMenu();
});
themeToggle.addEventListener("keydown", (event) => {
  if (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    openThemeMenu();
  }
});
themeMenu.addEventListener("keydown", (event) => {
  const options = [...themeMenu.querySelectorAll(".theme-option")];
  const index = options.indexOf(document.activeElement);
  if (event.key === "Escape") {
    event.preventDefault();
    closeThemeMenu();
    themeToggle.focus();
    return;
  }
  if (event.key === "ArrowDown") {
    event.preventDefault();
    options[(index + 1) % options.length]?.focus();
  }
  if (event.key === "ArrowUp") {
    event.preventDefault();
    options[(index - 1 + options.length) % options.length]?.focus();
  }
});
document.addEventListener("click", (event) => {
  if (!themePicker.contains(event.target)) closeThemeMenu();
  if (!widthPicker.contains(event.target) && !widthPanel.contains(event.target)) closeWidthPanel();
});

let data = { title: "Lantern", sections: [] };
let didDrag = false;
let dragTile = null;
let navObserver = null;
let activeNavSection = "";
let navScrollTarget = "";
let navScrollTimer = 0;
let contentWidth = 1440;
const SIDEBAR_EXPANDED_WIDTH = 220;
const SIDEBAR_COLLAPSED_WIDTH = 52;
const SHELL_CHROME_WIDTH = 56;
const WORKSPACE_GAP = 20;

function isSidebarCollapsed() {
  return localStorage.getItem(SIDEBAR_KEY) === "1";
}

function setSidebarCollapsed(collapsed) {
  sectionNav.classList.toggle("collapsed", collapsed);
  sectionNavToggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
  sectionNavToggle.title = collapsed ? "Expand sidebar" : "Collapse sidebar";
  localStorage.setItem(SIDEBAR_KEY, collapsed ? "1" : "0");
  document.documentElement.style.setProperty(
    "--sidebar-current-width",
    `${collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_EXPANDED_WIDTH}px`,
  );
}

function contentWidthLimit() {
  return Math.max(
    800,
    Math.floor(window.innerWidth - SHELL_CHROME_WIDTH - SIDEBAR_EXPANDED_WIDTH - WORKSPACE_GAP),
  );
}

function clampContentWidth(value) {
  const num = Number(value);
  const max = contentWidthLimit();
  if (!Number.isFinite(num)) return max;
  return Math.min(max, Math.max(800, Math.round(num / 10) * 10));
}

function previewContentWidth(value) {
  const width = clampContentWidth(value);
  if (contentWidthInput) {
    contentWidthInput.max = String(contentWidthLimit());
    contentWidthInput.setAttribute("aria-valuemax", String(contentWidthLimit()));
    contentWidthInput.setAttribute("aria-valuenow", String(width));
    contentWidthInput.setAttribute("aria-valuetext", `${width} px`);
  }
  if (contentWidthLabel) contentWidthLabel.textContent = `${width} px`;
}

function commitContentWidth(value, persist = true) {
  const width = clampContentWidth(value);
  contentWidth = width;
  if (contentWidthInput) contentWidthInput.value = String(width);
  previewContentWidth(width);
  document.documentElement.style.setProperty("--main-width", `${width}px`);
  if (persist) localStorage.setItem(CONTENT_WIDTH_KEY, String(width));
}

function initContentWidth() {
  const stored = localStorage.getItem(CONTENT_WIDTH_KEY);
  commitContentWidth(stored ?? contentWidthLimit(), false);
}

function widthPanelOpen() {
  return widthToggle.getAttribute("aria-expanded") === "true";
}

function openWidthPanel() {
  closeThemeMenu();
  previewContentWidth(contentWidth);
  if (contentWidthInput) contentWidthInput.value = String(contentWidth);
  widthPanel.hidden = false;
  widthToggle.setAttribute("aria-expanded", "true");
  widthPicker.classList.add("open");
  contentWidthInput?.focus();
}

function closeWidthPanel() {
  widthPanel.hidden = true;
  widthToggle.setAttribute("aria-expanded", "false");
  widthPicker.classList.remove("open");
  previewContentWidth(contentWidth);
  if (contentWidthInput) contentWidthInput.value = String(contentWidth);
}

function initSidebar() {
  setSidebarCollapsed(isSidebarCollapsed());
}

sectionNavToggle.addEventListener("click", () => {
  setSidebarCollapsed(!isSidebarCollapsed());
});

window.addEventListener("resize", () => {
  const width = clampContentWidth(contentWidth);
  if (width !== contentWidth) commitContentWidth(width, false);
  else previewContentWidth(contentWidth);
}, { passive: true });

function sectionElement(name) {
  return [...board.querySelectorAll(".section")].find(
    (el) => el.dataset.sectionName === name,
  );
}

function beginNavScroll(name) {
  navScrollTarget = name;
  window.clearTimeout(navScrollTimer);
  setActiveNavSection(name);
}

function endNavScroll() {
  if (!navScrollTarget) return;
  navScrollTarget = "";
  window.clearTimeout(navScrollTimer);
}

function scrollToSection(name) {
  const wrap = sectionElement(name);
  if (!wrap) return;
  beginNavScroll(name);
  if (isCollapsed(name)) toggleCollapsed(name);
  wrap.scrollIntoView({ behavior: "smooth", block: "start" });
  navScrollTimer = window.setTimeout(endNavScroll, 1400);
}

function setActiveNavSection(name) {
  if (name === activeNavSection) return;
  activeNavSection = name;
  for (const item of sectionNavList.querySelectorAll(".nav-item")) {
    const current = item.dataset.sectionName === name;
    item.setAttribute("aria-current", current ? "true" : "false");
  }
}

function renderNav(query) {
  sectionNavList.replaceChildren();
  let shown = 0;
  for (const section of data.sections || []) {
    const items = (section.items || []).filter((item) => matches(item, query));
    if (!items.length && query) continue;
    shown += 1;
    const li = document.createElement("li");
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "nav-item";
    btn.dataset.sectionName = section.name;
    const abbr = document.createElement("span");
    abbr.className = "nav-item-abbr";
    abbr.textContent = section.name.trim().charAt(0).toUpperCase() || "?";
    const label = document.createElement("span");
    label.className = "nav-item-label";
    label.textContent = section.name;
    const count = document.createElement("span");
    count.className = "nav-item-count";
    count.textContent = String(items.length);
    btn.append(abbr, label, count);
    btn.title = section.name;
    btn.setAttribute("aria-current", section.name === activeNavSection ? "true" : "false");
    btn.addEventListener("click", () => scrollToSection(section.name));
    li.append(btn);
    sectionNavList.append(li);
  }
  sectionNav.hidden = shown === 0;
}

function setupNavObserver() {
  if (navObserver) navObserver.disconnect();
  const sections = [...board.querySelectorAll(".section")];
  if (!sections.length) return;
  navObserver = new IntersectionObserver(
    (entries) => {
      if (navScrollTarget) {
        const target = entries.find(
          (entry) =>
            entry.isIntersecting &&
            entry.target.dataset.sectionName === navScrollTarget &&
            entry.intersectionRatio >= 0.35,
        );
        if (target) endNavScroll();
        return;
      }
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio);
      const top = visible[0]?.target;
      if (top?.dataset.sectionName) setActiveNavSection(top.dataset.sectionName);
    },
    { rootMargin: "-12% 0px -55% 0px", threshold: [0, 0.25, 0.5, 0.75, 1] },
  );
  for (const section of sections) navObserver.observe(section);
}

window.addEventListener("scrollend", endNavScroll, { passive: true });
widthToggle?.addEventListener("click", () => {
  if (widthPanelOpen()) closeWidthPanel();
  else openWidthPanel();
});
widthClose?.addEventListener("click", closeWidthPanel);
contentWidthInput?.addEventListener("input", () => {
  commitContentWidth(contentWidthInput.value);
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && widthPanelOpen()) {
    event.preventDefault();
    closeWidthPanel();
    widthToggle.focus();
  }
});
initContentWidth();
initSidebar();

function loadCollapsed() {
  try {
    const raw = localStorage.getItem(COLLAPSED_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed)
      ? parsed.filter((name) => typeof name === "string" && name)
      : [];
  } catch {
    return [];
  }
}

function saveCollapsed(names) {
  localStorage.setItem(COLLAPSED_KEY, JSON.stringify(names));
}

function isCollapsed(name) {
  return loadCollapsed().includes(name);
}

function toggleCollapsed(name) {
  if (searching()) return;
  const set = new Set(loadCollapsed());
  const willCollapse = !set.has(name);
  if (willCollapse) set.add(name);
  else set.delete(name);
  const known = new Set((data.sections || []).map((section) => section.name));
  saveCollapsed([...set].filter((entry) => known.has(entry)));

  const wrap = [...board.querySelectorAll(".section")].find(
    (el) => el.dataset.sectionName === name,
  );
  if (!wrap) {
    render();
    return;
  }
  wrap.classList.toggle("collapsed", willCollapse);
  const body = wrap.querySelector(".section-body");
  if (body) {
    body.toggleAttribute("inert", willCollapse);
    body.setAttribute("aria-hidden", willCollapse ? "true" : "false");
  }
  const toggle = wrap.querySelector(".section-toggle");
  if (toggle) {
    toggle.setAttribute("aria-expanded", willCollapse ? "false" : "true");
    toggle.title = willCollapse ? "Expand section" : "Collapse section";
  }
}

function layoutPayload(sections) {
  return {
    sections: (sections || []).map((section) => ({
      name: section.name,
      items: (section.items || []).map((item) => item.url),
    })),
  };
}

async function saveLayout(sections) {
  const res = await fetch("/api/layout", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(layoutPayload(sections)),
  });
  if (!res.ok) throw new Error("layout");
}

function matches(item, query) {
  if (!query) return true;
  const hay = (item.name || "").toLowerCase();
  return hay.includes(query);
}

function searching() {
  return Boolean((q.value || "").trim());
}

function setIcon(img, url) {
  img.onerror = () => {
    img.onerror = null;
    img.src = "/icons/__fallback";
  };
  img.src = url || "/icons/__fallback";
}

function tile(item, sectionName) {
  const wrap = document.createElement("div");
  wrap.className = "tile";
  wrap.dataset.url = item.url;
  wrap.dataset.section = sectionName;
  const handle = document.createElement("span");
  handle.className = "tile-drag";
  handle.title = "Drag to reorder";
  handle.setAttribute("role", "img");
  handle.setAttribute("aria-label", "Drag tile");
  handle.draggable = !searching();
  handle.innerHTML = ICON_GRIP;

  const menu = document.createElement("button");
  menu.type = "button";
  menu.className = "tile-menu";
  menu.title = "Settings";
  menu.setAttribute("aria-label", `Settings: ${item.name}`);
  menu.innerHTML = ICON_MORE;
  menu.addEventListener("click", (event) => {
    event.preventDefault();
    event.stopPropagation();
    openEditor(item, sectionName);
  });
  menu.addEventListener("mousedown", (event) => event.stopPropagation());
  menu.addEventListener("pointerdown", (event) => event.stopPropagation());

  const a = document.createElement("a");
  a.className = "tile-link";
  a.href = item.url;
  a.target = "_blank";
  a.rel = "noopener noreferrer";
  a.draggable = false;

  const img = document.createElement("img");
  img.alt = "";
  img.width = 48;
  img.height = 48;
  img.draggable = false;
  setIcon(img, item.icon);

  const iconWrap = document.createElement("span");
  iconWrap.className = "tile-icon";
  iconWrap.append(img);

  const strong = document.createElement("strong");
  strong.textContent = item.name;

  a.append(iconWrap, strong);

  a.addEventListener("click", onTileClick);
  handle.addEventListener("dragstart", (event) => onDragStart(event, wrap));
  handle.addEventListener("dragend", onDragEnd);
  wrap.append(handle, menu, a);
  return wrap;
}

function onTileClick(event) {
  if (!didDrag) return;
  event.preventDefault();
  didDrag = false;
}

function onDragStart(event, wrap) {
  if (searching()) {
    event.preventDefault();
    return;
  }
  dragTile = wrap;
  didDrag = false;
  dragTile.classList.add("dragging");
  event.dataTransfer.effectAllowed = "move";
  event.dataTransfer.setData("text/plain", dragTile.dataset.url || "");
  if (event.dataTransfer.setDragImage) {
    const tileRect = wrap.getBoundingClientRect();
    const handleRect = event.currentTarget.getBoundingClientRect();
    event.dataTransfer.setDragImage(
      wrap,
      handleRect.left - tileRect.left + handleRect.width / 2,
      handleRect.top - tileRect.top + handleRect.height / 2,
    );
  }
}

async function onDragEnd() {
  if (dragTile) dragTile.classList.remove("dragging");
  for (const grid of board.querySelectorAll(".grid")) {
    grid.classList.remove("drag-over");
  }
  if (didDrag) await syncFromDOM();
  dragTile = null;
}

function dropTarget(event) {
  const over = event.target.closest(".tile");
  const grid = event.target.closest(".grid");
  return { over, grid };
}

function onGridDragOver(event) {
  if (!dragTile) return;
  const { over, grid } = dropTarget(event);
  if (!grid) return;
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
  grid.classList.add("drag-over");
  didDrag = true;

  if (over && over !== dragTile) {
    const rect = over.getBoundingClientRect();
    const before = event.clientX < rect.left + rect.width / 2;
    const next = before ? over : over.nextSibling;
    if (dragTile !== next && dragTile.nextSibling !== next) {
      grid.insertBefore(dragTile, next);
    }
    return;
  }
  if (!over && dragTile.parentElement !== grid) {
    grid.append(dragTile);
  }
}

function onGridDragLeave(event) {
  const grid = event.currentTarget;
  if (!grid.contains(event.relatedTarget)) {
    grid.classList.remove("drag-over");
  }
}

function onGridDrop(event) {
  event.preventDefault();
  didDrag = true;
}

function itemByUrl(url) {
  for (const section of data.sections || []) {
    for (const item of section.items || []) {
      if (item.url === url) return item;
    }
  }
  return null;
}

function sectionsFromDOM() {
  const sections = [];
  for (const wrap of board.querySelectorAll(".section")) {
    const name = wrap.querySelector(".section-name")?.textContent || "";
    const items = [];
    for (const el of wrap.querySelectorAll(".tile")) {
      const item = itemByUrl(el.dataset.url);
      if (item) items.push(item);
    }
    sections.push({ name, items });
  }
  return sections;
}

async function syncFromDOM() {
  const sections = sectionsFromDOM();
  data.sections = sections;
  try {
    await saveLayout(sections);
  } catch {
    await boot();
  }
}

async function moveSection(index, dir) {
  if (searching()) return;
  const next = index + dir;
  const list = data.sections || [];
  if (index < 0 || next < 0 || next >= list.length) return;
  const copy = list.slice();
  const [section] = copy.splice(index, 1);
  copy.splice(next, 0, section);
  data.sections = copy;
  try {
    await saveLayout(copy);
    render();
  } catch {
    await boot();
  }
}

function moveButton(label, icon, disabled, onClick, extraClass) {
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = extraClass ? `section-move ${extraClass}` : "section-move";
  btn.title = label;
  btn.setAttribute("aria-label", label);
  btn.innerHTML = icon;
  btn.disabled = disabled;
  btn.addEventListener("click", onClick);
  return btn;
}

function render() {
  const query = (q.value || "").trim().toLowerCase();
  board.classList.remove("ready");
  board.replaceChildren();
  board.classList.toggle("searching", Boolean(query));
  let shown = 0;
  const total = (data.sections || []).length;
  (data.sections || []).forEach((section, index) => {
    const items = (section.items || []).filter((item) => matches(item, query));
    if (!items.length && query) return;
    const wrap = document.createElement("section");
    wrap.className = "section";
    wrap.dataset.sectionName = section.name;
    const collapsed = !query && isCollapsed(section.name);
    wrap.classList.toggle("collapsed", collapsed);
    const h = document.createElement("h2");
    const toggle = document.createElement("button");
    toggle.type = "button";
    toggle.className = "section-toggle";
    toggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
    toggle.title = collapsed ? "Expand section" : "Collapse section";
    toggle.innerHTML = ICON_CHEVRON;
    const label = document.createElement("span");
    label.className = "section-name";
    label.textContent = section.name;
    const count = document.createElement("span");
    count.className = "section-count";
    count.textContent = String(items.length);
    toggle.append(label, count);
    toggle.disabled = Boolean(query);
    toggle.addEventListener("click", () => toggleCollapsed(section.name));
    const rule = document.createElement("span");
    rule.className = "section-rule";
    rule.setAttribute("aria-hidden", "true");
    const moves = document.createElement("span");
    moves.className = "section-moves";
    moves.append(
      moveButton("Move section up", ICON_UP, index === 0, () => moveSection(index, -1)),
      moveButton("Move section down", ICON_DOWN, index === total - 1, () => moveSection(index, 1)),
      moveButton(`Add tile to ${section.name}`, ICON_PLUS, false, () => openCreate(section.name)),
      moveButton(`Delete section ${section.name}`, ICON_TRASH, false, () => openDeleteSection(section), "danger-icon"),
    );
    h.append(toggle, rule, moves);
    const grid = document.createElement("div");
    grid.className = "grid";
    grid.addEventListener("dragover", onGridDragOver);
    grid.addEventListener("dragleave", onGridDragLeave);
    grid.addEventListener("drop", onGridDrop);
    for (const item of items) {
      grid.append(tile(item, section.name));
      shown += 1;
    }
    const body = document.createElement("div");
    body.className = "section-body";
    const inner = document.createElement("div");
    inner.className = "section-body-inner";
    inner.append(grid);
    body.append(inner);
    if (collapsed) {
      body.setAttribute("aria-hidden", "true");
      body.setAttribute("inert", "");
    }
    wrap.append(h, body);
    board.append(wrap);
  });
  empty.hidden = shown !== 0 || !query;
  if (!query && !(data.sections || []).length) {
    empty.hidden = false;
    empty.textContent = "No sections yet. Add your first group.";
  } else if (query && shown === 0) {
    empty.textContent = "No matching services.";
  }
  renderNav(query);
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      board.classList.add("ready");
      setupNavObserver();
    });
  });
}

q.addEventListener("input", render);

function showItemError(message) {
  itemError.hidden = !message;
  itemError.textContent = message || "";
}

function clearPreviewObjectUrl() {
  if (previewObjectUrl) {
    URL.revokeObjectURL(previewObjectUrl);
    previewObjectUrl = "";
  }
}

function fillEditor(sectionName, item) {
  const isNew = !item;
  editing = {
    sectionName,
    isNew,
    originalUrl: item?.url || "",
    name: item?.name || "",
  };
  itemTitle.textContent = isNew ? "New tile" : "Edit tile";
  itemSave.textContent = isNew ? "Add" : "Save";
  itemDelete.hidden = isNew;
  itemName.value = item?.name || "";
  itemUrl.value = item?.url || "";
  itemIcon.value = item?.iconSource || "";
  itemIconFile.value = "";
  clearPreviewObjectUrl();
  setIcon(itemIconPreview, item?.icon);
  showItemError("");
  itemSave.disabled = false;
  editor.showModal();
  itemName.focus();
}

function openEditor(item, sectionName) {
  fillEditor(sectionName, item);
}

function openCreate(sectionName) {
  if (searching()) return;
  if (isCollapsed(sectionName)) toggleCollapsed(sectionName);
  fillEditor(sectionName, null);
}

function closeEditor() {
  clearPreviewObjectUrl();
  editing = null;
  itemForm.reset();
  showItemError("");
  if (editor.open) editor.close();
}

itemCancel.addEventListener("click", () => closeEditor());
editor.addEventListener("cancel", () => {
  clearPreviewObjectUrl();
  editing = null;
});
itemIconFile.addEventListener("change", () => {
  const file = itemIconFile.files && itemIconFile.files[0];
  clearPreviewObjectUrl();
  if (!file) return;
  previewObjectUrl = URL.createObjectURL(file);
  itemIconPreview.onerror = null;
  itemIconPreview.src = previewObjectUrl;
});

itemForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!editing) return;
  const name = itemName.value.trim();
  const url = itemUrl.value.trim();
  if (!name || !url) {
    showItemError("Name and URL are required.");
    return;
  }
  itemSave.disabled = true;
  showItemError("");
  const body = new FormData();
  body.set("section", editing.sectionName);
  if (!editing.isNew) body.set("originalUrl", editing.originalUrl);
  body.set("name", name);
  body.set("url", url);
  body.set("icon", itemIcon.value.trim());
  const file = itemIconFile.files && itemIconFile.files[0];
  if (file) body.set("iconFile", file);
  try {
    const res = await fetch("/api/item", { method: editing.isNew ? "POST" : "PUT", body });
    if (!res.ok) {
      const text = (await res.text()).trim() || "Could not save.";
      showItemError(text);
      itemSave.disabled = false;
      return;
    }
    closeEditor();
    await boot();
  } catch {
    showItemError("Could not save.");
    itemSave.disabled = false;
  }
});

function openDeleteItem() {
  if (!editing || editing.isNew) return;
  deletingItem = {
    sectionName: editing.sectionName,
    url: editing.originalUrl,
    name: itemName.value.trim() || editing.name || "this tile",
  };
  closeEditor();
  itemConfirmText.textContent = `Delete tile "${deletingItem.name}"? This cannot be undone.`;
  showSectionError(itemConfirmError, "");
  itemConfirmOk.disabled = false;
  itemConfirm.showModal();
}

function closeDeleteItem() {
  deletingItem = null;
  showSectionError(itemConfirmError, "");
  if (itemConfirm.open) itemConfirm.close();
}

itemDelete.addEventListener("click", openDeleteItem);
itemConfirmCancel.addEventListener("click", closeDeleteItem);
itemConfirm.addEventListener("cancel", closeDeleteItem);
itemConfirmForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!deletingItem) return;
  itemConfirmOk.disabled = true;
  showSectionError(itemConfirmError, "");
  try {
    const res = await fetch("/api/item", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ section: deletingItem.sectionName, url: deletingItem.url }),
    });
    if (!res.ok) {
      const text = (await res.text()).trim() || "Could not delete tile.";
      showSectionError(itemConfirmError, text);
      itemConfirmOk.disabled = false;
      return;
    }
    closeDeleteItem();
    await boot();
  } catch {
    showSectionError(itemConfirmError, "Could not delete tile.");
    itemConfirmOk.disabled = false;
  }
});

function showSectionError(el, message) {
  el.hidden = !message;
  el.textContent = message || "";
}

function openAddSection() {
  if (searching()) return;
  sectionNameInput.value = "";
  showSectionError(sectionError, "");
  sectionSave.disabled = false;
  sectionEditor.showModal();
  sectionNameInput.focus();
}

function closeAddSection() {
  sectionForm.reset();
  showSectionError(sectionError, "");
  if (sectionEditor.open) sectionEditor.close();
}

function tileCountLabel(count) {
  return count === 1 ? "1 tile" : `${count} tiles`;
}

function openDeleteSection(section) {
  if (searching()) return;
  deletingSection = section.name;
  const count = (section.items || []).length;
  sectionConfirmText.textContent = count
    ? `Delete section "${section.name}" and its ${tileCountLabel(count)}? This cannot be undone.`
    : `Delete empty section "${section.name}"?`;
  showSectionError(sectionConfirmError, "");
  sectionConfirmOk.disabled = false;
  sectionConfirm.showModal();
}

function closeDeleteSection() {
  deletingSection = null;
  showSectionError(sectionConfirmError, "");
  if (sectionConfirm.open) sectionConfirm.close();
}

addSectionBtn.addEventListener("click", openAddSection);
sectionCancel.addEventListener("click", closeAddSection);
sectionEditor.addEventListener("cancel", closeAddSection);
sectionForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const name = sectionNameInput.value.trim();
  if (!name) {
    showSectionError(sectionError, "Name is required.");
    return;
  }
  sectionSave.disabled = true;
  showSectionError(sectionError, "");
  try {
    const res = await fetch("/api/section", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    if (!res.ok) {
      const text = (await res.text()).trim() || "Could not add section.";
      showSectionError(sectionError, text);
      sectionSave.disabled = false;
      return;
    }
    closeAddSection();
    await boot();
  } catch {
    showSectionError(sectionError, "Could not add section.");
    sectionSave.disabled = false;
  }
});

sectionConfirmCancel.addEventListener("click", closeDeleteSection);
sectionConfirm.addEventListener("cancel", closeDeleteSection);
sectionConfirmForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!deletingSection) return;
  sectionConfirmOk.disabled = true;
  showSectionError(sectionConfirmError, "");
  try {
    const res = await fetch("/api/section", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: deletingSection }),
    });
    if (!res.ok) {
      const text = (await res.text()).trim() || "Could not delete section.";
      showSectionError(sectionConfirmError, text);
      sectionConfirmOk.disabled = false;
      return;
    }
    closeDeleteSection();
    await boot();
  } catch {
    showSectionError(sectionConfirmError, "Could not delete section.");
    sectionConfirmOk.disabled = false;
  }
});

async function boot() {
  const res = await fetch("/api/config");
  if (!res.ok) throw new Error("config");
  const incoming = await res.json();
  titleEl.textContent = incoming.title || "Lantern";
  document.title = incoming.title || "Lantern";
  data = {
    title: incoming.title || "Lantern",
    sections: incoming.sections || [],
  };
  render();
}

boot().catch(() => {
  empty.hidden = false;
  empty.textContent = "Could not load configuration.";
});
