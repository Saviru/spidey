import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';

function SidebarLogic() {
    const [projects, setProjects] = useState([]);
    const [activeId, setActiveId] = useState(null);
    const [newProjectName, setNewProjectName] = useState('');

    const fetchProjects = async () => {
        const res = await fetch('/api/projects');
        const data = await res.json();
        setProjects(data || []);
        
        // Auto-select first project if none active
        if (!activeId && data && data.length > 0) {
            selectProject(data[0].id);
        }
    };

    useEffect(() => {
        fetchProjects();
    }, []);

    const selectProject = (id) => {
        setActiveId(id);
        window.dispatchEvent(new CustomEvent('projectSelected', { detail: id }));
    };

    const handleCreate = async (e) => {
        e.preventDefault();
        if (!newProjectName.trim()) return;
        
        const res = await fetch('/api/projects', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: newProjectName })
        });
        
        if (res.ok) {
            setNewProjectName('');
            await fetchProjects();
        }
    };

    return (
        <div style={{
            width: '260px',
            background: 'var(--sidebar-bg)',
            borderRight: '1px solid var(--border)',
            display: 'flex',
            flexDirection: 'column',
            padding: '24px 16px'
        }}>
            <div style={{
                display: 'flex', alignItems: 'center', gap: '8px', 
                color: 'var(--text)', fontWeight: '600', fontSize: '14px', marginBottom: '24px', padding: '0 8px'
            }}>
                <div style={{ width: 16, height: 16, background: 'var(--accent)', borderRadius: 4 }}></div>
                Spidey Workspace
            </div>

            <div style={{ fontSize: '11px', fontWeight: '600', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '8px', padding: '0 8px' }}>
                Projects
            </div>

            <div style={{ flex: 1, overflowY: 'auto' }}>
                {projects.map(p => (
                    <div 
                        key={p.id}
                        onClick={() => selectProject(p.id)}
                        style={{
                            padding: '8px',
                            borderRadius: '6px',
                            cursor: 'pointer',
                            fontSize: '14px',
                            color: activeId === p.id ? 'var(--text)' : 'var(--text-muted)',
                            background: activeId === p.id ? 'var(--surface)' : 'transparent',
                            marginBottom: '2px',
                            fontWeight: activeId === p.id ? '500' : '400'
                        }}
                    >
                        # {p.name}
                    </div>
                ))}
            </div>

            <form onSubmit={handleCreate} style={{ marginTop: '20px' }}>
                <input 
                    value={newProjectName}
                    onChange={e => setNewProjectName(e.target.value)}
                    placeholder="New Project..."
                    style={{
                        width: '100%',
                        background: 'var(--surface)',
                        border: '1px solid var(--border)',
                        color: 'var(--text)',
                        padding: '8px',
                        borderRadius: '6px',
                        fontSize: '13px',
                        outline: 'none'
                    }}
                />
            </form>
        </div>
    );
}

export function mount(element) {
    const root = createRoot(element);
    root.render(<SidebarLogic />);
}
