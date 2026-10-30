// ==========================================
// Theme selector (user preference, persisted in localStorage)
// ==========================================
const THEME_KEY = "ts6viewer-theme";
const THEME_OPTIONS = ["default", "dark", "light"];

function resolveTheme(theme) {
    if (theme === "dark" || theme === "light") {
        return theme;
    }
    // "default" follows the system preference
    return (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches)
        ? "dark"
        : "light";
}

function applyTheme(theme) {
    const link = document.getElementById("themeStylesheet");
    if (link) {
        link.href = "/static/" + resolveTheme(theme) + ".css";
    }

    // Highlight the active option in the menu
    document.querySelectorAll(".theme-option").forEach((opt) => {
        opt.classList.toggle("active", opt.dataset.theme === theme);
    });
}

function initThemeSelector() {
    const button = document.getElementById("themeButton");
    const menu = document.getElementById("themeMenu");
    if (!button || !menu) return;

    // Restore saved preference (default: "default" = system)
    let saved = localStorage.getItem(THEME_KEY);
    if (!THEME_OPTIONS.includes(saved)) {
        saved = "default";
    }
    applyTheme(saved);

    // Toggle menu
    button.addEventListener("click", (e) => {
        e.stopPropagation();
        const open = !menu.hidden;
        menu.hidden = open;
        button.setAttribute("aria-expanded", String(!open));
    });

    // Select a theme
    menu.querySelectorAll(".theme-option").forEach((opt) => {
        opt.addEventListener("click", () => {
            const theme = opt.dataset.theme;
            localStorage.setItem(THEME_KEY, theme);
            applyTheme(theme);
            menu.hidden = true;
            button.setAttribute("aria-expanded", "false");
        });
    });

    // Close menu when clicking outside
    document.addEventListener("click", (e) => {
        if (!menu.hidden && !menu.contains(e.target) && e.target !== button) {
            menu.hidden = true;
            button.setAttribute("aria-expanded", "false");
        }
    });

    // Follow system preference changes while in "default" mode
    if (window.matchMedia) {
        window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
            if (localStorage.getItem(THEME_KEY) === "default") {
                applyTheme("default");
            }
        });
    }
}

window.addEventListener("load", initThemeSelector);

// ==========================================
// Refresh countdown with SESSIONSTORAGE persistence
// ==========================================

// Load counter from sessionStorage
let counter = Number(sessionStorage.getItem("refreshCounter"));

// If no valid value exists, reset to refreshTime
if (!counter || counter > refreshTime || counter < 0) {
    counter = refreshTime;
}

const refreshText = document.getElementById("refreshButtonText");

// Immediately update UI with stored value
refreshText.textContent = counter;

// Save initial value (in case it was missing)
sessionStorage.setItem("refreshCounter", counter);

// Countdown loop
setInterval(() => {
    counter--;
    refreshText.textContent = counter;

    // Save updated counter
    sessionStorage.setItem("refreshCounter", counter);

    // When countdown reaches zero → refresh data
    if (counter <= 0) {
        counter = refreshTime;
        refreshText.textContent = counter;
        sessionStorage.setItem("refreshCounter", counter);
        fetchViewerData(true);
    }
}, 1000);

// Save counter when leaving the page
window.addEventListener("beforeunload", () => {
    sessionStorage.setItem("refreshCounter", counter);
});

// ==========================================
// Spacer rendering helpers
// ==========================================
function repeatToWidth(element) {
    const pattern = element.dataset.pattern;

    element.style.display = "block";
    element.style.width = "100%";

    const rectWidth = element.getBoundingClientRect().width;
    const parentWidth = element.parentElement.getBoundingClientRect().width;
    const offsetWidth = element.offsetWidth;
    const clientWidth = element.clientWidth;

    const measure = document.createElement("span");
    measure.style.visibility = "hidden";
    measure.style.whiteSpace = "pre";
    measure.style.position = "absolute";
    measure.style.font = getComputedStyle(element).font;
    measure.textContent = pattern;
    document.body.appendChild(measure);

    const patternWidth = measure.getBoundingClientRect().width;

    measure.remove();

    const targetWidth = rectWidth || parentWidth || offsetWidth || clientWidth;
    
    const repeatCount = Math.ceil(targetWidth / patternWidth);

    element.textContent = pattern.repeat(repeatCount);
}

function getCharWidth(element) {
    const test = document.createElement("span");
    test.style.visibility = "hidden";
    test.style.whiteSpace = "pre";
    test.style.fontFamily = getComputedStyle(element).fontFamily;
    test.textContent = "_";
    document.body.appendChild(test);
    const width = test.getBoundingClientRect().width;
    test.remove();
    return width;
}

function updateAllSpacers() {
    document.querySelectorAll(".repeat").forEach(repeatToWidth);
}

