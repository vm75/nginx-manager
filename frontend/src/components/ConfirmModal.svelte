<script>
  import { createEventDispatcher } from 'svelte';

  export let show = false;
  export let title = 'Confirm';
  export let message = 'Are you sure?';
  export let confirmText = 'Confirm';
  export let cancelText = 'Cancel';

  const dispatch = createEventDispatcher();

  function handleConfirm() {
    dispatch('confirm');
  }

  function handleCancel() {
    dispatch('cancel');
  }
</script>

{#if show}
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-overlay"
    on:click={handleCancel}
    on:keydown={e => e.key === 'Escape' && handleCancel()}
    role="dialog"
    aria-modal="true"
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && handleCancel()}
      role="document"
    >
      <h3>{title}</h3>
      <p>{message}</p>
      <div class="modal-buttons">
        <button on:click={handleCancel}>{cancelText}</button>
        <button on:click={handleConfirm} class="confirm-btn">{confirmText}</button>
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

  .confirm-btn {
    background: #f44336;
  }

  .confirm-btn:hover {
    background: #d32f2f;
  }
</style>