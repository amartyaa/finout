'use client';

import { useState, useEffect, useCallback } from 'react';
import { insights } from '@/lib/api';

interface Anomaly {
    id: string;
    date: string;
    service: string;
    expected_amount: number;
    actual_amount: number;
    deviation_pct: number;
    confidence_score: number;
    narrative: string;
    status: string;
}

interface Recommendation {
    id: string;
    category: string;
    resource_type: string;
    resource_id: string;
    title: string;
    description: string;
    estimated_monthly_savings: number;
    risk_level: string;
    confidence_score: number;
    status: string;
}

interface ForecastData {
    available: boolean;
    predicted_total?: number;
    best_case?: number;
    worst_case?: number;
    accuracy_pct?: number;
    narrative?: string;
    message?: string;
}

export default function InsightsPage() {
    const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
    const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
    const [forecast, setForecast] = useState<ForecastData | null>(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<'recommendations' | 'anomalies' | 'forecast'>('recommendations');

    const loadData = useCallback(async () => {
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            const [anomalyRes, recRes, forecastRes] = await Promise.all([
                insights.anomalies(token, org.id),
                insights.recommendations(token, org.id),
                insights.forecast(token, org.id),
            ]);
            setAnomalies(anomalyRes.anomalies || []);
            setRecommendations(recRes.recommendations || []);
            setForecast(forecastRes);
        } catch (err) {
            console.error('Failed to load insights:', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadData();
    }, [loadData]);

    const handleDismiss = async (recId: string) => {
        const token = localStorage.getItem('finops_token');
        const orgData = localStorage.getItem('finops_org');
        if (!token || !orgData) return;

        const org = JSON.parse(orgData);
        try {
            await insights.dismissRecommendation(token, org.id, recId);
            setRecommendations(prev => prev.filter(r => r.id !== recId));
        } catch (err) {
            console.error('Failed to dismiss:', err);
        }
    };

    const formatCurrency = (val: number) => `$${val.toFixed(2)}`;

    const riskColors: Record<string, string> = {
        low: 'badge-green',
        medium: 'badge-amber',
        high: 'badge-red',
    };

    if (loading) {
        return <div className="loading"><div className="spinner" /></div>;
    }

    return (
        <div>
            <div className="page-header">
                <h1 className="page-title">AI Insights</h1>
                <p className="page-subtitle">AI-powered cost optimization recommendations and anomaly detection</p>
            </div>

            {/* Tabs */}
            <div style={{ display: 'flex', gap: '8px', marginBottom: '24px' }}>
                {[
                    { key: 'recommendations', label: `💡 Recommendations (${recommendations.length})` },
                    { key: 'anomalies', label: `⚠️ Anomalies (${anomalies.length})` },
                    { key: 'forecast', label: '📈 Forecast' },
                ].map(tab => (
                    <button
                        key={tab.key}
                        className={`btn ${activeTab === tab.key ? 'btn-primary' : 'btn-secondary'} btn-small`}
                        onClick={() => setActiveTab(tab.key as typeof activeTab)}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Recommendations Tab */}
            {activeTab === 'recommendations' && (
                <div>
                    {recommendations.length === 0 ? (
                        <div className="card">
                            <div className="empty-state">
                                <div className="empty-icon">✨</div>
                                <h3 className="empty-title">No recommendations yet</h3>
                                <p className="empty-text">Sync your AWS cost data to receive AI-powered optimization recommendations.</p>
                            </div>
                        </div>
                    ) : (
                        recommendations.map(rec => (
                            <div key={rec.id} className="card rec-card">
                                <div className="rec-header">
                                    <div className="rec-title">{rec.title}</div>
                                    <span className={`badge ${riskColors[rec.risk_level] || 'badge-blue'}`}>
                                        {rec.risk_level} risk
                                    </span>
                                </div>
                                <div className="rec-meta">
                                    <div className="rec-savings">{formatCurrency(rec.estimated_monthly_savings)}/mo</div>
                                    <span className="badge badge-blue">{rec.category.replace('_', ' ')}</span>
                                    <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                                        {rec.confidence_score.toFixed(0)}% confidence
                                    </span>
                                </div>
                                <div className="rec-description">
                                    {rec.description || `Optimization opportunity for ${rec.resource_type} resource ${rec.resource_id}`}
                                </div>
                                <div className="rec-actions">
                                    <button className="btn btn-secondary btn-small" onClick={() => handleDismiss(rec.id)}>
                                        Dismiss
                                    </button>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            )}

            {/* Anomalies Tab */}
            {activeTab === 'anomalies' && (
                <div className="card">
                    {anomalies.length === 0 ? (
                        <div className="empty-state">
                            <div className="empty-icon">🎉</div>
                            <h3 className="empty-title">No anomalies detected</h3>
                            <p className="empty-text">Your cloud spending patterns look normal. We&apos;ll alert you if we detect any unusual spikes.</p>
                        </div>
                    ) : (
                        anomalies.map(anomaly => (
                            <div key={anomaly.id} className="anomaly-item">
                                <div className="anomaly-indicator" />
                                <div className="anomaly-content">
                                    <div className="anomaly-header">
                                        <div className="anomaly-service">{anomaly.service}</div>
                                        <div className="anomaly-date">{anomaly.date}</div>
                                    </div>
                                    <div className="anomaly-amounts">
                                        <span>Expected: <strong>{formatCurrency(anomaly.expected_amount)}</strong></span>
                                        <span>Actual: <strong style={{ color: 'var(--accent-red)' }}>{formatCurrency(anomaly.actual_amount)}</strong></span>
                                        <span className="badge badge-red">+{anomaly.deviation_pct.toFixed(1)}%</span>
                                        <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                                            {anomaly.confidence_score.toFixed(0)}% confidence
                                        </span>
                                    </div>
                                    {anomaly.narrative && (
                                        <div className="anomaly-narrative">{anomaly.narrative}</div>
                                    )}
                                </div>
                            </div>
                        ))
                    )}
                </div>
            )}

            {/* Forecast Tab */}
            {activeTab === 'forecast' && (
                <div className="card">
                    {!forecast?.available ? (
                        <div className="empty-state">
                            <div className="empty-icon">📈</div>
                            <h3 className="empty-title">No forecast available</h3>
                            <p className="empty-text">{forecast?.message || 'Sync at least 7 days of cost data to generate a forecast.'}</p>
                        </div>
                    ) : (
                        <div>
                            <div className="kpi-grid">
                                <div className="kpi-card">
                                    <div className="kpi-label">Predicted Month-End Total</div>
                                    <div className="kpi-value">{formatCurrency(forecast.predicted_total || 0)}</div>
                                </div>
                                <div className="kpi-card">
                                    <div className="kpi-label">Best Case</div>
                                    <div className="kpi-value" style={{ color: 'var(--accent-green)' }}>{formatCurrency(forecast.best_case || 0)}</div>
                                </div>
                                <div className="kpi-card">
                                    <div className="kpi-label">Worst Case</div>
                                    <div className="kpi-value" style={{ color: 'var(--accent-red)' }}>{formatCurrency(forecast.worst_case || 0)}</div>
                                </div>
                                <div className="kpi-card">
                                    <div className="kpi-label">Model Accuracy (R²)</div>
                                    <div className="kpi-value">{(forecast.accuracy_pct || 0).toFixed(1)}%</div>
                                </div>
                            </div>

                            {forecast.narrative && (
                                <div style={{ marginTop: '20px' }}>
                                    <h3 style={{ fontSize: '16px', fontWeight: 600, marginBottom: '12px' }}>🤖 AI Forecast Analysis</h3>
                                    <p style={{ fontSize: '14px', color: 'var(--text-secondary)', lineHeight: 1.7 }}>
                                        {forecast.narrative}
                                    </p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
