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
const sectionExpandAllBtn = document.getElementById("section-expand-all");
const sectionCollapseAllBtn = document.getElementById("section-collapse-all");
const sectionNav = document.getElementById("section-nav");
const sectionNavList = document.getElementById("section-nav-list");
const sectionNavToggle = document.getElementById("section-nav-toggle");
const sectionNavEdge = document.getElementById("section-nav-edge");
const sectionNavBackdrop = document.getElementById("section-nav-backdrop");
const widthPicker = document.getElementById("width-picker");
const widthToggle = document.getElementById("width-toggle");
const widthPanel = document.getElementById("width-panel");
const widthClose = document.getElementById("width-close");
const gridColsPicker = document.getElementById("grid-cols-picker");

const THEME_KEY = "lantern-theme";
const COLLAPSED_KEY = "lantern-collapsed";
const SIDEBAR_KEY = "lantern-sidebar-collapsed";
const GRID_COLS_KEY = "lantern-grid-cols";
const LEGACY_CONTENT_WIDTH_KEY = "lantern-content-width";
const GRID_COLS_MIN = 4;
const GRID_COLS_MAX = 8;
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
let dropPlaceholder = null;
let dropHoverCell = null;
let dropCommitCell = null;
const GRID_CELL_MIN = 168;
const GRID_GAP = 14;
const GRID_PAD_X = 8;
const GRID_WIDTH_SAFETY = 16;
const GRID_ROW_HEIGHT = 148;
const gridObservers = new WeakMap();
let activeNavSection = "";
let navScrollTarget = "";
let navScrollTimer = 0;
let navSpyRaf = 0;
let gridCols = GRID_COLS_MIN;
let drawerOpen = false;
const SIDEBAR_EXPANDED_WIDTH = 220;
const SIDEBAR_COLLAPSED_WIDTH = 52;
const COMPACT_MQ = window.matchMedia("(max-width: 900px)");

function isCompactLayout() {
  return COMPACT_MQ.matches;
}

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

function setSidebarDrawer(open) {
  drawerOpen = open;
  document.body.classList.toggle("nav-drawer-open", open);
  sectionNav.classList.toggle("drawer-open", open);
  if (sectionNavBackdrop) sectionNavBackdrop.hidden = !open;
  if (sectionNavEdge) {
    sectionNavEdge.hidden = !isCompactLayout() || open || sectionNav.hidden;
    sectionNavEdge.setAttribute("aria-expanded", open ? "true" : "false");
  }
  sectionNavToggle.title = open ? "Close sections" : "Open sections";
  sectionNavToggle.setAttribute("aria-expanded", open ? "true" : "false");
  sectionNav.inert = !open;
  sectionNav.setAttribute("aria-hidden", open ? "false" : "true");
}

function updateSidebarLayout() {
  const compact = isCompactLayout();
  document.body.classList.toggle("layout-compact", compact);
  if (compact) {
    sectionNav.classList.remove("collapsed");
    document.documentElement.style.setProperty("--sidebar-current-width", "0px");
    setSidebarDrawer(drawerOpen);
    return;
  }
  document.body.classList.remove("nav-drawer-open");
  sectionNav.classList.remove("drawer-open");
  sectionNav.inert = false;
  sectionNav.removeAttribute("aria-hidden");
  if (sectionNavBackdrop) sectionNavBackdrop.hidden = true;
  if (sectionNavEdge) sectionNavEdge.hidden = true;
  setSidebarCollapsed(isSidebarCollapsed());
}

function clampGridCols(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) return GRID_COLS_MIN;
  return Math.min(GRID_COLS_MAX, Math.max(GRID_COLS_MIN, Math.round(num)));
}

function widthFromCols(cols) {
  const safeCols = clampGridCols(cols);
  return safeCols * GRID_CELL_MIN
    + (safeCols - 1) * GRID_GAP
    + GRID_PAD_X
    + GRID_WIDTH_SAFETY;
}

function gridInnerWidth(grid) {
  if (!grid) return widthFromCols(gridCols) - GRID_PAD_X;
  const styles = getComputedStyle(grid);
  const padL = Number.parseFloat(styles.paddingLeft) || 0;
  const padR = Number.parseFloat(styles.paddingRight) || 0;
  return Math.max(0, grid.clientWidth - padL - padR);
}

