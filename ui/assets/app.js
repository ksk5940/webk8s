let state = {
  namespaces: [],
  resourceTypes: [],
  namespace: "default",
  resource: "pods",
  resources: [],
  search: "",
  selectedPod: null,
  activeTab: "overview",
  sse: null,
};

function statusBadgePod(st) {
  const phase = st?.phase || "Unknown";
  if (phase === "Running") return `<span class="badge-run">Running</span>`;
  if (phase === "Pending") return `<span class="badge-warn">Pending</span>`;
  return `<span class="badge-bad">${escapeHtml(phase)}</span>`;
}

function statusBadgeWorkload(st) {
  const ready = st?.readyReplicas ?? 0;
  const rep = st?.replicas ?? 0;
  return `<span class="badge-soft">${ready}/${rep} ready</span>`;
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

      // ✅ Logs in NEW TAB
      if (tab === "logs") {
        if (!state.selectedPod) return;
        const url = `/logs.html?namespace=${encodeURIComponent(state.namespace)}&pod=${encodeURIComponent(state.selectedPod)}`;
        window.open(url, "_blank");
        return;
      }

      document.querySelectorAll(".drawer-tab").forEach(x => x.classList.remove("active"));
      a.classList.add("active");
      state.activeTab = tab;
      renderDrawer();
    };
  });

  state.namespaces = await apiGet("/api/namespaces");
  state.resourceTypes = await apiGet("/api/resources/types");

  UI.nsSel().innerHTML = state.namespaces.map(ns => `<option value="${ns}">${ns}</option>`).join("");

  if (state.namespaces.includes("default")) state.namespace = "default";
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
}

function buildResourceTabs() {
  const order = ["pods","deployments","replicasets","statefulsets","daemonsets","jobs","cronjobs","configmaps","secrets","services"];
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
  UI.title().textContent = capitalize(state.resource);
  state.resources = await apiGet(`/api/resources?namespace=${encodeURIComponent(state.namespace)}&type=${encodeURIComponent(state.resource)}`);
  renderTable();
}

function renderTable() {
  const items = state.resources.filter(x => !state.search || x.name.toLowerCase().includes(state.search));
  UI.countInfo().textContent = `${items.length} items`;

  const isPod = state.resource === "pods";

  UI.head().innerHTML = `
    <tr>
      <th style="width:42%">Name</th>
      <th>${isPod ? "Ready" : "Status"}</th>
      <th>${isPod ? "Status" : "Age"}</th>
      <th>${isPod ? "Restarts" : "Age"}</th>
      <th>Node</th>
      <th>Age</th>
    </tr>
  `;

  UI.body().innerHTML = items.map(x => {
    if (isPod) {
      const ready = x.status?.ready || "-";
      const st = statusBadgePod(x.status);
      const restarts = x.status?.restarts ?? 0;
      const node = x.status?.nodeName || "-";
      return `
        <tr class="pod-row" data-name="${escapeHtml(x.name)}">
          <td class="fw-bold">${escapeHtml(x.name)}</td>
          <td><span class="badge-soft">${escapeHtml(ready)}</span></td>
          <td>${st}</td>
          <td>${escapeHtml(restarts)}</td>
          <td class="fw-semibold">${escapeHtml(node)}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
        </tr>
      `;
    } else {
      const status = statusBadgeWorkload(x.status);
      return `
        <tr>
          <td class="fw-bold">${escapeHtml(x.name)}</td>
          <td>${status}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
          <td class="small-muted">-</td>
          <td class="small-muted">${fmtAge(x.creationTimestamp)}</td>
        </tr>
      `;
    }
  }).join("");

  if (isPod) {
    document.querySelectorAll(".pod-row").forEach(tr => {
      tr.onclick = () => openPod(tr.getAttribute("data-name"));
    });
  }
}

async function openPod(pod) {
  state.selectedPod = pod;
  state.activeTab = "overview";

  document.querySelectorAll(".drawer-tab").forEach(x => x.classList.remove("active"));
  document.querySelector(`.drawer-tab[data-tab="overview"]`).classList.add("active");

  UI.drawerTitle().textContent = `Pod Details: ${pod}`;
  UI.drawerSubtitle().textContent = `${state.namespace}`;
  openDrawer();

  await renderDrawer();
}

async function renderDrawer() {
  const ns = state.namespace;
  const pod = state.selectedPod;
  if (!pod) return;

  if (state.sse) { state.sse.close(); state.sse = null; }

  if (state.activeTab === "overview") {
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
      <div class="fw-bold mb-2">Containers</div>
      <ul class="mb-0">
        ${(d.containers||[]).map(c => `<li><b>${escapeHtml(c.name)}</b> <span class="small-muted">(Image: ${escapeHtml(c.image)})</span></li>`).join("")}
      </ul>
    `;

  } else if (state.activeTab === "events") {
    const events = await apiGet(`/api/pod/events?namespace=${encodeURIComponent(ns)}&pod=${encodeURIComponent(pod)}`);
    if (!events || events.length === 0) {
      UI.tabContent().innerHTML = `<div class="small-muted">No events found.</div>`;
      return;
    }

    UI.tabContent().innerHTML = `
      <div class="fw-bold mb-2">Events</div>
      <div class="table-responsive">
        <table class="table table-sm align-middle">
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

  } else if (state.activeTab === "metrics") {
    const m = await apiGet(`/api/pod/metrics?namespace=${encodeURIComponent(ns)}&pod=${encodeURIComponent(pod)}`);

    if (m.available === false) {
      UI.tabContent().innerHTML = `
        <div class="alert alert-warning">
          <b>Metrics not available</b><br/>
          ${escapeHtml(m.message || "")}
        </div>
      `;
      return;
    }

    const totals = calcMetricsTotals(m);
    UI.tabContent().innerHTML = `
      <div class="row g-3">
        <div class="col-6">
          <div class="p-3 border rounded-4 shadow-sm">
            <div class="small-muted">CPU</div>
            <div class="fs-3 fw-bold">${totals.cpu}</div>
          </div>
        </div>
        <div class="col-6">
          <div class="p-3 border rounded-4 shadow-sm">
            <div class="small-muted">Memory</div>
            <div class="fs-3 fw-bold">${totals.mem}</div>
          </div>
        </div>
      </div>
      <hr/>
      <div class="small-muted">Source: metrics.k8s.io</div>
    `;
  }
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
