<script>
  import { onMount } from 'svelte';
  import { apiFetch } from '../lib/api';
  import ConfirmModal from './ConfirmModal.svelte';
  import AlertModal from './AlertModal.svelte';

  let certificates = [];
  let loading = false;
  let showObtainModal = false;
  let obtaining = false;
  let deleting = null;
  let showDeleteConfirm = false;
  let deleteParams = null;
  let showAlertModal = false;
  let alertTitle = '';
  let alertMessage = '';

  // Form data
  let domains = [];
  let domainInput = '';
  let email = '';
  let challenge = 'http-01';
  let provider = '';
  let staging = false;
  let force = false;
  let envVars = [{ key: '', value: '' }];

  // Reactive: Check if any domain is a wildcard
  $: hasWildcard = domains.some(d => d.startsWith('*.'));
  $: if (hasWildcard && challenge !== 'dns-01') {
    challenge = 'dns-01';
  }

  // Common DNS providers for quick reference
  const commonProviders = [
    { name: 'Cloudflare', code: 'cloudflare', env: 'CLOUDFLARE_DNS_API_TOKEN or CLOUDFLARE_API_KEY + CLOUDFLARE_EMAIL' },
    { name: 'Route53 (AWS)', code: 'route53', env: 'AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION' },
    { name: 'Google Cloud DNS', code: 'gcloud', env: 'GCE_PROJECT, GCE_SERVICE_ACCOUNT_FILE' },
    { name: 'DigitalOcean', code: 'digitalocean', env: 'DO_AUTH_TOKEN' },
    { name: 'NameSilo', code: 'namesilo', env: 'NAMESILO_API_KEY' },
    { name: 'Namecheap', code: 'namecheap', env: 'NAMECHEAP_API_USER, NAMECHEAP_API_KEY' },
    { name: 'DuckDNS', code: 'duckdns', env: 'DUCKDNS_TOKEN' },
    { name: 'GoDaddy', code: 'godaddy', env: 'GODADDY_API_KEY, GODADDY_API_SECRET' },
    { name: 'Vultr', code: 'vultr', env: 'VULTR_API_KEY' },
    { name: 'Linode', code: 'linode', env: 'LINODE_TOKEN' },
  ];

  onMount(() => {
    loadCertificates();
  });

  async function loadCertificates() {
    loading = true;
    try {
      const response = await apiFetch('/api/certificates');
      certificates = await response.json();
    } catch (error) {
      console.error('Failed to load certificates:', error);
    } finally {
      loading = false;
    }
  }

  function deleteCertificate(domain, certFile, keyFile) {
    deleteParams = { domain, certFile, keyFile };
    showDeleteConfirm = true;
  }

  async function handleConfirmDelete() {
    const { domain, certFile, keyFile } = deleteParams;
    showDeleteConfirm = false;
    deleteParams = null;

    deleting = domain;
    try {
      const response = await apiFetch('/api/certificates/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ certFile, keyFile })
      });

      const result = await response.json();

      if (result.success) {
        showAlert('Success', 'Certificate deleted successfully');
        loadCertificates();
      } else {
        showAlert('Error', 'Failed to delete certificate: ' + result.error);
      }
    } catch (error) {
      showAlert('Error', 'Error: ' + error.message);
    } finally {
      deleting = null;
    }
  }

  function handleCancelDelete() {
    showDeleteConfirm = false;
    deleteParams = null;
  }

  function showAlert(title, message) {
    alertTitle = title;
    alertMessage = message;
    showAlertModal = true;
  }

  function handleAlertOk() {
    showAlertModal = false;
  }

  function addDomain() {
    const trimmed = domainInput.trim();
    if (!trimmed) return;

    if (domains.includes(trimmed)) {
      showAlert('Validation Error', 'Domain already added');
      return;
    }

    domains = [...domains, trimmed];
    domainInput = '';
  }

  function removeDomain(domain) {
    domains = domains.filter(d => d !== domain);
  }

  function handleDomainKeyPress(event) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addDomain();
    }
  }

  async function obtainCertificate() {
    if (domains.length === 0 || !email) {
      showAlert('Validation Error', 'Please add at least one domain and provide an email');
      return;
    }

    if (challenge === 'dns-01' && !provider) {
      showAlert('Validation Error', 'Please specify a DNS provider for DNS-01 challenge');
      return;
    }

    if (challenge === 'dns-01' && envVars.every(v => !v.key || !v.value)) {
      showAlert('Validation Error', 'Please provide at least one environment variable for your DNS provider');
      return;
    }

    obtaining = true;
    try {
      const payload = {
        domains,
        email,
        challenge,
        staging,
        force,
        provider: challenge === 'dns-01' ? provider : '',
        credentials: {}
      };

      // Add environment variables as credentials
      if (challenge === 'dns-01') {
        envVars.forEach(v => {
          if (v.key && v.value) {
            payload.credentials[v.key] = v.value;
          }
        });
      }

      const response = await apiFetch('/api/certificates/obtain', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      const result = await response.json();

      if (result.success) {
        showAlert('Success', 'Certificate obtained successfully!\n\nCert: ' + result.certFile + '\nKey: ' + result.keyFile);
        showObtainModal = false;
        resetForm();
        loadCertificates();
      } else {
        showAlert('Error', 'Failed to obtain certificate:\n\n' + (result.error || result.output));
      }
    } catch (error) {
      showAlert('Error', 'Error: ' + error.message);
    } finally {
      obtaining = false;
    }
  }

  function resetForm() {
    domains = [];
    domainInput = '';
    email = '';
    challenge = 'http-01';
    provider = '';
    staging = false;
    force = false;
    envVars = [{ key: '', value: '' }];
  }

  function addEnvVar() {
    envVars = [...envVars, { key: '', value: '' }];
  }

  function removeEnvVar(index) {
    envVars = envVars.filter((_, i) => i !== index);
  }

  function selectProvider(providerCode, envVarStr) {
    provider = providerCode;
    // Parse environment variables from the string
    const vars = envVarStr.split(',').map(v => v.trim());
    envVars = vars.map(v => ({ key: v, value: '' }));
  }

  function getCertificateStatus(daysLeft) {
    if (daysLeft < 0) return 'expired';
    if (daysLeft < 14) return 'warning';
    return 'valid';
  }

  function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    return new Date(dateStr).toLocaleDateString();
  }
