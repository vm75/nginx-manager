<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import { apiFetch } from '../lib/api';

  let activeTab = 'error';
  let accessLog = '';
  let errorLog = '';
  let certObtainLog = '';
  let autoRefresh = true;
  let refreshInterval;
  let logContentElement;

  onMount(() => {
    loadLogs();
    if (autoRefresh) {
      startAutoRefresh();
    }
  });

  onDestroy(() => {
    stopAutoRefresh();
  });

  $: if (autoRefresh) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }

  function startAutoRefresh() {
    stopAutoRefresh();
    refreshInterval = setInterval(loadLogs, 5000);
  }

  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }

  async function loadLogs() {
    try {
      if (activeTab === 'access') {
        const response = await apiFetch('/api/logs/access?lines=500');
        accessLog = await response.text();
      } else if (activeTab === 'error') {
        const response = await apiFetch('/api/logs/error?lines=500');
        errorLog = await response.text();
      } else if (activeTab === 'cert-obtain') {
        const response = await apiFetch('/api/logs/cert-obtain?lines=500');
        certObtainLog = await response.text();
      }
      await tick();
      scrollToBottom();
    } catch (error) {
      console.error('Error loading logs:', error);
    }
  }

  function scrollToBottom() {
    if (logContentElement) {
      logContentElement.scrollTop = logContentElement.scrollHeight;
    }
  }

  function switchTab(tab) {
    activeTab = tab;
    loadLogs();
  }

  function clearLog() {
    if (activeTab === 'access') {
      accessLog = '';
    } else if (activeTab === 'error') {
      errorLog = '';
    } else if (activeTab === 'cert-obtain') {
      certObtainLog = '';
    }
  }
</script>

<div class="logs-container">
  <div class="logs-header">
    <div class="tabs">
      <button
        class:active={activeTab === 'access'}
        on:click={() => switchTab('access')}
      >
        📊 Access Log
      </button>
      <button
        class:active={activeTab === 'error'}
        on:click={() => switchTab('error')}
      >
        ⚠️ Error Log
      </button>
      <button
        class:active={activeTab === 'cert-obtain'}
        on:click={() => switchTab('cert-obtain')}
      >
        🔐 Certificate Obtain
      </button>
    </div>

    <div class="log-actions">
      <label class="auto-refresh">
        <input type="checkbox" bind:checked={autoRefresh} />
        Auto-refresh (5s)
      </label>
      <button on:click={loadLogs}>🔄 Refresh</button>
      <button on:click={clearLog} class="secondary">🗑️ Clear</button>
    </div>
  </div>

  <div class="log-content" bind:this={logContentElement}>
    {#if activeTab === 'access'}
      <pre class="log-text">{accessLog || 'No access log entries'}</pre>
    {:else if activeTab === 'error'}
      <pre class="log-text">{errorLog || 'No error log entries'}</pre>
    {:else if activeTab === 'cert-obtain'}
      <pre class="log-text">{certObtainLog || 'No certificate obtain log entries'}</pre>
    {/if}
  </div>
</div>

<style>
  .logs-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: #1e1e1e;
  }

  .logs-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    background: #2d2d30;
    border-bottom: 1px solid #3c3c3c;
    flex-wrap: wrap;
    gap: 8px;
  }

  .tabs {
    display: flex;
    gap: 4px;
  }

  .tabs button {
    padding: 6px 12px;
    font-size: 13px;
    background: transparent;
  }

  .tabs button:hover {
    background: #3c3c3c;
  }

  .tabs button.active {
    background: #0e639c;
  }

  .log-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .auto-refresh {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #cccccc;
    cursor: pointer;
  }

  .auto-refresh input[type="checkbox"] {
    cursor: pointer;
    width: auto;
  }

  .log-content {
    flex: 1;
    overflow: auto;
    padding: 16px;
  }

  .log-text {
    font-family: 'Courier New', monospace;
    font-size: 12px;
    color: #cccccc;
    white-space: pre-wrap;
    word-wrap: break-word;
    line-height: 1.5;
    margin: 0;
  }

  @media (max-width: 768px) {
    .logs-header {
      flex-direction: column;
      align-items: stretch;
    }

    .tabs, .log-actions {
      justify-content: center;
    }
  }
</style>
