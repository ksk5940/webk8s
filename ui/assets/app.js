// API helper with better error handling
async function apiGet(path) {
  console.log("API GET:", path);
  try {
    const res = await fetch(path);
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`HTTP ${res.status}: ${text}`);
    }
    const data = await res.json();
    console.log("API response:", data);
    return data;
  } catch (err) {
    console.error("API error:", err);
    throw err;
  }
}

let state = {
  namespaces: [],
  resourceTypes: [],
  namespace: "default",
  resource: "pods",
  resources: [],
  search: "",
  selectedPod: null, // Keep for backward compatibility
  selectedResource: null, // New: { name, type }
  activeTab: "overview",
  sse: null,
};

function statusBadgePod(st) {
  const phase = st?.phase || "Unknown";
  if (phase === "Running") return `<span class="badge-run">Running</span>`;
  if (phase === "Pending") return `<span class="badge-warn">Pending</span>`;
  if (phase === "Completed") return `<span class="badge-run">Completed</span>`;
  return `<span class="badge-bad">${escapeHtml(phase)}</span>`;
}

function statusBadgeNode(st) {
  const ready = st?.ready || "Unknown";
  if (ready === "Ready") return `<span class="badge-run">Ready</span>`;
  return `<span class="badge-bad">NotReady</span>`;
}

function statusBadgeWorkload(st) {
  const ready = st?.readyReplicas ?? 0;
  const rep = st?.replicas ?? 0;
  const color = ready === rep ? "badge-run" : "badge-warn";
  return `<span class="${color}">${ready}/${rep}</span>`;
}

async function init() {
  document.getElementById("drawerClose").onclick = () => closeDrawer();
  document.getElementById("refreshBtn").onclick = () => refreshAll();

  UI.search().oninput = (e) => {
    state.search = e.target.value.trim().toLowerCase();
    renderTable();
  };

  // drawer tabs
  document.querySelectorAll(".drawer-tab").forEach(a => {
    a.onclick = (e) => {
      e.preventDefault();
      const tab = a.getAttribute("data-tab");

      // Logs in NEW TAB (only for pods)
      if (tab === "logs") {
        if (!state.selectedResource || state.selectedResource.type !== "pod") return;
        const url = `/logs.html?namespace=${encodeURIComponent(state.namespace)}&pod=${encodeURIComponent(state.selectedResource.name)}`;
        window.open(url, "_blank");
        return;
      }

      document.querySelectorAll(".drawer-tab").forEach(x => x.classList.remove("active"));
      a.classList.add("active");
      state.activeTab = tab;
      renderResourceDrawer();
    };
  });

  try {
    // Fetch namespaces with error handling
    console.log("Fetching namespaces...");
    state.namespaces = await apiGet("/api/namespaces");
    console.log("Namespaces received:", state.namespaces);
    
    state.resourceTypes = await apiGet("/api/resources/types");
    console.log("Resource types received:", state.resourceTypes);

    if (!state.namespaces || state.namespaces.length === 0) {
      alert("No namespaces found. Check RBAC permissions.");
      return;
    }

    UI.nsSel().innerHTML = state.namespaces.map(ns => `<option value="${ns}">${ns}</option>`).join("");

    if (state.namespaces.includes("default")) state.namespace = "default";
    else if (state.namespaces.includes("kube-system")) state.namespace = "kube-system";
    else if (state.namespaces.length > 0) state.namespace = state.namespaces[0];

    UI.nsSel().value = state.namespace;
    UI.nsSel().onchange = async (e) => {
      state.namespace = e.target.value;
      closeDrawer();
      await refreshAll();
    };

    buildResourceTabs();

    state.resource = "pods";
    setActiveResourceTab("pods");

    await refreshAll();
  } catch (err) {
    console.error("Initialization error:", err);
    alert("Failed to initialize: " + err.message);
  }
}