function debug(msg) {
    let box = document.getElementById("debugBox");
    if (!box) {
        box = document.createElement("div");
        box.id = "debugBox";
        box.style.position = "fixed";
        box.style.bottom = "0";
        box.style.left = "0";
        box.style.right = "0";
        box.style.background = "rgba(0,0,0,0.85)";
        box.style.color = "#0f0";
        box.style.fontSize = "12px";
        box.style.padding = "6px";
        box.style.zIndex = "999999";
        box.style.maxHeight = "40vh";
        box.style.overflowY = "auto";
        box.style.fontFamily = "monospace";
        document.body.appendChild(box);
    }
    box.innerHTML += msg + "<br>";
}

// ==========================================
// Fetch viewer data from backend
// ==========================================
async function fetchViewerData(force = false) {
    const url = force ? "/ts6viewer/data?force=1" : "/ts6viewer/data";

    try {
        const response = await fetch(url);
        const data = await response.json();

        updateServerInfo(data.VMServer);
        updateChannelTree(data.VMChannels);
        updateGraphPanel(data.VMHistory);
        updateStatsPanel(data.VMActivity);
        updateAllSpacers();
    } catch (err) {
        console.error("Polling error:", err);
    }
}

// ==========================================
// Update server info box
// ==========================================
function updateServerInfo(vmServer) {
    let bannerHtml = "";
    if (vmServer.HostBannerURL && vmServer.HostBannerURL.trim() !== "") {
        bannerHtml = `
        <div>
            <div class="banner-url">
                <a href="${vmServer.HostBannerURL}">
                    ${vmServer.HostBannerURL}
                </a>
            </div>
        </div>`;
    }

    let serverNameHtml = "";
    if (vmServer.HostConnectionLink && vmServer.HostConnectionLink.trim() !== "") {
        serverNameHtml = `
            <h1 id="server-name">
                <a href="teamspeak://${vmServer.HostConnectionLink}">
                    ${vmServer.Name}
                </a>
            </h1>`;
    } else {
        serverNameHtml = `<h1 id="server-name">${vmServer.Name}</h1>`;
    }

    document.querySelector(".server-info").innerHTML = `
        ${serverNameHtml}

        <div><span>User: </span> ${vmServer.ClientsOnline} / ${vmServer.MaxClients}</div>
        <div><span>Client Connections:</span> ${vmServer.ClientConnections}</div>
        <div><span>Uptime:</span> ${vmServer.UptimePretty}</div>
        <div><span>ChannelsOnline:</span> ${vmServer.ChannelsOnline}</div>

        ${bannerHtml}
    `;
}

