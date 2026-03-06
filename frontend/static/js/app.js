const API_BASE = '/api';

let currentPage = 1;
let currentCategory = '';
let currentTag = '';
let currentSearch = '';
let totalPages = 1;

// 记录访问
async function recordVisit() {
    try {
        // 使用 localStorage 判断是否是新访客（长期有效）
        const visitorKey = 'blog_visitor_id';
        const today = new Date().toDateString();
        const sessionKey = 'visit_recorded_' + today;

        // 检查今天是否已经记录过访问
        const lastRecorded = sessionStorage.getItem(sessionKey);
        if (lastRecorded) {
            return;
        }

        // 检查是否是新访客
        const visitorId = localStorage.getItem(visitorKey);
        const isNewVisitor = !visitorId;

        // 如果是新访客，生成一个唯一标识
        if (isNewVisitor) {
            localStorage.setItem(visitorKey, 'visitor_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9));
        }

        await api('/stats/record?new_visitor=' + isNewVisitor, { method: 'POST' });
        sessionStorage.setItem(sessionKey, 'true');
    } catch (error) {
        // 静默失败，不影响用户体验
        console.error('Failed to record visit:', error);
    }
}

function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
    document.body.setAttribute('data-theme', savedTheme);

    document.querySelectorAll('#theme-toggle').forEach(btn => {
        btn.addEventListener('click', () => {
            const current = localStorage.getItem('theme') || 'light';
            const newTheme = current === 'light' ? 'dark' : 'light';
            localStorage.setItem('theme', newTheme);
            document.documentElement.setAttribute('data-theme', newTheme);
            document.body.setAttribute('data-theme', newTheme);
        });
    });
}

async function api(endpoint, options = {}) {
    const defaults = {
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include'
    };

    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...defaults,
        ...options,
        headers: {
            ...defaults.headers,
            ...options.headers
        }
    });

    const data = await response.json();
    return { response, data };
}

async function checkSession() {
    try {
        const { response, data } = await api('/session');
        return data;
    } catch (error) {
        return { authenticated: false };
    }
}