function colsFromLegacyWidth(width) {
  const num = Number(width);
  if (!Number.isFinite(num)) return GRID_COLS_MIN;
  return clampGridCols(Math.floor((num + GRID_GAP) / (GRID_CELL_MIN + GRID_GAP)));
}

function updateGridColsPicker() {
  if (!gridColsPicker) return;
  for (const option of gridColsPicker.querySelectorAll(".grid-cols-option")) {
    const selected = Number(option.dataset.cols) === gridCols;
    option.setAttribute("aria-pressed", selected ? "true" : "false");
  }
}

function commitGridCols(value, persist = true) {
  const cols = clampGridCols(value);
  const prev = gridCols;
  gridCols = cols;
  document.documentElement.style.setProperty("--main-width", `${widthFromCols(cols)}px`);
  updateGridColsPicker();
  if (persist) localStorage.setItem(GRID_COLS_KEY, String(cols));
  if (prev !== cols && (data.sections || []).length) render();
}

function initGridColsPicker() {
  if (!gridColsPicker) return;
  gridColsPicker.replaceChildren();
  for (let cols = GRID_COLS_MIN; cols <= GRID_COLS_MAX; cols += 1) {
    const option = document.createElement("button");
    option.type = "button";
    option.className = "grid-cols-option";
    option.dataset.cols = String(cols);
    option.setAttribute("aria-pressed", "false");
    option.textContent = String(cols);
    option.addEventListener("click", () => commitGridCols(cols));
    gridColsPicker.append(option);
  }
}

function initGridCols() {
  initGridColsPicker();
  const storedCols = localStorage.getItem(GRID_COLS_KEY);
  if (storedCols) {
    commitGridCols(storedCols, false);
    return;
  }
  const legacyWidth = localStorage.getItem(LEGACY_CONTENT_WIDTH_KEY);
  if (legacyWidth) {
    commitGridCols(colsFromLegacyWidth(legacyWidth), true);
    return;
  }
  commitGridCols(GRID_COLS_MIN, false);
}

function widthPanelOpen() {
  return widthToggle.getAttribute("aria-expanded") === "true";
}

function openWidthPanel() {
  closeThemeMenu();
  updateGridColsPicker();
  widthPanel.hidden = false;
  widthToggle.setAttribute("aria-expanded", "true");
  widthPicker.classList.add("open");
  gridColsPicker?.querySelector(`[data-cols="${gridCols}"]`)?.focus();
}

function closeWidthPanel() {
  widthPanel.hidden = true;
  widthToggle.setAttribute("aria-expanded", "false");
  widthPicker.classList.remove("open");
  updateGridColsPicker();
}

function initSidebar() {
  drawerOpen = false;
  updateSidebarLayout();
}

sectionNavToggle.addEventListener("click", () => {
  if (isCompactLayout()) {
    setSidebarDrawer(false);
    return;
  }
  setSidebarCollapsed(!isSidebarCollapsed());
});

sectionNavEdge?.addEventListener("click", () => setSidebarDrawer(true));
sectionNavBackdrop?.addEventListener("click", () => setSidebarDrawer(false));
COMPACT_MQ.addEventListener("change", () => {
  drawerOpen = false;
  updateSidebarLayout();
});

function sectionElement(name) {
  return [...board.querySelectorAll(".section")].find(
    (el) => el.dataset.sectionName === name,
  );
}

function beginNavScroll(name) {
  navScrollTarget = name;
  window.clearTimeout(navScrollTimer);
  setActiveNavSection(name);
  navScrollTimer = window.setTimeout(endNavScroll, 1200);
}

function endNavScroll() {
  if (!navScrollTarget) return;
  navScrollTarget = "";
  window.clearTimeout(navScrollTimer);
  syncActiveNavFromScroll();
}

function scrollToSection(name) {
  const wrap = sectionElement(name);
  if (!wrap) return;
  beginNavScroll(name);
  if (isCollapsed(name)) toggleCollapsed(name);
  wrap.scrollIntoView({ behavior: "smooth", block: "start" });
}

function setActiveNavSection(name) {
  if (name === activeNavSection) return;
  activeNavSection = name;
  for (const item of sectionNavList.querySelectorAll(".nav-item")) {
    const current = item.dataset.sectionName === name;
    if (current) item.setAttribute("aria-current", "true");
    else item.removeAttribute("aria-current");
  }
}

function navSpyMarker() {
  return 80;
}