function buildResourceTabs() {
  const order = ["pods","nodes","deployments","replicasets","statefulsets","daemonsets","jobs","cronjobs","configmaps","services"];
  const typesByKey = {};
  state.resourceTypes.forEach(t => typesByKey[t.key] = t.label);

  const tabs = order.filter(k => typesByKey[k]).map(k => ({ key: k, label: typesByKey[k] }));

  UI.tabsContainer().innerHTML = tabs.map(t =>
    `<a href="#" class="tab-btn" data-key="${t.key}">${escapeHtml(t.label)}</a>`
  ).join("");

  document.querySelectorAll(".tab-btn").forEach(a => {
    a.onclick = async (e) => {
      e.preventDefault();
      const key = a.getAttribute("data-key");
      state.resource = key;
      closeDrawer();
      setActiveResourceTab(key);
      await refreshAll();
    };
  });
}

function setActiveResourceTab(key) {
  document.querySelectorAll(".tab-btn").forEach(x => x.classList.remove("active"));
  const el = document.querySelector(`.tab-btn[data-key="${key}"]`);
  if (el) el.classList.add("active");
}

async function refreshAll() {
  try {
    UI.title().textContent = capitalize(state.resource);
    state.resources = await apiGet(`/api/resources?namespace=${encodeURIComponent(state.namespace)}&type=${encodeURIComponent(state.resource)}`);
    renderTable();
  } catch (err) {
    console.error("Refresh error:", err);
    UI.countInfo().textContent = "Error loading data";
    UI.body().innerHTML = `<tr><td colspan="6" style="text-align:center;padding:40px;color:#dc3545;">Failed to load ${state.resource}: ${err.message}</td></tr>`;
  }
}

function renderTable() {
  const items = state.resources.filter(x => !state.search || x.name.toLowerCase().includes(state.search));
  UI.countInfo().textContent = `${items.length} items`;

  const isPod = state.resource === "pods";
  const isNode = state.resource === "nodes";

  if (isPod) {
    UI.head().innerHTML = `
      <tr>
        <th>Name</th>
        <th>Ready</th>
        <th>Status</th>
        <th>Restarts</th>
        <th>Node</th>
        <th>Age</th>
      </tr>
    `;

    UI.body().innerHTML = items.map(x => {
      const ready = x.status?.ready || "-";
      const st = statusBadgePod(x.status);
      const restarts = x.status?.restarts ?? 0;
      const node = x.status?.nodeName || "-";
      return `
        <tr class="resource-row" data-name="${escapeHtml(x.name)}" data-type="pod">
          <td class="fw-bold">${escapeHtml(x.name)}</td>
          <td>${escapeHtml(ready)}</td>
          <td>${st}</td>
          <td>${escapeHtml(restarts)}</td>
          <td class="small-muted">${escapeHtml(node)}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
        </tr>
      `;
    }).join("");

    document.querySelectorAll(".resource-row").forEach(tr => {
      tr.onclick = () => openResource(tr.getAttribute("data-name"), tr.getAttribute("data-type"));
    });
  } else if (isNode) {
    UI.head().innerHTML = `
      <tr>
        <th>Name</th>
        <th>Status</th>
        <th>Role</th>
        <th>Version</th>
        <th>Internal IP</th>
        <th>OS</th>
        <th>Age</th>
      </tr>
    `;

    UI.body().innerHTML = items.map(x => {
      const status = statusBadgeNode(x.status);
      const role = x.status?.role || "-";
      const version = x.status?.version || "-";
      const ip = x.status?.ip || "-";
      const os = x.status?.os || "-";
      return `
        <tr class="resource-row" data-name="${escapeHtml(x.name)}" data-type="node">
          <td class="fw-bold">${escapeHtml(x.name)}</td>
          <td>${status}</td>
          <td class="small-muted">${escapeHtml(role)}</td>
          <td class="small-muted">${escapeHtml(version)}</td>
          <td class="small-muted">${escapeHtml(ip)}</td>
          <td class="small-muted">${escapeHtml(os)}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
        </tr>
      `;
    }).join("");

    document.querySelectorAll(".resource-row").forEach(tr => {
      tr.onclick = () => openResource(tr.getAttribute("data-name"), tr.getAttribute("data-type"));
    });
  } else {
    UI.head().innerHTML = `
      <tr>
        <th>Name</th>
        <th>Ready</th>
        <th>Age</th>
      </tr>
    `;

    UI.body().innerHTML = items.map(x => {
      const status = statusBadgeWorkload(x.status);
      return `
        <tr class="resource-row" data-name="${escapeHtml(x.name)}" data-type="${state.resource}">
          <td class="fw-bold">${escapeHtml(x.name)}</td>
          <td>${status}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
        </tr>
      `;
    }).join("");

    document.querySelectorAll(".resource-row").forEach(tr => {
      tr.onclick = () => openResource(tr.getAttribute("data-name"), tr.getAttribute("data-type"));
    });
  }
}

