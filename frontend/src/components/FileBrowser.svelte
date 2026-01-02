<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import TreeNode from './TreeNode.svelte';
  import { apiFetch } from '../lib/api';
  import ConfirmModal from './ConfirmModal.svelte';
  import AlertModal from './AlertModal.svelte';

  export let refresh = 0;

  const dispatch = createEventDispatcher();

  let files = [];
  let expandedDirs = new Set(['/']);
  let draggedItem = null;
  let showCreateModal = false;
  let createIsDir = false;
  let createName = '';
  let createParentPath = '/';
  let contextMenu = null;
  let selectedPath = null;
  let showSymlinkModal = false;
  let symlinkName = '';
  let symlinkTarget = '';
  let symlinkParentPath = '/';
  let showMoveConfirm = false;
  let showDeleteConfirm = false;
  let moveParams = null;
  let deleteParams = null;
  let showAlertModal = false;
  let alertTitle = '';
  let alertMessage = '';

  onMount(() => {
    loadFiles('/');
  });

  // Removed refresh prop handling since refresh is now handled locally

  async function loadFiles(path) {
    try {
      const response = await apiFetch(`/api/files?path=${encodeURIComponent(path)}`);
      const data = await response.json();

      if (path === '/') {
        files = await buildTree(data || []);
      } else {
        await updateTreeNode(files, path, data || []);
      }
      files = files; // Trigger reactivity
    } catch (error) {
      console.error('Error loading files:', error);
    }
  }

  async function buildTree(items) {
    if (!items || !Array.isArray(items)) {
      return [];
    }
    const tree = await Promise.all(items
      .filter(item => item.name !== 'sites-enabled') // Hide sites-enabled folder
      .map(async item => {
        const node = {
          ...item,
          children: item.isDir ? [] : null,
          loaded: false
        };
        if (item.isDir && expandedDirs.has(item.path)) {
          node.loaded = true;
          node.children = await loadChildren(item.path);
        }
        return node;
      }));
    return tree.sort((a, b) => {
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
  }

  async function loadChildren(path) {
    try {
      const response = await apiFetch(`/api/files?path=${encodeURIComponent(path)}`);
      const data = await response.json();
      return await buildTree(data || []);
    } catch (error) {
      console.error('Error loading children:', error);
      return [];
    }
  }

  async function updateTreeNode(tree, path, items) {
    for (let node of tree) {
      if (node.path === path && node.isDir) {
        node.children = await buildTree(items);
        node.loaded = true;
        return true;
      }
      if (node.children && await updateTreeNode(node.children, path, items)) {
        return true;
      }
    }
    return false;
  }

  async function toggleDir(node) {
    if (!node.isDir) return;

    if (expandedDirs.has(node.path)) {
      expandedDirs.delete(node.path);
    } else {
      expandedDirs.add(node.path);
      if (!node.loaded) {
        await loadFiles(node.path);
      }
    }
    expandedDirs = expandedDirs;
  }

  function selectFile(node) {
    if (!node.isDir) {
      dispatch('fileSelect', node);
    }
  }

  function handleDragStart(event, node) {
    draggedItem = node;
    event.dataTransfer.effectAllowed = 'move';
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }

  async function handleDrop(event, targetNode) {
    event.preventDefault();
    if (!draggedItem || !targetNode.isDir) return;

    const itemName = draggedItem.path.split('/').pop();
    const targetName = targetNode.path.split('/').pop();
    moveParams = { draggedItem, targetNode, itemName, targetName };
    showMoveConfirm = true;
  }

  async function handleConfirmMove() {
    const { draggedItem: item, targetNode } = moveParams;
    showMoveConfirm = false;
    moveParams = null;

    try {
      const response = await apiFetch('/api/file/move', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sourcePath: ih,
          targetPath: targetNode.path
        })
      });

      if (response.ok) {
        loadFiles('/');
      } else {
        const error = await response.json();
        showAlert('Error', `Error: ${error.error}`);
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    }

    draggedItem = null;
  }

  function handleCancelMove() {
    showMoveConfirm = false;
    moveParams = null;
    draggedItem = null;
  }

  function showContextMenu(event, node) {
    event.preventDefault();
    selectedPath = node;
    contextMenu = {
      x: event.clientX,
      y: event.clientY
    };
  }

  function hideContextMenu() {
    contextMenu = null;
  }

  function openCreateModal(isDir, parentPath = '/') {
    createIsDir = isDir;
    createParentPath = parentPath;
    createName = '';
    showCreateModal = true;
    hideContextMenu();
  }

  async function handleCreate() {
    if (!createName) return;

    const path = createParentPath === '/'
      ? `/${createName}`
      : `${createParentPath}/${createName}`;

    try {
      const response = await apiFetch('/api/file/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, isDir: createIsDir })
      });

      if (response.ok) {
        showCreateModal = false;
        loadFiles('/');
      } else {
        const error = await response.json();
        showAlert('Error', `Error: ${error.error}`);
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    }
  }

  function deleteItem(node) {
    deleteParams = { node };
    showDeleteConfirm = true;
  }

  async function handleConfirmDelete() {
    const { node } = deleteParams;
    showDeleteConfirm = false;
    deleteParams = null;

    try {
      const response = await apiFetch('/api/file/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: node.path })
      });

      if (response.ok) {
        loadFiles('/');
      } else {
        const error = await response.json();
        showAlert('Error', `Error: ${error.error}`);
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    }
  }

  function handleCancelDelete() {
    showDeleteConfirm = false;
    deleteParams = null;
    hideContextMenu();
  }

  function showAlert(title, message) {
    alertTitle = title;
    alertMessage = message;
    showAlertModal = true;
  }

  function handleAlertOk() {
    showAlertModal = false;
  }

  async function renameItem(node) {
    const newName = prompt(`Rename ${node.name} to:`, node.name);
    if (!newName || newName === node.name) return;

    const dir = node.path.substring(0, node.path.lastIndexOf('/')) || '/';
    const newPath = dir === '/' ? `/${newName}` : `${dir}/${newName}`;

    try {
      const response = await apiFetch('/api/file/rename', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ oldPath: node.path, newPath })
      });

      if (response.ok) {
        loadFiles('/');
      } else {
        const error = await response.json();
        showAlert('Error', `Error: ${error.error}`);
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    }
    hideContextMenu();
  }

  function openSymlinkModal(parentPath = '/') {
    symlinkParentPath = parentPath;
    symlinkName = '';
    symlinkTarget = '';
    showSymlinkModal = true;
    hideContextMenu();
  }

  async function handleSymlinkCreate() {
    if (!symlinkName || !symlinkTarget) return;

    const linkPath = symlinkParentPath === '/'
      ? `/${symlinkName}`
      : `${symlinkParentPath}/${symlinkName}`;

    try {
      const response = await apiFetch('/api/file/symlink', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ linkPath, targetPath: symlinkTarget })
      });

      if (response.ok) {
        showSymlinkModal = false;
        loadFiles('/');
      } else {
        const error = await response.json();
        showAlert('Error', `Error: ${error.error}`);
      }
    } catch (error) {
      showAlert('Error', `Error: ${error.message}`);
    }
  }
