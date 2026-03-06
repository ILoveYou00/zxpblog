// Tech Blog Service Worker
const CACHE_NAME = 'tech-blog-v2';
const STATIC_CACHE = 'tech-blog-static-v2';
const DYNAMIC_CACHE = 'tech-blog-dynamic-v2';

// 需要缓存的静态资源
const STATIC_ASSETS = [
    '/',
    '/index.html',
    '/article.html',
    '/archives.html',
    '/about.html',
    '/login.html',
    '/admin.html',
    '/static/css/style.css',
    '/static/js/app.js',
    '/static/js/admin.js',
    '/static/js/ai.js',
    '/manifest.json'
];

// 需要缓存的 CDN 资源
const CDN_ASSETS = [
    'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css',
    'https://cdn.jsdelivr.net/npm/marked/marked.min.js',
    'https://cdn.quilljs.com/1.3.7/quill.min.js',
    'https://cdn.quilljs.com/1.3.7/quill.snow.css',
    'https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.js',
    'https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.css',
    'https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap'
];

// 安装事件 - 缓存静态资源
self.addEventListener('install', (event) => {
    console.log('[Service Worker] Installing...');
    event.waitUntil(
        Promise.all([
            // 缓存本地静态资源
            caches.open(STATIC_CACHE).then((cache) => {
                console.log('[Service Worker] Caching static assets');
                return cache.addAll(STATIC_ASSETS);
            }),
            // 缓存 CDN 资源
            caches.open(DYNAMIC_CACHE).then((cache) => {
                console.log('[Service Worker] Caching CDN assets');
                return Promise.all(
                    CDN_ASSETS.map((url) =>
                        fetch(url)
                            .then((response) => {
                                if (response.ok) {
                                    return cache.put(url, response);
                                }
                            })
                            .catch(() => console.log('[Service Worker] Failed to cache:', url))
                    )
                );
            })
        ]).then(() => {
            console.log('[Service Worker] Installation complete');
            return self.skipWaiting();
        })
    );
});

// 激活事件 - 清理旧缓存
self.addEventListener('activate', (event) => {
    console.log('[Service Worker] Activating...');
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    if (cacheName !== STATIC_CACHE && cacheName !== DYNAMIC_CACHE) {
                        console.log('[Service Worker] Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        }).then(() => {
            console.log('[Service Worker] Activation complete');
            return self.clients.claim();
        })
    );
});

// 请求拦截策略
self.addEventListener('fetch', (event) => {
    const { request } = event;
    const url = new URL(request.url);

    // API 请求 - 网络优先策略
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(networkFirst(request));
        return;
    }

    // 静态资源 - 缓存优先策略
    if (isStaticAsset(url)) {
        event.respondWith(cacheFirst(request));
        return;
    }

    // HTML 页面 - 网络优先策略
    if (request.headers.get('accept')?.includes('text/html')) {
        event.respondWith(networkFirst(request));
        return;
    }

    // 其他请求 - 网络优先
    event.respondWith(networkFirst(request));
});

// 缓存优先策略
async function cacheFirst(request) {
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
        return cachedResponse;
    }

    try {
        const networkResponse = await fetch(request);
        if (networkResponse.ok) {
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, networkResponse.clone());
        }
        return networkResponse;
    } catch (error) {
        console.log('[Service Worker] Network request failed:', request.url);
        return caches.match(request) || new Response('Offline', { status: 503 });
    }
}

// 网络优先策略
async function networkFirst(request) {
    try {
        const networkResponse = await fetch(request);
        if (networkResponse.ok) {
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, networkResponse.clone());
        }
        return networkResponse;
    } catch (error) {
        console.log('[Service Worker] Network request failed, trying cache:', request.url);
        const cachedResponse = await caches.match(request);
        if (cachedResponse) {
            return cachedResponse;
        }

        // 如果是 HTML 请求，返回离线页面
        if (request.headers.get('accept')?.includes('text/html')) {
            return caches.match('/index.html');
        }

        return new Response('Offline', { status: 503 });
    }
}

// 判断是否是静态资源
function isStaticAsset(url) {
    const staticExtensions = ['.css', '.js', '.png', '.jpg', '.jpeg', '.gif', '.svg', '.ico', '.woff', '.woff2', '.ttf', '.eot'];
    return staticExtensions.some((ext) => url.pathname.endsWith(ext)) || CDN_ASSETS.some((cdn) => url.href.startsWith(cdn));
}

// 推送通知支持
self.addEventListener('push', (event) => {
    if (!event.data) return;

    const data = event.data.json();
    const options = {
        body: data.body || '新文章发布',
        icon: '/static/icons/icon-192x192.png',
        badge: '/static/icons/icon-72x72.png',
        vibrate: [100, 50, 100],
        data: {
            url: data.url || '/'
        },
        actions: [
            { action: 'open', title: '查看' },
            { action: 'close', title: '关闭' }
        ]
    };

    event.waitUntil(self.registration.showNotification(data.title || 'Tech Blog', options));
});

// 通知点击事件
self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    if (event.action === 'close') return;

    const urlToOpen = event.notification.data?.url || '/';

    event.waitUntil(
        clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
            // 检查是否已有打开的窗口
            for (const client of windowClients) {
                if (client.url.includes(self.location.origin) && 'focus' in client) {
                    client.navigate(urlToOpen);
                    return client.focus();
                }
            }
            // 没有打开的窗口，打开新窗口
            if (clients.openWindow) {
                return clients.openWindow(urlToOpen);
            }
        })
    );
});

// 后台同步支持
self.addEventListener('sync', (event) => {
    console.log('[Service Worker] Background sync:', event.tag);

    if (event.tag === 'sync-articles') {
        event.waitUntil(syncArticles());
    }
});

async function syncArticles() {
    // 这里可以添加后台同步逻辑
    console.log('[Service Worker] Syncing articles...');
}

// 消息处理
self.addEventListener('message', (event) => {
    if (event.data && event.data.type === 'SKIP_WAITING') {
        self.skipWaiting();
    }

    if (event.data && event.data.type === 'CLEAR_CACHE') {
        event.waitUntil(
            caches.keys().then((cacheNames) => {
                return Promise.all(
                    cacheNames.map((cacheName) => caches.delete(cacheName))
                );
            })
        );
    }
});

console.log('[Service Worker] Loaded');