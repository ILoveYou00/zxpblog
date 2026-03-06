console.log('=== ADMIN.JS LOADED - VERSION 2 ===');

// 显示 toast 提示
function showToast(message, type = 'success') {
    // 移除已存在的 toast
    const existingToast = document.querySelector('.toast-notification');
    if (existingToast) {
        existingToast.remove();
    }

    const toast = document.createElement('div');
    toast.className = `toast-notification toast-${type}`;
    toast.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            ${type === 'success'
                ? '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline>'
                : '<circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line>'
            }
        </svg>
        <span>${message}</span>
    `;
    document.body.appendChild(toast);

    // 触发动画
    setTimeout(() => toast.classList.add('show'), 10);

    // 2秒后自动消失
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 2000);
}

// Admin specific code
let quill;
let easyMDE;
let editingArticleId = null;
let currentFormat = 'html'; // 当前编辑器格式: html 或 markdown
let allTags = []; // 存储所有标签
let selectedTagIds = []; // 存储选中的标签ID (文章)
let selectedHtmlPageTagIds = []; // 存储选中的标签ID (HTML页面)

// 自动保存相关变量
let autoSaveTimer = null;
let lastSavedContent = '';
const AUTO_SAVE_INTERVAL = 30000; // 30秒自动保存间隔
const DRAFT_KEY_PREFIX = 'blog_draft_';

// 自动保存状态更新
function updateAutoSaveStatus(status, message) {
    const statusEl = document.getElementById('auto-save-status');
    if (!statusEl) return;

    statusEl.className = 'auto-save-status ' + status;
    const textEl = statusEl.querySelector('.auto-save-text');
    if (textEl) {
        textEl.textContent = message || getAutoSaveMessage(status);
    }
}

function getAutoSaveMessage(status) {
    const messages = {
        '': '自动保存已就绪',
        'saving': '正在保存...',
        'saved': '已自动保存',
        'error': '保存失败'
    };
    return messages[status] || messages[''];
}

// 获取当前编辑器内容
function getCurrentEditorContent() {
    if (currentFormat === 'markdown') {
        return easyMDE ? easyMDE.value() : '';
    } else {
        return quill ? quill.root.innerHTML : '';
    }
}

// 获取草稿数据
function getDraftData() {
    const data = {
        title: document.getElementById('article-title')?.value || '',
        slug: document.getElementById('article-slug')?.value || '',
        summary: document.getElementById('article-summary')?.value || '',
        category_id: document.getElementById('article-category')?.value || '',
        cover_image: document.getElementById('article-cover')?.value || '',
        is_published: document.getElementById('article-published')?.checked || false,
        is_pinned: document.getElementById('article-pinned')?.checked || false,
        content_format: currentFormat,
        content: getCurrentEditorContent(),
        selected_tag_ids: selectedTagIds,
        timestamp: Date.now()
    };
    return data;
}

// 恢复草稿数据
function restoreDraftData(draft) {
    if (!draft) return false;

    document.getElementById('article-title').value = draft.title || '';
    document.getElementById('article-slug').value = draft.slug || '';
    document.getElementById('article-summary').value = draft.summary || '';
    document.getElementById('article-category').value = draft.category_id || '';
    document.getElementById('article-cover').value = draft.cover_image || '';
    document.getElementById('article-published').checked = draft.is_published || false;
    document.getElementById('article-pinned').checked = draft.is_pinned || false;

    // 恢复格式和内容
    if (draft.content_format === 'markdown') {
        document.getElementById('article-format').value = 'markdown';
        switchEditor('markdown');
        if (easyMDE && draft.content) {
            easyMDE.value(draft.content);
        }
    } else {
        document.getElementById('article-format').value = 'html';
        switchEditor('html');
        if (quill && draft.content) {
            quill.root.innerHTML = draft.content;
        }
    }

    // 恢复标签
    if (draft.selected_tag_ids) {
        selectedTagIds = draft.selected_tag_ids;
        updateTagsSelection();
    }

    lastSavedContent = draft.content || '';
    return true;
}

// 保存草稿到 localStorage
function saveDraftToLocal() {
    const draft = getDraftData();
    const key = editingArticleId ? DRAFT_KEY_PREFIX + editingArticleId : DRAFT_KEY_PREFIX + 'new';

    try {
        localStorage.setItem(key, JSON.stringify(draft));
        return true;
    } catch (e) {
        console.error('Failed to save draft:', e);
        return false;
    }
}

// 从 localStorage 加载草稿
function loadDraftFromLocal() {
    const key = editingArticleId ? DRAFT_KEY_PREFIX + editingArticleId : DRAFT_KEY_PREFIX + 'new';

    try {
        const draftStr = localStorage.getItem(key);
        return draftStr ? JSON.parse(draftStr) : null;
    } catch (e) {
        console.error('Failed to load draft:', e);
        return null;
    }
}

// 清除草稿
function clearDraft() {
    const key = editingArticleId ? DRAFT_KEY_PREFIX + editingArticleId : DRAFT_KEY_PREFIX + 'new';
    localStorage.removeItem(key);
}

// 自动保存检查
function autoSaveCheck() {
    const currentContent = getCurrentEditorContent();
    const currentTitle = document.getElementById('article-title')?.value || '';

    // 只有在内容有变化时才保存
    if (currentContent !== lastSavedContent || currentTitle !== lastSavedTitle) {
        performAutoSave();
    }
}

let lastSavedTitle = '';

// 执行自动保存
function performAutoSave() {
    updateAutoSaveStatus('saving');

    const success = saveDraftToLocal();

    if (success) {
        lastSavedContent = getCurrentEditorContent();
        lastSavedTitle = document.getElementById('article-title')?.value || '';
        updateAutoSaveStatus('saved');
    } else {
        updateAutoSaveStatus('error');
    }
}

// 启动自动保存
function startAutoSave() {
    if (autoSaveTimer) {
        clearInterval(autoSaveTimer);
    }
    autoSaveTimer = setInterval(autoSaveCheck, AUTO_SAVE_INTERVAL);
    lastSavedContent = getCurrentEditorContent();
    lastSavedTitle = document.getElementById('article-title')?.value || '';
}

// 停止自动保存
function stopAutoSave() {
    if (autoSaveTimer) {
        clearInterval(autoSaveTimer);
        autoSaveTimer = null;
    }
}

// 检查是否有未保存的草稿
function checkForDraft() {
    const draft = loadDraftFromLocal();
    if (draft && draft.timestamp) {
        // 显示草稿恢复横幅
        const banner = document.getElementById('draft-recovery-banner');
        if (banner) {
            banner.classList.remove('hidden');
        }
        return draft;
    }
    return null;
}

// Initialize admin page
document.addEventListener('DOMContentLoaded', async () => {
    // Check authentication
    const session = await checkSession();
    if (!session.authenticated) {
        window.location.href = '/login.html';
        return;
    }

    // 验证通过，移除加载状态
    document.getElementById('auth-loading')?.remove();

    // Show admin username
    const userEl = document.getElementById('admin-user');
    if (userEl && session.user) {
        userEl.textContent = `欢迎, ${session.user.username}`;
    }

    // Initialize Quill editor
    quill = new Quill('#editor', {
        theme: 'snow',
        placeholder: '开始编写文章内容...',
        modules: {
            toolbar: [
                [{ 'header': [1, 2, 3, false] }],
                ['bold', 'italic', 'underline', 'strike'],
                ['blockquote', 'code-block'],
                [{ 'list': 'ordered' }, { 'list': 'bullet' }],
                ['link', 'image'],
                ['clean']
            ]
        }
    });

    // Initialize EasyMDE editor
    easyMDE = new EasyMDE({
        element: document.getElementById('markdown-editor'),
        autofocus: false,
        spellChecker: false,
        placeholder: '使用 Markdown 编写文章内容...',
        toolbar: [
            'bold', 'italic', 'heading', '|',
            'quote', 'unordered-list', 'ordered-list', '|',
            'link', 'image', 'code', '|',
            'preview', 'side-by-side', 'fullscreen', '|',
            'guide'
        ],
        status: ['autosave', 'lines', 'words'],
        previewRender: function(plainText) {
            return marked.parse(plainText);
        },
        forceSync: true  // 强制同步到 textarea
    });

    // Format selector change handler
    document.getElementById('article-format').addEventListener('change', function() {
        const format = this.value;
        switchEditor(format);
    });

    // Load initial data
    loadAdminArticles();
    loadAdminHtmlPages();
    loadAdminCategories();
    loadAdminComments();
    loadAllTagsForSelect(); // 加载标签用于选择
    loadAdminTags();
    loadAdminAnnouncements();
    loadAdminFriends();
    loadAdminStats();
    loadAboutSettings(); // 加载关于页面设置

    // Setup event listeners
    console.log('Setting up event listeners...');
    setupTabs();
    console.log('Tabs setup complete');
    setupModals();
    console.log('Modals setup complete');
    setupForms();
    setupLogout();
});

// Setup tab navigation
function setupTabs() {
    const tabs = document.querySelectorAll('.admin-tab');
    const contents = document.querySelectorAll('.admin-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const tabId = tab.dataset.tab;

            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            contents.forEach(content => {
                content.classList.add('hidden');
                if (content.id === `${tabId}-tab`) {
                    content.classList.remove('hidden');
                }
            });

            // 切换到统计 tab 时加载数据
            if (tabId === 'stats') {
                loadStats();
            }
        });
    });
}

// Setup modals
function setupModals() {
    console.log('setupModals function called');

    // Article modal
    const articleModal = document.getElementById('article-modal');
    const newArticleBtn = document.getElementById('new-article-btn');
    const closeModalBtn = document.getElementById('close-modal');
    const cancelBtn = document.getElementById('cancel-btn');

    console.log('Article modal elements:', { articleModal, newArticleBtn });

    if (newArticleBtn) {
        newArticleBtn.addEventListener('click', () => {
            editingArticleId = null;
            document.getElementById('modal-title').textContent = '新建文章';
            document.getElementById('article-form').reset();
            document.getElementById('article-pinned').checked = false;

            // 重置编辑器格式为富文本
            document.getElementById('article-format').value = 'html';
            switchEditor('html');

            // 清空编辑器内容
            quill.setContents([]);
            easyMDE.value('');

            // 重置标签选择
            selectedTagIds = [];
            updateTagsSelection();

            // 重置自动保存状态
            updateAutoSaveStatus('', '自动保存已就绪');
            lastSavedContent = '';
            lastSavedTitle = '';

            // 隐藏草稿恢复横幅
            const banner = document.getElementById('draft-recovery-banner');
            if (banner) banner.classList.add('hidden');

            // 检查是否有未保存的草稿
            const draft = checkForDraft();
            if (draft) {
                // 草稿恢复按钮事件
                const restoreBtn = document.getElementById('restore-draft-btn');
                const discardBtn = document.getElementById('discard-draft-btn');

                if (restoreBtn) {
                    restoreBtn.onclick = () => {
                        restoreDraftData(draft);
                        banner.classList.add('hidden');
                        startAutoSave();
                        updateAutoSaveStatus('saved', '草稿已恢复');
                    };
                }

                if (discardBtn) {
                    discardBtn.onclick = () => {
                        clearDraft();
                        banner.classList.add('hidden');
                        startAutoSave();
                    };
                }
            } else {
                startAutoSave();
            }

            articleModal.classList.add('show');
        });
    }

    const closeModal = () => {
        articleModal.classList.remove('show');
        stopAutoSave();
        // 隐藏草稿恢复横幅
        const banner = document.getElementById('draft-recovery-banner');
        if (banner) banner.classList.add('hidden');
    };

    [closeModalBtn, cancelBtn].forEach(btn => {
        if (btn) {
            btn.addEventListener('click', closeModal);
        }
    });

    // Category modal
    setupModal('category', 'new-category-btn');

    // Tag modal
    setupModal('tag', 'new-tag-btn');

    // Announcement modal
    setupModal('announcement', 'new-announcement-btn');

    // Friend modal
    setupModal('friend', 'new-friend-btn');

    console.log('About to setup HTML Page modal');

    // HTML Page modal
    setupModal('htmlpage', 'new-htmlpage-btn');

    console.log('HTML Page modal setup complete');

    // Close on backdrop click
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('show');
                if (modal.id === 'article-modal') {
                    stopAutoSave();
                }
            }
        });
    });
}

function setupModal(name, newBtnId) {
    const modal = document.getElementById(`${name}-modal`);
    const newBtn = document.getElementById(newBtnId);
    const closeBtn = document.getElementById(`close-${name}-modal`);
    const cancelBtn = document.getElementById(`cancel-${name}-btn`);

    console.log(`setupModal called for ${name}, modal:`, modal, 'btn:', newBtn);

    if (!modal) {
        console.warn(`Modal not found: ${name}-modal`);
        return;
    }

    if (newBtn) {
        newBtn.addEventListener('click', () => {
            console.log(`Button clicked for ${name}`);
            const titleEl = document.getElementById(`${name}-modal-title`);
            const formEl = document.getElementById(`${name}-form`);
            const idEl = document.getElementById(`${name}-id`);

            if (titleEl) titleEl.textContent = `新建${getChineseName(name)}`;
            if (formEl) formEl.reset();
            if (idEl) idEl.value = '';

            // 重置HTML页面的分类和标签选择
            if (name === 'htmlpage') {
                selectedHtmlPageTagIds = [];
                updateHtmlPageTagsSelection();
            }

            modal.classList.add('show');
        });
    } else {
        console.warn(`Button not found: ${newBtnId}`);
    }

    [closeBtn, cancelBtn].forEach(btn => {
        if (btn) {
            btn.addEventListener('click', () => {
                modal.classList.remove('show');
            });
        }
    });
}

function getChineseName(name) {
    const names = {
        'category': '分类',
        'tag': '标签',
        'announcement': '公告',
        'friend': '友链',
        'htmlpage': 'HTML页面'
    };
    return names[name] || name;
}

// Switch between Quill and EasyMDE editors
function switchEditor(format) {
    const quillGroup = document.getElementById('quill-editor-group');
    const mdGroup = document.getElementById('markdown-editor-group');

    if (format === 'markdown') {
        quillGroup.classList.add('hidden');
        mdGroup.classList.remove('hidden');
        currentFormat = 'markdown';
    } else {
        mdGroup.classList.add('hidden');
        quillGroup.classList.remove('hidden');
        currentFormat = 'html';
    }
}

// Setup forms
function setupForms() {
    // Article form
    document.getElementById('article-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        // 获取内容，根据当前格式选择编辑器
        let content;
        if (currentFormat === 'markdown') {
            content = easyMDE.value();
        } else {
            content = quill.root.innerHTML;
        }

        const articleData = {
            title: document.getElementById('article-title').value,
            slug: document.getElementById('article-slug').value,
            summary: document.getElementById('article-summary').value,
            content: content,
            content_format: currentFormat,
            cover_image: document.getElementById('article-cover').value,
            category_id: parseInt(document.getElementById('article-category').value) || 1,
            tags: selectedTagIds.map(id => {
                const tag = allTags.find(t => t.id === id);
                return tag ? tag.name : '';
            }).filter(Boolean).join(','),
            is_published: document.getElementById('article-published').checked,
            is_pinned: document.getElementById('article-pinned').checked
        };

        // 调试：输出提交的数据
        console.log('Submitting article with content_format:', currentFormat, articleData);

        try {
            let response;
            if (editingArticleId) {
                response = await api(`/admin/articles/${editingArticleId}`, {
                    method: 'PUT',
                    body: JSON.stringify(articleData)
                });
                // 更新文章标签关联
                await api(`/admin/articles/${editingArticleId}/tags`, {
                    method: 'PUT',
                    body: JSON.stringify({ tag_ids: selectedTagIds })
                });
            } else {
                response = await api('/admin/articles', {
                    method: 'POST',
                    body: JSON.stringify(articleData)
                });
                // 如果是新文章，创建后设置标签
                if (response.data && response.data.data && response.data.data.id) {
                    await api(`/admin/articles/${response.data.data.id}/tags`, {
                        method: 'PUT',
                        body: JSON.stringify({ tag_ids: selectedTagIds })
                    });
                }
            }

            if (response.response.ok) {
                showToast(editingArticleId ? '文章更新成功！' : '文章创建成功！');
                // 清除草稿并停止自动保存
                clearDraft();
                stopAutoSave();
                document.getElementById('article-modal').classList.remove('show');
                loadAdminArticles();
            } else {
                showToast('操作失败：' + (response.data.error || '未知错误'), 'error');
            }
        } catch (error) {
            showToast('网络错误，请重试', 'error');
        }
    });

    // Category form
    document.getElementById('category-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleFormSubmit('category');
    });

    // Tag form
    document.getElementById('tag-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleFormSubmit('tag');
    });

    // Announcement form
    document.getElementById('announcement-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = document.getElementById('announcement-id').value;
        const startValue = document.getElementById('announcement-start').value;
        const endValue = document.getElementById('announcement-end').value;

        // 将 datetime-local 格式转换为 RFC3339 格式
        const data = {
            title: document.getElementById('announcement-title').value,
            content: document.getElementById('announcement-content').value,
            is_active: document.getElementById('announcement-active').checked,
            start_time: startValue ? new Date(startValue).toISOString() : new Date().toISOString(),
            end_time: endValue ? new Date(endValue).toISOString() : new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString()
        };
        await handleFormSubmit('announcement', id, data);
    });

    // Friend form
    document.getElementById('friend-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = document.getElementById('friend-id').value;
        const data = {
            name: document.getElementById('friend-name').value,
            url: document.getElementById('friend-url').value,
            logo: document.getElementById('friend-logo').value,
            desc: document.getElementById('friend-desc').value,
            sort_order: parseInt(document.getElementById('friend-sort').value) || 0,
            is_active: document.getElementById('friend-active').checked
        };
        await handleFormSubmit('friend', id, data, 'friend-links');
    });

    // HTML Page form
    document.getElementById('htmlpage-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = document.getElementById('htmlpage-id').value;

        // 获取标签名称
        const tagNames = selectedHtmlPageTagIds.map(id => {
            const tag = allTags.find(t => t.id === id);
            return tag ? tag.name : '';
        }).filter(Boolean).join(',');

        const data = {
            title: document.getElementById('htmlpage-title').value,
            slug: document.getElementById('htmlpage-slug').value,
            summary: document.getElementById('htmlpage-summary').value,
            cover_image: document.getElementById('htmlpage-cover').value,
            content: document.getElementById('htmlpage-content').value,
            category_id: parseInt(document.getElementById('htmlpage-category').value) || 0,
            tags: tagNames,
            is_published: document.getElementById('htmlpage-published').checked
        };
        await handleFormSubmit('htmlpage', id, data, 'html-pages');
    });
}

async function handleFormSubmit(name, id = null, customData = null, apiPath = null) {
    const form = document.getElementById(`${name}-form`);
    const idField = document.getElementById(`${name}-id`);
    const actualId = id || idField.value;

    // 特殊处理路径名
    let path;
    if (apiPath) {
        path = apiPath;
    } else if (name === 'category') {
        path = 'categories';
    } else {
        path = `${name}s`;
    }

    const formData = customData || {
        name: document.getElementById(`${name}-name`).value,
        slug: document.getElementById(`${name}-slug`).value
    };

    try {
        let response;
        if (actualId) {
            response = await api(`/admin/${path}/${actualId}`, {
                method: 'PUT',
                body: JSON.stringify(formData)
            });
        } else {
            response = await api(`/admin/${path}`, {
                method: 'POST',
                body: JSON.stringify(formData)
            });
        }

        if (response.response.ok) {
            showToast('保存成功！');
            document.getElementById(`${name}-modal`).classList.remove('show');
            if (name === 'category') loadAdminCategories();
            else if (name === 'tag') {
                loadAdminTags();
                loadAllTagsForSelect(); // 刷新文章编辑器的标签选择
            }
            else if (name === 'announcement') loadAdminAnnouncements();
            else if (name === 'friend') loadAdminFriends();
            else if (name === 'htmlpage') loadAdminHtmlPages();
        } else {
            showToast('操作失败：' + (response.data.error || '未知错误'), 'error');
        }
    } catch (error) {
        showToast('网络错误，请重试', 'error');
    }
}

// Setup logout
function setupLogout() {
    document.getElementById('logout-btn').addEventListener('click', async () => {
        try {
            await api('/auth/logout', { method: 'POST' });
            window.location.href = '/login.html';
        } catch (error) {
            console.error('Logout failed:', error);
        }
    });
}

// Load admin articles
async function loadAdminArticles(page = 1) {
    const tbody = document.getElementById('articles-table-body');
    if (!tbody) return;

    try {
        const { data } = await api(`/admin/articles?page=${page}&page_size=10`);

        tbody.innerHTML = data.data.map(article => `
            <tr>
                <td>${article.id}</td>
                <td>${article.title}</td>
                <td>${article.category ? article.category.name : '未分类'}</td>
                <td>
                    <span class="status-badge ${article.is_published ? 'status-published' : 'status-draft'}">
                        ${article.is_published ? '已发布' : '草稿'}
                    </span>
                </td>
                <td>${article.is_pinned ? '是' : '否'}</td>
                <td>${article.view_count || 0} / ${article.like_count || 0}</td>
                <td>${new Date(article.created_at).toLocaleDateString('zh-CN')}</td>
                <td>
                    <div class="action-btns">
                        <button class="btn btn-sm btn-outline" onclick="previewArticle(${article.id})">预览</button>
                        <button class="btn btn-sm btn-outline" onclick="editArticle(${article.id})">编辑</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteArticle(${article.id})">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');

        // Update pagination
        const paginationEl = document.getElementById('articles-pagination');
        if (paginationEl && data.pagination) {
            let paginationHtml = '';
            for (let i = 1; i <= data.pagination.total_page; i++) {
                paginationHtml += `<button class="${i === page ? 'active' : ''}" onclick="loadAdminArticles(${i})">${i}</button>`;
            }
            paginationEl.innerHTML = paginationHtml;
        }
    } catch (error) {
        console.error('Failed to load articles:', error);
    }
}