async function openResource(name, type) {
  state.selectedResource = { name, type };
  state.activeTab = "overview";

  document.querySelectorAll(".drawer-tab").forEach(x => x.classList.remove("active"));
  document.querySelector(`.drawer-tab[data-tab="overview"]`).classList.add("active");

  // Update drawer tabs visibility based on resource type
  const logsTab = document.querySelector(`.drawer-tab[data-tab="logs"]`);
  const eventsTab = document.querySelector(`.drawer-tab[data-tab="events"]`);
  const metricsTab = document.querySelector(`.drawer-tab[data-tab="metrics"]`);

  if (type === "pod") {
    // Pods show all tabs
    logsTab.style.display = "block";
    eventsTab.style.display = "block";
    metricsTab.style.display = "block";
    state.selectedPod = name; // Keep backward compatibility
  } else {
    // Other resources don't show logs, events, metrics
    logsTab.style.display = "none";
    eventsTab.style.display = "none";
    metricsTab.style.display = "none";
  }

  UI.drawerTitle().textContent = `${capitalize(type)} Details: ${name}`;
  UI.drawerSubtitle().textContent = type === "node" ? "Cluster Node" : `Namespace: ${state.namespace}`;
  openDrawer();

  await renderResourceDrawer();
}

// Keep backward compatibility for pods
async function openPod(pod) {
  await openResource(pod, "pod");
}

async function renderResourceDrawer() {
  const { name, type } = state.selectedResource;
  if (!name || !type) return;

  if (state.sse) { state.sse.close(); state.sse = null; }

  if (state.activeTab === "overview") {
    if (type === "pod") {
      await renderPodOverview(name);
    } else if (type === "node") {
      await renderNodeOverview(name);
    } else {
      await renderGenericOverview(name, type);
    }
  } else if (state.activeTab === "events" && type === "pod") {
    await renderPodEvents(name);
  } else if (state.activeTab === "metrics" && type === "pod") {
    await renderPodMetrics(name);
  }
}

async function renderPodOverview(pod) {
  const ns = state.namespace;
  const d = await apiGet(`/api/pod?namespace=${encodeURIComponent(ns)}&pod=${encodeURIComponent(pod)}`);
  UI.tabContent().innerHTML = `
    <div class="kv-grid">
      <div class="kv-key">Node</div><div class="kv-val">${escapeHtml(d.node)}</div>
      <div class="kv-key">Status</div><div class="kv-val">${escapeHtml(d.phase)} ${d.reason ? "(" + escapeHtml(d.reason) + ")" : ""}</div>
      <div class="kv-key">IP</div><div class="kv-val">${escapeHtml(d.podIP || "-")}</div>
      <div class="kv-key">Ready</div><div class="kv-val">${escapeHtml(d.ready || "-")}</div>
      <div class="kv-key">Restarts</div><div class="kv-val">${escapeHtml(d.restarts || 0)}</div>
      <div class="kv-key">Started</div><div class="kv-val">${escapeHtml(d.startTime || "-")}</div>
    </div>
    <hr/>
    <div style="font-weight: 700; margin-bottom: 16px;">Containers</div>
    <ul style="list-style: none; padding: 0; margin: 0;">
      ${(d.containers||[]).map(c => `
        <li style="padding: 8px 0; border-bottom: 1px solid var(--border);">
          <div style="font-weight: 600;">${escapeHtml(c.name)}</div>
          <div class="small-muted">${escapeHtml(c.image)}</div>
        </li>
      `).join("")}
    </ul>
  `;
}

