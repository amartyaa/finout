'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { orgs } from '@/lib/api';

interface Org {
    id: string;
    name: string;
    slug: string;
    role: string;
}

interface User {
    id: string;
    name: string;
    email: string;
}

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();
    const [user, setUser] = useState<User | null>(null);
    const [orgList, setOrgList] = useState<Org[]>([]);
    const [currentOrg, setCurrentOrg] = useState<Org | null>(null);

    const loadData = useCallback(async () => {
        const token = localStorage.getItem('finops_token');
        const userData = localStorage.getItem('finops_user');

        if (!token || !userData) {
            router.push('/login');
            return;
        }

        setUser(JSON.parse(userData));

        try {
            const res = await orgs.list(token);
            setOrgList(res.organizations || []);

            const savedOrg = localStorage.getItem('finops_org');
            if (savedOrg) {
                const org = JSON.parse(savedOrg);
                const found = (res.organizations || []).find((o: Org) => o.id === org.id);
                if (found) {
                    setCurrentOrg(found);
                    return;
                }
            }

            if (res.organizations?.length > 0) {
                setCurrentOrg(res.organizations[0]);
                localStorage.setItem('finops_org', JSON.stringify(res.organizations[0]));
            }
        } catch {
            // If API fails, might be token expired
        }
    }, [router]);

    useEffect(() => {
        loadData();
    }, [loadData]);

    const handleOrgChange = (orgId: string) => {
        const org = orgList.find(o => o.id === orgId);
        if (org) {
            setCurrentOrg(org);
            localStorage.setItem('finops_org', JSON.stringify(org));
            window.location.reload();
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('finops_token');
        localStorage.removeItem('finops_user');
        localStorage.removeItem('finops_org');
        router.push('/login');
    };

    const navItems = [
        { href: '/dashboard', icon: '📊', label: 'Overview' },
        { href: '/dashboard/insights', icon: '🧠', label: 'AI Insights' },
        { href: '/dashboard/connections', icon: '☁️', label: 'AWS Connection' },
    ];

    return (
        <div className="dashboard-layout">
            <aside className="sidebar">
                <div className="sidebar-logo">⚡ FinOps AI</div>

                {orgList.length > 0 && (
                    <div className="org-selector">
                        <select
                            value={currentOrg?.id || ''}
                            onChange={(e) => handleOrgChange(e.target.value)}
                            style={{ width: '100%' }}
                        >
                            {orgList.map(org => (
                                <option key={org.id} value={org.id}>{org.name}</option>
                            ))}
                        </select>
                    </div>
                )}

                <nav className="sidebar-nav">
                    {navItems.map(item => (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={`nav-link ${pathname === item.href ? 'active' : ''}`}
                        >
                            <span className="nav-icon">{item.icon}</span>
                            {item.label}
                        </Link>
                    ))}
                </nav>

                <div className="sidebar-footer">
                    {user && (
                        <div className="user-info">
                            <div className="user-avatar">{user.name?.charAt(0).toUpperCase()}</div>
                            <div>
                                <div className="user-name">{user.name}</div>
                                <div className="user-email">{user.email}</div>
                            </div>
                        </div>
                    )}
                    <button
                        className="btn btn-secondary btn-small"
                        style={{ width: '100%', marginTop: '12px' }}
                        onClick={handleLogout}
                    >
                        Sign Out
                    </button>
                </div>
            </aside>

            <main className="main-content">
                {children}
            </main>
        </div>
    );
}
