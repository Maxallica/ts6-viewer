// ==========================================
// Refresh countdown (also controls data polling)
// ==========================================
let counter = refreshTime;

const refreshText = document.getElementById("refreshButtonText");

// Update countdown every second
setInterval(() => {
    counter--;
    refreshText.textContent = counter;

    // When countdown reaches zero → refresh data
    if (counter <= 0) {
        counter = refreshTime;
        fetchViewerData();
    }
}, 1000);


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
async function fetchViewerData() {
    try {
        const response = await fetch("/ts6viewer/data");
        const data = await response.json();

        updateServerInfo(data.Server);
        updateChannelTree(data.ChannelTree);

        updateAllSpacers();
    } catch (err) {
        console.error("Polling error:", err);
    }
}


// ==========================================
// Update server info box
// ==========================================
function updateServerInfo(server) {
    document.querySelector(".server-info").innerHTML = `
        <h1 id='server-name'>${server.Name}</h1>

        <div><span>User: </span> ${server.ClientsOnline} / ${server.MaxClients}</div>
        <div><span>Client Connections:</span> ${server.ClientConnections}</div>
        <div><span>Uptime:</span> ${server.UptimePretty}</div>
        <div><span>ChannelsOnline:</span> ${server.ChannelsOnline}</div>
        <div><a href="${server.HostBannerURL}">${server.HostBannerURL}</a></div>
    `;
}

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
            html += '<div class="row client"><span class="status-dot"></span>' +
                    c.Nickname +
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
// Initial load
// ==========================================
window.addEventListener("load", () => {
    fetchViewerData();
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
