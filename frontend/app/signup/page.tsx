'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { auth, orgs } from '@/lib/api';

export default function SignupPage() {
    const router = useRouter();
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [orgName, setOrgName] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const res = await auth.signup({ email, password, name });
            localStorage.setItem('finops_token', res.token);
            localStorage.setItem('finops_user', JSON.stringify(res.user));

            // Auto-create org if name provided
            if (orgName.trim()) {
                const org = await orgs.create(res.token, { name: orgName });
                localStorage.setItem('finops_org', JSON.stringify(org));
            }

            router.push('/dashboard');
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Signup failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-container">
            <div className="auth-card card">
                <div style={{ textAlign: 'center', fontSize: '40px', marginBottom: '8px' }}>🚀</div>
                <h1 className="auth-title">Get Started</h1>
                <p className="auth-subtitle">Create your FinOps AI account and organization</p>

                {error && <div className="error-message">{error}</div>}

                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label className="form-label">Full Name</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="Jane Smith"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                        />
                    </div>

                    <div className="form-group">
                        <label className="form-label">Email</label>
                        <input
                            type="email"
                            className="form-input"
                            placeholder="you@company.com"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                        />
                    </div>

                    <div className="form-group">
                        <label className="form-label">Password</label>
                        <input
                            type="password"
                            className="form-input"
                            placeholder="Min 8 characters"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                            minLength={8}
                        />
                    </div>

                    <div className="form-group">
                        <label className="form-label">Organization Name</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="Acme Inc."
                            value={orgName}
                            onChange={(e) => setOrgName(e.target.value)}
                        />
                        <span style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px', display: 'block' }}>
                            Optional — you can create one later
                        </span>
                    </div>

                    <button type="submit" className="btn btn-primary" style={{ width: '100%' }} disabled={loading}>
                        {loading ? 'Creating account...' : 'Create Account'}
                    </button>
                </form>

                <div className="auth-link">
                    Already have an account? <a href="/login">Sign in</a>
                </div>
            </div>
        </div>
    );
}