async function loadCategories() {
    try {
        const { data } = await api('/categories');
        const container = document.getElementById('categories');
        if (!container) return;

        container.innerHTML = `
            <button class="category-btn active" data-id="">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="3" width="7" height="7"></rect>
                    <rect x="14" y="3" width="7" height="7"></rect>
                    <rect x="14" y="14" width="7" height="7"></rect>
                    <rect x="3" y="14" width="7" height="7"></rect>
                </svg>
                全部
            </button>
        `;

        data.data.forEach(category => {
            const btn = document.createElement('button');
            btn.className = 'category-btn';
            btn.dataset.id = category.id;
            btn.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                </svg>
                ${category.name}
            `;
            container.appendChild(btn);
        });

        container.querySelectorAll('.category-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                container.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                currentCategory = btn.dataset.id;
                currentTag = ''; // 重置标签筛选
                document.querySelectorAll('.tag-btn').forEach(b => b.classList.remove('active'));
                currentPage = 1;
                loadArticles();
            });
        });
    } catch (error) {
        console.error('Failed to load categories:', error);
    }
}

// 加载标签云
async function loadTags() {
    const container = document.getElementById('tags-cloud');
    if (!container) return;

    try {
        const { data } = await api('/tags');
        if (!data.data || data.data.length === 0) {
            container.style.display = 'none';
            return;
        }

        container.innerHTML = data.data.map(tag => `
            <button class="tag-btn" data-id="${tag.id}">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
                    <line x1="7" y1="7" x2="7.01" y2="7"></line>
                </svg>
                ${tag.name}
                <span class="tag-count">${tag.article_count || 0}</span>
            </button>
        `).join('');

        container.querySelectorAll('.tag-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                // 切换标签选中状态
                const isActive = btn.classList.contains('active');
                container.querySelectorAll('.tag-btn').forEach(b => b.classList.remove('active'));
                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));

                if (isActive) {
                    currentTag = '';
                    document.querySelector('.category-btn[data-id=""]').classList.add('active');
                } else {
                    btn.classList.add('active');
                    currentTag = btn.dataset.id;
                }
                currentCategory = '';
                currentPage = 1;
                loadArticles();
            });
        });
    } catch (error) {
        console.error('Failed to load tags:', error);
    }
}

// 图片懒加载观察器
let lazyImageObserver = null;

function initLazyLoad() {
    if ('IntersectionObserver' in window) {
        lazyImageObserver = new IntersectionObserver((entries, observer) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    if (img.dataset.src) {
                        img.src = img.dataset.src;
                        img.classList.add('loaded');
                        img.removeAttribute('data-src');
                        observer.unobserve(img);
                    }
                }
            });
        }, {
            rootMargin: '100px 0px',
            threshold: 0.1
        });
    }
}

function observeLazyImages() {
    if (lazyImageObserver) {
        document.querySelectorAll('.lazy-image[data-src]').forEach(img => {
            lazyImageObserver.observe(img);
        });
    } else {
        // 降级处理：直接加载所有图片
        document.querySelectorAll('.lazy-image[data-src]').forEach(img => {
            img.src = img.dataset.src;
            img.removeAttribute('data-src');
        });
    }
}

async function loadArticles() {
    const container = document.getElementById('articles-container');
    const loading = document.getElementById('loading');
    const loadMoreBtn = document.getElementById('load-more-btn');

    if (!container) return;

    loading.classList.add('show');

    try {
        let url;
        if (currentTag) {
            // 按标签筛选 - API 返回格式不同
            url = `/tags/${currentTag}/articles?page=${currentPage}&page_size=9`;
        } else {
            // 按分类或搜索筛选
            url = `/articles?page=${currentPage}&page_size=9`;
            if (currentCategory) url += `&category_id=${currentCategory}`;
            if (currentSearch) url += `&search=${encodeURIComponent(currentSearch)}`;
        }

        const { data } = await api(url);

        // 同时加载 HTML 页面（仅在第一页且没有筛选条件时）
        let htmlPages = [];
        if (currentPage === 1 && !currentTag && !currentCategory && !currentSearch) {
            try {
                const htmlRes = await api('/html-pages?page=1&page_size=20');
                htmlPages = (htmlRes.data.data || []).map((hp) => ({
                    ...hp,
                    _type: 'htmlpage'
                }));
            } catch (e) {
                console.error('Failed to load HTML pages:', e);
            }
        }

        if (currentPage === 1) {
            container.innerHTML = '';
        }

        // 统一处理不同的 API 返回格式
        let articles = data.data;
        let pagination = data.pagination;

        // 标签 API 返回的是 data 直接是数组
        if (currentTag && Array.isArray(data.data)) {
            articles = data.data;
            pagination = data.pagination;
        }

        // 为文章添加类型标识
        articles = (articles || []).map(a => ({ ...a, _type: 'article' }));

        // 合并文章和 HTML 页面，按时间排序
        let allItems = [...articles, ...htmlPages];
        allItems.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

        if (allItems && allItems.length > 0) {
            allItems.forEach((item, index) => {
                const card = item._type === 'htmlpage'
                    ? createHtmlPageCard(item)
                    : createArticleCard(item);
                card.style.animationDelay = `${index * 0.1}s`;
                container.appendChild(card);
            });
            totalPages = pagination.total_page;
            // 观察新加载的图片
            observeLazyImages();
        } else {
            if (currentPage === 1) {
                container.innerHTML = `
                    <div style="grid-column: 1 / -1; text-align: center; padding: 80px 20px;">
                        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="margin: 0 auto 20px; opacity: 0.3;">
                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                            <polyline points="14 2 14 8 20 8"></polyline>
                            <line x1="16" y1="13" x2="8" y2="13"></line>
                            <line x1="16" y1="17" x2="8" y2="17"></line>
                            <polyline points="10 9 9 9 8 9"></polyline>
                        </svg>
                        <p style="color: var(--text-muted); font-size: 1.1rem;">暂无文章</p>
                        <p style="color: var(--text-muted); font-size: 0.9rem; margin-top: 8px;">开始创作你的第一篇文章吧！</p>
                    </div>
                `;
            }
        }

        if (loadMoreBtn) {
            loadMoreBtn.style.display = currentPage >= totalPages ? 'none' : 'flex';
        }
    } catch (error) {
        console.error('Failed to load articles:', error);
        container.innerHTML = `
            <div style="grid-column: 1 / -1; text-align: center; padding: 80px 20px;">
                <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="margin: 0 auto 20px; opacity: 0.3;">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <p style="color: var(--text-muted); font-size: 1.1rem;">加载失败</p>
                <p style="color: var(--text-muted); font-size: 0.9rem; margin-top: 8px;">请刷新页面重试</p>
            </div>
        `;
    }

    loading.classList.remove('show');
}

function createArticleCard(article) {
    const card = document.createElement('article');
    card.className = 'article-card' + (article.is_pinned ? ' pinned' : '');
    card.style.cursor = 'pointer';

    // 点击时保存滚动位置
    card.onclick = () => {
        sessionStorage.setItem('scrollPosition', window.pageYOffset);
        sessionStorage.setItem('scrollPage', currentPage);
        sessionStorage.setItem('scrollCategory', currentCategory);
        sessionStorage.setItem('scrollTag', currentTag);
        sessionStorage.setItem('scrollSearch', currentSearch);
        window.location.href = `/article.html?id=${article.id}`;
    };

    // 图片懒加载 - 使用 data-src
    const coverHtml = article.cover_image
        ? `<img data-src="${article.cover_image}" alt="${article.title}" class="article-cover lazy-image" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 400 200'%3E%3Crect fill='%23f1f5f9' width='400' height='200'/%3E%3C/svg%3E">`
        : `<div class="article-cover-placeholder">${article.title.charAt(0)}</div>`;

    const categoryName = article.category ? article.category.name : '未分类';
    const date = new Date(article.created_at).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });

    card.innerHTML = `
        ${coverHtml}
        <div class="article-content">
            <span class="article-category">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                </svg>
                ${categoryName}
            </span>
            <h2 class="article-title">${article.title}</h2>
            <p class="article-summary">${article.summary || ''}</p>
            <div class="article-meta">
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    ${date}
                </span>
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                        <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                    ${article.view_count || 0}
                </span>
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"></path>
                    </svg>
                    ${article.like_count || 0}
                </span>
                ${article.read_time ? `<span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <polyline points="12 6 12 12 16 14"></polyline>
                    </svg>
                    ${article.read_time} 分钟
                </span>` : ''}
            </div>
        </div>
    `;

    return card;
}

// 创建 HTML 页面卡片
function createHtmlPageCard(htmlpage) {
    const card = document.createElement('article');
    card.className = 'article-card';
    card.style.cursor = 'pointer';

    // 点击时在新窗口打开
    card.onclick = () => {
        window.open(`/html-viewer.html?id=${htmlpage.id}`, '_blank');
    };

    // 图片懒加载
    const coverHtml = htmlpage.cover_image
        ? `<img data-src="${htmlpage.cover_image}" alt="${htmlpage.title}" class="article-cover lazy-image" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 400 200'%3E%3Crect fill='%23f1f5f9' width='400' height='200'/%3E%3C/svg%3E">`
        : `<div class="article-cover-placeholder">${htmlpage.title.charAt(0)}</div>`;

    const categoryName = htmlpage.category ? htmlpage.category.name : 'HTML';
    const date = new Date(htmlpage.created_at).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });

    card.innerHTML = `
        ${coverHtml}
        <div class="article-content">
            <span class="article-category">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                </svg>
                ${categoryName}
            </span>
            <h2 class="article-title">${htmlpage.title}</h2>
            <div class="article-summary-wrapper">
                <p class="article-summary">${htmlpage.summary || ''}</p>
                ${htmlpage.summary && htmlpage.summary.length > 60 ? `<div class="article-summary-tooltip">${htmlpage.summary}</div>` : ''}
            </div>
            <div class="article-meta">
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    ${date}
                </span>
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                        <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                    ${htmlpage.view_count || 0}
                </span>
            </div>
        </div>
    `;

    return card;
}

