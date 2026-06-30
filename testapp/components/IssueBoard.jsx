import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';

function IssueBoardLogic() {
    const [issues, setIssues] = useState([]);
    const [activeProjectId, setActiveProjectId] = useState(null);

    const fetchIssues = async (projectId) => {
        if (!projectId) return;
        const res = await fetch(`/api/projects/${projectId}/issues`);
        const data = await res.json();
        setIssues(data || []);
    };

    useEffect(() => {
        const handler = (e) => {
            setActiveProjectId(e.detail);
            fetchIssues(e.detail);
        };
        const updateHandler = (e) => {
            if (e.detail === activeProjectId) fetchIssues(activeProjectId);
        };
        
        window.addEventListener('projectSelected', handler);
        window.addEventListener('issueAdded', updateHandler);
        
        return () => {
            window.removeEventListener('projectSelected', handler);
            window.removeEventListener('issueAdded', updateHandler);
        };
    }, [activeProjectId]);

    const moveIssue = async (id, currentStatus) => {
        const nextStatus = currentStatus === 'todo' ? 'in_progress' : 'done';
        if (currentStatus === 'done') return;

        await fetch(`/api/issues/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: nextStatus })
        });
        fetchIssues(activeProjectId);
    };

    const deleteIssue = async (id) => {
        await fetch(`/api/issues/${id}`, { method: 'DELETE' });
        fetchIssues(activeProjectId);
    };

    if (!activeProjectId) return null;

    const columns = [
        { id: 'todo', title: 'To Do', color: '#e3e3e8' },
        { id: 'in_progress', title: 'In Progress', color: '#f1c40f' },
        { id: 'done', title: 'Done', color: '#2ecc71' }
    ];

    return (
        <div style={{ display: 'flex', gap: '24px', height: '100%' }}>
            {columns.map(col => {
                const colIssues = issues.filter(i => i.status === col.id);
                return (
                    <div key={col.id} style={{
                        flex: 1,
                        background: 'var(--sidebar-bg)',
                        borderRadius: '12px',
                        padding: '16px',
                        display: 'flex',
                        flexDirection: 'column',
                        minWidth: '280px'
                    }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '16px', fontSize: '13px', fontWeight: '600' }}>
                            <div style={{ width: 8, height: 8, borderRadius: '50%', background: col.color }}></div>
                            {col.title} <span style={{ color: 'var(--text-muted)' }}>{colIssues.length}</span>
                        </div>
                        
                        <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                            {colIssues.map(issue => (
                                <div key={issue.id} style={{
                                    background: 'var(--surface)',
                                    padding: '16px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--border)',
                                    cursor: 'pointer',
                                    boxShadow: 'var(--shadow)',
                                    transition: 'transform 0.1s',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: '12px'
                                }}
                                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-2px)'}
                                onMouseLeave={e => e.currentTarget.style.transform = 'translateY(0)'}
                                >
                                    <div style={{ fontSize: '14px', fontWeight: '500', lineHeight: '1.4' }}>
                                        {issue.title}
                                    </div>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                        <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                                            ISSUE-{issue.id}
                                        </div>
                                        <div style={{ display: 'flex', gap: '6px' }}>
                                            {col.id !== 'done' && (
                                                <button onClick={() => moveIssue(issue.id, issue.status)} style={{
                                                    background: 'transparent', color: 'var(--accent)', border: 'none', cursor: 'pointer', fontSize: '12px', padding: 0
                                                }}>Move &rarr;</button>
                                            )}
                                            <button onClick={() => deleteIssue(issue.id)} style={{
                                                background: 'transparent', color: 'var(--danger)', border: 'none', cursor: 'pointer', fontSize: '12px', padding: 0
                                            }}>Delete</button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}

export function mount(element) {
    const root = createRoot(element);
    root.render(<IssueBoardLogic />);
}
