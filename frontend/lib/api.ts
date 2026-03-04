const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface FetchOptions extends RequestInit {
    token?: string;
}

async function apiFetch(path: string, options: FetchOptions = {}) {
    const { token, ...fetchOptions } = options;

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(options.headers as Record<string, string> || {}),
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
        ...fetchOptions,
        headers,
    });

    if (!res.ok) {
        const error = await res.json().catch(() => ({ error: 'Request failed' }));
        throw new Error(error.error || 'Request failed');
    }

    return res.json();
}

// Auth
export const auth = {
    signup: (data: { email: string; password: string; name: string }) =>
        apiFetch('/api/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
    login: (data: { email: string; password: string }) =>
        apiFetch('/api/auth/login', { method: 'POST', body: JSON.stringify(data) }),
    me: (token: string) =>
        apiFetch('/api/auth/me', { token }),
};

// Organizations
export const orgs = {
    create: (token: string, data: { name: string }) =>
        apiFetch('/api/orgs', { method: 'POST', body: JSON.stringify(data), token }),
    list: (token: string) =>
        apiFetch('/api/orgs', { token }),
};

// Projects
export const projects = {
    create: (token: string, orgId: string, data: { name: string; description?: string }) =>
        apiFetch(`/api/orgs/${orgId}/projects`, { method: 'POST', body: JSON.stringify(data), token }),
    list: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/projects`, { token }),
};

// AWS
export const aws = {
    connect: (token: string, orgId: string, data: { role_arn: string }) =>
        apiFetch(`/api/orgs/${orgId}/aws/connect`, { method: 'POST', body: JSON.stringify(data), token }),
    status: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/aws/status`, { token }),
    sync: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/aws/sync`, { method: 'POST', token }),
};

// Azure
export const azure = {
    connect: (token: string, orgId: string, data: { tenant_id: string; client_id: string; client_secret: string; subscription_id: string }) =>
        apiFetch(`/api/orgs/${orgId}/azure/connect`, { method: 'POST', body: JSON.stringify(data), token }),
    status: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/azure/status`, { token }),
    sync: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/azure/sync`, { method: 'POST', token }),
};

// Insights
export const insights = {
    overview: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/insights/overview`, { token }),
    anomalies: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/insights/anomalies`, { token }),
    forecast: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/insights/forecast`, { token }),
    recommendations: (token: string, orgId: string) =>
        apiFetch(`/api/orgs/${orgId}/insights/recommendations`, { token }),
    dismissRecommendation: (token: string, orgId: string, recId: string) =>
        apiFetch(`/api/orgs/${orgId}/insights/recommendations/${recId}/dismiss`, { method: 'POST', token }),
};
