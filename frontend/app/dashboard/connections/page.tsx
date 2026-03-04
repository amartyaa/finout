'use client';

import { useState, useEffect, useCallback } from 'react';
import { aws as awsApi } from '@/lib/api';

interface ConnectionStatus {
    connected: boolean;
    status: string;
    role_arn?: string;
    external_id?: string;
    error?: string;
    last_sync_at?: string;
}

export default function ConnectionsPage() {
    const [status, setStatus] = useState<ConnectionStatus | null>(null);
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
            console.error('Failed to load status:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadStatus();
    }, [loadStatus]);

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
            if (res.status === 'connected') {
                setMessage('✅ AWS connection validated successfully!');
            } else {
                setMessage(`❌ Connection failed: ${res.error}`);
            }
            loadStatus();
        } catch (err) {
            setMessage(`❌ Failed to connect: ${err instanceof Error ? err.message : 'Unknown error'}`);
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
            setMessage('🔄 Cost sync job queued. Data will appear in your dashboard shortly.');
            loadStatus();
        } catch (err) {
            setMessage(`❌ Sync failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
            setSyncing(false);
        }
    };

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

    if (loading) {
        return <div className="loading"><div className="spinner" /></div>;
    }

    return (
        <div>
            <div className="page-header">
                <h1 className="page-title">AWS Connection</h1>
                <p className="page-subtitle">Connect your AWS account using cross-account IAM role assumption</p>
            </div>

            {/* Connection Status */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px' }}>
                    <span className={`status-dot ${getStatusDot(status?.status || 'not_configured')}`} />
                    <div>
                        <div style={{ fontWeight: 600, fontSize: '15px' }}>
                            {status?.connected ? 'Connected' : status?.status === 'syncing' ? 'Syncing...' : status?.status === 'error' ? 'Connection Error' : 'Not Connected'}
                        </div>
                        {status?.role_arn && (
                            <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>
                                Role: {status.role_arn}
                            </div>
                        )}
                        {status?.last_sync_at && (
                            <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                                Last sync: {new Date(status.last_sync_at).toLocaleString()}
                            </div>
                        )}
                    </div>
                </div>

                {status?.connected && (
                    <button
                        className="btn btn-primary btn-small"
                        onClick={handleSync}
                        disabled={syncing || status?.status === 'syncing'}
                    >
                        {syncing || status?.status === 'syncing' ? '🔄 Syncing...' : '🔄 Sync Now'}
                    </button>
                )}

                {status?.error && (
                    <div className="error-message" style={{ marginTop: '12px' }}>
                        {status.error}
                    </div>
                )}
            </div>

            {message && (
                <div className="card" style={{ marginBottom: '24px', padding: '16px' }}>
                    <p style={{ fontSize: '14px' }}>{message}</p>
                </div>
            )}

            {/* Setup Instructions */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="chart-title">Setup Instructions</div>
                <div className="connection-steps">
                    <div className="step">
                        <div className="step-number">1</div>
                        <div className="step-content">
                            <h3>Create an IAM Role in your AWS Account</h3>
                            <p>
                                Go to <strong>IAM → Roles → Create Role</strong>. Select <strong>Another AWS Account</strong> as the trusted entity type.
                                Enter SaaS Account ID: <code>{process.env.NEXT_PUBLIC_AWS_SAAS_ACCOUNT_ID || '123456789012'}</code>
                            </p>
                        </div>
                    </div>

                    <div className="step">
                        <div className="step-number">2</div>
                        <div className="step-content">
                            <h3>Attach Cost Explorer Permissions</h3>
                            <p>
                                Attach the managed policy <code>ViewOnlyAccess</code> or create a custom policy with at minimum
                                <code>ce:GetCostAndUsage</code>, <code>ce:GetCostForecast</code>, and <code>cloudwatch:GetMetricData</code>.
                            </p>
                        </div>
                    </div>

                    <div className="step">
                        <div className="step-number">3</div>
                        <div className="step-content">
                            <h3>Copy the Role ARN</h3>
                            <p>
                                Once created, copy the Role ARN (e.g. <code>arn:aws:iam::123456789012:role/FinOpsSaaSRole</code>) and paste it below.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            {/* ARN Input Form */}
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
        </div>
    );
}
