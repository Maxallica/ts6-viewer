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

// Format a unix timestamp as an absolute date + clock time, mirroring the
// Go MakeDateTime helper: "DD.MM. HH:MM" for the current year,
// "DD.MM.YYYY HH:MM" for an earlier year. The date is always shown so the
// merged activity list stays unambiguous.
function lastSeenTime(unixSeconds) {
    if (!unixSeconds) {
        return "";
    }
    const t = new Date(unixSeconds * 1000);
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
        timeHtml = c.ConnectedAt
            ? '<span class="stat-de online">since <b>' + escapeHtml(c.ConnectedAt) + '</b></span>'
            : "";
    } else {
        const lastSeen = c.LastSeen || lastSeenTime(c.LastSeenAt);
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
// history payload) per range. Before drawing, the raw samples are
// aggregated into fixed wall-clock buckets (average online count per
// bucket) so the shape of the curve no longer depends on how often the
// page happened to be polled: the 0/1 oscillation of quiet hours
// averages out, and fast evening jumps are spread over their bucket
// instead of drawing as near-vertical steps. Buckets are anchored at
// epoch boundaries, so a sample always lands in the same bucket no
// matter when the graph is rendered.
const GRAPH_BUCKETS = {
    "24h": 15 * 60,
    "7d": 3600,
    "14d": 2 * 3600,
    "1m": 8 * 3600,
};

const GRAPH_RANGE_KEY = "ts6viewer-graph-range";
const GRAPH_RANGE_DEFAULT = "24h";
let graphRange = GRAPH_RANGE_DEFAULT; // narrowed to the server's list on load
let lastHistory = []; // latest history payload, re-rendered on range change

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

// Smooth a polyline into a monotone cubic curve (Fritsch–Carlson): the
// result passes exactly through every point, has no overshoot (never
// leaves the data's own min/max), and — unlike a step chart or a plain
// polyline — shows no flat steps or sharp corners: each pair of
// consecutive samples is joined by one smooth, flowing span.
// pts: [[x, y], ...] (x strictly increasing). Returns an SVG path d.
function smoothPathD(pts) {
    const n = pts.length;
    if (n < 2) {
        return "M" + pts[0][0].toFixed(1) + " " + pts[0][1].toFixed(1);
    }

    // 1) Secant slopes between consecutive points.
    const dx = [];
    const m = [];
    for (let i = 0; i < n - 1; i++) {
        dx[i] = pts[i + 1][0] - pts[i][0];
        m[i] = (pts[i + 1][1] - pts[i][1]) / dx[i];
    }

    // 2) Interior tangent slopes: flat at direction changes, otherwise the
    //    average of the neighboring secants, then clamped per span so the
    //    cubic Hermite stays monotone (inside the data's range).
    const t = new Array(n);
    t[0] = m[0]; // endpoint tangent: free to differ from the interior ones
    t[n - 1] = m[n - 2];
    for (let i = 1; i < n - 1; i++) {
        t[i] = m[i - 1] * m[i] <= 0 ? 0 : (m[i - 1] + m[i]) / 2;
    }
    for (let i = 0; i < n - 1; i++) {
        if (m[i] === 0) {
            if (i > 0) t[i] = 0;
            if (i + 1 < n - 1) t[i + 1] = 0;
        } else {
            const a = t[i] / m[i];
            const b = t[i + 1] / m[i];
            const s = a * a + b * b;
            if (s > 9) {
                const f = 3 / Math.sqrt(s);
                if (i > 0) t[i] = f * a * m[i];
                if (i + 1 < n - 1) t[i + 1] = f * b * m[i];
            }
        }
    }

    // 3) Emit cubic Beziers; control points sit at 1/3 and 2/3 of the span.
    let d = "M" + pts[0][0].toFixed(1) + " " + pts[0][1].toFixed(1);
    for (let i = 0; i < n - 1; i++) {
        const h = dx[i] / 3;
        const tin = t[i];
        const tout = t[i + 1];
        d += "C" +
            (pts[i][0] + h).toFixed(1) + " " + (pts[i][1] + h * tin).toFixed(1) + " " +
            (pts[i + 1][0] - h).toFixed(1) + " " + (pts[i + 1][1] - h * tout).toFixed(1) + " " +
            pts[i + 1][0].toFixed(1) + " " + pts[i + 1][1].toFixed(1);
    }
    return d;
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
// boundary such as 14:15:00) with the bucket's average online count,
// rounded to a whole number — there is no such thing as 0.9 of a person,
// and a whole count keeps the drawn point and the tooltip value
// consistent (the point lands exactly on its grid line).
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
            b.sum += s.N;
            b.cnt += 1;
        } else {
            byBucket.set(key, { sum: s.N, cnt: 1 });
        }
    }
    const pts = [];
    byBucket.forEach((b, key) => {
        pts.push({ T: key * bucketSec, N: Math.round(b.sum / b.cnt) });
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

// Tooltip value line: the drawn points are bucket averages, so the count
// can be fractional — but there is no such thing as 0.9 of a person, so
// the tooltip always shows a whole number (rounded to the nearest one).
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

    // Window samples, then the points actually drawn: one averaged point
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
        // Smooth line through the bucket points: one monotone cubic span
        // per consecutive pair (passes every point, no corners, no
        // overshoot). Even, uniform span widths keep the curve calm no
        // matter how densely the raw samples were spaced.
        pts = samples.map((p) => [p.T, sx(p.T * 1000), sy(p.N)]);
        const raw = pts.map((p) => [p[1], p[2]]);
        const d = smoothPathD(raw);
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
    list.innerHTML = (activity || []).map(renderStatCell).join("");
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