function sectionAtMarker() {
  const sections = [...board.querySelectorAll(".section")];
  if (!sections.length) return "";
  const maxScroll = document.documentElement.scrollHeight - window.innerHeight;
  if (maxScroll > 0 && window.scrollY >= maxScroll - 8) {
    return sections[sections.length - 1].dataset.sectionName || "";
  }
  if (window.scrollY <= 8) {
    return sections[0].dataset.sectionName || "";
  }
  const marker = navSpyMarker();
  let current = sections[0];
  for (const section of sections) {
    if (section.getBoundingClientRect().top <= marker) current = section;
    else break;
  }
  return current.dataset.sectionName || "";
}

function syncActiveNavFromScroll() {
  if (navScrollTarget) return;
  const name = sectionAtMarker();
  if (name) setActiveNavSection(name);
}

function scheduleNavSpy() {
  if (navSpyRaf) return;
  navSpyRaf = requestAnimationFrame(() => {
    navSpyRaf = 0;
    syncActiveNavFromScroll();
  });
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
    if (section.name === activeNavSection) btn.setAttribute("aria-current", "true");
    btn.addEventListener("click", () => {
      scrollToSection(section.name);
      if (isCompactLayout()) setSidebarDrawer(false);
    });
    li.append(btn);
    sectionNavList.append(li);
  }
  sectionNav.hidden = shown === 0;
  if (sectionNavEdge) {
    sectionNavEdge.hidden = !isCompactLayout() || drawerOpen || shown === 0;
  }
}

function setupNavObserver() {
  syncActiveNavFromScroll();
}

window.addEventListener("scroll", scheduleNavSpy, { passive: true });
window.addEventListener("resize", scheduleNavSpy);
window.addEventListener("scrollend", endNavScroll, { passive: true });
window.addEventListener("wheel", () => {
  if (navScrollTarget) endNavScroll();
}, { passive: true });
window.addEventListener("touchstart", () => {
  if (navScrollTarget) endNavScroll();
}, { passive: true });
widthToggle?.addEventListener("click", () => {
  if (widthPanelOpen()) closeWidthPanel();
  else openWidthPanel();
});
widthClose?.addEventListener("click", closeWidthPanel);
document.addEventListener("keydown", (event) => {
  if (event.key !== "Escape") return;
  if (widthPanelOpen()) {
    event.preventDefault();
    closeWidthPanel();
    widthToggle.focus();
    return;
  }
  if (isCompactLayout() && drawerOpen) {
    event.preventDefault();
    setSidebarDrawer(false);
    sectionNavEdge?.focus();
  }
});
initGridCols();
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

function applySectionCollapsed(wrap, collapsed) {
  wrap.classList.toggle("collapsed", collapsed);
  const body = wrap.querySelector(".section-body");
  if (body) {
    body.toggleAttribute("inert", collapsed);
    body.setAttribute("aria-hidden", collapsed ? "true" : "false");
  }
  const toggle = wrap.querySelector(".section-toggle");
  if (toggle) {
    toggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
    toggle.title = collapsed ? "Expand section" : "Collapse section";
  }
}

function setAllSectionsCollapsed(collapsed) {
  if (searching()) return;
  const names = (data.sections || []).map((section) => section.name);
  saveCollapsed(collapsed ? names : []);
  const sections = [...board.querySelectorAll(".section")];
  if (!sections.length) {
    render();
    return;
  }
  for (const wrap of sections) applySectionCollapsed(wrap, collapsed);
  scheduleNavSpy();
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
  applySectionCollapsed(wrap, willCollapse);
  scheduleNavSpy();
}

function itemHasPosition(item) {
  return item.col > 0 && item.row > 0;
}

function cellIndex(col, row, cols) {
  const safeCols = Math.max(1, cols);
  return (row - 1) * safeCols + (col - 1);
}

function indexToCell(index, cols) {
  const safeCols = Math.max(1, cols);
  const safeIndex = Math.max(0, index);
  return {
    col: (safeIndex % safeCols) + 1,
    row: Math.floor(safeIndex / safeCols) + 1,
  };
}

