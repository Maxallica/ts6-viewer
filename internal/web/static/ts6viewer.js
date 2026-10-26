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

    const samples = (history || []).filter((p) => p && p.T >= t0 / 1000);
    let maxN = 0;
    for (const p of samples) {
        if (p.N > maxN) maxN = p.N;
    }
    // Even ceiling (>= 1) so the 0 / mid / max ticks are always integers.
    const yMax = maxN <= 1 ? 1 : Math.ceil(maxN / 2) * 2;

    const sx = (ts) => x0 + ((ts - t0) / rangeMs) * (x1 - x0);
    const sy = (n) => y1 - (Math.max(0, Math.min(n, yMax)) / yMax) * (y1 - y0);

    // Grid + axis labels: horizontal at 0/mid/max, vertical per range step.
    let grid = "";
    const yTicks = [0, yMax];
    if (yMax / 2 >= 1) {
        yTicks.splice(1, 0, yMax / 2);
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
    if (samples.length === 1) {
        const px = sx(samples[0].T * 1000).toFixed(1);
        const py = sy(samples[0].N).toFixed(1);
        series += '<circle class="graph-dot" cx="' + px + '" cy="' + py + '" r="2.5"/>';
    } else if (samples.length >= 2) {
        let d = "";
        for (let i = 0; i < samples.length; i++) {
            const px = sx(samples[i].T * 1000).toFixed(1);
            const py = sy(samples[i].N).toFixed(1);
            d += (i === 0 ? "M" : "L") + px + " " + py;
        }
        const first = sx(samples[0].T * 1000).toFixed(1);
        const last = sx(samples[samples.length - 1].T * 1000).toFixed(1);
        series += '<path class="graph-area" d="' + d + "L" + last + " " + y1 + "L" + first + " " + y1 + 'Z"/>';
        series += '<path class="graph-line" d="' + d + '"/>';
    }

    return '<svg class="stats-graph-svg" viewBox="0 0 ' + W + " " + H + '" role="img" aria-label="Online clients in the last ' + graphRange + '">' +
        grid + series +
        "</svg>";
}

// Range buttons: show the ranges the server allows, persist the user's
// choice, re-render the graph on change.
function initGraphRanges() {
    const panel = document.getElementById("graphPanel");
    const holder = document.getElementById("graphRanges");
    if (!panel || !holder) {
        return;
    }

    const allowed = (panel.dataset.ranges || "")
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

    if (allowed.length === 1) {
        // Single range: no buttons needed.
        return;
    }

    holder.innerHTML = allowed
        .map((r) =>
            '<button type="button" class="graph-range" data-range="' + r +
            '" aria-pressed="' + (r === graphRange) + '">' + r + "</button>"
        )
        .join("");

    holder.addEventListener("click", (e) => {
        const btn = e.target.closest(".graph-range");
        if (!btn || btn.dataset.range === graphRange) {
            return;
        }
        graphRange = btn.dataset.range;
        localStorage.setItem(GRAPH_RANGE_KEY, graphRange);
        holder.querySelectorAll(".graph-range").forEach((b) => {
            b.classList.toggle("active", b === btn);
            b.setAttribute("aria-pressed", String(b === btn));
        });
        renderGraphBody();
    });

    holder.querySelector('[data-range="' + graphRange + '"]').classList.add("active");
}

function renderGraphBody() {
    const body = document.getElementById("graphBody");
    if (!body) {
        return;
    }
    body.innerHTML = renderHistoryGraph(lastHistory);
}

function updateGraphPanel(history) {
    const body = document.getElementById("graphBody");
    if (!body) {
        return; // graph section disabled by config
    }
    lastHistory = history || [];
    renderGraphBody();
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
        toggle.title = showStats ? "Back to channels" : "Toggle statistics";
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