</script>

<svelte:window on:click={hideContextMenu} />

<div class="file-browser">
  <div class="browser-header">
    <h3>📁 Files</h3>
    <div class="header-actions">
      <button on:click={() => loadFiles('/')} title="Refresh file tree">🔄</button>
      <button on:click={() => openCreateModal(false, '/')} title="New File">📄</button>
      <button on:click={() => openCreateModal(true, '/')} title="New Folder">📁</button>
    </div>
  </div>

  <div class="file-tree">
    {#each files as node}
      <TreeNode
        {node}
        level={0}
        {expandedDirs}
        on:toggle={e => toggleDir(e.detail)}
        on:select={e => selectFile(e.detail)}
        on:dragstart={e => handleDragStart(e.detail.event, e.detail.node)}
        on:drop={e => handleDrop(e.detail.event, e.detail.node)}
        on:contextmenu={e => showContextMenu(e.detail.event, e.detail.node)}
      />
    {/each}
  </div>
</div>

{#if contextMenu && selectedPath}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="context-menu"
    style="left: {contextMenu.x}px; top: {contextMenu.y}px"
    on:click|stopPropagation
    on:keydown={e => e.key === 'Escape' && hideContextMenu()}
    role="menu"
    tabindex="-1"
  >
    <button on:click={() => renameItem(selectedPath)}>✏️ Rename</button>
    <button on:click={() => deleteItem(selectedPath)} class="danger">🗑️ Delete</button>
    {#if selectedPath.isDir}
      <hr>
      <button on:click={() => openCreateModal(false, selectedPath.path)}>📄 New File</button>
      <button on:click={() => openCreateModal(true, selectedPath.path)}>📁 New Folder</button>
      <button on:click={() => openSymlinkModal(selectedPath.path)}>🔗 New Symlink</button>
    {/if}
  </div>
{/if}

{#if showCreateModal}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-overlay"
    on:click={() => showCreateModal = false}
    on:keydown={e => e.key === 'Escape' && (showCreateModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && (showCreateModal = false)}
      role="document"
    >
      <h3>Create {createIsDir ? 'Folder' : 'File'}</h3>
      <input
        type="text"
        bind:value={createName}
        placeholder={createIsDir ? 'Folder name' : 'File name'}
        on:keydown={e => e.key === 'Enter' && handleCreate()}
      />
      <div class="modal-buttons">
        <button class="secondary" on:click={() => showCreateModal = false}>Cancel</button>
        <button on:click={handleCreate}>Create</button>
      </div>
    </div>
  </div>
{/if}

{#if showSymlinkModal}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-overlay"
    on:click={() => showSymlinkModal = false}
    on:keydown={e => e.key === 'Escape' && (showSymlinkModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && (showSymlinkModal = false)}
      role="document"
    >
      <h3>Create Symlink</h3>
      <div class="form-group">
        <label for="symlink-name">Link name:</label>
        <input
          id="symlink-name"
          type="text"
          bind:value={symlinkName}
          placeholder="e.g., site.conf"
          on:keydown={e => e.key === 'Enter' && symlinkTarget && handleSymlinkCreate()}
        />
      </div>
      <div class="form-group">
        <label for="symlink-target">Target path:</label>
        <input
          id="symlink-target"
          type="text"
          bind:value={symlinkTarget}
          placeholder="../sites-available/site.conf or /sites-available/site.conf"
          on:keydown={e => e.key === 'Enter' && symlinkName && handleSymlinkCreate()}
        />
        <small>Use relative path (../) or absolute path (/path)</small>
      </div>
      <div class="modal-buttons">
        <button class="secondary" on:click={() => showSymlinkModal = false}>Cancel</button>
        <button on:click={handleSymlinkCreate}>Create</button>
      </div>
    </div>
  </div>
{/if}

{#if showMoveConfirm}
  <ConfirmModal
    title="Move File"
    message={`Move "${moveParams?.itemName}" to "${moveParams?.targetName}"?`}
    confirmText="Move"
    cancelText="Cancel"
    on:confirm={handleConfirmMove}
    on:cancel={handleCancelMove}
    bind:show={showMoveConfirm}
  />
{/if}

{#if showDeleteConfirm}
  <ConfirmModal
    title="Delete File"
    message={`Are you sure you want to delete ${deleteParams?.node?.name}?`}
    confirmText="Delete"
    cancelText="Cancel"
    on:confirm={handleConfirmDelete}
    on:cancel={handleCancelDelete}
    bind:show={showDeleteConfirm}
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
  .file-browser {
    display: flex;
    flex-direction: column;
    height: 100%;
    color: #cccccc;
  }

  .browser-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    background: #2d2d30;
    border-bottom: 1px solid #3c3c3c;
  }

  .browser-header h3 {
    font-size: 14px;
    font-weight: 600;
    margin: 0;
  }

  .header-actions {
    display: flex;
    gap: 4px;
  }

  .header-actions button {
    padding: 4px 8px;
    font-size: 16px;
    background: transparent;
  }

  .header-actions button:hover {
    background: #3c3c3c;
  }

  .file-tree {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }

  .context-menu {
    position: fixed;
    background: #2d2d30;
    border: 1px solid #3c3c3c;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    z-index: 1000;
    padding: 4px 0;
    min-width: 150px;
  }

  .context-menu button {
    width: 100%;
    text-align: left;
    background: transparent;
    padding: 8px 12px;
    border-radius: 0;
    font-size: 13px;
  }

  .context-menu button:hover {
    background: #3c3c3c;
  }

  .context-menu hr {
    border: none;
    border-top: 1px solid #3c3c3c;
    margin: 4px 0;
  }

  input {
    width: 100%;
    margin: 16px 0;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    margin-bottom: 6px;
    font-size: 13px;
    color: #cccccc;
    font-weight: 500;
  }

  .form-group input {
    margin: 0;
  }

  .form-group small {
    display: block;
    margin-top: 4px;
    font-size: 11px;
    color: #888;
  }
</style>
