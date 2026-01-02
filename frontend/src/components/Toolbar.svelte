<script>
  import { createEventDispatcher } from 'svelte';

  export let currentView;

  const dispatch = createEventDispatcher();

  let testOutput = '';
  let showTestModal = false;
  let testing = false;
  let reloading = false;

  async function testNginx() {
    testing = true;
    try {
      const response = await fetch('/api/nginx/test', { method: 'POST' });
      const result = await response.json();
      testOutput = result.output;
      showTestModal = true;
    } catch (error) {
      testOutput = `Error: ${error.message}`;
      showTestModal = true;
    } finally {
      testing = false;
    }
  }

  async function reloadNginx() {
    if (!confirm('Are you sure you want to reload nginx?')) return;

    reloading = true;
    try {
      const response = await fetch('/api/nginx/reload', { method: 'POST' });
      const result = await response.json();
      alert(result.success ? 'Nginx reloaded successfully' : `Error: ${result.output}`);
    } catch (error) {
      alert(`Error: ${error.message}`);
    } finally {
      reloading = false;
    }
  }

  function refresh() {
    dispatch('refresh');
  }

  function switchView(view) {
    dispatch('viewChange', view);
  }
</script>

<div class="toolbar">
  <div class="toolbar-left">
    <h1>🔧 Nginx Config Editor</h1>
  </div>

  <div class="toolbar-center">
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
    <button on:click={refresh} title="Refresh file tree">
      🔄 Refresh
    </button>
    <button on:click={testNginx} disabled={testing}>
      {testing ? '⏳ Testing...' : '✓ Test Config'}
    </button>
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