// ==========================================
// Statistics panel (toggle + stat cells)
// ==========================================
function escapeHtml(s) {
    return String(s)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

// Format a unix timestamp in the VIEWER's own local timezone (the user's
// time zone, DST included) — never the server's — mirroring the
// MakeLastSeen contract: "HH:MM" for the current local day,
// "DD.MM. HH:MM" for an earlier day of the same local year,
// "DD.MM.YYYY HH:MM" for an earlier year.
// Returns an empty string for a missing or invalid timestamp.
function lastSeenTime(unixSeconds) {
    if (!unixSeconds || unixSeconds <= 0) {
        return "";
    }
    const t = new Date(unixSeconds * 1000);
    if (isNaN(t.getTime())) {
        return "";
    }
    const now = new Date();
    const clock = String(t.getHours()).padStart(2, "0") + ":" + String(t.getMinutes()).padStart(2, "0");
    if (t.getFullYear() === now.getFullYear() &&
        t.getMonth() === now.getMonth() &&
        t.getDate() === now.getDate()) {
        return clock;
    }
    const day = String(t.getDate()).padStart(2, "0") + "." + String(t.getMonth() + 1).padStart(2, "0") + ".";
    if (t.getFullYear() === now.getFullYear()) {
        return day + " " + clock;
    }
    return day + t.getFullYear() + " " + clock;
}

// Format a unix timestamp in the viewer's own local timezone with the date
// always shown, mirroring the MakeDateTime contract used for "since":
// "DD.MM. HH:MM" for the current year, "DD.MM.YYYY HH:MM" for an earlier
// year. Returns an empty string for a missing or invalid timestamp.
function connectedAtTime(unixSeconds) {
    if (!unixSeconds || unixSeconds <= 0) {
        return "";
    }
    const t = new Date(unixSeconds * 1000);
    if (isNaN(t.getTime())) {
        return "";
    }
    const now = new Date();
    const clock = String(t.getHours()).padStart(2, "0") + ":" + String(t.getMinutes()).padStart(2, "0");
    const day = String(t.getDate()).padStart(2, "0") + "." + String(t.getMonth() + 1).padStart(2, "0") + ".";
    if (t.getFullYear() === now.getFullYear()) {
        return day + " " + clock;
    }
    return day + t.getFullYear() + " " + clock;
}

function badgeHtmlFor(c) {
    if (c.Badge === "admin") {
        return '<span class="badge admin">ADMIN</span>';
    }
    if (c.Badge === "vip") {
        return '<span class="badge vip">VIP</span>';
    }
    return "";
}

function flagHtmlFor(c) {
    return c.Country
        ? '<span class="flag">' + escapeHtml(c.Country) + '</span>'
        : "";
}

function renderStatCell(c) {
    const badgeHtml = c.Badge
        ? badgeHtmlFor(c)
        : '<span class="stat-none">—</span>';

    let timeHtml;
    if (c.Online) {
        // Format the raw timestamp in the viewer's own local timezone
        // (the user's time, DST included). The pre-formatted Go string is
        // only a fallback when no unix value is available.
        const since = connectedAtTime(c.ConnectedAtUnix) || c.ConnectedAt;
        timeHtml = since
            ? '<span class="stat-de online">since <b>' + escapeHtml(since) + '</b></span>'
            : "";
    } else {
        const lastSeen = lastSeenTime(c.LastSeenAt) || c.LastSeen;
        timeHtml = '<span class="stat-de offline">last seen <b>' +
            escapeHtml(lastSeen) + '</b></span>';
    }

    // Log-like rows: the status color goes on the name and the timestamp
    // (green = online, red = offline).
    const nameClass = c.Online ? "stat-name online" : "stat-name offline";

    return '<div class="stat-grid">' +
        '<span class="stat-cell">' +
        '<span class="' + nameClass + '">' + escapeHtml(c.Nickname) + '</span>' +
        badgeHtml +
        '<span class="stat-pipe">|</span>' +
        flagHtmlFor(c) +
        timeHtml +
        '</span>' +
        '</div>';
}

// ==========================================
// Online-count graph (responsive SVG, TeamSpeak blue line + area)
// Time range selectable via buttons (24h / 7d / 14d / 1m), persisted in
// localStorage, capped by the server-provided maximum range.
// ==========================================
const GRAPH = {
    width: 640,
    height: 190,
    padL: 26,
    padR: 10,
    padT: 10,
    padB: 24,
    blue: "#0370DB",
};

const GRAPH_RANGES = {
    "24h": 24 * 3600 * 1000,
    "7d": 7 * 24 * 3600 * 1000,
    "14d": 14 * 24 * 3600 * 1000,
    "1m": 31 * 24 * 3600 * 1000,
};
// Bucket size (seconds, matching the unix-seconds timestamps of the
// history payload) per range. 0 means "no bucketing — draw the raw
// samples at their exact times".
//
// The 24h (recent) view uses NO bucketing: the statistics are captured
// in real time (every join and leave is folded in the moment it happens,
// plus a 30s sampler), so the samples are already time-accurate. Drawing
// them raw means a brief visit shows as a spike at its exact moment AND
// a departure drops the line back immediately — a coarse 15-minute
// bucket would hold the peak and keep the line "online" for up to 15
// minutes after the client left.
//
// The long ranges (7d/14d/1m) aggregate into fixed wall-clock buckets
// carrying the PEAK online count (max per bucket), so a month of
// history stays small while a brief visit is still visible as the bucket
// peak. Buckets are anchored at epoch boundaries, so a sample always
// lands in the same bucket no matter when the graph is rendered.
const GRAPH_BUCKETS = {
    "24h": 0,
    "7d": 3600,
    "14d": 2 * 3600,
    "1m": 8 * 3600,
};

const GRAPH_RANGE_KEY = "ts6viewer-graph-range";
const GRAPH_RANGE_DEFAULT = "24h";
let graphRange = GRAPH_RANGE_DEFAULT; // narrowed to the server's list on load
let lastHistory = []; // latest history payload, re-rendered on range change

const ACTIVITY_RANGE_KEY = "ts6viewer-activity-range";
const ACTIVITY_RANGE_DEFAULT = "24h";
let activityRange = ACTIVITY_RANGE_DEFAULT; // narrowed to the server's list on load
let statsActivityLimit = 0; // recent_activity_limit, 0 = unlimited (from the template)
let lastActivity = []; // latest activity payload, re-rendered on range change

function graphClock(ts) {
    const t = new Date(ts);
    return String(t.getHours()).padStart(2, "0") + ":" + String(t.getMinutes()).padStart(2, "0");
}

function graphDayLabel(ts) {
    const t = new Date(ts);
    return String(t.getDate()).padStart(2, "0") + "." + String(t.getMonth() + 1).padStart(2, "0") + ".";
}

// Vertical grid lines + labels for the selected range: every 3 hours for
// 24h, every 1-2 days otherwise (always whole-day marks).
function graphTicks(rangeMs, now) {
    const day = 24 * 3600 * 1000;
    const ticks = [];
    const push = (d) => {
        const ts = now - d * day;
        ticks.push({ ts, label: graphDayLabel(ts) });
    };
    if (rangeMs <= 24 * 3600 * 1000) {
        for (let h = 24; h >= 0; h -= 3) {
            const ts = now - h * 3600 * 1000;
            ticks.push({ ts, label: graphClock(ts) });
        }
    } else if (rangeMs <= 7 * day) {
        for (let d = 7; d >= 0; d--) push(d);
    } else if (rangeMs <= 14 * day) {
        for (let d = 14; d >= 0; d -= 2) push(d);
    } else {
        for (let d = 31; d >= 0; d -= 7) push(d);
    }
    return ticks;
}

// Round the max count up to a "nice" scale (steps of 1/2/5 × 10^n, always
// integer and at least 1) so grid lines and labels land on clean values and
// the mid ticks are never odd.
function niceGraphScale(maxN) {
    if (maxN <= 1) return { max: 1, step: 1 };
    const rough = maxN / 4;
    const mag = Math.pow(10, Math.floor(Math.log10(rough)));
    const norm = rough / mag;
    let step;
    if (norm <= 1) step = 1;
    else if (norm <= 2) step = 2;
    else if (norm <= 5) step = 5;
    else step = 10;
    step = Math.max(1, step * mag);
    return { max: Math.ceil(maxN / step) * step, step };
}

// Aggregate the samples of one graph window into fixed time buckets and
// return the points to draw, oldest first: one point per non-empty
// bucket, positioned at the bucket's start (an epoch-aligned wall-clock
// boundary such as 14:15:00) with the bucket's PEAK online count (max
// over its samples). The peak — not an average — is the meaningful
// statistic for "who was online": a client online for only a few seconds
// really was online, so the bucket must reach that count instead of
// diluting the spike into the surrounding samples. The result is always
// a whole count (a max of whole counts), so the drawn point and the
// tooltip value are consistent and land exactly on their grid line.
// Empty buckets (no page opened to poll) are dropped entirely, so gaps
// between the points are visible instead of being filled with "0".
// With bucketSec <= 0 the samples are returned unchanged.
function bucketHistorySamples(samples, bucketSec) {
    samples = samples || [];
    if (!bucketSec || bucketSec <= 0) {
        return samples.map((s) => ({ T: s.T, N: s.N }));
    }
    const byBucket = new Map();
    for (const s of samples) {
        const key = Math.floor(s.T / bucketSec);
        const b = byBucket.get(key);
        if (b) {
            if (s.N > b.max) {
                b.max = s.N;
            }
        } else {
            byBucket.set(key, { max: s.N });
        }
    }
    const pts = [];
    byBucket.forEach((b, key) => {
        pts.push({ T: key * bucketSec, N: b.max });
    });
    pts.sort((a, b) => a.T - b.T);
    return pts;
}

// ---------------------------------------------------------------------------
// Hover tooltip (pure helpers, DOM-free so they can be unit-tested locally)
// ---------------------------------------------------------------------------

// Index of the sample whose x coordinate is closest to the given x
// (ties go to the earlier sample).
function graphNearestIndex(samples, x) {
    let best = 0;
    let bestDist = Infinity;
    for (let i = 0; i < samples.length; i++) {
        const d = Math.abs(samples[i].x - x);
        if (d < bestDist) {
            bestDist = d;
            best = i;
        }
    }
    return best;
}

// Tooltip value line: the drawn points are bucket peaks — already whole
// counts — but there is no such thing as 0.9 of a person, so the tooltip
// always shows a whole number (rounding stays as a defensive no-op).
function graphOnlineLabel(n) {
    return Math.round(n) + " online";
}

// Tooltip time line: clock time for the 24h range (plus the date when the
// sample falls on a different day), date + clock for the longer ranges.
function graphHoverTimeLabel(ts, rangeMs) {
    const clock = graphClock(ts);
    if (rangeMs <= 24 * 3600 * 1000) {
        const t = new Date(ts);
        const now = new Date();
        const sameDay = t.getDate() === now.getDate() &&
            t.getMonth() === now.getMonth() &&
            t.getFullYear() === now.getFullYear();
        return sameDay ? clock : graphDayLabel(ts) + " " + clock;
    }
    return graphDayLabel(ts) + " " + clock;
}

// Tooltip box geometry: fixed size, centered on the point, 12px above it —
// flipped below the point when there is no room on top, clamped to the
// graph's edges.
function graphHoverBox(px, py, W, H) {
    const w = 108;
    const h = 40;
    const x = Math.max(2, Math.min(px - w / 2, W - w - 2));
    const above = py - h - 12 > 0;
    const y = above ? py - h - 12 : py + 12;
    return { w, h, x, y, above, vY: y + 16, tY: y + 31 };
}

function renderHistoryGraph(history) {
    const W = GRAPH.width;
    const H = GRAPH.height;
    const x0 = GRAPH.padL;
    const x1 = W - GRAPH.padR;
    const y0 = GRAPH.padT;
    const y1 = H - GRAPH.padB;

    const rangeMs = GRAPH_RANGES[graphRange] || GRAPH_RANGES["24h"];
    const now = Date.now();
    const t0 = now - rangeMs;

    // Window samples, then the points actually drawn: one peak point
    // per non-empty bucket of the range's bucket size (see GRAPH_BUCKETS).
    const windowSamples = (history || []).filter((p) => p && p.T >= t0 / 1000);
    const bucketSec = GRAPH_BUCKETS[graphRange] || GRAPH_BUCKETS["24h"];
    const samples = bucketHistorySamples(windowSamples, bucketSec);
    let maxN = 0;
    for (const p of samples) {
        if (p.N > maxN) maxN = p.N;
    }
    // Nice scale (>= 1) so grid lines land on clean integer steps.
    const scale = niceGraphScale(maxN);
    const yMax = scale.max;

    const sx = (ts) => x0 + ((ts - t0) / rangeMs) * (x1 - x0);
    const sy = (n) => y1 - (Math.max(0, Math.min(n, yMax)) / yMax) * (y1 - y0);

    // Grid + axis labels: horizontal every step, vertical per range step.
    let grid = "";
    const yTicks = [];
    for (let v = 0; v <= yMax + 1e-9; v += scale.step) {
        yTicks.push(Math.round(v));
    }
    for (const v of yTicks) {
        const y = sy(v).toFixed(1);
        grid += '<line class="graph-grid" x1="' + x0 + '" x2="' + x1 + '" y1="' + y + '" y2="' + y + '"/>';
        grid += '<text class="graph-label" x="' + (x0 - 6) + '" y="' + (sy(v) + 3).toFixed(1) + '" text-anchor="end">' + v + "</text>";
    }
    for (const tick of graphTicks(rangeMs, now)) {
        const x = sx(tick.ts).toFixed(1);
        grid += '<line class="graph-grid" x1="' + x + '" x2="' + x + '" y1="' + y0 + '" y2="' + y1 + '"/>';
        grid += '<text class="graph-label" x="' + x + '" y="' + (y1 + 16) + '" text-anchor="middle">' + tick.label + "</text>";
    }

    // Line + area fill.
    let series = "";
    let pts = [];
    if (samples.length === 1) {
        const px = sx(samples[0].T * 1000).toFixed(1);
        const py = sy(samples[0].N).toFixed(1);
        series += '<circle class="graph-dot" cx="' + px + '" cy="' + py + '" r="2.5"/>';
        pts = [[samples[0].T, +px, +py]];
    } else if (samples.length >= 2) {
        // Straight line through the bucket points: one direct segment per
        // consecutive pair (passes every point). Deliberately NOT smoothed —
        // a curve between the points would bulge above or below the data
        // and suggest values nobody was actually online.
        pts = samples.map((p) => [p.T, sx(p.T * 1000), sy(p.N)]);
        const raw = pts.map((p) => [p[1], p[2]]);
        let d = "M" + raw[0][0].toFixed(1) + " " + raw[0][1].toFixed(1);
        for (let i = 1; i < raw.length; i++) {
            d += "L" + raw[i][0].toFixed(1) + " " + raw[i][1].toFixed(1);
        }
        const first = raw[0][0].toFixed(1);
        const last = raw[raw.length - 1][0].toFixed(1);
        series += '<path class="graph-area" d="' + d + "L" + last + " " + y1 + "L" + first + " " + y1 + 'Z"/>';
        series += '<path class="graph-line" d="' + d + '"/>';
    }

    // Hidden hover tooltip (guideline, dot, box with the exact value and
    // time). Populated on mousemove by initGraphHover(); the sample list is
    // carried on the svg so the handler needs no other state.
    let hover = "";
    if (pts.length > 0) {
        const data = samples.map((s, i) => ({
            x: +pts[i][1].toFixed(2),
            y: +pts[i][2].toFixed(2),
            n: s.N,
            t: s.T,
        }));
        const attr = JSON.stringify(data);
        hover = '<g class="graph-hover" style="display:none" aria-hidden="true">' +
            '<line class="graph-hover-line" x1="0" x2="0" y1="' + y0 + '" y2="' + y1 + '"/>' +
            '<circle class="graph-hover-dot" cx="0" cy="0" r="3.5"/>' +
            '<rect class="graph-hover-box" x="0" y="0" width="108" height="40" rx="3"/>' +
            '<text class="graph-hover-value" x="0" y="0" text-anchor="middle">0 online</text>' +
            '<text class="graph-hover-time" x="0" y="0" text-anchor="middle">--:--</text>' +
            "</g>";
        // JSON contains only numbers here, so it is safe inside a
        // single-quoted HTML attribute.
        return '<svg class="stats-graph-svg" viewBox="0 0 ' + W + " " + H + '" role="img" aria-label="Online clients in the last ' + graphRange + '" data-samples=\'' + attr + '\'">' +
            grid + series + hover +
            "</svg>";
    }

    return '<svg class="stats-graph-svg" viewBox="0 0 ' + W + " " + H + '" role="img" aria-label="Online clients in the last ' + graphRange + '">' +
        grid + series +
        "</svg>";
}

// Range buttons: show the ranges the server allows, persist the user's
// choice, re-render the graph on change. The graph may appear in several
// places (main page and/or the statistics view); all instances share the
// same selected range.
function initGraphRanges() {
    const panels = Array.from(document.querySelectorAll(".graph-panel"));
    if (panels.length === 0) {
        return;
    }

    const allowed = (panels[0].dataset.ranges || "")
        .split(",")
        .map((r) => r.trim())
        .filter((r) => GRAPH_RANGES[r]);
    if (allowed.length === 0) {
        return;
    }

    // Restore the saved range if still allowed, otherwise the default.
    const saved = localStorage.getItem(GRAPH_RANGE_KEY);
    graphRange = allowed.includes(saved) ? saved : GRAPH_RANGE_DEFAULT;
    if (!allowed.includes(graphRange)) {
        graphRange = allowed[0];
    }

    const syncButtons = () => {
        document.querySelectorAll(".graph-ranges").forEach((holder) => {
            if (allowed.length === 1) {
                holder.innerHTML = "";
                return;
            }
            holder.innerHTML = allowed
                .map((r) =>
                    '<button type="button" class="graph-range" data-range="' + r +
                    '" aria-pressed="' + (r === graphRange) + '">' + r + "</button>"
                )
                .join("");
            holder.querySelector('[data-range="' + graphRange + '"]').classList.add("active");
        });
    };

    syncButtons();

    document.querySelectorAll(".graph-ranges").forEach((holder) => {
        holder.addEventListener("click", (e) => {
            const btn = e.target.closest(".graph-range");
            if (!btn || btn.dataset.range === graphRange) {
                return;
            }
            graphRange = btn.dataset.range;
            localStorage.setItem(GRAPH_RANGE_KEY, graphRange);
            syncButtons();
            renderGraphBodies();
        });
    });
}

// Range buttons for the Recent Activity header: same pattern as the graph
// (show the ranges the server allows, persist the user's choice), but the
// choice only narrows the already-fetched list instead of re-querying.
function initActivityRanges() {
    const holders = Array.from(document.querySelectorAll(".activity-ranges"));
    if (holders.length === 0) {
        return; // recent-activity section disabled by config
    }

    statsActivityLimit = Number(holders[0].dataset.limit) || 0; // 0 = unlimited

    const allowed = (holders[0].dataset.ranges || "")
        .split(",")
        .map((r) => r.trim())
        .filter((r) => GRAPH_RANGES[r]);
    if (allowed.length === 0) {
        return;
    }

    // Restore the saved range if still allowed, otherwise the default.
    const saved = localStorage.getItem(ACTIVITY_RANGE_KEY);
    activityRange = allowed.includes(saved) ? saved : ACTIVITY_RANGE_DEFAULT;
    if (!allowed.includes(activityRange)) {
        activityRange = allowed[0];
    }

    const syncButtons = () => {
        // Unlike the graph panel, always render the buttons even when only one
        // range is allowed (default config) so the active range stays visible.
        document.querySelectorAll(".activity-ranges").forEach((holder) => {
            holder.innerHTML = allowed
                .map((r) =>
                    '<button type="button" class="graph-range" data-range="' + r +
                    '" aria-pressed="' + (r === activityRange) + '">' + r + "</button>"
                )
                .join("");
            holder.querySelector('[data-range="' + activityRange + '"]').classList.add("active");
        });
    };

    syncButtons();

    document.querySelectorAll(".activity-ranges").forEach((holder) => {
        holder.addEventListener("click", (e) => {
            const btn = e.target.closest(".graph-range");
            if (!btn || btn.dataset.range === activityRange) {
                return;
            }
            activityRange = btn.dataset.range;
            localStorage.setItem(ACTIVITY_RANGE_KEY, activityRange);
            syncButtons();
            renderActivityList();
        });
    });
}

function renderGraphBodies() {
    // Fill every graph instance (main page and/or statistics view) with the
    // current series.
    document.querySelectorAll(".stats-graph").forEach((body) => {
        body.innerHTML = renderHistoryGraph(lastHistory);
    });
}

// Hover tooltip: show the exact sample value + time under the cursor.
// Delegated on document so it keeps working after every re-render (range
// change, data refresh) without re-binding.
function hideAllGraphHovers() {
    document.querySelectorAll(".stats-graph-svg").forEach((svg) => {
        const g = svg.querySelector(".graph-hover");
        if (g) {
            g.style.display = "none";
        }
    });
}

function showGraphHover(svg, clientX) {
    const g = svg.querySelector(".graph-hover");
    if (!g) {
        return;
    }
    let samples = [];
    try {
        samples = JSON.parse(svg.getAttribute("data-samples") || "[]");
    } catch (err) {
        return; // corrupt attribute: keep the tooltip hidden
    }
    if (!samples.length) {
        return;
    }

    // Convert the cursor's clientX into viewBox coordinates.
    const rect = svg.getBoundingClientRect();
    if (!rect || !rect.width) {
        return;
    }
    const svgX = ((clientX - rect.left) / rect.width) * GRAPH.width;
    const s = samples[graphNearestIndex(samples, svgX)];

    const line = g.querySelector(".graph-hover-line");
    const dot = g.querySelector(".graph-hover-dot");
    const box = g.querySelector(".graph-hover-box");
    const value = g.querySelector(".graph-hover-value");
    const time = g.querySelector(".graph-hover-time");
    if (!line || !dot || !box || !value || !time) {
        return;
    }

    const bx = graphHoverBox(s.x, s.y, GRAPH.width, GRAPH.height);
    const cx = (bx.x + bx.w / 2).toFixed(1);
    line.setAttribute("x1", s.x.toFixed(1));
    line.setAttribute("x2", s.x.toFixed(1));
    dot.setAttribute("cx", s.x.toFixed(1));
    dot.setAttribute("cy", s.y.toFixed(1));
    box.setAttribute("x", bx.x.toFixed(1));
    box.setAttribute("y", bx.y.toFixed(1));
    value.setAttribute("x", cx);
    value.setAttribute("y", bx.vY.toFixed(1));
    time.setAttribute("x", cx);
    time.setAttribute("y", bx.tY.toFixed(1));
    value.textContent = graphOnlineLabel(s.n);
    time.textContent = graphHoverTimeLabel(s.t * 1000, GRAPH_RANGES[graphRange] || GRAPH_RANGES["24h"]);
    g.style.display = "";
}

function initGraphHover() {
    document.addEventListener("mousemove", (e) => {
        const svg = e.target.closest ? e.target.closest(".stats-graph-svg") : null;
        if (svg) {
            showGraphHover(svg, e.clientX);
        }
    });

    document.addEventListener("mouseout", (e) => {
        const svg = e.target.closest ? e.target.closest(".stats-graph-svg") : null;
        if (!svg) {
            return;
        }
        // Only hide when actually leaving the svg (not when moving between
        // its child elements).
        const to = e.relatedTarget;
        if (!to || !svg.contains(to)) {
            const g = svg.querySelector(".graph-hover");
            if (g) {
                g.style.display = "none";
            }
        }
    });

    // Hide on scroll so the tooltip never lags behind the content.
    window.addEventListener("scroll", hideAllGraphHovers, true);
}

function updateGraphPanel(history) {
    if (!document.querySelector(".graph-panel")) {
        return; // graph section disabled by config
    }
    lastHistory = history || [];
    renderGraphBodies();
}

function updateStatsPanel(activity) {
    const panel = document.getElementById("statsPanel");
    if (!panel) {
        return; // recent-activity section disabled by config
    }

    const list = document.getElementById("activityList");
    if (!list) {
        return;
    }
    lastActivity = activity || [];
    renderActivityList();
}

// Timestamp (unix seconds) that orders an activity entry: the connection
// start for online clients, the last confirmed-online time for offline
// clients (both what is displayed) — mirrors the server's eventTime.
function activityEventTime(a) {
    return (a.Online && a.ConnectedAtUnix) || a.LastSeenAt || 0;
}

// Render the retained activity list narrowed to the selected range button,
// then apply the configured display cap (0 = unlimited). The server already
// keeps only the retention of the largest enabled range.
function renderActivityList() {
    const list = document.getElementById("activityList");
    if (!list) {
        return;
    }
    // GRAPH_RANGES are milliseconds, but activityEventTime() returns unix
    // SECONDS — convert the range to seconds so the comparison is consistent.
    const rangeSec = (GRAPH_RANGES[activityRange] || GRAPH_RANGES["24h"]) / 1000;
    const cutoff = Date.now() / 1000 - rangeSec;
    const visible = lastActivity
        .filter((a) => activityEventTime(a) >= cutoff)
        .slice(0, statsActivityLimit > 0 ? statsActivityLimit : undefined);
    list.innerHTML = visible.map(renderStatCell).join("");
}

function initStatsToggle() {
    const toggle = document.getElementById("statsToggle");
    const channels = document.getElementById("channels");
    const stats = document.getElementById("statsPanel");
    if (!toggle || !channels || !stats) return;

    function setView(showStats) {
        stats.hidden = !showStats;
        channels.hidden = showStats;
        toggle.classList.toggle("active", showStats);
        toggle.setAttribute("aria-pressed", String(showStats));
        toggle.title = showStats ? "Back to main page" : "Toggle statistics";
        // Swap icon: chart <-> channel tree
        toggle.querySelector("i").className = showStats
            ? "fa-solid fa-sitemap"
            : "fa-solid fa-chart-column";
    }

    toggle.addEventListener("click", () => {
        setView(stats.hidden);
    });
}

window.addEventListener("load", initStatsToggle);

// ==========================================
// Render channel tree
// ==========================================
function updateChannelTree(tree) {
    const container = document.getElementById("channels");
    container.innerHTML = renderChannels(tree);
    requestAnimationFrame(updateAllSpacers);
}

function renderChannels(nodes) {
    let html = "";
    for (const node of nodes) {
        html += renderChannel(node);
    }
    return html;
}

function renderChannel(ch) {
    if (ch.Type === 8) {
        return '<div class="row spacer blank-spacer"></div>';
    }

    const typeClass = ch.Type === 0 ? "channel" : "spacer";

    const alignClassName =
        ch.Align === 0 ? "spacer-left" :
        ch.Align === 1 ? "spacer-center" :
                         "spacer-right";

    let html = '<div class="row ' + typeClass;

    if (ch.Repeat) {
        html += ' repeat spacer-mono';
    }

    html += ' ' + alignClassName + '"';

    if (ch.Repeat) {
        html += ' data-pattern="' + ch.Name + '"';
    }

    html += '>' + ch.Name + '</div>';
    
    if (ch.Clients && ch.Clients.length > 0) {
        html += '<div class="children">';
        for (const c of ch.Clients) {
            let icon = '<i class="fa-solid fa-circle status-online"></i>';

            if (c.OutputMuted) {
                icon = '<i class="fa-solid fa-volume-xmark status-audio"></i>';
            } else if (c.MicMuted) {
                icon = '<i class="fa-solid fa-microphone-slash status-mic"></i>';
            } else {
                icon = '<i class="fa-solid fa-circle status-online"></i>';
            }

            html += '<div class="row client" data-online-time="' + (c.OnlineTimePretty || "") + '">' +
                    icon +
                    '<span class="client-name">' + c.Nickname + '</span>' +
                    badgeHtmlFor(c) +
                    flagHtmlFor(c) +
                    '</div>';

        }
        html += '</div>';
    }

    if (ch.Children && ch.Children.length > 0) {
        html += '<div class="children">';
        for (const child of ch.Children) {
            html += renderChannel(child);
        }
        html += '</div>';
    }

    return html;
}

// ==========================================
// Online-time hover tooltip (event delegation)
// ==========================================
function getOrCreateTooltip() {
    let tip = document.getElementById("onlineTooltip");
    if (!tip) {
        tip = document.createElement("div");
        tip.id = "onlineTooltip";
        document.body.appendChild(tip);
    }
    return tip;
}

function showOnlineTooltip(row, x, y) {
    const onlineTime = row.getAttribute("data-online-time");
    if (!onlineTime) {
        hideOnlineTooltip();
        return;
    }

    const tip = getOrCreateTooltip();
    tip.textContent = "Online: " + onlineTime;
    tip.style.display = "block";

    // Position close to the cursor, keep inside the viewport
    const rect = tip.getBoundingClientRect();
    let left = x + 10;
    let top = y + 12;

    if (left + rect.width > window.innerWidth - 8) {
        left = x - rect.width - 10;
    }
    if (top + rect.height > window.innerHeight - 8) {
        top = y - rect.height - 12;
    }

    tip.style.left = Math.max(8, left) + "px";
    tip.style.top = Math.max(8, top) + "px";
}

function hideOnlineTooltip() {
    const tip = document.getElementById("onlineTooltip");
    if (tip) {
        tip.style.display = "none";
    }
}

function initOnlineTooltip() {
    const channels = document.getElementById("channels");
    if (!channels) return;

    channels.addEventListener("mouseover", (e) => {
        const row = e.target.closest(".row.client");
        if (row) {
            showOnlineTooltip(row, e.clientX, e.clientY);
        }
    });

    channels.addEventListener("mousemove", (e) => {
        const row = e.target.closest(".row.client");
        if (row) {
            showOnlineTooltip(row, e.clientX, e.clientY);
        }
    });

    channels.addEventListener("mouseout", (e) => {
        const row = e.target.closest(".row.client");
        if (row && !row.contains(e.relatedTarget)) {
            hideOnlineTooltip();
        }
    });

    // Hide on scroll so the tooltip never lags behind the content
    window.addEventListener("scroll", hideOnlineTooltip, true);
}

// ==========================================
// Initial load
// ==========================================
window.addEventListener("load", () => {
    initGraphRanges();
    initActivityRanges();
    fetchViewerData();
    initGraphHover();
    initOnlineTooltip();
    requestAnimationFrame(updateAllSpacers);
});

let resizeTimer = null;

window.addEventListener("resize", () => {
    if (resizeTimer) clearTimeout(resizeTimer);

    resizeTimer = setTimeout(() => {
        requestAnimationFrame(() => {
            updateAllSpacers();
        });
    }, 150);
});
