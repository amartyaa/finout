'use client';

import { useState, useEffect, useCallback } from 'react';
import { aws as awsApi, azure as azureApi } from '@/lib/api';

// ── Types ──────────────────────────────────────────

interface AWSStatus {
    connected: boolean;
    status: string;
    role_arn?: string;
    external_id?: string;
    error?: string;
    last_sync_at?: string;
}

interface AzureStatus {
    connected: boolean;
    status: string;
    tenant_id?: string;
    client_id?: string;
    subscription_id?: string;
    error?: string;
    last_sync_at?: string;
}

// ── Main Page ──────────────────────────────────────

export default function ConnectionsPage() {
    const [activeTab, setActiveTab] = useState<'aws' | 'azure'>('aws');

    return (
        <div>
            <div className="page-header">
                <h1 className="page-title">Cloud Connections</h1>
                <p className="page-subtitle">Connect your cloud accounts to ingest cost data</p>
            </div>

            {/* Provider Tabs */}
            <div style={{ display: 'flex', gap: '8px', marginBottom: '24px' }}>
                <button
                    className={`btn ${activeTab === 'aws' ? 'btn-primary' : 'btn-secondary'} btn-small`}
                    onClick={() => setActiveTab('aws')}
                >
                    🟧 Amazon Web Services
                </button>
                <button
                    className={`btn ${activeTab === 'azure' ? 'btn-primary' : 'btn-secondary'} btn-small`}
                    onClick={() => setActiveTab('azure')}
                >
                    🔷 Microsoft Azure
                </button>
            </div>

            {activeTab === 'aws' ? <AWSTab /> : <AzureTab />}
        </div>
    );
}

// ── AWS Tab ─────────────────────────────────────────

