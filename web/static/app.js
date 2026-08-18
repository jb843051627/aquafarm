const API = '/api';

function fetchJSON(url) {
    return fetch(url).then(r => r.json());
}

function initTabs() {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById('tab-' + btn.dataset.tab).classList.add('active');
        });
    });
}

function loadTanks() {
    fetchJSON(`${API}/tanks`).then(tanks => {
        const el = document.getElementById('tank-list');
        if (!tanks || tanks.length === 0) {
            el.innerHTML = '<p>No tanks found.</p>';
            return;
        }
        el.innerHTML = tanks.map(t => `
            <div class="card">
                <h3>${t.name}</h3>
                <div class="meta">Species: ${t.species}</div>
                <div class="meta">Capacity: ${t.capacity}L | Stock: ${t.stock_qty}</div>
                <div class="meta">Status: <span class="badge ${t.status}">${t.status}</span></div>
            </div>
        `).join('');
    });
}

function loadAlerts() {
    fetchJSON(`${API}/alerts/unresolved`).then(alerts => {
        const el = document.getElementById('alert-list');
        if (!alerts || alerts.length === 0) {
            el.innerHTML = '<p>No active alerts.</p>';
            return;
        }
        el.innerHTML = alerts.map(a => `
            <div class="alert-item ${a.severity}">
                <strong>${a.severity.toUpperCase()}</strong> - Tank #${a.tank_id} ${a.sensor_type}=${a.value}
                <p>${a.message}</p>
                <small>${new Date(a.created_at).toLocaleString()}</small>
            </div>
        `).join('');
    });
}

function loadFeedPlans() {
    fetchJSON(`${API}/feed-plans`).then(plans => {
        const el = document.getElementById('feed-list');
        if (!plans || plans.length === 0) {
            el.innerHTML = '<p>No feed plans found.</p>';
            return;
        }
        el.innerHTML = `<table><thead><tr>
            <th>ID</th><th>Tank</th><th>Type</th><th>Amount</th><th>Schedule</th><th>Active</th>
        </tr></thead><tbody>${plans.map(p => `
            <tr>
                <td>${p.id}</td><td>#${p.tank_id}</td><td>${p.feed_type}</td>
                <td>${p.amount}g</td><td>${p.schedule}</td>
                <td>${p.active ? 'Yes' : 'No'}</td>
            </tr>`).join('')}</tbody></table>`;
    });
}

function loadEquipment() {
    fetchJSON(`${API}/equipment`).then(items => {
        const el = document.getElementById('equipment-list');
        if (!items || items.length === 0) {
            el.innerHTML = '<p>No equipment found.</p>';
            return;
        }
        el.innerHTML = items.map(e => `
            <div class="card">
                <h3>${e.name}</h3>
                <div class="meta">Type: ${e.type} | Tank #${e.tank_id}</div>
                <div class="meta">Power: ${e.power_rating}W</div>
                <div class="meta">Status: <span class="badge ${e.status}">${e.status}</span></div>
            </div>
        `).join('');
    });
}

function loadOverview() {
    fetchJSON(`${API}/overview`).then(o => {
        const el = document.getElementById('overview-stats');
        el.innerHTML = `
            <div class="stat-grid">
                <div class="stat-card"><div class="value">${o.total_tanks || 0}</div><div class="label">Total Tanks</div></div>
                <div class="stat-card"><div class="value">${o.total_stock || 0}</div><div class="label">Total Stock</div></div>
                <div class="stat-card"><div class="value">${o.unresolved_alerts || 0}</div><div class="label">Active Alerts</div></div>
                <div class="stat-card"><div class="value">${o.equipment_running || 0}</div><div class="label">Running Equipment</div></div>
                <div class="stat-card"><div class="value">${o.equipment_fault || 0}</div><div class="label">Faulted Equipment</div></div>
                <div class="stat-card"><div class="value">${o.overdue_tasks || 0}</div><div class="label">Overdue Tasks</div></div>
            </div>`;
    });
}

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    loadTanks();
    loadAlerts();
    loadFeedPlans();
    loadEquipment();
    loadOverview();
});
