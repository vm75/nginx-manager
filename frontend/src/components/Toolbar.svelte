<script>
  import { createEventDispatcher } from 'svelte';
  import { apiFetch } from '../lib/api';
  import ConfirmModal from './ConfirmModal.svelte';
  import AlertModal from './AlertModal.svelte';
  import logo from '../../logo.png';

  export let currentView;
  export let configStatus = 'unknown';
  export let testConfig = null;

  const dispatch = createEventDispatcher();

  let testOutput = '';
  let showTestModal = false;
  let testing = false;
  let reloading = false;
  let showConfirmModal = false;
  let showAlertModal = false;
  let alertTitle = '';
  let alertMessage = '';

  async function testNginx() {
    testing = true;
    try {
      const response = await apiFetch('/api/nginx/test', { method: 'POST' });
      const result = await response.json();
      testOutput = result.output;
      showTestModal = true;
      if (testConfig) {
        testConfig(result.success ? 'ok' : 'error');
      }
    } catch (error) {
      testOutput = `Error: ${error.message}`;
      showTestModal = true;
      if (testConfig) {
        testConfig('error');
      }
    } finally {
      testing = false;
    }
  }

  async function reloadNginx() {
    showConfirmModal = true;
  }

  async function handleConfirmReload() {
    showConfirmModal = false;
    reloading = true;
    try {
      const response = await apiFetch('/api/nginx/reload', { method: 'POST' });
      const result = await response.json();
      showAlert(result.success ? 'Success' : 'Error', result.success ? 'Nginx reloaded successfully' : `Error: ${result.output}`);
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    } finally {
      reloading = false;
    }
  }

  function handleCancelReload() {
    showConfirmModal = false;
  }

  function showAlert(title, message) {
    alertTitle = title;
    alertMessage = message;
    showAlertModal = true;
  }

  function handleAlertOk() {
    showAlertModal = false;
  }

  function switchView(view) {
    dispatch('viewChange', view);
  }
</script>

<div class="toolbar">
  <div class="toolbar-left">
    <h1><img src={logo} alt="Logo" style="height: 18px; vertical-align: middle;" /> Nginx Manager</h1>
  </div>

  <div class="toolbar-center">
    <button
      class:active={currentView === 'dashboard'}
      on:click={() => switchView('dashboard')}
    >
      🏠 Dashboard
    </button>
    <button
      class:active={currentView === 'editor'}
      on:click={() => switchView('editor')}
    >
      📝 Editor
    </button>
    <button
      class:active={currentView === 'logs'}
      on:click={() => switchView('logs')}
    >
      📊 Logs
    </button>
    <button
      class:active={currentView === 'certificates'}
      on:click={() => switchView('certificates')}
    >
      🔐 Certificates
    </button>
  </div>

  <div class="toolbar-right">
    <div class="config-status">
      <span class="status-label">Nginx Config Status:</span>
      <button
        class="status-icon {configStatus}"
        on:click={testNginx}
        disabled={testing}
        title={configStatus === 'ok' ? 'Config OK - Click to retest' : configStatus === 'error' ? 'Config Error - Click to retest' : 'Click to test config'}
      >
        {#if testing}
          ⏳
        {:else if configStatus === 'ok'}
          🟢
        {:else if configStatus === 'error'}
          🔴
        {:else}
          ⚪
        {/if}
      </button>
    </div>
    <button on:click={reloadNginx} disabled={reloading} class="reload-btn">
      {reloading ? '⏳ Reloading...' : '↻ Reload Nginx'}
    </button>
  </div>
</div>

{#if showTestModal}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-overlay"
    on:click={() => showTestModal = false}
    on:keydown={e => e.key === 'Escape' && (showTestModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && (showTestModal = false)}
      role="document"
    >
      <h3>Nginx Configuration Test</h3>
      <pre class="test-output">{testOutput}</pre>
      <div class="modal-buttons">
        <button on:click={() => showTestModal = false}>Close</button>
      </div>
    </div>
  </div>
{/if}

{#if showConfirmModal}
  <ConfirmModal
    title="Confirm Reload"
    message="Are you sure you want to reload nginx?"
    confirmText="Reload"
    cancelText="Cancel"
    on:confirm={handleConfirmReload}
    on:cancel={handleCancelReload}
    bind:show={showConfirmModal}
  />
{/if}

{#if showAlertModal}
  <AlertModal
    title={alertTitle}
    message={alertMessage}
    on:ok={handleAlertOk}
    bind:show={showAlertModal}
  />
{/if}

<style>
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #2d2d30;
    border-bottom: 1px solid #3c3c3c;
    padding: 8px 16px;
    gap: 16px;
    flex-wrap: wrap;
  }

  .toolbar-left h1 {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    color: #fff;
  }

  .toolbar-center, .toolbar-right {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .config-status {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .status-label {
    font-size: 13px;
    color: #fff;
    white-space: nowrap;
  }

  .status-icon {
    border: none;
    background: none;
    cursor: pointer;
    font-size: 16px;
    padding: 2px;
    border-radius: 3px;
    transition: background-color 0.2s;
  }

  .status-icon:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .status-icon:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  .status-icon.ok {
    color: #4caf50;
  }

  .status-icon.error {
    color: #f44336;
  }

  .status-icon.unknown {
    color: #9e9e9e;
  }

  button {
    padding: 6px 12px;
    font-size: 13px;
  }

  button.active {
    background: #0e639c;
  }

  button.reload-btn {
    background: #16825d;
  }

  button.reload-btn:hover {
    background: #1a9970;
  }

  .modal-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
  }

  .confirm-btn {
    background: #f44336;
  }

  .confirm-btn:hover {
    background: #d32f2f;
  }

  .test-output {
    background: #1e1e1e;
    padding: 16px;
    border-radius: 4px;
    color: #e0e0e0;
    overflow-x: auto;
    white-space: pre-wrap;
    font-family: 'Courier New', monospace;
    font-size: 13px;
    max-height: 400px;
    overflow-y: auto;
  }

  @media (max-width: 768px) {
    .toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-left, .toolbar-center, .toolbar-right {
      justify-content: center;
    }
  }
</style>