function AWSTab() {
    const [status, setStatus] = useState<AWSStatus | null>(null);
    const [roleArn, setRoleArn] = useState('');
    const [loading, setLoading] = useState(true);
    const [connecting, setConnecting] = useState(false);
    const [syncing, setSyncing] = useState(false);
    const [message, setMessage] = useState('');

    const loadStatus = useCallback(async () => {
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const res = await awsApi.status(token, org.id);
            setStatus(res);
            if (res.role_arn) setRoleArn(res.role_arn);
        } catch (err) {
            console.error('Failed to load AWS status:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => { loadStatus(); }, [loadStatus]);

    const handleConnect = async (e: React.FormEvent) => {
        e.preventDefault();
        setMessage('');
        setConnecting(true);

        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const res = await awsApi.connect(token, org.id, { role_arn: roleArn });
            setMessage(res.status === 'connected' ? '✅ AWS connection validated!' : `❌ ${res.error}`);
            loadStatus();
        } catch (err) {
            setMessage(`❌ Failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
            setConnecting(false);
        }
    };

    const handleSync = async () => {
        setSyncing(true);
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            await awsApi.sync(token, org.id);
            setMessage('🔄 AWS cost sync queued.');
            loadStatus();
        } catch (err) {
            setMessage(`❌ Sync failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
            setSyncing(false);
        }
    };

    if (loading) return <div className="loading"><div className="spinner" /></div>;

    return (
        <>
            <ConnectionStatusCard
                provider="AWS"
                connected={status?.connected || false}
                statusText={status?.status || 'not_configured'}
                detail={status?.role_arn ? `Role: ${status.role_arn}` : undefined}
                lastSync={status?.last_sync_at}
                error={status?.error}
                syncing={syncing}
                onSync={status?.connected ? handleSync : undefined}
            />

            {message && (
                <div className="card" style={{ marginBottom: '24px', padding: '16px' }}>
                    <p style={{ fontSize: '14px' }}>{message}</p>
                </div>
            )}

            {/* Setup Instructions */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="chart-title">AWS Setup Instructions</div>
                <div className="connection-steps">
                    <div className="step">
                        <div className="step-number">1</div>
                        <div className="step-content">
                            <h3>Create an IAM Role in your AWS Account</h3>
                            <p>
                                Go to <strong>IAM → Roles → Create Role</strong>. Select <strong>Another AWS Account</strong>.
                                Enter SaaS Account ID: <code>{process.env.NEXT_PUBLIC_AWS_SAAS_ACCOUNT_ID || '123456789012'}</code>
                            </p>
                        </div>
                    </div>
                    <div className="step">
                        <div className="step-number">2</div>
                        <div className="step-content">
                            <h3>Attach Cost Explorer Permissions</h3>
                            <p>
                                Attach <code>ViewOnlyAccess</code> or a custom policy with
                                <code>ce:GetCostAndUsage</code>, <code>ce:GetCostForecast</code>, <code>cloudwatch:GetMetricData</code>.
                            </p>
                        </div>
                    </div>
                    <div className="step">
                        <div className="step-number">3</div>
                        <div className="step-content">
                            <h3>Copy the Role ARN</h3>
                            <p>Paste below, e.g. <code>arn:aws:iam::123456789012:role/FinOpsSaaSRole</code></p>
                        </div>
                    </div>
                </div>
            </div>

            {/* ARN Input */}
            <div className="card">
                <div className="chart-title">Connect AWS Account</div>
                <form onSubmit={handleConnect}>
                    <div className="form-group">
                        <label className="form-label">IAM Role ARN</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="arn:aws:iam::123456789012:role/YourRole"
                            value={roleArn}
                            onChange={(e) => setRoleArn(e.target.value)}
                            required
                            pattern="arn:aws:iam::\d{12}:role/.+"
                        />
                        <span style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px', display: 'block' }}>
                            Must be a valid IAM Role ARN format
                        </span>
                    </div>
                    <button type="submit" className="btn btn-primary" disabled={connecting}>
                        {connecting ? 'Validating...' : status?.connected ? 'Update Connection' : 'Connect AWS'}
                    </button>
                </form>
            </div>
        </>
    );
}

// ── Azure Tab ───────────────────────────────────────

function AzureTab() {
    const [status, setStatus] = useState<AzureStatus | null>(null);
    const [tenantId, setTenantId] = useState('');
    const [clientId, setClientId] = useState('');
    const [clientSecret, setClientSecret] = useState('');
    const [subscriptionId, setSubscriptionId] = useState('');
    const [loading, setLoading] = useState(true);
    const [connecting, setConnecting] = useState(false);
    const [syncing, setSyncing] = useState(false);
    const [message, setMessage] = useState('');

    const loadStatus = useCallback(async () => {
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const res = await azureApi.status(token, org.id);
            setStatus(res);
            if (res.tenant_id) setTenantId(res.tenant_id);
            if (res.client_id) setClientId(res.client_id);
            if (res.subscription_id) setSubscriptionId(res.subscription_id);
        } catch (err) {
            console.error('Failed to load Azure status:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => { loadStatus(); }, [loadStatus]);

    const handleConnect = async (e: React.FormEvent) => {
        e.preventDefault();
        setMessage('');
        setConnecting(true);

        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const res = await azureApi.connect(token, org.id, {
                tenant_id: tenantId,
                client_id: clientId,
                client_secret: clientSecret,
                subscription_id: subscriptionId,
            });
            setMessage(res.status === 'connected' ? '✅ Azure connection validated!' : `❌ ${res.error}`);
            loadStatus();
        } catch (err) {
            setMessage(`❌ Failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
            setConnecting(false);
        }
    };

    const handleSync = async () => {
        setSyncing(true);
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            await azureApi.sync(token, org.id);
            setMessage('🔄 Azure cost sync queued.');
            loadStatus();
        } catch (err) {
            setMessage(`❌ Sync failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
            setSyncing(false);
        }
    };

    if (loading) return <div className="loading"><div className="spinner" /></div>;

    return (
        <>
            <ConnectionStatusCard
                provider="Azure"
                connected={status?.connected || false}
                statusText={status?.status || 'not_configured'}
                detail={status?.subscription_id ? `Subscription: ${status.subscription_id}` : undefined}
                lastSync={status?.last_sync_at}
                error={status?.error}
                syncing={syncing}
                onSync={status?.connected ? handleSync : undefined}
            />

            {message && (
                <div className="card" style={{ marginBottom: '24px', padding: '16px' }}>
                    <p style={{ fontSize: '14px' }}>{message}</p>
                </div>
            )}

            {/* Setup Instructions */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="chart-title">Azure Setup Instructions</div>
                <div className="connection-steps">
                    <div className="step">
                        <div className="step-number">1</div>
                        <div className="step-content">
                            <h3>Register an App in Azure AD</h3>
                            <p>
                                Go to <strong>Azure Portal → App registrations → New registration</strong>.
                                Name it e.g. &quot;FinOps SaaS Connector&quot;.
                            </p>
                        </div>
                    </div>
                    <div className="step">
                        <div className="step-number">2</div>
                        <div className="step-content">
                            <h3>Create a Client Secret</h3>
                            <p>
                                Under <strong>Certificates &amp; secrets</strong>, create a new client secret and copy its value immediately.
                            </p>
                        </div>
                    </div>
                    <div className="step">
                        <div className="step-number">3</div>
                        <div className="step-content">
                            <h3>Assign Cost Management Reader Role</h3>
                            <p>
                                Go to your <strong>Subscription → Access control (IAM) → Add role assignment</strong>.
                                Assign <code>Cost Management Reader</code> to the app you created.
                            </p>
                        </div>
                    </div>
                    <div className="step">
                        <div className="step-number">4</div>
                        <div className="step-content">
                            <h3>Copy Credentials</h3>
                            <p>
                                You&apos;ll need: <strong>Tenant ID</strong>, <strong>Client ID</strong> (Application ID),
                                <strong> Client Secret</strong>, and <strong>Subscription ID</strong>.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            {/* Credential Form */}
            <div className="card">
                <div className="chart-title">Connect Azure Account</div>
                <form onSubmit={handleConnect}>
                    <div className="form-group">
                        <label className="form-label">Tenant ID (Directory ID)</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                            value={tenantId}
                            onChange={(e) => setTenantId(e.target.value)}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label className="form-label">Client ID (Application ID)</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                            value={clientId}
                            onChange={(e) => setClientId(e.target.value)}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label className="form-label">Client Secret</label>
                        <input
                            type="password"
                            className="form-input"
                            placeholder="Enter your client secret"
                            value={clientSecret}
                            onChange={(e) => setClientSecret(e.target.value)}
                            required
                        />
                        <span style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px', display: 'block' }}>
                            Encrypted at rest — never displayed after saving
                        </span>
                    </div>
                    <div className="form-group">
                        <label className="form-label">Subscription ID</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                            value={subscriptionId}
                            onChange={(e) => setSubscriptionId(e.target.value)}
                            required
                        />
                    </div>
                    <button type="submit" className="btn btn-primary" disabled={connecting}>
                        {connecting ? 'Validating...' : status?.connected ? 'Update Connection' : 'Connect Azure'}
                    </button>
                </form>
            </div>
        </>
    );
}

// ── Shared Status Card Component ────────────────────

function ConnectionStatusCard({
    provider,
    connected,
    statusText,
    detail,
    lastSync,
    error,
    syncing,
    onSync,
}: {
    provider: string;
    connected: boolean;
    statusText: string;
    detail?: string;
    lastSync?: string;
    error?: string;
    syncing: boolean;
    onSync?: () => void;
}) {
    const getStatusDot = (s: string) => {
        const classes: Record<string, string> = {
            connected: 'status-connected',
            error: 'status-error',
            syncing: 'status-syncing',
            pending: 'status-pending',
            not_configured: 'status-pending',
        };
        return classes[s] || 'status-pending';
    };

    return (
        <div className="card" style={{ marginBottom: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px' }}>
                <span className={`status-dot ${getStatusDot(statusText)}`} />
                <div>
                    <div style={{ fontWeight: 600, fontSize: '15px' }}>
                        {provider}: {connected ? 'Connected' : statusText === 'syncing' ? 'Syncing...' : statusText === 'error' ? 'Connection Error' : 'Not Connected'}
                    </div>
                    {detail && (
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>{detail}</div>
                    )}
                    {lastSync && (
                        <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                            Last sync: {new Date(lastSync).toLocaleString()}
                        </div>
                    )}
                </div>
            </div>

            {onSync && (
                <button
                    className="btn btn-primary btn-small"
                    onClick={onSync}
                    disabled={syncing || statusText === 'syncing'}
                >
                    {syncing || statusText === 'syncing' ? '🔄 Syncing...' : '🔄 Sync Now'}
                </button>
            )}

            {error && (
                <div className="error-message" style={{ marginTop: '12px' }}>{error}</div>
            )}
        </div>
    );
}