async function loadArticle(id) {
    const container = document.getElementById('article-content');
    if (!container) return;

    try {
        const { data } = await api(`/articles/${id}`);
        const article = data.data;

        const date = new Date(article.created_at).toLocaleDateString('zh-CN', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
        const categoryName = article.category ? article.category.name : '未分类';

        // 封面图片
        const coverHtml = article.cover_image
            ? `<div class="article-cover-wrapper">
                <img src="${article.cover_image}" alt="${article.title}" class="article-cover-image">
               </div>`
            : '';

        // 标签
        const tagsHtml = article.tags
            ? `<div class="article-tags">
                ${article.tags.split(',').map(tag => `
                    <span class="article-tag">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
                            <line x1="7" y1="7" x2="7.01" y2="7"></line>
                        </svg>
                        ${tag.trim()}
                    </span>
                `).join('')}
               </div>`
            : '';

        // 根据内容格式渲染内容
        let contentHtml;
        const isMarkdown = article.content_format === 'markdown';

        // 调试：输出内容格式
        console.log('Article content_format:', article.content_format, 'isMarkdown:', isMarkdown);

        if (isMarkdown && article.content) {
            // Markdown 渲染 - 配置 marked 支持 GFM 表格
            marked.setOptions({
                breaks: true,
                gfm: true,
                headerIds: true,
                mangle: false
            });
            // 确保表格渲染（GFM 风格）
            contentHtml = marked.parse(article.content);
        } else {
            // HTML 直接输出
            contentHtml = article.content || '';
        }

        container.innerHTML = `
            ${coverHtml}
            <header class="article-detail-header">
                <h1 class="article-detail-title">${article.title}</h1>
                <div class="article-detail-meta">
                    <span>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                        </svg>
                        ${categoryName}
                    </span>
                    <span>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                            <line x1="16" y1="2" x2="16" y2="6"></line>
                            <line x1="8" y1="2" x2="8" y2="6"></line>
                            <line x1="3" y1="10" x2="21" y2="10"></line>
                        </svg>
                        ${date}
                    </span>
                    <span>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                            <circle cx="12" cy="12" r="3"></circle>
                        </svg>
                        ${article.view_count || 0} 阅读
                    </span>
                    ${article.read_time ? `
                    <span>
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10"></circle>
                            <polyline points="12 6 12 12 16 14"></polyline>
                        </svg>
                        ${article.read_time} 分钟
                    </span>
                    ` : ''}
                </div>
                ${tagsHtml}
            </header>
            <div class="article-detail-content">
                ${contentHtml}
            </div>
        `;

        // 为标题添加 ID，用于大纲跳转
        addHeadingIds();

        // 代码高亮
        document.querySelectorAll('.article-detail-content pre code').forEach(block => {
            hljs.highlightElement(block);
        });

        // 生成大纲
        generateTOC();

        document.title = `${article.title} - Tech Blog`;

        // 更新 SEO meta 标签
        updateSEOMeta(article);

        // 更新底部点赞数
        const likeCountEl = document.getElementById('like-count');
        if (likeCountEl) {
            likeCountEl.textContent = article.like_count || 0;
        }
    } catch (error) {
        console.error('Failed to load article:', error);
        container.innerHTML = `
            <div style="text-align: center; padding: 80px 20px;">
                <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="margin: 0 auto 20px; opacity: 0.3;">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <p style="color: var(--text-muted); font-size: 1.1rem;">文章加载失败</p>
            </div>
        `;
    }
}

// 更新 SEO Meta 标签
function updateSEOMeta(article) {
    const pageUrl = window.location.href;
    const categoryName = article.category ? article.category.name : '未分类';

    // 更新 Open Graph 标签
    const ogTitle = document.getElementById('og-title');
    const ogDescription = document.getElementById('og-description');
    const ogUrl = document.getElementById('og-url');
    const ogImage = document.getElementById('og-image');
    const ogPublishedTime = document.getElementById('og-published-time');
    const ogModifiedTime = document.getElementById('og-modified-time');
    const ogSection = document.getElementById('og-section');
    const ogTag = document.getElementById('og-tag');

    if (ogTitle) ogTitle.setAttribute('content', article.title);
    if (ogDescription) ogDescription.setAttribute('content', article.summary || article.title);
    if (ogUrl) ogUrl.setAttribute('content', pageUrl);
    if (ogImage) ogImage.setAttribute('content', article.cover_image || '');
    if (ogPublishedTime) ogPublishedTime.setAttribute('content', article.created_at);
    if (ogModifiedTime) ogModifiedTime.setAttribute('content', article.updated_at || article.created_at);
    if (ogSection) ogSection.setAttribute('content', categoryName);
    if (ogTag) ogTag.setAttribute('content', article.tags || '');

    // 更新 Twitter Card 标签
    const twitterTitle = document.getElementById('twitter-title');
    const twitterDescription = document.getElementById('twitter-description');
    const twitterImage = document.getElementById('twitter-image');

    if (twitterTitle) twitterTitle.setAttribute('content', article.title);
    if (twitterDescription) twitterDescription.setAttribute('content', article.summary || article.title);
    if (twitterImage) twitterImage.setAttribute('content', article.cover_image || '');

    // 更新 JSON-LD
    const jsonLdScript = document.getElementById('json-ld-article');
    if (jsonLdScript) {
        const jsonLd = {
            "@context": "https://schema.org",
            "@type": "Article",
            "headline": article.title,
            "description": article.summary || article.title,
            "image": article.cover_image || '',
            "datePublished": article.created_at,
            "dateModified": article.updated_at || article.created_at,
            "author": {
                "@type": "Person",
                "name": "Tech Blog Author"
            },
            "publisher": {
                "@type": "Organization",
                "name": "Tech Blog",
                "logo": {
                    "@type": "ImageObject",
                    "url": window.location.origin + '/static/img/logo.png'
                }
            },
            "mainEntityOfPage": {
                "@type": "WebPage",
                "@id": pageUrl
            },
            "articleSection": categoryName,
            "keywords": article.tags || ''
        };
        jsonLdScript.textContent = JSON.stringify(jsonLd);
    }
}

// 为标题添加 ID
function addHeadingIds() {
    const content = document.querySelector('.article-detail-content');
    if (!content) return;

    const headings = content.querySelectorAll('h1, h2, h3, h4, h5, h6');
    headings.forEach((heading, index) => {
        if (!heading.id) {
            heading.id = 'heading-' + index;
        }
    });
}

// 生成目录大纲
function generateTOC() {
    const content = document.querySelector('.article-detail-content');
    const tocNav = document.getElementById('toc-nav');
    const tocSidebar = document.getElementById('toc-sidebar');

    if (!content || !tocNav) return;

    const headings = content.querySelectorAll('h2, h3, h4, h5, h6');

    if (headings.length === 0) {
        if (tocSidebar) tocSidebar.style.display = 'none';
        return;
    }

    let tocHtml = '';
    headings.forEach((heading, index) => {
        const level = heading.tagName.toLowerCase();
        const id = heading.id || 'heading-' + index;
        const text = heading.textContent.trim();

        tocHtml += `<a href="#${id}" class="toc-${level}" data-target="${id}">${text}</a>`;
    });

    tocNav.innerHTML = tocHtml;

    // 绑定点击事件
    tocNav.querySelectorAll('a').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const targetId = link.getAttribute('data-target');
            const targetEl = document.getElementById(targetId);
            if (targetEl) {
                const headerOffset = 100;
                const elementPosition = targetEl.getBoundingClientRect().top;
                const offsetPosition = elementPosition + window.pageYOffset - headerOffset;

                window.scrollTo({
                    top: offsetPosition,
                    behavior: 'smooth'
                });
            }
        });
    });

    // 滚动监听，高亮当前标题
    setupScrollSpy(headings);
}