async function renderNodeOverview(nodeName) {
  // Find node details from current resources list
  const node = state.resources.find(r => r.name === nodeName);
  if (!node) {
    UI.tabContent().innerHTML = `<div class="small-muted">Node details not found.</div>`;
    return;
  }

  const status = node.status || {};
  const labels = node.labels || {};
  
  // Extract key labels
  const roleLabels = Object.keys(labels).filter(k => k.includes('node-role.kubernetes.io')).map(k => k.replace('node-role.kubernetes.io/', '')).join(', ') || 'worker';
  
  UI.tabContent().innerHTML = `
    <div class="kv-grid">
      <div class="kv-key">Status</div><div class="kv-val">${statusBadgeNode(status)}</div>
      <div class="kv-key">Role</div><div class="kv-val">${escapeHtml(roleLabels)}</div>
      <div class="kv-key">Internal IP</div><div class="kv-val">${escapeHtml(status.ip || "-")}</div>
      <div class="kv-key">Version</div><div class="kv-val">${escapeHtml(status.version || "-")}</div>
      <div class="kv-key">OS</div><div class="kv-val">${escapeHtml(status.os || "-")}</div>
      <div class="kv-key">Created</div><div class="kv-val">${fmtAge(node.creationTimestamp)}</div>
    </div>
    <hr/>
    <div style="font-weight: 700; margin-bottom: 16px;">Labels</div>
    <div style="display: flex; flex-wrap: wrap; gap: 8px;">
      ${Object.entries(labels).slice(0, 10).map(([k, v]) => `
        <span style="background: #e3f2fd; color: #1976d2; padding: 4px 8px; border-radius: 4px; font-size: 12px;">
          ${escapeHtml(k)}: ${escapeHtml(v)}
        </span>
      `).join("")}
    </div>
  `;
}

async function renderGenericOverview(name, type) {
  // Find resource details from current resources list
  const resource = state.resources.find(r => r.name === name);
  if (!resource) {
    UI.tabContent().innerHTML = `<div class="small-muted">Resource details not found.</div>`;
    return;
  }

  const status = resource.status || {};
  const labels = resource.labels || {};
  
  UI.tabContent().innerHTML = `
    <div class="kv-grid">
      <div class="kv-key">Name</div><div class="kv-val">${escapeHtml(resource.name)}</div>
      <div class="kv-key">Namespace</div><div class="kv-val">${escapeHtml(resource.namespace || state.namespace)}</div>
      <div class="kv-key">Created</div><div class="kv-val">${fmtAge(resource.creationTimestamp)}</div>
      ${Object.entries(status).map(([k, v]) => `
        <div class="kv-key">${escapeHtml(capitalize(k))}</div>
        <div class="kv-val">${typeof v === 'object' ? JSON.stringify(v) : escapeHtml(String(v))}</div>
      `).join("")}
    </div>
    <hr/>
    <div style="font-weight: 700; margin-bottom: 16px;">Labels</div>
    <div style="display: flex; flex-wrap: wrap; gap: 8px;">
      ${Object.entries(labels).slice(0, 10).map(([k, v]) => `
        <span style="background: #e3f2fd; color: #1976d2; padding: 4px 8px; border-radius: 4px; font-size: 12px;">
          ${escapeHtml(k)}: ${escapeHtml(v)}
        </span>
      `).join("")}
      ${Object.keys(labels).length > 10 ? `<span class="small-muted">+${Object.keys(labels).length - 10} more</span>` : ''}
    </div>
  `;
}

