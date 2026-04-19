let currentFile = "";
let config = {};

// Initialization
document.addEventListener('DOMContentLoaded', async () => {
    await refreshConfig();
    await refreshAppsList();
});

async function refreshConfig() {
    config = await window.GetConfig();
    document.getElementById('set-desktop-icon').checked = config.desktop_icon;
    document.getElementById('set-no-extract').checked = config.no_extract_icon;
    document.getElementById('current-desktop-path').innerText = config.desktop_path;
}

async function refreshAppsList() {
    const list = document.getElementById('apps-list');
    list.innerHTML = '<div class="loading">Lade Apps...</div>';
    
    const apps = await window.GetInstalledApps();
    
    if (!apps || apps.length === 0) {
        list.innerHTML = '<div class="empty-state">Keine installierten AppImages gefunden.</div>';
        return;
    }

    list.innerHTML = '';
    apps.forEach(app => {
        const iconSrc = app.IconPath ? `file://${app.IconPath}` : 'icon-placeholder.png';
        const card = document.createElement('div');
        card.className = 'app-card';
        card.innerHTML = `
            <img src="${iconSrc}" class="app-icon" onerror="this.src='https://cdn-icons-png.flaticon.com/512/25/25231.png'">
            <div class="app-info">
                <h4>${app.Name}</h4>
                <p>${app.DesktopFile}</p>
            </div>
            <div class="app-actions">
                <button class="btn-delete" title="Deinstallieren" onclick="uninstallApp('${app.Name}')">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-trash-2"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>
                </button>
            </div>
        `;
        list.appendChild(card);
    });
}

function showSection(id) {
    document.querySelectorAll('section').forEach(s => s.classList.add('hidden'));
    document.getElementById(`section-${id}`).classList.remove('hidden');
    
    document.querySelectorAll('.sidebar nav button').forEach(b => b.classList.remove('active'));
    document.getElementById(`nav-${id}`).classList.add('active');
    
    if (id === 'apps') refreshAppsList();
}

async function pickFile() {
    const path = await window.SelectAppImage();
    if (path) {
        currentFile = path;
        document.getElementById('selected-file-label').innerText = path.split('/').pop();
        document.getElementById('install-form').classList.remove('hidden');
        
        // Pre-fill name
        let name = path.split('/').pop().replace(/\.AppImage$/i, '');
        document.getElementById('app-name').value = name;
    }
}

async function startInstall() {
    const name = document.getElementById('app-name').value;
    const desc = document.getElementById('app-desc').value;
    const cats = document.getElementById('app-cats').value;
    const args = document.getElementById('app-args').value;

    if (!name) {
        showToast("Bitte gib einen Namen an", "error");
        return;
    }

    showToast("Installiere...", "");
    const result = await window.InstallApp(currentFile, name, desc, "", args, cats);
    
    if (result.startsWith("FEHLER:")) {
        showToast(result, "error");
    } else {
        showToast(`${result} erfolgreich installiert!`, "success");
        resetForm();
        showSection('apps');
    }
}

function resetForm() {
    currentFile = "";
    document.getElementById('selected-file-label').innerText = "AppImage auswählen oder hierher ziehen";
    document.getElementById('install-form').classList.add('hidden');
    document.getElementById('app-name').value = "";
    document.getElementById('app-desc').value = "";
    document.getElementById('app-cats').value = "";
    document.getElementById('app-args').value = "";
}

async function uninstallApp(name) {
    if (confirm(`Möchtest du "${name}" wirklich deinstallieren?`)) {
        const res = await window.UninstallApp(name);
        if (res === "OK") {
            showToast("Erfolgreich deinstalliert", "success");
            refreshAppsList();
        } else {
            showToast(res, "error");
        }
    }
}

async function pickDesktopFolder() {
    const path = await window.SelectFolder();
    if (path) {
        config.desktop_path = path;
        document.getElementById('current-desktop-path').innerText = path;
    }
}

async function saveSettings() {
    config.desktop_icon = document.getElementById('set-desktop-icon').checked;
    config.no_extract_icon = document.getElementById('set-no-extract').checked;
    
    const res = await window.SaveConfig(config);
    if (res === "OK") {
        showToast("Einstellungen gespeichert", "success");
    } else {
        showToast(res, "error");
    }
}

function showToast(msg, type) {
    const t = document.getElementById('toast');
    t.innerText = msg;
    t.className = `toast ${type}`;
    t.classList.remove('hidden');
    
    setTimeout(() => {
        t.classList.add('hidden');
    }, 3000);
}
