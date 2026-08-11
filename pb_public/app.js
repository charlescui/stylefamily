async function loadMembers() {
    const res = await fetch('/api/family_members');
    const members = await res.json();
    const list = document.getElementById('members-list');
    list.innerHTML = members.map(m => `
        <div class="member-card" data-id="${m.member_id}">
            <h3>${m.display_name}</h3>
            <div class="form-group">
                <label>昵称</label>
                <input type="text" class="display_name" value="${m.display_name}">
            </div>
            <div class="form-group">
                <label>风格偏好（逗号分隔）</label>
                <input type="text" class="style_preferences" value="${(m.style_preferences || []).join('，')}">
            </div>
            <div class="form-group">
                <label>喜欢颜色</label>
                <input type="text" class="favorite_colors" value="${(m.favorite_colors || []).join('，')}">
            </div>
            <div class="form-group">
                <label>避免颜色</label>
                <input type="text" class="avoid_colors" value="${(m.avoid_colors || []).join('，')}">
            </div>
            <div class="form-group">
                <label>尺码</label>
                <input type="text" class="size" value="${m.size || ''}">
            </div>
            <div class="form-group">
                <label>形象描述</label>
                <textarea class="model_prompt" rows="3">${m.model_prompt || ''}</textarea>
            </div>
        </div>
    `).join('');
}

document.getElementById('save-btn').addEventListener('click', async () => {
    const card = document.querySelector('.member-card');
    const id = card.dataset.id;
    const body = {
        id,
        display_name: card.querySelector('.display_name').value,
        style_preferences: card.querySelector('.style_preferences').value.split(/[,，]/).map(s => s.trim()).filter(Boolean),
        favorite_colors: card.querySelector('.favorite_colors').value.split(/[,，]/).map(s => s.trim()).filter(Boolean),
        avoid_colors: card.querySelector('.avoid_colors').value.split(/[,，]/).map(s => s.trim()).filter(Boolean),
        size: card.querySelector('.size').value,
        model_prompt: card.querySelector('.model_prompt').value,
    };
    await fetch('/api/family_members', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
    });
    alert('保存成功');
});

document.getElementById('generate-btn').addEventListener('click', async () => {
    const btn = document.getElementById('generate-btn');
    btn.disabled = true;
    btn.innerText = '生成中...';
    const res = await fetch('/api/generate_outfit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
    });
    const data = await res.json();
    btn.disabled = false;
    btn.innerText = '立即生成本周穿搭';
    alert(data.message);
});

loadMembers();
