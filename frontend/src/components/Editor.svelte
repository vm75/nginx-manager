<script>
  import { onMount, onDestroy } from 'svelte';
  import { createEventDispatcher } from 'svelte';
  import { apiFetch } from '../lib/api';
  import * as monaco from 'monaco-editor';
  import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
  import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
  import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
  import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
  import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';

  export let file = null;

  const dispatch = createEventDispatcher();

  let editorContainer;
  let editor;
  let currentContent = '';
  let saving = false;
  let saveStatus = '';

  onMount(() => {
    // Configure Monaco Editor workers
    self.MonacoEnvironment = {
      getWorker(_, label) {
        if (label === 'json') {
          return new jsonWorker();
        }
        if (label === 'css' || label === 'scss' || label === 'less') {
          return new cssWorker();
        }
        if (label === 'html' || label === 'handlebars' || label === 'razor') {
          return new htmlWorker();
        }
        if (label === 'typescript' || label === 'javascript') {
          return new tsWorker();
        }
        return new editorWorker();
      }
    };

    // Configure Monaco Editor
    monaco.editor.defineTheme('nginx-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [],
      colors: {
        'editor.background': '#1e1e1e',
      }
    });

    // Register nginx language if not already registered
    if (!monaco.languages.getLanguages().some(lang => lang.id === 'nginx')) {
      monaco.languages.register({ id: 'nginx' });

      monaco.languages.setMonarchTokensProvider('nginx', {
        keywords: [
          'server', 'location', 'upstream', 'http', 'events', 'stream',
          'if', 'set', 'return', 'rewrite', 'break', 'proxy_pass',
          'listen', 'server_name', 'root', 'index', 'error_page',
          'access_log', 'error_log', 'include', 'proxy_set_header',
          'add_header', 'try_files', 'fastcgi_pass', 'proxy_redirect'
        ],
        tokenizer: {
          root: [
            [/#.*$/, 'comment'],
            [/"([^"\\]|\\.)*$/, 'string.invalid'],
            [/'([^'\\]|\\.)*$/, 'string.invalid'],
            [/"/, 'string', '@string_double'],
            [/'/, 'string', '@string_single'],
            [/[{}()\[\]]/, '@brackets'],
            [/[;,]/, 'delimiter'],
            [/\d+/, 'number'],
            [/[a-zA-Z_][\w]*/, {
              cases: {
                '@keywords': 'keyword',
                '@default': 'variable'
              }
            }]
          ],
          string_double: [
            [/[^\\"]+/, 'string'],
            [/"/, 'string', '@pop']
          ],
          string_single: [
            [/[^\\']+/, 'string'],
            [/'/, 'string', '@pop']
          ]
        }
      });
    }

    // Create editor instance
    editor = monaco.editor.create(editorContainer, {
      value: 'Select a file to edit',
      language: 'nginx',
      theme: 'nginx-dark',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 14,
      lineNumbers: 'on',
      roundedSelection: false,
      scrollBeyondLastLine: false,
      readOnly: true,
      renderWhitespace: 'selection',
      tabSize: 2,
    });

    // Auto-save on content change (debounced)
    let saveTimeout;
    editor.onDidChangeModelContent(() => {
      clearTimeout(saveTimeout);
      saveTimeout = setTimeout(() => {
        if (file && editor.getValue() !== currentContent) {
          saveFile();
        }
      }, 1000);
    });
  });

  onDestroy(() => {
    if (editor) {
      editor.dispose();
    }
  });

  $: if (file && editor) {
    loadFile(file);
  }

  async function loadFile(fileInfo) {
    try {
      const response = await apiFetch(`/api/file/read?path=${encodeURIComponent(fileInfo.path)}`);
      currentContent = await response.text();

      editor.setValue(currentContent);
      editor.updateOptions({ readOnly: false });
      // Force LF line endings
      editor.getModel().setEOL(monaco.editor.EndOfLineSequence.LF);

      // Detect language based on file extension
      const ext = fileInfo.name.split('.').pop().toLowerCase();
      let language = 'nginx';

      if (ext === 'conf') language = 'nginx';
      else if (ext === 'json') language = 'json';
      else if (ext === 'js') language = 'javascript';
      else if (ext === 'html') language = 'html';
      else if (ext === 'css') language = 'css';
      else if (ext === 'sh') language = 'shell';
      else if (ext === 'md') language = 'markdown';

      monaco.editor.setModelLanguage(editor.getModel(), language);

      saveStatus = '';
    } catch (error) {
      editor.setValue(`Error loading file: ${error.message}`);
      editor.updateOptions({ readOnly: true });
    }
  }

  async function saveFile() {
    if (!file || saving) return;

    saving = true;
    saveStatus = 'Saving...';

    try {
      // Ensure LF line endings before saving
      editor.getModel().setEOL(monaco.editor.EndOfLineSequence.LF);
      const response = await apiFetch('/api/file/write', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: file.path,
          content: editor.getValue()
        })
      });

      if (response.ok) {
        currentContent = editor.getValue();
        saveStatus = '✓ Saved';
        setTimeout(() => saveStatus = '', 2000);
        dispatch('configSaved');
      } else {
        const error = await response.json();
        saveStatus = `✗ Error: ${error.error}`;
      }
    } catch (error) {
      saveStatus = `✗ Error: ${error.message}`;
    } finally {
      saving = false;
    }
  }

  function handleSave() {
    saveFile();
  }
</script>

<div class="editor-container">
  {#if file}
    <div class="editor-header">
      <div class="file-info">
        <span class="file-icon">📄</span>
        <span class="file-path">{file.path}</span>
      </div>
      <div class="editor-actions">
        {#if saveStatus}
          <span class="save-status">{saveStatus}</span>
        {/if}
        <button on:click={handleSave} disabled={saving}>
          💾 Save
        </button>
      </div>
    </div>
  {/if}

  <div class="editor" bind:this={editorContainer}></div>
</div>

<style>
  .editor-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: #1e1e1e;
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    background: #2d2d30;
    border-bottom: 1px solid #3c3c3c;
    min-height: 40px;
  }

  .file-info {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #cccccc;
  }

  .file-icon {
    font-size: 16px;
  }

  .file-path {
    font-size: 13px;
    font-family: 'Courier New', monospace;
  }

  .editor-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .save-status {
    font-size: 13px;
    color: #4ec9b0;
  }

  .editor {
    flex: 1;
    overflow: hidden;
  }
</style>
