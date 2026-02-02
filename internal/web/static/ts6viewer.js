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
    if (!pattern) return;

    const width = element.clientWidth;
    const charWidth = getCharWidth(element);
    const repeatCount = Math.ceil(width / (pattern.length * charWidth));

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
    document.querySelectorAll(".full-width").forEach(repeatToWidth);
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
    document.getElementById("server-name").textContent = server.Name;

    document.querySelector(".server-info").innerHTML = `
        <div><span>User: </span> ${server.ClientsOnline} / ${server.MaxClients}</div>
        <div><span>Client Connections:</span> ${server.ClientConnections}</div>
        <div><span>Uptime:</span> ${server.UptimePretty}</div>
        <div><span>ChannelsOnline:</span> ${server.ChannelsOnline}</div>
        <div><span>HostBannerURL:</span> <a href="${server.HostBannerURL}">${server.HostBannerURL}</a></div>
    `;
}


// ==========================================
// Render channel tree
// ==========================================
function updateChannelTree(tree) {
    const container = document.getElementById("channels");
    container.innerHTML = renderChannels(tree);
}

function renderChannels(nodes) {
    let html = "";
    for (const node of nodes) {
        html += renderChannel(node);
    }
    return html;
}

function renderChannel(ch) {
    const typeClass = ch.Type === 0 ? "channel" : "spacer";

    const alignClassName =
        ch.Align === 0 ? "spacer-left" :
        ch.Align === 1 ? "spacer-center" :
                         "spacer-right";

    let html = '<div class="row ' + typeClass;

    if (ch.FullWidth) {
        html += ' full-width spacer-mono';
    }

    html += ' ' + alignClassName + '"';

    if (ch.FullWidth) {
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
    updateAllSpacers();
});

window.addEventListener("resize", updateAllSpacers);
