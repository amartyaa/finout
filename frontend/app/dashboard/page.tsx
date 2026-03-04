'use client';

import { useState, useEffect, useCallback } from 'react';
import { insights } from '@/lib/api';
import {
    BarChart, Bar, XAxis, YAxis, CartesianGrid,
    Tooltip, ResponsiveContainer, Area, AreaChart
} from 'recharts';

interface OverviewData {
    total_spend_mtd: number;
    forecast: {
        predicted_total: number;
        best_case: number;
        worst_case: number;
        narrative: string;
    };
    anomaly_count: number;
    potential_savings: number;
    top_services: Array<{ service: string; amount: number }>;
    cost_trend: Array<{ date: string; amount: number }>;
}

export default function OverviewPage() {
    const [data, setData] = useState<OverviewData | null>(null);
    const [loading, setLoading] = useState(true);

    const loadData = useCallback(async () => {
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const overview = await insights.overview(token, org.id);
            setData(overview);
        } catch (err) {
            console.error('Failed to load overview:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadData();
    }, [loadData]);

    if (loading) {
        return <div className="loading"><div className="spinner" /></div>;
    }

    const formatCurrency = (val: number) => {
        if (val >= 1000) return `$${(val / 1000).toFixed(1)}k`;
        return `$${val.toFixed(2)}`;
    };

    const hasData = data && (data.total_spend_mtd > 0 || data.cost_trend.length > 0);

    return (
        <div>
            <div className="page-header">
                <h1 className="page-title">Overview</h1>
                <p className="page-subtitle">Your cloud cost intelligence at a glance</p>
            </div>

            {/* KPI Cards */}
            <div className="kpi-grid">
                <div className="kpi-card">
                    <div className="kpi-label">💰 Month-to-Date Spend</div>
                    <div className="kpi-value">{formatCurrency(data?.total_spend_mtd || 0)}</div>
                    <div className="kpi-subtitle">Current billing period</div>
                </div>
                <div className="kpi-card">
                    <div className="kpi-label">📈 Forecasted Total</div>
                    <div className="kpi-value">{formatCurrency(data?.forecast?.predicted_total || 0)}</div>
                    <div className="kpi-subtitle">
                        {data?.forecast?.best_case ? `${formatCurrency(data.forecast.best_case)} — ${formatCurrency(data.forecast.worst_case)}` : 'End of month estimate'}
                    </div>
                </div>
                <div className="kpi-card">
                    <div className="kpi-label">⚠️ Active Anomalies</div>
                    <div className="kpi-value" style={{ color: (data?.anomaly_count || 0) > 0 ? 'var(--accent-red)' : undefined }}>
                        {data?.anomaly_count || 0}
                    </div>
                    <div className="kpi-subtitle">Cost spikes detected</div>
                </div>
                <div className="kpi-card">
                    <div className="kpi-label">✨ Potential Savings</div>
                    <div className="kpi-value" style={{ color: 'var(--accent-green)' }}>
                        {formatCurrency(data?.potential_savings || 0)}
                    </div>
                    <div className="kpi-subtitle">Monthly optimization opportunities</div>
                </div>
            </div>

            {hasData ? (
                <>
                    {/* Charts */}
                    <div className="charts-grid">
                        <div className="card chart-card">
                            <div className="chart-title">Daily Cost Trend (30 Days)</div>
                            <ResponsiveContainer width="100%" height={260}>
                                <AreaChart data={data?.cost_trend || []}>
                                    <defs>
                                        <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                                            <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                                        </linearGradient>
                                    </defs>
                                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                                    <XAxis
                                        dataKey="date"
                                        stroke="var(--text-muted)"
                                        fontSize={11}
                                        tickFormatter={(val) => val.slice(5)}
                                    />
                                    <YAxis stroke="var(--text-muted)" fontSize={11} tickFormatter={(val) => `$${val}`} />
                                    <Tooltip
                                        contentStyle={{
                                            background: 'var(--bg-secondary)',
                                            border: '1px solid var(--border-color)',
                                            borderRadius: '8px',
                                            color: 'var(--text-primary)',
                                        }}
                                        formatter={(value) => [`$${Number(value ?? 0).toFixed(2)}`, 'Cost']}
                                    />
                                    <Area type="monotone" dataKey="amount" stroke="#6366f1" fill="url(#colorCost)" strokeWidth={2} />
                                </AreaChart>
                            </ResponsiveContainer>
                        </div>

                        <div className="card chart-card">
                            <div className="chart-title">Top Services by Cost</div>
                            <ResponsiveContainer width="100%" height={260}>
                                <BarChart data={data?.top_services || []} layout="vertical">
                                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                                    <XAxis type="number" stroke="var(--text-muted)" fontSize={11} tickFormatter={(val) => `$${val}`} />
                                    <YAxis
                                        type="category"
                                        dataKey="service"
                                        stroke="var(--text-muted)"
                                        fontSize={11}
                                        width={180}
                                        tickFormatter={(val) => val.length > 25 ? val.slice(0, 25) + '...' : val}
                                    />
                                    <Tooltip
                                        contentStyle={{
                                            background: 'var(--bg-secondary)',
                                            border: '1px solid var(--border-color)',
                                            borderRadius: '8px',
                                            color: 'var(--text-primary)',
                                        }}
                                        formatter={(value) => [`$${Number(value ?? 0).toFixed(2)}`, 'Cost']}
                                    />
                                    <Bar dataKey="amount" fill="#6366f1" radius={[0, 4, 4, 0]} />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* AI Forecast Narrative */}
                    {data?.forecast?.narrative && (
                        <div className="card" style={{ marginBottom: '24px' }}>
                            <div className="chart-title">🤖 AI Forecast Summary</div>
                            <p style={{ fontSize: '14px', color: 'var(--text-secondary)', lineHeight: 1.7 }}>
                                {data.forecast.narrative}
                            </p>
                        </div>
                    )}
                </>
            ) : (
                <div className="card">
                    <div className="empty-state">
                        <div className="empty-icon">☁️</div>
                        <h3 className="empty-title">No cost data yet</h3>
                        <p className="empty-text">
                            Connect your AWS account and sync billing data to see your cloud cost dashboard come to life.
                        </p>
                        <a href="/dashboard/connections" className="btn btn-primary" style={{ marginTop: '20px' }}>
                            Connect AWS →
                        </a>
                    </div>
                </div>
            )}
        </div>
    );
}
