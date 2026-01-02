<script>
  import { createEventDispatcher } from 'svelte';

  export let show = false;
  export let title = 'Alert';
  export let message = '';
  export let okText = 'OK';

  const dispatch = createEventDispatcher();

  function handleOk() {
    dispatch('ok');
  }
</script>

{#if show}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-overlay"
    on:click={handleOk}
    on:keydown={e => e.key === 'Escape' && handleOk()}
    role="dialog"
    aria-modal="true"
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && handleOk()}
      role="document"
    >
      <h3>{title}</h3>
      <p>{message}</p>
      <div class="modal-buttons">
        <button on:click={handleOk} class="ok-btn">{okText}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: #2d2d30;
    border-radius: 8px;
    padding: 20px;
    max-width: 500px;
    width: 90%;
    color: #fff;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .modal h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
  }

  .modal p {
    margin: 0 0 16px 0;
    line-height: 1.5;
  }

  .modal-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .modal-buttons button {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    background: #3c3c3c;
    color: #fff;
  }

  .modal-buttons button:hover {
    background: #4c4c4c;
  }

  .ok-btn {
    background: #2196f3;
  }

  .ok-btn:hover {
    background: #1976d2;
  }
</style>