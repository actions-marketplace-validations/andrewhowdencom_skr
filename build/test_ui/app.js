document.addEventListener('DOMContentLoaded', () => {
    const grid = document.getElementById('skills-grid');
    const searchInput = document.getElementById('search-input');
    const emptyState = document.getElementById('empty-state');

    let allSkills = [];

    // Fetch skills
    fetch('skills.json')
        .then(response => {
            if (!response.ok) return fetch('/api/skills');
            return response;
        })
        .then(response => response.json())
        .then(data => {
            allSkills = Array.isArray(data) ? data : (data.skills || []);
            renderSkills(allSkills);
        })
        .catch(err => {
            console.error('Failed to load skills:', err);
            grid.innerHTML = '<p style="color:var(--md-sys-color-error); text-align:center; grid-column: 1/-1;">Error loading skills.</p>';
        });

    // Search
    searchInput.addEventListener('input', (e) => {
        const term = e.target.value.toLowerCase();
        const filtered = allSkills.filter(skill => {
            return (skill.name && skill.name.toLowerCase().includes(term)) ||
                (skill.description && skill.description.toLowerCase().includes(term)) ||
                (skill.metadata && skill.metadata.author && skill.metadata.author.toLowerCase().includes(term));
        });
        renderSkills(filtered);
    });

    function renderSkills(skills) {
        grid.innerHTML = '';

        if (skills.length === 0) {
            emptyState.classList.remove('hidden');
            return;
        } else {
            emptyState.classList.add('hidden');
        }

        skills.forEach((skill, index) => {
            const card = document.createElement('a');
            card.className = 'card';

            let link = '#';
            if (skill.path) link = skill.path;
            card.href = link;
            if (link.startsWith('http')) {
                card.target = '_blank';
                card.rel = 'noopener noreferrer';
            }

            const author = (skill.metadata && skill.metadata.author) || 'Unknown Author';
            const version = (skill.metadata && skill.metadata.version) || 'v?';

            card.innerHTML = `
                <div class="card-content">
                    <div class="card-title">${escapeHtml(skill.name)}</div>
                    <div class="card-subhead">${escapeHtml(version)} • ${escapeHtml(author)}</div>
                    <div class="card-text">${escapeHtml(skill.description)}</div>
                    <div class="chips-container">
                        <span class="chip">${escapeHtml(author)}</span>
                    </div>
                </div>
            `;
            grid.appendChild(card);
        });
    }

    function escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.innerText = str;
        return div.innerHTML;
    }
});
