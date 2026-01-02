<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { apiFetch } from '../lib/api';
  import AlertModal from './AlertModal.svelte';

  export let node;
  export let level;
  export let expandedDirs;

  const dispatch = createEventDispatcher();

  let isEnabled = false;
  let loading = false;
  let showAlertModal = false;
  let alertTitle = '';
  let alertMessage = '';

  // Check if this is a file in sites-available
  $: isSiteConfig = !node.isDir && node.path.includes('/sites-available/');

  onMount(() => {
    if (isSiteConfig) {
      checkSiteStatus();
    }
  });

  async function checkSiteStatus() {
    try {
      // Get the filename
      const filename = node.path.split('/').pop();
      // Check if symlink exists in sites-enabled
      const response = await apiFetch(`/api/files?path=/sites-enabled`);
      const data = await response.json();
      isEnabled = data.some(file => file.name === filename && file.isSymlink);
    } catch (error) {
      console.error('Error checking site status:', error);
      isEnabled = false;
    }
  }

  async function toggleSite(event) {
    event.stopPropagation();
    loading = true;

    try {
      const filename = node.path.split('/').pop();
      const symlinkPath = `/sites-enabled/${filename}`;

      if (!isEnabled) {
        // Enable: Create symlink
        const response = await apiFetch('/api/file/symlink', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            linkPath: symlinkPath,
            targetPath: `../sites-available/${filename}`
          })
        });

        if (response.ok) {
          isEnabled = true;
        } else {
          const error = await response.json();
          showAlert('Error', `Error enabling site: ${error.error}`);
        }
      } else {
        // Disable: Delete symlink
        const response = await apiFetch('/api/file/delete', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            path: symlinkPath
          })
        });

        if (response.ok) {
          isEnabled = false;
        } else {
          const error = await response.json();
          showAlert('Error', `Error disabling site: ${error.error}`);
        }
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    } finally {
      loading = false;
    }
  }

  function showAlert(title, message) {
    alertTitle = title;
    alertMessage = message;
    showAlertModal = true;
  }

  function handleAlertOk() {
    showAlertModal = false;
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }
</script>

<div
  class="tree-node"
  style="padding-left: {level * 16}px"
  draggable="true"
  on:dragstart={e => dispatch('dragstart', { event: e, node })}
  on:dragover={handleDragOver}
  on:drop={e => dispatch('drop', { event: e, node })}
  on:contextmenu={e => dispatch('contextmenu', { event: e, node })}
  role="treeitem"
  aria-selected="false"
  tabindex="0"
>
  <div
    class="node-content"
    on:click={() => node.isDir ? dispatch('toggle', node) : dispatch('select', node)}
    on:keydown={e => e.key === 'Enter' && (node.isDir ? dispatch('toggle', node) : dispatch('select', node))}
    role="button"
    tabindex="0"
    title={node.isSymlink ? `→ ${node.linkTarget}` : ''}
  >
    {#if node.isSymlink}
      <span class="icon">🔗</span>
    {:else if node.isDir}
      <span class="icon">{expandedDirs.has(node.path) ? '📂' : '📁'}</span>
    {:else}
      <span class="icon">📄</span>
    {/if}
    <span class="node-name" class:symlink={node.isSymlink}>{node.name}</span>
    {#if isSiteConfig}
      <button
        class="toggle-btn"
        class:enabled={isEnabled}
        class:disabled={!isEnabled}
        class:loading={loading}
        on:click={toggleSite}
        disabled={loading}
        title={isEnabled ? 'Disable site' : 'Enable site'}
      >
        {#if loading}
          ⏳
        {:else if isEnabled}
          ✓
        {:else}
          ○
        {/if}
      </button>
    {/if}
  </div>
</div>

{#if node.isDir && expandedDirs.has(node.path) && node.children}
  {#each node.children as child}
    <svelte:self
      node={child}
      level={level + 1}
      {expandedDirs}
      on:toggle
      on:select
      on:dragstart
      on:drop
      on:contextmenu
    />
  {/each}
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
  .tree-node {
    user-select: none;
    cursor: pointer;
  }

  .node-content {
    display: flex;
    align-items: center;
    padding: 4px 8px;
    gap: 6px;
    transition: background 0.1s;
    cursor: pointer;
  }

  .node-content:hover {
    background: #2a2d2e;
  }

  .node-content:focus {
    outline: 1px solid #0e639c;
    outline-offset: -1px;
  }

  .icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .node-name {
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .node-name.symlink {
    color: #4ec9b0;
    font-style: italic;
  }

  .toggle-btn {
    margin-left: auto;
    padding: 2px 8px;
    font-size: 12px;
    border-radius: 3px;
    background: transparent;
    border: 1px solid #555;
    cursor: pointer;
    transition: all 0.2s;
    min-width: 24px;
  }

  .toggle-btn:hover:not(:disabled) {
    background: #3c3c3c;
    border-color: #888;
  }

  .toggle-btn.enabled {
    background: #0e7a0d;
    border-color: #0e7a0d;
    color: white;
  }

  .toggle-btn.enabled:hover:not(:disabled) {
    background: #0c6b0c;
  }

  .toggle-btn.disabled {
    background: #3c3c3c;
    border-color: #555;
  }

  .toggle-btn.loading {
    opacity: 0.6;
    cursor: wait;
  }

  .toggle-btn:disabled {
    cursor: not-allowed;
  }
</style>