function resolvePositions(items, cols) {
  const safeCols = Math.max(1, cols);
  const out = (items || []).map((item) => ({ ...item }));
  let maxIndex = -1;
  for (const item of out) {
    if (itemHasPosition(item)) {
      const col = Math.min(item.col, safeCols);
      item.col = col;
      const idx = cellIndex(col, item.row, safeCols);
      if (idx > maxIndex) maxIndex = idx;
      continue;
    }
    const nextIndex = maxIndex + 1;
    const cell = indexToCell(nextIndex, safeCols);
    item.col = cell.col;
    item.row = cell.row;
    maxIndex = nextIndex;
  }
  return out;
}

function maxColsForWidth(width) {
  return Math.max(1, Math.floor((width + GRID_GAP) / (GRID_CELL_MIN + GRID_GAP)));
}

function estimateGridWidth() {
  const main = document.querySelector(".workspace-main");
  if (main?.clientWidth) return Math.max(0, main.clientWidth - GRID_PAD_X);
  return widthFromCols(gridCols) - GRID_PAD_X;
}

function effectiveGridCols(grid) {
  const width = grid ? gridInnerWidth(grid) : estimateGridWidth();
  return Math.min(gridCols, maxColsForWidth(width));
}

function toDisplayPosition(item, displayCols) {
  if (!itemHasPosition(item)) return { col: item.col, row: item.row };
  const idx = cellIndex(item.col, item.row, gridCols);
  return indexToCell(idx, displayCols);
}

function toCanonicalPosition(col, row, displayCols) {
  const idx = cellIndex(col, row, displayCols);
  return indexToCell(idx, gridCols);
}

function gridMetrics(grid) {
  const cols = effectiveGridCols(grid);
  const styles = getComputedStyle(grid);
  const gap = Number.parseFloat(styles.columnGap) || GRID_GAP;
  const width = gridInnerWidth(grid);
  const cellWidth = (width - gap * (cols - 1)) / cols;
  const rowHeight = Number.parseFloat(styles.getPropertyValue("--grid-row-height")) || GRID_ROW_HEIGHT;
  return { cols, gap, cellWidth, rowHeight };
}

function gridMaxRow(items, displayCols) {
  const resolved = resolvePositions(items, gridCols);
  let maxRow = 1;
  for (const item of resolved) {
    const pos = toDisplayPosition(item, displayCols);
    if (pos.row > maxRow) maxRow = pos.row;
  }
  return maxRow;
}

function setGridRowCount(grid, rowCount) {
  const rows = Math.max(1, rowCount);
  grid.style.gridTemplateRows = `repeat(${rows}, minmax(${GRID_ROW_HEIGHT}px, auto))`;
  grid.style.minHeight = `${rows * GRID_ROW_HEIGHT + Math.max(0, rows - 1) * GRID_GAP + 4}px`;
}

function applyGridSizing(grid, items) {
  const cols = effectiveGridCols(grid);
  grid.style.setProperty("--grid-cols", String(cols));
  const canonical = resolvePositions(items, gridCols);
  setGridRowCount(grid, gridMaxRow(canonical, cols));
  return canonical.map((item) => {
    const pos = toDisplayPosition(item, cols);
    return { ...item, col: pos.col, row: pos.row };
  });
}

function clampGridCol(col, cols) {
  return Math.min(cols, Math.max(1, col));
}

function sectionItemsForGrid(grid) {
  const sectionName = grid.closest(".section")?.dataset.sectionName;
  if (!sectionName) return [];
  const section = (data.sections || []).find((entry) => entry.name === sectionName);
  return section?.items || [];
}

function beginGridDrop(grid) {
  const cols = effectiveGridCols(grid);
  setGridRowCount(grid, gridMaxRow(sectionItemsForGrid(grid), cols) + 1);
  grid.classList.add("is-dropping");
}

function endGridDrop() {
  if (dropPlaceholder) dropPlaceholder.hidden = true;
  dropHoverCell = null;
  dropCommitCell = null;
  for (const grid of board.querySelectorAll(".grid.is-dropping")) {
    grid.classList.remove("is-dropping");
    const sectionName = grid.closest(".section")?.dataset.sectionName;
    const section = (data.sections || []).find((entry) => entry.name === sectionName);
    if (section) applyGridSizing(grid, section.items || []);
  }
}

function syncTilePositions(grid, sectionName, items) {
  const cols = effectiveGridCols(grid);
  grid.style.setProperty("--grid-cols", String(cols));
  const canonical = resolvePositions(items, gridCols);
  setGridRowCount(grid, gridMaxRow(canonical, cols));
  for (const item of canonical) {
    const pos = toDisplayPosition(item, cols);
    const tile = grid.querySelector(`.tile[data-url="${CSS.escape(item.url)}"]`);
    if (!tile) continue;
    tile.dataset.col = String(pos.col);
    tile.dataset.row = String(pos.row);
    tile.style.gridColumn = String(pos.col);
    tile.style.gridRow = String(pos.row);
  }
}