async function renderPodEvents(pod) {
  const ns = state.namespace;
  const events = await apiGet(`/api/pod/events?namespace=${encodeURIComponent(ns)}&pod=${encodeURIComponent(pod)}`);
  if (!events || events.length === 0) {
    UI.tabContent().innerHTML = `<div class="small-muted">No events found.</div>`;
    return;
  }

  UI.tabContent().innerHTML = `
    <div style="font-weight: 700; margin-bottom: 16px;">Events</div>
    <div style="overflow-x: auto;">
      <table class="table">
        <thead class="table-head">
          <tr><th>Type</th><th>Reason</th><th>Message</th><th>Time</th></tr>
        </thead>
        <tbody>
          ${events.map(x => `
            <tr>
              <td>${escapeHtml(x.type || "")}</td>
              <td class="fw-bold">${escapeHtml(x.reason || "")}</td>
              <td>${escapeHtml(x.message || "")}</td>
              <td class="small-muted">${escapeHtml(x.lastTimestamp || x.eventTime || x.firstTimestamp || "")}</td>
            </tr>
          `).join("")}
        </tbody>
      </table>
    </div>
  `;
}

async function renderPodMetrics(pod) {
  const ns = state.namespace;
  const m = await apiGet(`/api/pod/metrics?namespace=${encodeURIComponent(ns)}&pod=${encodeURIComponent(pod)}`);

  if (m.available === false) {
    UI.tabContent().innerHTML = `
      <div style="padding: 16px; background: #fff3cd; border: 1px solid #ffc107; border-radius: 6px; color: #856404;">
        <strong>Metrics not available</strong><br/>
        ${escapeHtml(m.message || "")}
      </div>
    `;
    return;
  }

  const totals = calcMetricsTotals(m);
  UI.tabContent().innerHTML = `
    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 24px;">
      <div style="padding: 20px; background: #f8f9fa; border: 1px solid var(--border); border-radius: 8px;">
        <div class="small-muted" style="margin-bottom: 8px;">CPU Usage</div>
        <div style="font-size: 28px; font-weight: 700; color: var(--text-primary);">${totals.cpu}</div>
      </div>
      <div style="padding: 20px; background: #f8f9fa; border: 1px solid var(--border); border-radius: 8px;">
        <div class="small-muted" style="margin-bottom: 8px;">Memory Usage</div>
        <div style="font-size: 28px; font-weight: 700; color: var(--text-primary);">${totals.mem}</div>
      </div>
    </div>
    <div class="small-muted">Source: metrics.k8s.io/v1beta1</div>
  `;
}

function calcMetricsTotals(metricsObj) {
  let cpuNano = 0, memBytes = 0;
  (metricsObj.containers || []).forEach(c => {
    cpuNano += parseCPUToNano(c.usage?.cpu);
    memBytes += parseMemToBytes(c.usage?.memory);
  });
  return { cpu: formatCPU(cpuNano), mem: formatMem(memBytes) };
}

function parseCPUToNano(s) {
  if (!s) return 0;
  if (s.endsWith("n")) return Number(s.slice(0, -1));
  if (s.endsWith("u")) return Number(s.slice(0, -1)) * 1000;
  if (s.endsWith("m")) return Number(s.slice(0, -1)) * 1000000;
  return Number(s) * 1000000000;
}

function parseMemToBytes(s) {
  if (!s) return 0;
  const num = Number(s.replace(/[a-zA-Z]/g, ""));
  if (s.endsWith("Ki")) return num * 1024;
  if (s.endsWith("Mi")) return num * 1024 * 1024;
  if (s.endsWith("Gi")) return num * 1024 * 1024 * 1024;
  return num;
}

function formatCPU(nano) {
  const millicores = nano / 1000000;
  return `${millicores.toFixed(0)}m`;
}

function formatMem(bytes) {
  const mib = bytes / (1024 * 1024);
  return `${mib.toFixed(0)}Mi`;
}

function capitalize(s) {
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
}

init();