</script>

<div class="certificates-container">
  <div class="certificates-header">
    <h2>🔐 SSL Certificates</h2>
    <div class="header-actions">
      <button class="btn-secondary" on:click={loadCertificates} disabled={loading} title="Refresh certificate list">
        🔄 Refresh
      </button>
      <button class="btn-primary" on:click={() => showObtainModal = true}>
        ➕ Obtain Certificate
      </button>
    </div>
  </div>

  {#if loading}
    <div class="loading">Loading certificates...</div>
  {:else if certificates.length === 0}
    <div class="empty-state">
      <p>No certificates found</p>
      <p class="hint">Click "Obtain Certificate" to get started with Let's Encrypt</p>
    </div>
  {:else}
    <div class="certificates-list">
      {#each certificates as cert}
        <div class="certificate-card" class:expired={getCertificateStatus(cert.daysLeft) === 'expired'}>
          <div class="cert-header">
            <h3>
              {#if cert.isWildcard}
                🌐 {cert.domain}
              {:else}
                🔒 {cert.domain}
              {/if}
            </h3>
            <div class="cert-actions">
              <span class="cert-status status-{getCertificateStatus(cert.daysLeft)}">
                {#if cert.daysLeft < 0}
                  Expired
                {:else if cert.daysLeft < 14}
                  Expiring Soon
                {:else}
                  Valid
                {/if}
              </span>
              <button
                class="btn-icon btn-delete"
                on:click={() => deleteCertificate(cert.domain, cert.certFile, cert.keyFile)}
                disabled={deleting === cert.domain}
                title="Delete certificate"
              >
                🗑️
              </button>
            </div>
          </div>

          <div class="cert-details">
            <div class="cert-detail">
              <span class="label">Valid From:</span>
              <span class="value">{formatDate(cert.notBefore)}</span>
            </div>
            <div class="cert-detail">
              <span class="label">Expires:</span>
              <span class="value">{formatDate(cert.notAfter)}</span>
            </div>
            <div class="cert-detail">
              <span class="label">Days Left:</span>
              <span class="value">{cert.daysLeft >= 0 ? cert.daysLeft : 0} days</span>
            </div>
            <div class="cert-detail">
              <span class="label">Certificate:</span>
              <span class="value path">{cert.certFile}</span>
            </div>
            {#if cert.keyFile}
              <div class="cert-detail">
                <span class="label">Private Key:</span>
                <span class="value path">{cert.keyFile}</span>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showObtainModal}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal-overlay" on:click={() => showObtainModal = false}>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>Obtain SSL Certificate</h3>
        <button class="close-btn" on:click={() => showObtainModal = false}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="domainInput">Domains *</label>
          <div class="domain-input-container">
            <input
              id="domainInput"
              type="text"
              bind:value={domainInput}
              on:keypress={handleDomainKeyPress}
              placeholder="example.com or *.example.com"
              disabled={obtaining}
            />
            <button
              type="button"
              class="btn-add-domain"
              on:click={addDomain}
              disabled={obtaining || !domainInput.trim()}
            >
              Add
            </button>
          </div>
          <p class="hint">Add each domain separately. Use *.domain.com for wildcard domains. Press Enter or click Add.</p>

          {#if domains.length > 0}
            <div class="domain-pills">
              {#each domains as domain}
                <div class="domain-pill">
                  <span class="pill-text">{domain}</span>
                  <button
                    type="button"
                    class="pill-remove"
                    on:click={() => removeDomain(domain)}
                    disabled={obtaining}
                    title="Remove domain"
                  >
                    ×
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <div class="form-group">
          <label for="email">Email *</label>
          <input
            id="email"
            type="email"
            bind:value={email}
            placeholder="admin@example.com"
            disabled={obtaining}
          />
        </div>

        <div class="form-group">
          <label>
            <input type="checkbox" bind:checked={staging} disabled={obtaining} />
            Use Let's Encrypt Staging (for testing)
          </label>
          {#if staging}
            <p class="hint">⚠️ Staging certificates are NOT trusted by browsers. Use only for testing to avoid rate limits.</p>
          {:else}
            <p class="hint">💡 Enable staging to test without hitting Let's Encrypt rate limits (50 certs/week)</p>
          {/if}
        </div>

        <div class="form-group">
          <label>
            <input type="checkbox" bind:checked={force} disabled={obtaining} />
            Force renewal (re-issue existing certificates)
          </label>
          {#if force}
            <p class="hint">⚠️ Will re-issue the certificate even if it already exists and hasn't expired.</p>
          {:else}
            <p class="hint">💡 Enable to replace existing certificates (e.g., when switching from staging to production)</p>
          {/if}
        </div>

        <div class="form-group">
          <label for="challenge">Challenge Type *</label>
          <select id="challenge" bind:value={challenge} disabled={obtaining || hasWildcard}>
            <option value="http-01">HTTP-01 (Port 80 required)</option>
            <option value="dns-01">DNS-01 (DNS API required)</option>
            <option value="tls-alpn-01">TLS-ALPN-01 (Port 443 required)</option>
          </select>
          {#if hasWildcard}
            <p class="hint">🔒 DNS-01 is required for wildcard domains</p>
          {:else}
            <p class="hint">💡 Use DNS-01 for wildcard domains (*.example.com)</p>
          {/if}
        </div>

        {#if challenge === 'dns-01'}
          <div class="form-group">
            <label for="provider">DNS Provider Code *
              <a href="https://github.com/acmesh-official/acme.sh/wiki/dnsapi" target="_blank" rel="noopener" style="font-size: 12px; margin-left: 5px;">
                📚 View all 150+ providers
              </a>
            </label>
            <input
              id="provider"
              type="text"
              bind:value={provider}
              placeholder="e.g., cloudflare, route53, gcloud, digitalocean"
              disabled={obtaining}
            />
            <small>Enter the acme.sh DNS provider code (lowercase, without dns_ prefix)</small>
          </div>

          <div class="common-providers">
            <h4>Common Providers (click to auto-fill):</h4>
            <div class="provider-grid">
              {#each commonProviders as p}
                <button
                  type="button"
                  class="provider-btn"
                  on:click={() => selectProvider(p.code, p.env)}
                  disabled={obtaining}
                >
                  {p.name}
                </button>
              {/each}
            </div>
          </div>

          <div class="credentials-section">
            <h4>Environment Variables</h4>
            <p class="hint">Add the environment variables required by your DNS provider</p>

            {#each envVars as envVar, i}
              <div class="env-var-row">
                <input
                  type="text"
                  bind:value={envVar.key}
                  placeholder="VARIABLE_NAME"
                  disabled={obtaining}
                  class="env-key"
                />
                <input
                  type="password"
                  bind:value={envVar.value}
                  placeholder="value"
                  disabled={obtaining}
                  class="env-value"
                />
                {#if envVars.length > 1}
                  <button
                    type="button"
                    class="btn-remove"
                    on:click={() => removeEnvVar(i)}
                    disabled={obtaining}
                  >
                    ×
                  </button>
                {/if}
              </div>
            {/each}

            <button
              type="button"
              class="btn-add-env"
              on:click={addEnvVar}
              disabled={obtaining}
            >
              ➕ Add Variable
            </button>
          </div>
        {/if}

        <div class="info-box">
          <h4>ℹ️ Information</h4>
          <ul>
            {#if challenge === 'http-01'}
              <li>HTTP-01 requires port 80 accessible from internet</li>
              <li>Domain must point to this server's IP address</li>
              <li>Cannot be used for wildcard certificates</li>
            {:else if challenge === 'tls-alpn-01'}
              <li>TLS-ALPN-01 requires port 443 accessible from internet</li>
              <li>Domain must point to this server's IP address</li>
              <li>Cannot be used for wildcard certificates</li>
            {:else}
              <li>DNS-01 requires DNS provider API access</li>
              <li>Can issue wildcard certificates (*.domain.com)</li>
              <li>DNS propagation may take a few minutes</li>
              <li>Supports all 150+ DNS providers in acme.sh</li>
            {/if}
            {#if domains.length > 1}
              <li><strong>Multiple domains:</strong> Will generate a single certificate for all {domains.length} domains</li>
            {/if}
            <li>Certificates saved to ssl/ directory</li>
            <li>Rate limit: 50 certificates per domain per week</li>
          </ul>
        </div>
      </div>

      <div class="modal-footer">
        <button
          class="btn-secondary"
          on:click={() => showObtainModal = false}
          disabled={obtaining}
        >
          Cancel
        </button>
        <button
          class="btn-primary"
          on:click={obtainCertificate}
          disabled={obtaining}
        >
          {obtaining ? '⏳ Obtaining...' : '🔐 Obtain Certificate'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDeleteConfirm}
  <ConfirmModal
    title="Delete Certificate"
    message={`Are you sure you want to delete the certificate for ${deleteParams?.domain}?\n\nThis will delete:\n- ${deleteParams?.certFile}\n- ${deleteParams?.keyFile || 'No key file'}`}
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
  .certificates-container {
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
  }

  .certificates-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 2px solid #e0e0e0;
  }

  .certificates-header h2 {
    margin: 0;
    font-size: 24px;
    color: #333;
  }

  .header-actions {
    display: flex;
    gap: 10px;
  }

  .header-actions {
    display: flex;
    gap: 10px;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: #666;
  }

  .empty-state {
    text-align: center;
    padding: 60px 20px;
    color: #666;
  }

  .empty-state p {
    margin: 10px 0;
  }

  .hint {
    font-size: 14px;
    color: #999;
    margin-top: 5px;
  }

  .certificates-list {
    display: grid;
    gap: 20px;
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  }

  .certificate-card {
    background: white;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: box-shadow 0.2s;
  }

  .certificate-card:hover {
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  }

  .certificate-card.expired {
    border-color: #ff4444;
    background: #fff5f5;
  }

  .cert-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 15px;
    padding-bottom: 10px;
    border-bottom: 1px solid #eee;
  }

  .cert-title-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .cert-header h3 {
    margin: 0;
    font-size: 18px;
    color: #333;
    word-break: break-all;
  }

  .wildcard-badge {
    display: inline-block;
    padding: 2px 8px;
    background: #fff3cd;
    color: #856404;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
  }

  .cert-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .cert-status {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-valid {
    background: #d4edda;
    color: #155724;
  }

  .status-warning {
    background: #fff3cd;
    color: #856404;
  }

  .status-expired {
    background: #f8d7da;
    color: #721c24;
  }

  .cert-details {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .cert-detail {
    display: flex;
    justify-content: space-between;
    font-size: 14px;
  }

  .cert-detail .label {
    color: #666;
    font-weight: 500;
  }

  .cert-detail .value {
    color: #333;
    text-align: right;
  }

  .cert-detail .value.path {
    font-family: monospace;
    font-size: 12px;
    color: #0066cc;
    word-break: break-all;
  }

  .btn-primary, .btn-secondary {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #0066cc;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0052a3;
  }

  .btn-primary:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: #f0f0f0;
    color: #333;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #e0e0e0;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    border-radius: 8px;
    max-width: 600px;
    width: 90%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px;
    border-bottom: 1px solid #eee;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 20px;
    color: #333;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 28px;
    color: #999;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    color: #666;
  }

  .modal-body {
    padding: 20px;
  }

  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    margin-bottom: 8px;
    color: #333;
    font-weight: 500;
    font-size: 14px;
  }

  .form-group input[type="text"],
  .form-group input[type="email"],
  .form-group select {
    width: 100%;
    padding: 10px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    box-sizing: border-box;
  }

  .form-group input[type="checkbox"] {
    margin-right: 8px;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #0066cc;
  }

  .domain-input-container {
    display: flex;
    gap: 8px;
  }

  .domain-input-container input {
    flex: 1;
  }

  .btn-add-domain {
    padding: 10px 20px;
    background: #0066cc;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
  }

  .btn-add-domain:hover:not(:disabled) {
    background: #0052a3;
  }

  .btn-add-domain:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  .domain-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .domain-pill {
    display: inline-flex;
    align-items: center;
    background: #e3f2fd;
    border: 1px solid #90caf9;
    border-radius: 20px;
    padding: 6px 12px;
    font-size: 14px;
    color: #1565c0;
    gap: 8px;
  }

  .pill-text {
    font-weight: 500;
  }

  .pill-remove {
    background: transparent;
    border: none;
    color: #1565c0;
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .pill-remove:hover:not(:disabled) {
    background: #1565c0;
    color: white;
  }

  .pill-remove:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .credentials-section {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 6px;
    margin-top: 10px;
  }

  .credentials-section h4 {
    margin: 0 0 15px 0;
    font-size: 16px;
    color: #333;
  }

  .common-providers {
    margin: 15px 0;
    padding: 15px;
    background: #f5f5f5;
    border-radius: 6px;
  }

  .common-providers h4 {
    margin: 0 0 10px 0;
    font-size: 14px;
    color: #666;
  }

  .provider-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 8px;
  }

  .provider-btn {
    padding: 8px 12px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 13px;
    color: #333;
    cursor: pointer;
    transition: all 0.2s;
    font-weight: 500;
  }

  .provider-btn:hover:not(:disabled) {
    background: #0066cc;
    color: white;
    border-color: #0066cc;
  }

  .provider-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .env-var-row {
    display: flex;
    gap: 8px;
    margin-bottom: 10px;
    align-items: center;
  }

  .env-key {
    flex: 1;
    min-width: 0;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-family: monospace;
    font-size: 13px;
  }

  .env-value {
    flex: 2;
    min-width: 0;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 13px;
  }

  .btn-remove {
    padding: 4px 10px;
    background: #ff4444;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 18px;
    line-height: 1;
  }

  .btn-remove:hover:not(:disabled) {
    background: #cc0000;
  }

  .btn-add-env {
    margin-top: 10px;
    padding: 8px 16px;
    background: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
  }

  .btn-add-env:hover:not(:disabled) {
    background: #218838;
  }

  .info-box {
    background: #e3f2fd;
    padding: 15px;
    border-radius: 6px;
    border-left: 4px solid #2196f3;
    margin-top: 20px;
  }

  .info-box h4 {
    margin: 0 0 10px 0;
    font-size: 16px;
    color: #1976d2;
  }

  .info-box ul {
    margin: 0;
    padding-left: 20px;
  }

  .info-box li {
    margin: 5px 0;
    font-size: 14px;
    color: #1565c0;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 20px;
    border-top: 1px solid #eee;
  }

  .btn-icon {
    padding: 6px 10px;
    background: transparent;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    transition: all 0.2s;
  }

  .btn-icon:hover:not(:disabled) {
    background: #f5f5f5;
  }

  .btn-delete:hover:not(:disabled) {
    background: #fff5f5;
    border-color: #ff4444;
  }

  .btn-icon:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