function observeGrid(grid, sectionName) {
  if (gridObservers.has(grid)) return;
  const observer = new ResizeObserver(() => {
    if (searching()) return;
    const section = (data.sections || []).find((entry) => entry.name === sectionName);
    if (!section) return;
    syncTilePositions(grid, sectionName, section.items || []);
  });
  observer.observe(grid);
  gridObservers.set(grid, observer);
}

function occupantAt(posMap, col, row, skipUrl = "") {
  for (const [url, pos] of posMap) {
    if (url === skipUrl) continue;
    if (pos.col === col && pos.row === row) return url;
  }
  return "";
}

function sectionDisplayMap(section, grid) {
  const { cols } = gridMetrics(grid);
  const canonical = resolvePositions(section.items || [], gridCols);
  return {
    cols,
    posMap: new Map(canonical.map((item) => [item.url, toDisplayPosition(item, cols)])),
  };
}

function writePositions(section, grid, posMap, cols) {
  for (const item of section.items || []) {
    const pos = posMap.get(item.url);
    if (!pos) continue;
    const canon = toCanonicalPosition(pos.col, pos.row, cols);
    item.col = canon.col;
    item.row = canon.row;
  }
  syncTilePositions(grid, section.name, section.items || []);
}

function repackSection(sectionName) {
  const section = (data.sections || []).find((entry) => entry.name === sectionName);
  const grid = sectionElement(sectionName)?.querySelector(".grid");
  if (!section || !grid) return;
  section.items = resolvePositions(section.items || [], gridCols);
  syncTilePositions(grid, sectionName, section.items);
  applyGridSizing(grid, section.items);
}

function moveItemBetweenSections(
  sourceSectionName,
  targetSectionName,
  url,
  col,
  row,
  tileEl,
  targetGrid,
) {
  const sourceSection = (data.sections || []).find((entry) => entry.name === sourceSectionName);
  const targetSection = (data.sections || []).find((entry) => entry.name === targetSectionName);
  if (!sourceSection || !targetSection) return;

  const sourceGrid = sectionElement(sourceSectionName)?.querySelector(".grid");
  const itemIndex = (sourceSection.items || []).findIndex((item) => item.url === url);
  if (itemIndex === -1) return;

  const sourceOrigin = sourceGrid
    ? sectionDisplayMap(sourceSection, sourceGrid).posMap.get(url)
    : null;
  const targetMap = sectionDisplayMap(targetSection, targetGrid);
  const occupantUrl = occupantAt(targetMap.posMap, col, row, url);

  const [item] = sourceSection.items.splice(itemIndex, 1);
  targetSection.items = targetSection.items || [];
  targetSection.items.push(item);

  tileEl.dataset.section = targetSectionName;
  targetGrid.append(tileEl);

  if (occupantUrl && sourceOrigin && sourceGrid) {
    const occupantIndex = targetSection.items.findIndex((entry) => entry.url === occupantUrl);
    if (occupantIndex !== -1) {
      const [occupant] = targetSection.items.splice(occupantIndex, 1);
      sourceSection.items.push(occupant);
      const occupantTile = targetGrid.querySelector(`.tile[data-url="${CSS.escape(occupantUrl)}"]`);
      if (occupantTile) {
        occupantTile.dataset.section = sourceSectionName;
        sourceGrid.append(occupantTile);
      }
      placeWithSwap(sourceSectionName, occupantUrl, sourceOrigin.col, sourceOrigin.row);
    }
  }

  placeWithSwap(targetSectionName, url, col, row);
  if (!occupantUrl) repackSection(sourceSectionName);

  const sourceCount = sectionElement(sourceSectionName)?.querySelector(".section-count");
  const targetCount = sectionElement(targetSectionName)?.querySelector(".section-count");
  if (sourceCount) sourceCount.textContent = String(sourceSection.items.length);
  if (targetCount) targetCount.textContent = String(targetSection.items.length);
}

