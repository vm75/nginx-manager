<script>
  import FileBrowser from './components/FileBrowser.svelte';
  import Editor from './components/Editor.svelte';
  import Logs from './components/Logs.svelte';
  import Certificates from './components/Certificates.svelte';
  import Toolbar from './components/Toolbar.svelte';

  let currentFile = null;
  let currentView = 'editor'; // 'editor', 'logs', or 'certificates'
  let refreshBrowser = 0;

  function handleFileSelect(event) {
    currentFile = event.detail;
    currentView = 'editor';
  }

  function handleRefresh() {
    refreshBrowser++;
  }

  function handleViewChange(view) {
    currentView = view;
  }
</script>

<div class="app-container">
  <Toolbar
    on:refresh={handleRefresh}
    on:viewChange={(e) => handleViewChange(e.detail)}
    currentView={currentView}
  />

  <div class="main-content">
    <aside class="sidebar">
      <FileBrowser on:fileSelect={handleFileSelect} refresh={refreshBrowser} />
    </aside>

    <main class="content">
      {#if currentView === 'editor'}
        <Editor file={currentFile} />
      {:else if currentView === 'logs'}
        <Logs />
      {:else if currentView === 'certificates'}
        <Certificates />
      {/if}
    </main>
  </div>
</div>

<style>
  .app-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: #1e1e1e;
  }

  .main-content {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .sidebar {
    width: 300px;
    min-width: 200px;
    max-width: 500px;
    background: #252526;
    border-right: 1px solid #3c3c3c;
    overflow-y: auto;
    resize: horizontal;
  }

  .content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  @media (max-width: 768px) {
    .main-content {
      flex-direction: column;
    }

    .sidebar {
      width: 100%;
      max-height: 40vh;
      resize: vertical;
    }
  }
</style>