// Load admin categories
async function loadAdminCategories() {
    const tbody = document.getElementById('categories-table-body');
    const select = document.getElementById('article-category');
    const htmlpageSelect = document.getElementById('htmlpage-category');

    try {
        const { data } = await api('/categories');

        if (tbody) {
            tbody.innerHTML = data.data.map(category => `
                <tr>
                    <td>${category.id}</td>
                    <td>${category.name}</td>
                    <td>${category.slug || '-'}</td>
                    <td>0</td>
                    <td>
                        <div class="action-btns">
                            <button class="btn btn-sm btn-outline" onclick="editCategory(${category.id}, '${category.name}', '${category.slug || ''}')">编辑</button>
                            <button class="btn btn-sm btn-danger" onclick="deleteCategory(${category.id})">删除</button>
                        </div>
                    </td>
                </tr>
            `).join('');
        }

        const categoryOptions = data.data.map(category =>
            `<option value="${category.id}">${category.name}</option>`
        ).join('');

        if (select) {
            select.innerHTML = categoryOptions;
        }

        if (htmlpageSelect) {
            htmlpageSelect.innerHTML = categoryOptions;
        }
    } catch (error) {
        console.error('Failed to load categories:', error);
    }
}

// Load admin tags
async function loadAdminTags() {
    const tbody = document.getElementById('tags-table-body');
    if (!tbody) return;

    try {
        const { data } = await api('/tags');
        tbody.innerHTML = data.data.map(tag => `
            <tr>
                <td>${tag.id}</td>
                <td>${tag.name}</td>
                <td>${tag.slug || '-'}</td>
                <td>${tag.article_count || 0}</td>
                <td>
                    <div class="action-btns">
                        <button class="btn btn-sm btn-outline" onclick="editTag(${tag.id}, '${tag.name}', '${tag.slug || ''}')">编辑</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteTag(${tag.id})">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load tags:', error);
    }
}

// Load admin comments
async function loadAdminComments(page = 1) {
    const tbody = document.getElementById('comments-table-body');
    if (!tbody) return;

    try {
        const { data } = await api(`/admin/comments?page=${page}&page_size=10`);

        tbody.innerHTML = data.data.map(comment => `
            <tr>
                <td>${comment.id}</td>
                <td>${comment.article ? comment.article.title : '未知文章'}</td>
                <td>${comment.nickname}</td>
                <td>${comment.content.substring(0, 50)}${comment.content.length > 50 ? '...' : ''}</td>
                <td>
                    <span class="status-badge ${comment.is_approved ? 'status-published' : 'status-draft'}">
                        ${comment.is_approved ? '已审核' : '待审核'}
                    </span>
                </td>
                <td>${new Date(comment.created_at).toLocaleString('zh-CN')}</td>
                <td>
                    <div class="action-btns">
                        ${!comment.is_approved ? `<button class="btn btn-sm btn-primary" onclick="approveComment(${comment.id})">通过</button>` : ''}
                        <button class="btn btn-sm btn-danger" onclick="deleteComment(${comment.id})">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load comments:', error);
    }
}

// Load admin announcements
async function loadAdminAnnouncements() {
    const tbody = document.getElementById('announcements-table-body');
    if (!tbody) return;

    try {
        const { data } = await api('/admin/announcements/all');
        tbody.innerHTML = data.data.map(ann => `
            <tr>
                <td>${ann.id}</td>
                <td>${ann.title}</td>
                <td>${ann.content.substring(0, 50)}${ann.content.length > 50 ? '...' : ''}</td>
                <td>
                    <span class="status-badge ${ann.is_active ? 'status-published' : 'status-draft'}">
                        ${ann.is_active ? '启用' : '禁用'}
                    </span>
                </td>
                <td>${new Date(ann.start_time).toLocaleDateString('zh-CN')} - ${new Date(ann.end_time).toLocaleDateString('zh-CN')}</td>
                <td>
                    <div class="action-btns">
                        <button class="btn btn-sm btn-outline" onclick="editAnnouncement(${ann.id})">编辑</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteAnnouncement(${ann.id})">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load announcements:', error);
    }
}

// Load admin friends
async function loadAdminFriends() {
    const tbody = document.getElementById('friends-table-body');
    if (!tbody) return;

    try {
        const { data } = await api('/admin/friend-links/all');
        tbody.innerHTML = data.data.map(friend => `
            <tr>
                <td>${friend.id}</td>
                <td>${friend.name}</td>
                <td><a href="${friend.url}" target="_blank">${friend.url}</a></td>
                <td>${friend.desc || '-'}</td>
                <td>${friend.sort_order}</td>
                <td>
                    <span class="status-badge ${friend.is_active ? 'status-published' : 'status-draft'}">
                        ${friend.is_active ? '启用' : '禁用'}
                    </span>
                </td>
                <td>
                    <div class="action-btns">
                        <button class="btn btn-sm btn-outline" onclick="editFriend(${friend.id})">编辑</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteFriend(${friend.id})">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load friends:', error);
    }
}

// Load admin HTML pages
async function loadAdminHtmlPages(page = 1) {
    const tbody = document.getElementById('htmlpages-table-body');
    if (!tbody) return;

    tbody.innerHTML = '<tr><td colspan="7" style="text-align: center;">加载中...</td></tr>';

    try {
        const { data } = await api(`/admin/html-pages/all?page=${page}&page_size=10`);

        if (data.data && data.data.length > 0) {
            tbody.innerHTML = data.data.map(htmlpage => `
                <tr>
                    <td>${htmlpage.id}</td>
                    <td>${htmlpage.title}</td>
                    <td>${htmlpage.category ? htmlpage.category.name : '-'}</td>
                    <td>
                        <span class="status-badge ${htmlpage.is_published ? 'status-published' : 'status-draft'}">
                            ${htmlpage.is_published ? '已发布' : '草稿'}
                        </span>
                    </td>
                    <td>${htmlpage.view_count || 0}</td>
                    <td>${new Date(htmlpage.created_at).toLocaleDateString('zh-CN')}</td>
                    <td>
                        <div class="action-btns">
                            <button class="btn btn-sm btn-outline" onclick="previewHtmlPage(${htmlpage.id})">预览</button>
                            <button class="btn btn-sm btn-outline" onclick="editHtmlPage(${htmlpage.id})">编辑</button>
                            <button class="btn btn-sm btn-danger" onclick="deleteHtmlPage(${htmlpage.id})">删除</button>
                        </div>
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: var(--text-muted);">暂无HTML页面</td></tr>';
        }

        // Update pagination
        const paginationEl = document.getElementById('htmlpages-pagination');
        if (paginationEl && data.pagination) {
            let paginationHtml = '';
            for (let i = 1; i <= data.pagination.total_page; i++) {
                paginationHtml += `<button class="${i === page ? 'active' : ''}" onclick="loadAdminHtmlPages(${i})">${i}</button>`;
            }
            paginationEl.innerHTML = paginationHtml;
        }
    } catch (error) {
        console.error('Failed to load HTML pages:', error);
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: var(--text-muted);">暂无HTML页面</td></tr>';
    }
}

// Preview HTML page
function previewHtmlPage(id) {
    window.open(`/html-viewer.html?id=${id}`, '_blank');
}

// Edit HTML page
async function editHtmlPage(id) {
    try {
        const { data } = await api(`/html-pages/${id}`);
        const htmlpage = data.data;

        document.getElementById('htmlpage-modal-title').textContent = '编辑HTML页面';
        document.getElementById('htmlpage-id').value = id;
        document.getElementById('htmlpage-title').value = htmlpage.title;
        document.getElementById('htmlpage-slug').value = htmlpage.slug || '';
        document.getElementById('htmlpage-summary').value = htmlpage.summary || '';
        document.getElementById('htmlpage-cover').value = htmlpage.cover_image || '';
        document.getElementById('htmlpage-content').value = htmlpage.content || '';
        document.getElementById('htmlpage-category').value = htmlpage.category_id || '';
        document.getElementById('htmlpage-published').checked = htmlpage.is_published;

        // 恢复标签选择
        selectedHtmlPageTagIds = [];
        if (htmlpage.tags) {
            const tagNames = htmlpage.tags.split(',').map(t => t.trim()).filter(Boolean);
            selectedHtmlPageTagIds = tagNames.map(name => {
                const tag = allTags.find(t => t.name === name);
                return tag ? tag.id : null;
            }).filter(id => id !== null);
        }
        updateHtmlPageTagsSelection();

        document.getElementById('htmlpage-modal').classList.add('show');
    } catch (error) {
        console.error('Failed to load HTML page:', error);
        showToast('加载HTML页面失败', 'error');
    }
}

// Delete HTML page
async function deleteHtmlPage(id) {
    if (!confirm('确定要删除这个HTML页面吗？')) return;

    try {
        const { response } = await api(`/admin/html-pages/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('HTML页面已删除');
            loadAdminHtmlPages();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// 导入 HTML 文件
function importHtmlFile() {
    const fileInput = document.getElementById('htmlpage-file');
    const file = fileInput.files[0];

    if (!file) {
        showToast('请先选择一个HTML文件', 'error');
        return;
    }

    // 检查文件类型
    if (!file.name.match(/\.(html|htm)$/i)) {
        showToast('请选择 .html 或 .htm 文件', 'error');
        return;
    }

    const reader = new FileReader();
    reader.onload = function(e) {
        const content = e.target.result;

        // 填充内容到文本框
        document.getElementById('htmlpage-content').value = content;

        // 尝试从 HTML 中提取标题
        const titleMatch = content.match(/<title[^>]*>([^<]+)<\/title>/i);
        if (titleMatch && titleMatch[1]) {
            const titleField = document.getElementById('htmlpage-title');
            if (!titleField.value) {
                titleField.value = titleMatch[1].trim();
            }
        }

        showToast('文件导入成功！');
    };
    reader.onerror = function() {
        showToast('文件读取失败', 'error');
    };
    reader.readAsText(file);
}

// Load admin stats
async function loadAdminStats() {
    const container = document.getElementById('stats-grid');
    if (!container) return;

    try {
        const { data } = await api('/stats');
        const stats = data.data;
        container.innerHTML = `
            <div class="stat-card">
                <div class="stat-value">${stats.article_count}</div>
                <div class="stat-label">文章总数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.published_count}</div>
                <div class="stat-label">已发布</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.comment_count}</div>
                <div class="stat-label">评论数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.view_count}</div>
                <div class="stat-label">总浏览量</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.like_count}</div>
                <div class="stat-label">总点赞数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.tag_count}</div>
                <div class="stat-label">标签数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.category_count}</div>
                <div class="stat-label">分类数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${stats.today_views}</div>
                <div class="stat-label">今日访问</div>
            </div>
        `;
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}

// Preview article
function previewArticle(id) {
    window.open(`/article.html?id=${id}`, '_blank');
}

// Edit article
async function editArticle(id) {
    try {
        const { data } = await api(`/articles/${id}`);
        const article = data.data;

        // 获取文章的标签
        const tagsRes = await api(`/admin/articles/${id}/tags`);
        selectedTagIds = tagsRes.data.data ? tagsRes.data.data.map(t => t.id) : [];
        updateTagsSelection();

        editingArticleId = id;
        document.getElementById('modal-title').textContent = '编辑文章';
        document.getElementById('article-id').value = id;
        document.getElementById('article-title').value = article.title;
        document.getElementById('article-slug').value = article.slug || '';
        document.getElementById('article-summary').value = article.summary || '';
        document.getElementById('article-cover').value = article.cover_image || '';
        document.getElementById('article-category').value = article.category_id || '';
        document.getElementById('article-published').checked = article.is_published;
        document.getElementById('article-pinned').checked = article.is_pinned;

        // 根据内容格式设置编辑器
        const contentFormat = article.content_format || 'html';
        document.getElementById('article-format').value = contentFormat;

        // 先显示 modal，确保编辑器可见
        document.getElementById('article-modal').classList.add('show');

        // 等待 modal 显示后再切换编辑器和设置内容
        requestAnimationFrame(() => {
            switchEditor(contentFormat);

            // 使用 setTimeout 确保编辑器完全渲染后再设置内容
            setTimeout(() => {
                if (contentFormat === 'markdown') {
                    if (easyMDE) {
                        easyMDE.value(article.content || '');
                        easyMDE.codemirror.refresh();
                    }
                } else {
                    if (quill) {
                        quill.root.innerHTML = article.content || '';
                    }
                }
            }, 100);
        });

        // 重置自动保存状态并启动
        updateAutoSaveStatus('', '自动保存已就绪');
        lastSavedContent = article.content || '';
        lastSavedTitle = article.title || '';

        // 隐藏草稿恢复横幅
        const banner = document.getElementById('draft-recovery-banner');
        if (banner) banner.classList.add('hidden');

        // 检查是否有该文章的草稿
        const draft = checkForDraft();
        if (draft) {
            const restoreBtn = document.getElementById('restore-draft-btn');
            const discardBtn = document.getElementById('discard-draft-btn');

            if (restoreBtn) {
                restoreBtn.onclick = () => {
                    restoreDraftData(draft);
                    banner.classList.add('hidden');
                    startAutoSave();
                    updateAutoSaveStatus('saved', '草稿已恢复');
                };
            }

            if (discardBtn) {
                discardBtn.onclick = () => {
                    clearDraft();
                    banner.classList.add('hidden');
                    startAutoSave();
                };
            }
        } else {
            startAutoSave();
        }
    } catch (error) {
        console.error('Failed to load article:', error);
        showToast('加载文章失败', 'error');
    }
}

// Delete article
async function deleteArticle(id) {
    if (!confirm('确定要删除这篇文章吗？此操作不可恢复。')) return;

    try {
        const { response } = await api(`/admin/articles/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('文章已删除');
            loadAdminArticles();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// Load stats dashboard
async function loadStats() {
    try {
        const { data } = await api('/stats');
        const stats = data.data;

        // 更新概览卡片
        document.getElementById('stat-total-views').textContent = formatNumber(stats.view_count || 0);
        document.getElementById('stat-articles').textContent = stats.published_count || 0;
        document.getElementById('stat-likes').textContent = formatNumber(stats.like_count || 0);
        document.getElementById('stat-comments').textContent = stats.comment_count || 0;

        // 更新今日统计
        document.getElementById('stat-today-views').textContent = stats.today_views || 0;
        document.getElementById('stat-today-visitors').textContent = stats.today_visitors || 0;
        document.getElementById('stat-categories').textContent = stats.category_count || 0;
        document.getElementById('stat-tags').textContent = stats.tag_count || 0;

        // 渲染访问趋势图表
        renderViewsChart(stats.weekly_stats || []);

        // 渲染热门文章
        renderHotArticles(stats.hot_articles || []);

        // 渲染最新文章
        renderLatestArticles(stats.latest_articles || []);

    } catch (error) {
        console.error('Failed to load stats:', error);
        showToast('加载统计数据失败', 'error');
    }
}

// 格式化数字
function formatNumber(num) {
    if (num >= 10000) {
        return (num / 10000).toFixed(1) + 'w';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'k';
    }
    return num.toString();
}

// 渲染访问趋势图表
function renderViewsChart(weeklyStats) {
    const container = document.getElementById('chart-bars');
    if (!container) return;

    const days = ['日', '一', '二', '三', '四', '五', '六'];
    const today = new Date().getDay();

    // 获取最近7天的数据
    const chartData = [];
    for (let i = 6; i >= 0; i--) {
        const date = new Date();
        date.setDate(date.getDate() - i);
        const dateStr = date.toISOString().split('T')[0];
        const stat = weeklyStats.find(s => s.date.split('T')[0] === dateStr);
        chartData.push({
            day: days[(today - i + 7) % 7],
            views: stat ? stat.view_count : 0
        });
    }

    const maxViews = Math.max(...chartData.map(d => d.views), 1);

    container.innerHTML = chartData.map(d => `
        <div class="chart-bar-wrapper">
            <div class="chart-bar-value">${d.views}</div>
            <div class="chart-bar" style="height: ${Math.max((d.views / maxViews) * 150, 4)}px"></div>
            <div class="chart-bar-label">周${d.day}</div>
        </div>
    `).join('');
}

// 渲染热门文章
function renderHotArticles(articles) {
    const container = document.getElementById('hot-articles-list');
    if (!container) return;

    if (articles.length === 0) {
        container.innerHTML = '<div style="text-align: center; color: var(--text-muted); padding: 40px;">暂无数据</div>';
        return;
    }

    container.innerHTML = articles.map((article, index) => `
        <div class="hot-article-item">
            <div class="hot-article-info">
                <span class="article-rank ${index < 3 ? 'rank-' + (index + 1) : ''}">${index + 1}</span>
                <span class="hot-article-title">${article.title}</span>
            </div>
            <span class="hot-article-views">${article.view_count || 0} 浏览</span>
        </div>
    `).join('');
}

// 渲染最新文章
function renderLatestArticles(articles) {
    const container = document.getElementById('latest-articles-list');
    if (!container) return;

    if (articles.length === 0) {
        container.innerHTML = '<div style="text-align: center; color: var(--text-muted); padding: 40px;">暂无数据</div>';
        return;
    }

    container.innerHTML = articles.map(article => `
        <div class="latest-article-item">
            <div class="latest-article-info">
                <span class="latest-article-title">${article.title}</span>
            </div>
            <span class="latest-article-date">${new Date(article.created_at).toLocaleDateString('zh-CN')}</span>
        </div>
    `).join('');
}

// Edit category
function editCategory(id, name, slug) {
    document.getElementById('category-modal-title').textContent = '编辑分类';
    document.getElementById('category-id').value = id;
    document.getElementById('category-name').value = name;
    document.getElementById('category-slug').value = slug;
    document.getElementById('category-modal').classList.add('show');
}

// Delete category
async function deleteCategory(id) {
    if (!confirm('确定要删除这个分类吗？')) return;

    try {
        const { response } = await api(`/admin/categories/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('分类已删除');
            loadAdminCategories();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// Edit tag
function editTag(id, name, slug) {
    document.getElementById('tag-modal-title').textContent = '编辑标签';
    document.getElementById('tag-id').value = id;
    document.getElementById('tag-name').value = name;
    document.getElementById('tag-slug').value = slug;
    document.getElementById('tag-modal').classList.add('show');
}

// Delete tag
async function deleteTag(id) {
    if (!confirm('确定要删除这个标签吗？')) return;

    try {
        const { response } = await api(`/admin/tags/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('标签已删除');
            loadAdminTags();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// Edit announcement
async function editAnnouncement(id) {
    try {
        const { data } = await api(`/admin/announcements/all`);
        const ann = data.data.find(a => a.id === id);
        if (!ann) return;

        document.getElementById('announcement-modal-title').textContent = '编辑公告';
        document.getElementById('announcement-id').value = id;
        document.getElementById('announcement-title').value = ann.title;
        document.getElementById('announcement-content').value = ann.content;
        document.getElementById('announcement-active').checked = ann.is_active;
        // 设置时间字段（转换为 datetime-local 格式）
        if (ann.start_time) {
            document.getElementById('announcement-start').value = ann.start_time.slice(0, 16);
        }
        if (ann.end_time) {
            document.getElementById('announcement-end').value = ann.end_time.slice(0, 16);
        }
        document.getElementById('announcement-modal').classList.add('show');
    } catch (error) {
        showToast('加载失败', 'error');
    }
}

// Delete announcement
async function deleteAnnouncement(id) {
    if (!confirm('确定要删除这个公告吗？')) return;

    try {
        const { response } = await api(`/admin/announcements/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('公告已删除');
            loadAdminAnnouncements();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// Edit friend
async function editFriend(id) {
    try {
        const { data } = await api('/admin/friend-links/all');
        const friend = data.data.find(f => f.id === id);
        if (!friend) return;

        document.getElementById('friend-modal-title').textContent = '编辑友链';
        document.getElementById('friend-id').value = id;
        document.getElementById('friend-name').value = friend.name;
        document.getElementById('friend-url').value = friend.url;
        document.getElementById('friend-logo').value = friend.logo || '';
        document.getElementById('friend-desc').value = friend.desc || '';
        document.getElementById('friend-sort').value = friend.sort_order || 0;
        document.getElementById('friend-active').checked = friend.is_active;
        document.getElementById('friend-modal').classList.add('show');
    } catch (error) {
        showToast('加载失败', 'error');
    }
}

// Delete friend
async function deleteFriend(id) {
    if (!confirm('确定要删除这个友链吗？')) return;

    try {
        const { response } = await api(`/admin/friend-links/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('友链已删除');
            loadAdminFriends();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// Approve comment
async function approveComment(id) {
    try {
        const { response } = await api(`/admin/comments/${id}/approve`, { method: 'PUT' });
        if (response.ok) {
            showToast('评论已审核通过');
            loadAdminComments();
        }
    } catch (error) {
        showToast('操作失败', 'error');
    }
}

// Delete comment
async function deleteComment(id) {
    if (!confirm('确定要删除这条评论吗？')) return;

    try {
        const { response } = await api(`/admin/comments/${id}`, { method: 'DELETE' });
        if (response.ok) {
            showToast('评论已删除');
            loadAdminComments();
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// 加载所有标签用于文章编辑时的选择
async function loadAllTagsForSelect() {
    try {
        const { data } = await api('/tags');
        allTags = data.data || [];
        renderTagsSelection();
        renderHtmlPageTagsSelection();
    } catch (error) {
        console.error('Failed to load tags for select:', error);
    }
}

// 渲染标签选择器
function renderTagsSelection() {
    const container = document.getElementById('article-tags-container');
    if (!container) return;

    if (allTags.length === 0) {
        container.innerHTML = '<span style="color: var(--text-muted); font-size: 0.85rem;">暂无标签，请先在标签管理中创建</span>';
        return;
    }

    container.innerHTML = allTags.map(tag => `
        <label class="tag-checkbox ${selectedTagIds.includes(tag.id) ? 'checked' : ''}" data-tag-id="${tag.id}">
            <input type="checkbox" value="${tag.id}" ${selectedTagIds.includes(tag.id) ? 'checked' : ''}>
            <svg class="check-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            ${tag.name}
        </label>
    `).join('');

    // 绑定点击事件
    container.querySelectorAll('.tag-checkbox').forEach(label => {
        label.addEventListener('click', (e) => {
            e.preventDefault();
            const tagId = parseInt(label.dataset.tagId);
            toggleTagSelection(tagId);
        });
    });
}

// 切换标签选择
function toggleTagSelection(tagId) {
    const index = selectedTagIds.indexOf(tagId);
    if (index > -1) {
        selectedTagIds.splice(index, 1);
    } else {
        selectedTagIds.push(tagId);
    }
    updateTagsSelection();
}

// 更新标签选择的UI状态
function updateTagsSelection() {
    const container = document.getElementById('article-tags-container');
    if (!container) return;

    container.querySelectorAll('.tag-checkbox').forEach(label => {
        const tagId = parseInt(label.dataset.tagId);
        if (selectedTagIds.includes(tagId)) {
            label.classList.add('checked');
            label.querySelector('input').checked = true;
        } else {
            label.classList.remove('checked');
            label.querySelector('input').checked = false;
        }
    });
}

// 渲染HTML页面标签选择器
function renderHtmlPageTagsSelection() {
    const container = document.getElementById('htmlpage-tags-container');
    if (!container) return;

    if (allTags.length === 0) {
        container.innerHTML = '<span style="color: var(--text-muted); font-size: 0.85rem;">暂无标签，请先在标签管理中创建</span>';
        return;
    }

    container.innerHTML = allTags.map(tag => `
        <label class="tag-checkbox ${selectedHtmlPageTagIds.includes(tag.id) ? 'checked' : ''}" data-tag-id="${tag.id}">
            <input type="checkbox" value="${tag.id}" ${selectedHtmlPageTagIds.includes(tag.id) ? 'checked' : ''}>
            <svg class="check-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            ${tag.name}
        </label>
    `).join('');

    // 绑定点击事件
    container.querySelectorAll('.tag-checkbox').forEach(label => {
        label.addEventListener('click', (e) => {
            e.preventDefault();
            const tagId = parseInt(label.dataset.tagId);
            toggleHtmlPageTagSelection(tagId);
        });
    });
}

// 切换HTML页面标签选择
function toggleHtmlPageTagSelection(tagId) {
    const index = selectedHtmlPageTagIds.indexOf(tagId);
    if (index > -1) {
        selectedHtmlPageTagIds.splice(index, 1);
    } else {
        selectedHtmlPageTagIds.push(tagId);
    }
    updateHtmlPageTagsSelection();
}

// 更新HTML页面标签选择的UI状态
function updateHtmlPageTagsSelection() {
    const container = document.getElementById('htmlpage-tags-container');
    if (!container) return;

    container.querySelectorAll('.tag-checkbox').forEach(label => {
        const tagId = parseInt(label.dataset.tagId);
        if (selectedHtmlPageTagIds.includes(tagId)) {
            label.classList.add('checked');
            label.querySelector('input').checked = true;
        } else {
            label.classList.remove('checked');
            label.querySelector('input').checked = false;
        }
    });
}

// 加载关于页面设置
async function loadAboutSettings() {
    try {
        const { data } = await api('/about');
        const about = data.data;

        document.getElementById('about-name').value = about.name || '';
        document.getElementById('about-bio').value = about.bio || '';
        document.getElementById('about-avatar').value = about.avatar || '';
        document.getElementById('about-content').value = about.about_me || '';
        document.getElementById('about-skills').value = about.skills || '';
        document.getElementById('about-email').value = about.email || '';
        document.getElementById('about-github').value = about.github || '';
        document.getElementById('about-twitter').value = about.twitter || '';
        document.getElementById('about-website').value = about.website || '';
    } catch (error) {
        console.error('Failed to load about settings:', error);
    }
}

// 保存关于页面设置
async function saveAboutSettings(e) {
    e.preventDefault();

    const aboutData = {
        name: document.getElementById('about-name').value,
        bio: document.getElementById('about-bio').value,
        avatar: document.getElementById('about-avatar').value,
        about_me: document.getElementById('about-content').value,
        skills: document.getElementById('about-skills').value,
        email: document.getElementById('about-email').value,
        github: document.getElementById('about-github').value,
        twitter: document.getElementById('about-twitter').value,
        website: document.getElementById('about-website').value
    };

    try {
        const { response } = await api('/admin/about', {
            method: 'PUT',
            body: JSON.stringify(aboutData)
        });

        if (response.ok) {
            showToast('关于页面设置已保存！');
        } else {
            showToast('保存失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请重试', 'error');
    }
}

// 在表单设置中添加关于表单的事件监听
document.addEventListener('DOMContentLoaded', function() {
    const aboutForm = document.getElementById('about-form');
    if (aboutForm) {
        aboutForm.addEventListener('submit', saveAboutSettings);
    }

    // 初始化 AI 写作助手
    initAIWritingAssistant();
});

// AI 写作助手相关
let aiAssistant = null;

async function initAIWritingAssistant() {
    aiAssistant = new AIWritingAssistant();
}

// 富文本编辑器 AI 助手
async function aiWritingAssist(type) {
    if (!aiAssistant || !aiAssistant.enabled) {
        showToast('AI 功能未启用', 'error');
        return;
    }

    const content = quill.root.innerHTML;
    const plainText = quill.getText();

    if (!plainText.trim() && type !== 'title') {
        showToast('请先输入一些内容', 'error');
        return;
    }

    // 设置加载状态
    const btn = event.target;
    const originalText = btn.textContent;
    btn.textContent = '处理中...';
    btn.classList.add('loading');
    btn.disabled = true;

    try {
        const result = await aiAssistant.assist(type, plainText, content);

        if (type === 'title') {
            // 生成标题，填充到标题字段
            const titles = result.split('\n').filter(t => t.trim());
            if (titles.length > 0) {
                document.getElementById('article-title').value = titles[0].replace(/^[0-9.、\s]+/, '');
                showToast('标题已生成');
            }
        } else {
            // 其他类型，追加或替换内容
            quill.root.innerHTML += '\n\n' + result.replace(/\n/g, '<br>');
            showToast('内容已添加');
        }
    } catch (error) {
        showToast('AI 处理失败：' + error.message, 'error');
    } finally {
        btn.textContent = originalText;
        btn.classList.remove('loading');
        btn.disabled = false;
    }
}

// Markdown 编辑器 AI 助手
async function aiWritingAssistMD(type) {
    if (!aiAssistant || !aiAssistant.enabled) {
        showToast('AI 功能未启用', 'error');
        return;
    }

    const content = easyMDE.value();

    if (!content.trim() && type !== 'title') {
        showToast('请先输入一些内容', 'error');
        return;
    }

    // 设置加载状态
    const btn = event.target;
    const originalText = btn.textContent;
    btn.textContent = '处理中...';
    btn.classList.add('loading');
    btn.disabled = true;

    try {
        const result = await aiAssistant.assist(type, content);

        if (type === 'title') {
            const titles = result.split('\n').filter(t => t.trim());
            if (titles.length > 0) {
                document.getElementById('article-title').value = titles[0].replace(/^[0-9.、\s]+/, '');
                showToast('标题已生成');
            }
        } else {
            easyMDE.value(content + '\n\n' + result);
            showToast('内容已添加');
        }
    } catch (error) {
        showToast('AI 处理失败：' + error.message, 'error');
    } finally {
        btn.textContent = originalText;
        btn.classList.remove('loading');
        btn.disabled = false;
    }
}

// AI 生成摘要
async function aiGenerateSummary() {
    if (!aiAssistant || !aiAssistant.enabled) {
        showToast('AI 功能未启用', 'error');
        return;
    }

    // 获取当前编辑器内容
    let content;
    if (currentFormat === 'markdown') {
        content = easyMDE.value();
    } else {
        content = quill.getText();
    }

    if (!content.trim()) {
        showToast('请先输入文章内容', 'error');
        return;
    }

    // 设置加载状态
    const btn = event.target;
    const originalHTML = btn.innerHTML;
    btn.innerHTML = '生成中...';
    btn.disabled = true;

    try {
        const summary = await aiAssistant.generateSummary(content);
        document.getElementById('article-summary').value = summary;
        showToast('摘要已生成');
    } catch (error) {
        showToast('生成摘要失败：' + error.message, 'error');
    } finally {
        btn.innerHTML = originalHTML;
        btn.disabled = false;
    }
}

// AI 生成 HTML 页面摘要
async function aiGenerateHtmlPageSummary() {
    if (!aiAssistant || !aiAssistant.enabled) {
        showToast('AI 功能未启用', 'error');
        return;
    }

    const content = document.getElementById('htmlpage-content').value;

    if (!content.trim()) {
        showToast('请先输入HTML内容', 'error');
        return;
    }

    // 设置加载状态
    const btn = event.target;
    const originalHTML = btn.innerHTML;
    btn.innerHTML = '生成中...';
    btn.disabled = true;

    try {
        // 从 HTML 内容中提取纯文本用于生成摘要
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = content;
        const plainText = tempDiv.textContent || tempDiv.innerText || '';

        if (!plainText.trim()) {
            showToast('HTML内容中没有可提取的文本', 'error');
            return;
        }

        const summary = await aiAssistant.generateSummary(plainText);
        document.getElementById('htmlpage-summary').value = summary;
        showToast('摘要已生成');
    } catch (error) {
        showToast('生成摘要失败：' + error.message, 'error');
    } finally {
        btn.innerHTML = originalHTML;
        btn.disabled = false;
    }
}