function placeWithSwap(sectionName, draggedUrl, col, row) {
  const section = (data.sections || []).find((entry) => entry.name === sectionName);
  const grid = sectionElement(sectionName)?.querySelector(".grid");
  if (!section || !grid) return;
  const { cols, posMap } = sectionDisplayMap(section, grid);
  const origin = posMap.get(draggedUrl);
  const occupant = occupantAt(posMap, col, row, draggedUrl);
  if (occupant && origin) posMap.set(occupant, { col: origin.col, row: origin.row });
  posMap.set(draggedUrl, { col, row });
  writePositions(section, grid, posMap, cols);
}

function layoutPayload(sections) {
  return {
    sections: (sections || []).map((section) => ({
      name: section.name,
      items: (section.items || []).map((item) => ({
        url: item.url,
        col: item.col || 0,
        row: item.row || 0,
      })),
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
  handle.title = "Drag to any section";
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
  dropCommitCell = null;
  dropHoverCell = null;
  dragTile.classList.add("dragging");
  const grid = wrap.parentElement;
  if (grid?.classList.contains("grid")) beginGridDrop(grid);
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

function ensureDropPlaceholder(grid) {
  if (dropPlaceholder?.parentElement === grid) return dropPlaceholder;
  const el = document.createElement("div");
  el.className = "grid-drop-placeholder";
  el.setAttribute("aria-hidden", "true");
  el.hidden = true;
  grid.append(el);
  dropPlaceholder = el;
  return el;
}

function showDropPlaceholder(grid, col, row) {
  if (dropPlaceholder?.parentElement && dropPlaceholder.parentElement !== grid) {
    dropPlaceholder.hidden = true;
    dropPlaceholder.remove();
    dropPlaceholder = null;
  }
  const { cols } = gridMetrics(grid);
  const targetCol = clampGridCol(col, cols);
  const targetRow = Math.max(1, row);
  const items = sectionItemsForGrid(grid);
  setGridRowCount(grid, Math.max(gridMaxRow(items, cols), targetRow + 1));
  if (!grid.classList.contains("is-dropping")) beginGridDrop(grid);
  const placeholder = ensureDropPlaceholder(grid);
  placeholder.style.gridColumn = String(targetCol);
  placeholder.style.gridRow = String(targetRow);
  placeholder.hidden = false;
  dropHoverCell = { col: targetCol, row: targetRow, grid };
}

function cellFromPointer(event, grid) {
  const rect = grid.getBoundingClientRect();
  const styles = getComputedStyle(grid);
  const padding = Number.parseFloat(styles.paddingLeft) || 2;
  const { cols, gap, cellWidth, rowHeight } = gridMetrics(grid);
  const x = event.clientX - rect.left - padding;
  const y = event.clientY - rect.top - padding;
  const col = Math.min(cols, Math.max(1, Math.floor(x / (cellWidth + gap)) + 1));
  const row = Math.max(1, Math.floor(y / (rowHeight + gap)) + 1);
  return { col, row };
}

async function onDragEnd() {
  const tile = dragTile;
  const commit = dropCommitCell ?? dropHoverCell;
  const target = commit?.grid
    ? commit
    : commit && dropHoverCell?.grid
      ? { ...commit, grid: dropHoverCell.grid }
      : null;
  if (tile) tile.classList.remove("dragging");
  endGridDrop();
  if (didDrag && tile && target?.grid) {
    const url = tile.dataset.url;
    const sourceSectionName = tile.dataset.section;
    const targetSectionName = target.grid.closest(".section")?.dataset.sectionName;
    if (sourceSectionName && targetSectionName) {
      if (sourceSectionName === targetSectionName) {
        placeWithSwap(sourceSectionName, url, target.col, target.row);
      } else {
        moveItemBetweenSections(
          sourceSectionName,
          targetSectionName,
          url,
          target.col,
          target.row,
          tile,
          target.grid,
        );
        renderNav((q.value || "").trim().toLowerCase());
      }
      await syncFromDOM();
    }
  }
  dragTile = null;
}

function onGridDragOver(event) {
  if (!dragTile) return;
  const grid = event.target.closest(".grid");
  if (!grid) return;
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
  didDrag = true;
  const cell = cellFromPointer(event, grid);
  showDropPlaceholder(grid, cell.col, cell.row);
}

function onGridDragLeave(event) {
  const grid = event.currentTarget;
  if (!grid.contains(event.relatedTarget) && dropHoverCell?.grid === grid) {
    if (dropPlaceholder) dropPlaceholder.hidden = true;
    dropHoverCell = null;
  }
}

function onGridDrop(event) {
  event.preventDefault();
  didDrag = true;
  const grid = event.currentTarget;
  dropCommitCell = { ...cellFromPointer(event, grid), grid };
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
      if (!item) continue;
      items.push({
        ...item,
        col: Number(el.dataset.col) || item.col || 0,
        row: Number(el.dataset.row) || item.row || 0,
      });
    }
    sections.push({ name, items });
  }
  return sections;
}

async function syncFromDOM() {
  try {
    await saveLayout(data.sections);
  } catch {
    await boot();
  }
}

function updateSectionMoveButtons() {
  const sections = [...board.querySelectorAll(".section")];
  sections.forEach((section, index) => {
    section.querySelector('[data-section-direction="up"]').disabled = index === 0;
    section.querySelector('[data-section-direction="down"]').disabled = index === sections.length - 1;
  });
}

async function moveSection(name, dir) {
  if (searching()) return;
  const list = data.sections || [];
  const index = list.findIndex((section) => section.name === name);
  const next = index + dir;
  if (index < 0 || next < 0 || next >= list.length) return;
  const neighbor = list[next];
  const copy = list.slice();
  const [section] = copy.splice(index, 1);
  copy.splice(next, 0, section);
  try {
    await saveLayout(copy);
    data.sections = copy;
    const moved = sectionElement(name);
    const adjacent = sectionElement(neighbor.name);
    if (moved && adjacent && moved !== adjacent) {
      if (dir < 0) board.insertBefore(moved, adjacent);
      else adjacent.after(moved);
    }
    updateSectionMoveButtons();
    renderNav("");
    setupNavObserver();
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
  dropPlaceholder = null;
  dropHoverCell = null;
  dropCommitCell = null;
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
    const moveUp = moveButton("Move section up", ICON_UP, index === 0, () => moveSection(section.name, -1));
    moveUp.dataset.sectionDirection = "up";
    const moveDown = moveButton("Move section down", ICON_DOWN, index === total - 1, () => moveSection(section.name, 1));
    moveDown.dataset.sectionDirection = "down";
    moves.append(
      moveUp,
      moveDown,
      moveButton(`Add tile to ${section.name}`, ICON_PLUS, false, () => openCreate(section.name)),
      moveButton(`Delete section ${section.name}`, ICON_TRASH, false, () => openDeleteSection(section), "danger-icon"),
    );
    h.append(toggle, rule, moves);
    const grid = document.createElement("div");
    grid.className = "grid";
    grid.style.setProperty("--grid-row-height", `${GRID_ROW_HEIGHT}px`);
    grid.addEventListener("dragover", onGridDragOver);
    grid.addEventListener("dragleave", onGridDragLeave);
    grid.addEventListener("drop", onGridDrop);
    const displayCols = query ? gridCols : effectiveGridCols(null);
    const resolvedItems = query ? items : resolvePositions(items, gridCols);
    for (const item of resolvedItems) {
      const el = tile(item, section.name);
      if (!query) {
        const pos = toDisplayPosition(item, displayCols);
        el.dataset.col = String(pos.col);
        el.dataset.row = String(pos.row);
        el.style.gridColumn = String(pos.col);
        el.style.gridRow = String(pos.row);
      }
      grid.append(el);
      shown += 1;
    }
    if (!query) {
      applyGridSizing(grid, items);
      observeGrid(grid, section.name);
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
sectionExpandAllBtn?.addEventListener("click", () => setAllSectionsCollapsed(false));
sectionCollapseAllBtn?.addEventListener("click", () => setAllSectionsCollapsed(true));
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
  applyConfig(await res.json());
}

function applyConfig(incoming) {
  titleEl.textContent = incoming.title || "Lantern";
  document.title = incoming.title || "Lantern";
  data = {
    title: incoming.title || "Lantern",
    sections: incoming.sections || [],
  };
  render();
}

const initialConfig = document.getElementById("initial-config");
if (initialConfig?.textContent) {
  try {
    applyConfig(JSON.parse(initialConfig.textContent));
  } catch {
    boot().catch(() => {
      empty.hidden = false;
      empty.textContent = "Could not load configuration.";
    });
  }
} else {
  boot().catch(() => {
    empty.hidden = false;
    empty.textContent = "Could not load configuration.";
  });
}