// 滚动监听
function setupScrollSpy(headings) {
    const tocLinks = document.querySelectorAll('#toc-nav a');

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.id;
                tocLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('data-target') === id) {
                        link.classList.add('active');
                    }
                });
            }
        });
    }, {
        rootMargin: '-100px 0px -70% 0px'
    });

    headings.forEach(heading => observer.observe(heading));
}

async function loadComments(articleId) {
    const container = document.getElementById('comments-list');
    if (!container) return;

    try {
        const { data } = await api(`/comments?article_id=${articleId}`);

        if (data.data && data.data.length > 0) {
            container.innerHTML = data.data.map(comment => `
                <div class="comment-item">
                    <div class="comment-header">
                        <span class="comment-author">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                                <circle cx="12" cy="7" r="4"></circle>
                            </svg>
                            ${comment.nickname}
                        </span>
                        <span class="comment-date">${new Date(comment.created_at).toLocaleString('zh-CN')}</span>
                    </div>
                    <div class="comment-content">${comment.content}</div>
                </div>
            `).join('');
        } else {
            container.innerHTML = `
                <div style="text-align: center; padding: 40px 20px; color: var(--text-muted);">
                    <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="margin: 0 auto 16px; opacity: 0.3;">
                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                    </svg>
                    <p>暂无评论，来抢沙发吧！</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Failed to load comments:', error);
    }
}

async function submitComment(articleId, nickname, email, content) {
    try {
        const { response, data } = await api('/comments', {
            method: 'POST',
            body: JSON.stringify({
                article_id: parseInt(articleId),
                nickname,
                email,
                content
            })
        });

        if (response.ok) {
            loadComments(articleId);
            return true;
        }
        return false;
    } catch (error) {
        console.error('Failed to submit comment:', error);
        return false;
    }
}

function initSearch() {
    const searchInput = document.getElementById('search-input');
    const searchBtn = document.getElementById('search-btn');

    if (!searchInput || !searchBtn) return;

    const doSearch = () => {
        currentSearch = searchInput.value.trim();
        currentPage = 1;
        loadArticles();
    };

    searchBtn.addEventListener('click', doSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') doSearch();
    });
}

function initLoadMore() {
    const loadMoreBtn = document.getElementById('load-more-btn');
    if (!loadMoreBtn) return;

    loadMoreBtn.addEventListener('click', () => {
        currentPage++;
        loadArticles();
    });
}

function initCommentForm() {
    const submitBtn = document.getElementById('submit-comment');
    if (!submitBtn) return;

    submitBtn.addEventListener('click', async () => {
        const urlParams = new URLSearchParams(window.location.search);
        const articleId = urlParams.get('id');
        if (!articleId) return;

        const nickname = document.getElementById('comment-nickname').value.trim();
        const email = document.getElementById('comment-email').value.trim();
        const content = document.getElementById('comment-content').value.trim();

        if (!nickname || !content) {
            showToast('请填写昵称和评论内容', 'error');
            return;
        }

        const success = await submitComment(articleId, nickname, email, content);
        if (success) {
            document.getElementById('comment-content').value = '';
            showToast('评论发表成功！');
        } else {
            showToast('评论发表失败，请重试', 'error');
        }
    });
}

function initScrollAnimations() {
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    document.querySelectorAll('.article-card, .about-card, .comments-section').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(20px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });
}

function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

function initParallaxEffect() {
    const hero = document.querySelector('.hero');
    if (!hero) return;

    window.addEventListener('scroll', () => {
        const scrolled = window.pageYOffset;
        const rate = scrolled * -0.3;
        hero.style.transform = `translateY(${rate}px)`;
    });
}

document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initLazyLoad(); // 初始化图片懒加载
    initSearch();
    initLoadMore();
    initCommentForm();
    initSmoothScroll();
    initParallaxEffect();
    recordVisit(); // 记录访问

    if (document.getElementById('articles-container')) {
        // 检查是否有保存的滚动位置
        const savedPosition = sessionStorage.getItem('scrollPosition');
        const savedPage = sessionStorage.getItem('scrollPage');
        const savedCategory = sessionStorage.getItem('scrollCategory');
        const savedTag = sessionStorage.getItem('scrollTag');
        const savedSearch = sessionStorage.getItem('scrollSearch');

        // 清除保存的位置（只恢复一次）
        sessionStorage.removeItem('scrollPosition');
        sessionStorage.removeItem('scrollPage');
        sessionStorage.removeItem('scrollCategory');
        sessionStorage.removeItem('scrollTag');
        sessionStorage.removeItem('scrollSearch');

        // 如果有保存的状态，先恢复筛选条件
        if (savedPosition !== null) {
            currentCategory = savedCategory || '';
            currentTag = savedTag || '';
            currentSearch = savedSearch || '';
        }

        loadCategories();
        loadTags();

        // 如果需要恢复到第N页，需要依次加载所有页面
        const targetPage = savedPosition !== null ? (parseInt(savedPage) || 1) : 1;

        const loadAllPages = async () => {
            for (let p = 1; p <= targetPage; p++) {
                currentPage = p;
                await loadArticles();
            }
            // 恢复滚动位置
            if (savedPosition !== null) {
                setTimeout(() => {
                    window.scrollTo({
                        top: parseInt(savedPosition),
                        behavior: 'instant'
                    });
                }, 100);
            }
        };

        loadAllPages();
        loadAnnouncements();
        loadFriendLinks();
        setTimeout(initScrollAnimations, 500);
    }
});

if ('scrollRestoration' in history) {
    history.scrollRestoration = 'manual';
}

// 加载公告
async function loadAnnouncements() {
    try {
        const { data } = await api('/announcements');
        if (data.data && data.data.length > 0) {
            const announcement = data.data[0];
            const banner = document.getElementById('announcement-banner');
            if (banner) {
                banner.innerHTML = `
                    <div class="announcement-content">
                        <div class="announcement-title">${announcement.title}</div>
                        <div class="announcement-text">${announcement.content}</div>
                    </div>
                    <button class="announcement-close" onclick="this.parentElement.style.display='none'">&times;</button>
                `;
                banner.style.display = 'flex';
            }
        }
    } catch (error) {
        console.error('Failed to load announcements:', error);
    }
}

// 加载友情链接
async function loadFriendLinks() {
    try {
        const { data } = await api('/friend-links');
        if (data.data && data.data.length > 0) {
            const section = document.getElementById('friend-links-section');
            const list = document.getElementById('friend-links-list');
            if (section && list) {
                list.innerHTML = data.data.map(link => {
                    const logoHtml = link.logo
                        ? `<img src="${link.logo}" alt="${link.name}" class="friend-link-logo">`
                        : `<div class="friend-link-logo-placeholder">${link.name.charAt(0).toUpperCase()}</div>`;
                    return `
                        <a href="${link.url}" target="_blank" rel="noopener noreferrer" class="friend-link-item" title="${link.desc || link.name}">
                            ${logoHtml}
                            <span>${link.name}</span>
                        </a>
                    `;
                }).join('');
                section.style.display = 'block';
            }
        }
    } catch (error) {
        console.error('Failed to load friend links:', error);
    }
}

// 创建文章卡片（更新版，支持置顶和点赞数）
function createArticleCard(article) {
    const card = document.createElement('article');
    card.className = 'article-card' + (article.is_pinned ? ' pinned' : '');
    card.onclick = () => window.location.href = `/article.html?id=${article.id}`;
    card.style.cursor = 'pointer';

    const coverHtml = article.cover_image
        ? `<img src="${article.cover_image}" alt="${article.title}" class="article-cover">`
        : `<div class="article-cover-placeholder">${article.title.charAt(0)}</div>`;

    const categoryName = article.category ? article.category.name : '未分类';
    const date = new Date(article.created_at).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });

    // 文章标签显示
    const tagsHtml = article.tags
        ? `<div class="card-tags">${article.tags.split(',').slice(0, 3).map(tag =>
            `<span class="card-tag" onclick="event.stopPropagation(); filterByTag('${tag.trim()}')">${tag.trim()}</span>`
        ).join('')}</div>`
        : '';

    card.innerHTML = `
        ${coverHtml}
        <div class="article-content">
            <span class="article-category">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                </svg>
                ${categoryName}
            </span>
            <h2 class="article-title">${article.title}</h2>
            <div class="article-summary-wrapper">
                <p class="article-summary">${article.summary || ''}</p>
                ${article.summary && article.summary.length > 60 ? `<div class="article-summary-tooltip">${article.summary}</div>` : ''}
            </div>
            ${tagsHtml}
            <div class="article-meta">
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    ${date}
                </span>
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                        <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                    ${article.view_count || 0}
                </span>
                <span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"></path>
                    </svg>
                    ${article.like_count || 0}
                </span>
                ${article.read_time ? `<span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <polyline points="12 6 12 12 16 14"></polyline>
                    </svg>
                    ${article.read_time} 分钟
                </span>` : ''}
            </div>
        </div>
    `;

    return card;
}

// 通过标签名筛选文章
async function filterByTag(tagName) {
    const container = document.getElementById('tags-cloud');
    if (!container) return;

    // 查找标签按钮
    const tagBtn = Array.from(container.querySelectorAll('.tag-btn')).find(btn =>
        btn.textContent.includes(tagName)
    );

    if (tagBtn) {
        tagBtn.click();
        // 滚动到文章列表
        document.getElementById('articles-container').scrollIntoView({ behavior: 'smooth' });
    }
}
