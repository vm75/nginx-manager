<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiFetch } from '../lib/api';
  import ConfirmModal from './ConfirmModal.svelte';
  import AlertModal from './AlertModal.svelte';

  let systemStats = {
    cpu: { usage: 0, cores: 0 },
    memory: { used: 0, total: 0, percent: 0 },
    disk: { used: 0, total: 0, percent: 0 },
    network: { rx: 0, tx: 0 }
  };

  let appIcons = [];
  let containerApps = {
    docker: [],
    podman: [],
    incus: []
  };

  let statsInterval;
  let containerStatsInterval;
  let showEditModal = false;
  let showDeleteModal = false;
  let showAlertModal = false;
  let alertTitle = '';
  let alertMessage = '';
  let containerRefreshSeconds = 5;
  let resourcesCollapsed = true;
  let smallIcons = false;

  let editingIcon = null;
  let iconForm = {
    id: null,
    name: '',
    url: '',
    icon: '',
    type: 'custom'
  };

  let deletingIcon = null;
  let selectedApp = null;

  onMount(async () => {
    // Load preferences
    const savedSmallIcons = localStorage.getItem('dashboard-smallIcons');
    if (savedSmallIcons !== null) {
      smallIcons = savedSmallIcons === 'true';
    }

    await loadSystemStats();
    await loadAppIcons();
    await loadContainerApps();

    // Update stats every 5 seconds
    statsInterval = setInterval(loadSystemStats, 5000);
    // Update container stats based on refresh interval
    containerStatsInterval = setInterval(loadContainerApps, containerRefreshSeconds * 1000);
  });

  onDestroy(() => {
    if (statsInterval) {
      clearInterval(statsInterval);
    }
    if (containerStatsInterval) {
      clearInterval(containerStatsInterval);
    }
  });

  async function loadSystemStats() {
    try {
      const response = await apiFetch('/api/system/stats');
      systemStats = await response.json();
    } catch (error) {
      console.error('Failed to load system stats:', error);
    }
  }

  async function loadAppIcons() {
    try {
      const response = await apiFetch('/api/dashboard/icons');
      const data = await response.json();
      appIcons = Array.isArray(data) ? data : [];
    } catch (error) {
      console.error('Failed to load app icons:', error);
      appIcons = [];
    }
  }

  async function loadContainerApps() {
    try {
      const response = await apiFetch('/api/containers/list');
      const data = await response.json();
      console.log('Container apps loaded:', data);
      // Ensure all arrays are initialized even if null
      containerApps = {
        docker: data.docker || [],
        podman: data.podman || [],
        incus: data.incus || []
      };
    } catch (error) {
      console.error('Failed to load container apps:', error);
      containerApps = {
        docker: [],
        podman: [],
        incus: []
      };
    }
  }

  function openEditModal(icon = null) {
    if (icon) {
      editingIcon = icon;
      iconForm = { ...icon };
    } else {
      editingIcon = null;
      iconForm = {
        id: null,
        name: '',
        url: '',
        icon: '🚀',
        type: 'custom'
      };
    }
    showEditModal = true;
  }

  function closeEditModal() {
    showEditModal = false;
    editingIcon = null;
  }

  async function saveIcon() {
    try {
      const endpoint = editingIcon ? '/api/dashboard/icons/update' : '/api/dashboard/icons/create';
      const response = await apiFetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(iconForm)
      });

      if (response.ok) {
        await loadAppIcons();
        closeEditModal();
      } else {
        const error = await response.json();
        showAlert('Error', error.error || 'Failed to save icon');
      }
    } catch (error) {
      showAlert('Error', error.message);
    }
  }

  function confirmDeleteIcon(icon) {
    deletingIcon = icon;
    showDeleteModal = true;
  }

  async function deleteIcon() {
    try {
      const response = await apiFetch('/api/dashboard/icons/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: deletingIcon.id })
      });

      if (response.ok) {
        await loadAppIcons();
        showDeleteModal = false;
        deletingIcon = null;
      } else {
        const error = await response.json();
        showAlert('Error', error.error || 'Failed to delete icon');
      }
    } catch (error) {
      showAlert('Error', error.message);
    }
  }

  async function performContainerAction(type, name, action) {
    try {
      const response = await apiFetch(`/api/containers/${type}/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name })
      });

      if (response.ok) {
        await loadContainerApps();
        showAlert('Success', `${action} operation completed successfully`);
      } else {
        const error = await response.json();
        showAlert('Error', error.error || `Failed to ${action} container`);
      }
    } catch (error) {
      showAlert('Error', error.message);
    }
  }

  function showAlert(title, message) {
    alertTitle = title;
    alertMessage = message;
    showAlertModal = true;
  }

  function openApp(url) {
    window.open(url, '_blank');
  }

  function toggleIconSize() {
    smallIcons = !smallIcons;
    localStorage.setItem('dashboard-smallIcons', smallIcons.toString());
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function getStatusColor(status) {
    if (status === 'running' || status === 'Running') return '#4caf50';
    if (status === 'stopped' || status === 'Stopped' || status === 'exited') return '#f44336';
    if (status === 'paused' || status === 'Paused') return '#ff9800';
    return '#9e9e9e';
  }
</script>

<div class="dashboard">
  <!-- System Resources Section -->
  <section class="section resources-section">
    <div class="section-header-collapsible" on:click={() => resourcesCollapsed = !resourcesCollapsed}>
      {#if !resourcesCollapsed}
        <h2>💻 System Resources</h2>
      {/if}
      {#if resourcesCollapsed}
        <div class="compact-stats">
          <span class="compact-stat">💻 {systemStats.cpu.usage.toFixed(1)}%</span>
          <span class="compact-stat">🧠 {systemStats.memory.percent.toFixed(1)}%</span>
          <span class="compact-stat">💾 {systemStats.disk.percent.toFixed(1)}%</span>
          <span class="compact-stat">🌐 ↓{formatBytes(systemStats.network.rx)}/s ↑{formatBytes(systemStats.network.tx)}/s</span>
        </div>
      {/if}
      <button class="collapse-btn" title={resourcesCollapsed ? 'Expand' : 'Collapse'}>
        {resourcesCollapsed ? '▼' : '▲'}
      </button>
    </div>
    {#if !resourcesCollapsed}
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-icon">💻</span>
          <span class="stat-title">CPU</span>
        </div>
        <div class="stat-value">{systemStats.cpu.usage.toFixed(1)}%</div>
        <div class="stat-bar">
          <div class="stat-bar-fill" style="width: {systemStats.cpu.usage}%; background: #2196f3;"></div>
        </div>
        <div class="stat-detail">{systemStats.cpu.cores} cores</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-icon">🧠</span>
          <span class="stat-title">Memory</span>
        </div>
        <div class="stat-value">{systemStats.memory.percent.toFixed(1)}%</div>
        <div class="stat-bar">
          <div class="stat-bar-fill" style="width: {systemStats.memory.percent}%; background: #4caf50;"></div>
        </div>
        <div class="stat-detail">{formatBytes(systemStats.memory.used)} / {formatBytes(systemStats.memory.total)}</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-icon">💾</span>
          <span class="stat-title">Disk</span>
        </div>
        <div class="stat-value">{systemStats.disk.percent.toFixed(1)}%</div>
        <div class="stat-bar">
          <div class="stat-bar-fill" style="width: {systemStats.disk.percent}%; background: #ff9800;"></div>
        </div>
        <div class="stat-detail">{formatBytes(systemStats.disk.used)} / {formatBytes(systemStats.disk.total)}</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-icon">🌐</span>
          <span class="stat-title">Network</span>
        </div>
        <div class="stat-value">
          <div class="network-stat">↓ {formatBytes(systemStats.network.rx)}/s</div>
          <div class="network-stat">↑ {formatBytes(systemStats.network.tx)}/s</div>
        </div>
      </div>
    </div>
    {/if}
  </section>

  <!-- App Icons Section -->
  <section class="section icons-section">
    <div class="section-header">
      <h2>Quick Launch</h2>
      <div class="section-header-actions">
        <button class="btn btn-secondary" on:click={toggleIconSize} title={smallIcons ? 'Use large icons' : 'Use small icons'}>
          {smallIcons ? '🔲' : '⬜'}
        </button>
        <button class="btn btn-primary" on:click={() => openEditModal()}>
          <span>+ Add App</span>
        </button>
      </div>
    </div>

    <div class="icons-grid" class:small-icons={smallIcons}>
      {#each appIcons as icon}
        <div class="app-icon">
          <button class="icon-button" on:click={() => openApp(icon.url)} title={icon.name}>
            <span class="icon-emoji">{icon.icon}</span>
            <span class="icon-name">{icon.name}</span>
          </button>
          <div class="icon-actions">
            <button class="icon-action" on:click={() => openEditModal(icon)} title="Edit">✏️</button>
            <button class="icon-action" on:click={() => confirmDeleteIcon(icon)} title="Delete">🗑️</button>
          </div>
        </div>
      {/each}

      {#if appIcons.length === 0}
        <div class="empty-state">
          <p>No apps configured. Click "Add App" to create shortcuts.</p>
        </div>
      {/if}
    </div>
  </section>

  <!-- Container Management Section -->
  <section class="section containers-section">
    <h2>Container Management</h2>

    <div class="container-tabs">
      {#if containerApps.docker && containerApps.docker.length > 0}
        <div class="container-group">
          <h3>🐳 Docker</h3>
          <div class="container-list">
            {#each containerApps.docker as container}
              <div class="container-item">
                <div class="container-info">
                  <div class="container-name">{container.name}</div>
                  <div class="container-meta">
                    <span class="container-status" style="background: {getStatusColor(container.status)}">
                      {container.status}
                    </span>
                    {#if container.image}
                      <span class="container-image">{container.image}</span>
                    {/if}
                  </div>
                  {#if container.ports && container.ports.length > 0}
                    <div class="container-ports">
                      <span class="port-label">Ports:</span>
                      {#each container.ports as port}
                        <span class="port-badge">{port}</span>
                      {/each}
                    </div>
                  {/if}
                  {#if container.size || container.virtualSize}
                    <div class="container-disk">
                      {#if container.size}
                        <span class="disk-item">💾 {container.size}</span>
                      {/if}
                      {#if container.virtualSize}
                        <span class="disk-item">📦 Virtual: {container.virtualSize}</span>
                      {/if}
                    </div>
                  {/if}
                  {#if container.status === 'running' && container.cpu !== undefined}
                    <div class="container-stats">
                      <span class="stat-item">💻 {container.cpu.toFixed(1)}%</span>
                      <span class="stat-item">🧠 {formatBytes(container.memory)} ({container.memUsage.toFixed(1)}%)</span>
                      <span class="stat-item">📊 ↓{formatBytes(container.diskRead)} ↑{formatBytes(container.diskWrite)}</span>
                    </div>
                  {/if}
                </div>
                <div class="container-actions">
                  {#if container.status === 'running'}
                    <button class="btn-action btn-stop" on:click={() => performContainerAction('docker', container.name, 'stop')}>
                      Stop
                    </button>
                    <button class="btn-action btn-restart" on:click={() => performContainerAction('docker', container.name, 'restart')}>
                      Restart
                    </button>
                  {:else}
                    <button class="btn-action btn-start" on:click={() => performContainerAction('docker', container.name, 'start')}>
                      Start
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      {#if containerApps.podman && containerApps.podman.length > 0}
        <div class="container-group">
          <h3>📦 Podman</h3>
          <div class="container-list">
            {#each containerApps.podman as container}
              <div class="container-item">
                <div class="container-info">
                  <div class="container-name">{container.name}</div>
                  <div class="container-meta">
                    <span class="container-status" style="background: {getStatusColor(container.status)}">
                      {container.status}
                    </span>
                    {#if container.image}
                      <span class="container-image">{container.image}</span>
                    {/if}
                  </div>
                </div>
                <div class="container-actions">
                  {#if container.status === 'running'}
                    <button class="btn-action btn-stop" on:click={() => performContainerAction('podman', container.name, 'stop')}>
                      Stop
                    </button>
                    <button class="btn-action btn-restart" on:click={() => performContainerAction('podman', container.name, 'restart')}>
                      Restart
                    </button>
                  {:else}
                    <button class="btn-action btn-start" on:click={() => performContainerAction('podman', container.name, 'start')}>
                      Start
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      {#if containerApps.incus && containerApps.incus.length > 0}
        <div class="container-group">
          <h3>🖥️ Incus</h3>
          <div class="container-list">
            {#each containerApps.incus as container}
              <div class="container-item">
                <div class="container-info">
                  <div class="container-name">{container.name}</div>
                  <div class="container-meta">
                    <span class="container-status" style="background: {getStatusColor(container.status)}">
                      {container.status}
                    </span>
                    {#if container.type}
                      <span class="container-type">{container.type}</span>
                    {/if}
                  </div>
                  {#if container.status === 'running'}
                    {#if container.memory}
                      <div class="container-stats">
                        {#if container.cpu !== undefined}
                          <span class="stat-item">💻 CPU: {container.cpu.toFixed(1)}s</span>
                        {/if}
                        <span class="stat-item">🧠 {formatBytes(container.memory)}</span>
                        {#if container.processes}
                          <span class="stat-item">⚙️ {container.processes} processes</span>
                        {/if}
                      </div>
                    {/if}
                    {#if container.ipv4 || container.ipv6}
                      <div class="container-network">
                        {#if container.ipv4}
                          <span class="network-item">🌐 IPv4: {container.ipv4}</span>
                        {/if}
                        {#if container.ipv6}
                          <span class="network-item">🌐 IPv6: {container.ipv6}</span>
                        {/if}
                      </div>
                    {/if}
                  {/if}
                </div>
                <div class="container-actions">
                  {#if container.status === 'running'}
                    <button class="btn-action btn-stop" on:click={() => performContainerAction('incus', container.name, 'stop')}>
                      Stop
                    </button>
                    <button class="btn-action btn-restart" on:click={() => performContainerAction('incus', container.name, 'restart')}>
                      Restart
                    </button>
                  {:else}
                    <button class="btn-action btn-start" on:click={() => performContainerAction('incus', container.name, 'start')}>
                      Start
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      {#if (!containerApps.docker || containerApps.docker.length === 0) &&
           (!containerApps.podman || containerApps.podman.length === 0) &&
           (!containerApps.incus || containerApps.incus.length === 0)}
        <div class="empty-state">
          <p>No containers found. Make sure Docker, Podman, or Incus is installed and accessible.</p>
        </div>
      {/if}
    </div>
  </section>
</div>

<!-- Edit Icon Modal -->
{#if showEditModal}
  <div class="modal-overlay" on:click={closeEditModal}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>{editingIcon ? 'Edit App' : 'Add App'}</h3>
        <button class="close-btn" on:click={closeEditModal}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>Name</label>
          <input type="text" bind:value={iconForm.name} placeholder="App Name" />
        </div>
        <div class="form-group">
          <label>URL</label>
          <input type="text" bind:value={iconForm.url} placeholder="https://example.com" />
        </div>
        <div class="form-group">
          <label>Icon (emoji or URL)</label>
          <input type="text" bind:value={iconForm.icon} placeholder="🚀" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" on:click={closeEditModal}>Cancel</button>
        <button class="btn btn-primary" on:click={saveIcon}>Save</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
  <ConfirmModal
    title="Delete App"
    message="Are you sure you want to delete '{deletingIcon?.name}'?"
    on:confirm={deleteIcon}
    on:cancel={() => { showDeleteModal = false; deletingIcon = null; }}
  />
{/if}

<!-- Alert Modal -->
{#if showAlertModal}
  <AlertModal
    title={alertTitle}
    message={alertMessage}
    on:ok={() => { showAlertModal = false; }}
  />
{/if}

<style>
  .dashboard {
    padding: 20px;
    overflow-y: auto;
    height: 100%;
    background: #1e1e1e;
  }

  .section {
    background: #252526;
    border-radius: 8px;
    padding: 24px;
    margin-bottom: 24px;
    border: 1px solid #3c3c3c;
  }

  .section h2 {
    margin: 0 0 20px 0;
    color: #e0e0e0;
    font-size: 20px;
    font-weight: 600;
  }

  .section h3 {
    margin: 0 0 16px 0;
    color: #e0e0e0;
    font-size: 16px;
    font-weight: 500;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .section-header h2 {
    margin: 0;
  }

  .section-header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .section-header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .section-header-collapsible {
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
    padding: 4px 0;
    user-select: none;
  }

  .section-header-collapsible h2 {
    margin: 0;
    font-size: 18px;
    flex-shrink: 0;
  }

  .compact-stats {
    display: flex;
    gap: 16px;
    flex: 1;
    font-size: 0.9em;
    color: #999;
    flex-wrap: wrap;
  }

  .compact-stat {
    white-space: nowrap;
    font-size: 0.85em;
  }

  @media (max-width: 768px) {
    .compact-stats {
      gap: 8px;
      font-size: 0.8em;
    }

    .compact-stat {
      font-size: 0.75em;
    }
  }

  .collapse-btn {
    background: none;
    border: none;
    color: #808080;
    font-size: 16px;
    cursor: pointer;
    padding: 8px;
    transition: color 0.2s;
    flex-shrink: 0;
  }

  .collapse-btn:hover {
    color: #e0e0e0;
  }

  /* System Resources */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 16px;
    margin-top: 20px;
  }

  .stat-card {
    background: #2d2d2d;
    border-radius: 6px;
    padding: 16px;
    border: 1px solid #3c3c3c;
  }

  .stat-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .stat-icon {
    font-size: 24px;
  }

  .stat-title {
    color: #b0b0b0;
    font-size: 14px;
    font-weight: 500;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    color: #e0e0e0;
    margin-bottom: 8px;
  }

  .stat-bar {
    height: 6px;
    background: #1e1e1e;
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .stat-bar-fill {
    height: 100%;
    transition: width 0.3s ease;
    border-radius: 3px;
  }

  .stat-detail {
    color: #808080;
    font-size: 12px;
  }

  .network-stat {
    font-size: 16px;
    margin: 4px 0;
  }

  /* App Icons */
  .icons-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 16px;
  }

  .icons-grid.small-icons {
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
    gap: 12px;
  }

  .app-icon {
    position: relative;
  }

  .icon-button {
    width: 100%;
    background: #2d2d2d;
    border: 1px solid #3c3c3c;
    border-radius: 8px;
    padding: 20px 12px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .small-icons .icon-button {
    padding: 12px 8px;
    gap: 6px;
  }

  .icon-button:hover {
    background: #333333;
    border-color: #4a4a4a;
    transform: translateY(-2px);
  }

  .icon-emoji {
    font-size: 40px;
  }

  .small-icons .icon-emoji {
    font-size: 28px;
  }

  .icon-name {
    color: #e0e0e0;
    font-size: 13px;
    text-align: center;
    word-break: break-word;
  }

  .small-icons .icon-name {
    font-size: 11px;
  }

  .icon-actions {
    display: flex;
    justify-content: center;
    gap: 8px;
    margin-top: 8px;
  }

  .small-icons .icon-actions {
    gap: 4px;
    margin-top: 4px;
  }

  .icon-action {
    background: #2d2d2d;
    border: 1px solid #3c3c3c;
    border-radius: 4px;
    padding: 4px 8px;
    cursor: pointer;
    font-size: 16px;
    transition: all 0.2s;
  }

  .small-icons .icon-action {
    padding: 2px 6px;
    font-size: 14px;
  }

  .icon-action:hover {
    background: #333333;
    border-color: #4a4a4a;
  }

  /* Container Management */
  .container-tabs {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .container-group {
    background: #2d2d2d;
    border-radius: 6px;
    padding: 16px;
    border: 1px solid #3c3c3c;
  }

  .container-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .container-item {
    background: #1e1e1e;
    border: 1px solid #3c3c3c;
    border-radius: 6px;
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  .container-info {
    flex: 1;
    min-width: 0;
  }

  .container-name {
    color: #e0e0e0;
    font-weight: 500;
    font-size: 14px;
    margin-bottom: 6px;
  }

  .container-meta {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .container-status {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 12px;
    color: white;
    font-weight: 500;
  }

  .container-image,
  .container-type {
    font-size: 12px;
    color: #808080;
  }

  .container-stats {
    display: flex;
    gap: 12px;
    margin-top: 8px;
    flex-wrap: wrap;
  }

  .stat-item {
    font-size: 11px;
    color: #b0b0b0;
    background: #1e1e1e;
    padding: 4px 8px;
    border-radius: 4px;
    white-space: nowrap;
  }

  .container-ports {
    display: flex;
    gap: 6px;
    margin-top: 6px;
    flex-wrap: wrap;
    align-items: center;
  }

  .port-label {
    font-size: 11px;
    color: #808080;
    font-weight: 500;
  }

  .port-badge {
    font-size: 11px;
    color: #4caf50;
    background: rgba(76, 175, 80, 0.15);
    padding: 2px 8px;
    border-radius: 4px;
    border: 1px solid rgba(76, 175, 80, 0.3);
    font-weight: 500;
  }

  .container-disk {
    display: flex;
    gap: 12px;
    margin-top: 6px;
    flex-wrap: wrap;
  }

  .disk-item {
    font-size: 11px;
    color: #b0b0b0;
    background: rgba(255, 152, 0, 0.15);
    padding: 2px 8px;
    border-radius: 4px;
    border: 1px solid rgba(255, 152, 0, 0.3);
  }

  .container-network {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 6px;
  }

  .network-item {
    font-size: 11px;
    color: #b0b0b0;
    background: rgba(33, 150, 243, 0.15);
    padding: 2px 8px;
    border-radius: 4px;
    border: 1px solid rgba(33, 150, 243, 0.3);
    display: inline-block;
  }

  .container-actions {
    display: flex;
    gap: 8px;
  }

  .btn-action {
    padding: 6px 12px;
    border: 1px solid #3c3c3c;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 500;
    transition: all 0.2s;
  }

  .btn-start {
    background: #4caf50;
    color: white;
  }

  .btn-start:hover {
    background: #45a049;
  }

  .btn-stop {
    background: #f44336;
    color: white;
  }

  .btn-stop:hover {
    background: #da190b;
  }

  .btn-restart {
    background: #ff9800;
    color: white;
  }

  .btn-restart:hover {
    background: #fb8c00;
  }

  /* Buttons */
  .btn {
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s;
    border: none;
  }

  .btn-primary {
    background: #0e639c;
    color: white;
  }

  .btn-primary:hover {
    background: #1177bb;
  }

  .btn-secondary {
    background: #3c3c3c;
    color: #e0e0e0;
    border: 1px solid #4a4a4a;
  }

  .btn-secondary:hover {
    background: #4a4a4a;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: #252526;
    border-radius: 8px;
    border: 1px solid #3c3c3c;
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    padding: 16px 20px;
    border-bottom: 1px solid #3c3c3c;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .modal-header h3 {
    margin: 0;
    color: #e0e0e0;
    font-size: 18px;
  }

  .close-btn {
    background: none;
    border: none;
    color: #808080;
    font-size: 28px;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
  }

  .close-btn:hover {
    background: #3c3c3c;
    color: #e0e0e0;
  }

  .modal-body {
    padding: 20px;
    overflow-y: auto;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    color: #b0b0b0;
    font-size: 13px;
    margin-bottom: 6px;
    font-weight: 500;
  }

  .form-group input {
    width: 100%;
    background: #1e1e1e;
    border: 1px solid #3c3c3c;
    border-radius: 4px;
    padding: 8px 12px;
    color: #e0e0e0;
    font-size: 14px;
  }

  .form-group input:focus {
    outline: none;
    border-color: #0e639c;
  }

  .modal-footer {
    padding: 16px 20px;
    border-top: 1px solid #3c3c3c;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  /* Empty State */
  .empty-state {
    padding: 40px 20px;
    text-align: center;
    color: #808080;
  }

  /* Responsive */
  @media (max-width: 768px) {
    .dashboard {
      padding: 12px;
    }

    .section {
      padding: 16px;
      margin-bottom: 16px;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .icons-grid {
      grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
      gap: 12px;
    }

    .container-item {
      flex-direction: column;
      align-items: flex-start;
    }

    .container-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
