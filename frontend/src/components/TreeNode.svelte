<script>
  import { createEventDispatcher } from 'svelte';

  export let node;
  export let level;
  export let expandedDirs;

  const dispatch = createEventDispatcher();

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
</style>
