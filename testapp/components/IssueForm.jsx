import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';

function IssueFormLogic() {
    const [title, setTitle] = useState('');
    const [activeProjectId, setActiveProjectId] = useState(null);

    useEffect(() => {
        const handler = (e) => setActiveProjectId(e.detail);
        window.addEventListener('projectSelected', handler);
        return () => window.removeEventListener('projectSelected', handler);
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!title.trim() || !activeProjectId) return;

        const res = await fetch('/api/issues', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ projectId: activeProjectId, title })
        });
        
        if (res.ok) {
            setTitle('');
            window.dispatchEvent(new CustomEvent('issueAdded', { detail: activeProjectId }));
        }
    };

    if (!activeProjectId) return <div style={{ color: 'var(--text-muted)' }}>Select a project to view issues</div>;

    return (
        <form onSubmit={handleSubmit} style={{ width: '100%', display: 'flex', gap: '12px', alignItems: 'center' }}>
            <input 
                type="text" 
                value={title}
                onChange={e => setTitle(e.target.value)}
                placeholder="Create new issue..."
                style={{
                    flex: 1,
                    background: 'var(--surface)',
                    border: '1px solid var(--border)',
                    color: 'var(--text)',
                    padding: '10px 16px',
                    borderRadius: '8px',
                    fontSize: '14px',
                    outline: 'none',
                    boxShadow: 'var(--shadow)'
                }}
            />
            <button style={{
                background: 'var(--accent)',
                color: 'white',
                border: 'none',
                padding: '10px 20px',
                borderRadius: '8px',
                fontSize: '14px',
                fontWeight: '600',
                cursor: 'pointer'
            }}>
                Create Issue
            </button>
        </form>
    );
}

export function mount(element) {
    const root = createRoot(element);
    root.render(<IssueFormLogic />);
}
