document.addEventListener('DOMContentLoaded', () => {
    const grid = document.getElementById('skills-grid');
    const searchInput = document.getElementById('search-input');
    const emptyState = document.getElementById('empty-state');

    let allSkills = [];

    // Fetch skills
    // Fetch skills
    fetch('skills.json')
        .then(response => {
            if (response.ok) return response;
            return fetch('/api/skills');
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            if (!data) {
                allSkills = [];
            } else {
                allSkills = Array.isArray(data) ? data : (data.skills || []);
            }
            renderSkills(allSkills);
        })
        .catch(err => {
            console.error('Failed to load skills:', err);
            grid.innerHTML = '<p style="color:var(--md-sys-color-error); text-align:center; grid-column: 1/-1;">Error loading skills. Please check the server logs.</p>';
        });

    // Search
    searchInput.addEventListener('input', (e) => {
        const term = e.target.value.toLowerCase();
        const filtered = allSkills.filter(skill => {
            return (skill.name && skill.name.toLowerCase().includes(term)) ||
                (skill.description && skill.description.toLowerCase().includes(term)) ||
                (skill.author && skill.author.toLowerCase().includes(term));
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
            const card = document.createElement('div');
            card.className = 'card';

            const author = skill.author || 'Unknown Author';
            const versions = skill.versions || [];

            // Create version chips
            const versionChips = versions.map(v =>
                `<span class="chip" title="${v.tag}">${escapeHtml(v.version)}</span>`
            ).join('');

            card.innerHTML = `
                <div class="card-content">
                    <div class="card-title">${escapeHtml(skill.name)}</div>
                    <div class="card-subhead">${escapeHtml(author)}</div>
                    <div class="card-text">${escapeHtml(skill.description)}</div>
                    <div class="chips-container">
                        ${versionChips}
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
