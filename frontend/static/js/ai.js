// AI Chat Widget
class AIChatWidget {
    constructor() {
        this.enabled = false;
        this.history = [];
        this.isOpen = false;
        this.isLoading = false;
        this.init();
    }

    async init() {
        // 检查 AI 是否启用
        try {
            const { data } = await api('/ai/status');
            this.enabled = data.enabled;
        } catch (error) {
            console.error('Failed to check AI status:', error);
            this.enabled = false;
        }

        this.render();
    }

    render() {
        // 创建 AI 聊天组件
        const widget = document.createElement('div');
        widget.className = 'ai-chat-widget';
        widget.id = 'ai-chat-widget';

        widget.innerHTML = `
            <div class="ai-chat-window">
                <div class="ai-chat-header">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 0 1 7 7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-1H2a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a7 7 0 0 1 7-7h1V5.73c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2z"/>
                        <circle cx="7.5" cy="14.5" r="1.5"/>
                        <circle cx="16.5" cy="14.5" r="1.5"/>
                    </svg>
                    <div>
                        <h4>AI 助手</h4>
                        <span>随时为您解答</span>
                    </div>
                </div>
                <div class="ai-chat-messages" id="ai-chat-messages">
                    ${this.enabled ? `
                        <div class="ai-message assistant">
                            你好！我是博客的 AI 助手，有什么可以帮助你的吗？
                        </div>
                    ` : `
                        <div class="ai-chat-disabled">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                                <circle cx="12" cy="12" r="10"/>
                                <line x1="12" y1="8" x2="12" y2="12"/>
                                <line x1="12" y1="16" x2="12.01" y2="16"/>
                            </svg>
                            <p>AI 功能暂未启用</p>
                        </div>
                    `}
                </div>
                <div class="ai-chat-input">
                    <input type="text" id="ai-chat-input" placeholder="输入消息..." ${!this.enabled ? 'disabled' : ''}>
                    <button id="ai-chat-send" ${!this.enabled ? 'disabled' : ''}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="22" y1="2" x2="11" y2="13"/>
                            <polygon points="22 2 15 22 11 13 2 9 22 2"/>
                        </svg>
                    </button>
                </div>
            </div>
            <button class="ai-chat-trigger" title="AI 助手">
                <svg class="chat-icon" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                </svg>
                <svg class="close-icon" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
            </button>
        `;

        document.body.appendChild(widget);

        // 绑定事件
        this.bindEvents();
    }

    bindEvents() {
        const widget = document.getElementById('ai-chat-widget');
        const trigger = widget.querySelector('.ai-chat-trigger');
        const input = document.getElementById('ai-chat-input');
        const sendBtn = document.getElementById('ai-chat-send');

        // 切换聊天窗口
        trigger.addEventListener('click', () => {
            this.isOpen = !this.isOpen;
            widget.classList.toggle('open', this.isOpen);
            if (this.isOpen && input) {
                input.focus();
            }
        });

        if (this.enabled && input && sendBtn) {
            // 发送消息
            sendBtn.addEventListener('click', () => this.sendMessage());

            // 回车发送
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    this.sendMessage();
                }
            });
        }
    }

    async sendMessage() {
        const input = document.getElementById('ai-chat-input');
        const message = input.value.trim();

        if (!message || this.isLoading) return;

        // 添加用户消息
        this.addMessage(message, 'user');
        input.value = '';

        // 显示加载状态
        this.isLoading = true;
        const loadingEl = this.addMessage('思考中...', 'loading');

        try {
            const response = await fetch(`${API_BASE}/ai/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    message: message,
                    history: this.history
                })
            });

            const data = await response.json();

            // 移除加载状态
            loadingEl.remove();

            if (!response.ok) {
                throw new Error(data.error || `请求失败 (${response.status})`);
            }

            // 添加 AI 回复
            this.addMessage(data.message, 'assistant');

            // 更新历史
            this.history.push(
                { role: 'user', content: message },
                { role: 'assistant', content: data.message }
            );

            // 限制历史长度
            if (this.history.length > 20) {
                this.history = this.history.slice(-20);
            }
        } catch (error) {
            loadingEl.remove();
            console.error('AI chat error:', error);
            this.addMessage(`抱歉，发生了错误：${error.message}`, 'assistant error');
        } finally {
            this.isLoading = false;
        }
    }

    addMessage(content, type) {
        const container = document.getElementById('ai-chat-messages');
        const messageEl = document.createElement('div');
        messageEl.className = `ai-message ${type}`;

        if (type === 'assistant' || type === 'assistant error') {
            // 渲染 Markdown 内容
            messageEl.innerHTML = this.renderMarkdown(content);
            // 高亮代码块
            messageEl.querySelectorAll('pre code').forEach(block => {
                if (typeof hljs !== 'undefined') {
                    hljs.highlightElement(block);
                }
            });
            // 添加代码复制按钮
            messageEl.querySelectorAll('pre').forEach(pre => {
                this.addCodeCopyButton(pre);
            });
        } else {
            messageEl.textContent = content;
        }

        container.appendChild(messageEl);
        container.scrollTop = container.scrollHeight;
        return messageEl;
    }

    renderMarkdown(text) {
        // 简单的 Markdown 渲染
        let html = text
            // 转义 HTML
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            // 代码块
            .replace(/```(\w*)\n([\s\S]*?)```/g, (match, lang, code) => {
                return `<pre class="ai-code-block"><code class="language-${lang || 'text'}">${code.trim()}</code></pre>`;
            })
            // 行内代码
            .replace(/`([^`]+)`/g, '<code class="ai-inline-code">$1</code>')
            // 粗体
            .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
            // 斜体
            .replace(/\*([^*]+)\*/g, '<em>$1</em>')
            // 标题
            .replace(/^### (.+)$/gm, '<h4 class="ai-h4">$1</h4>')
            .replace(/^## (.+)$/gm, '<h3 class="ai-h3">$1</h3>')
            .replace(/^# (.+)$/gm, '<h2 class="ai-h2">$1</h2>')
            // 列表
            .replace(/^\d+\. (.+)$/gm, '<li class="ai-list-item">$1</li>')
            .replace(/^- (.+)$/gm, '<li class="ai-list-item">$1</li>')
            // 换行
            .replace(/\n/g, '<br>');

        // 包裹连续的列表项
        html = html.replace(/(<li class="ai-list-item">.*?<\/li>(<br>)?)+/g, (match) => {
            return '<ul class="ai-list">' + match.replace(/<br>/g, '') + '</ul>';
        });

        return html;
    }

    addCodeCopyButton(pre) {
        const btn = document.createElement('button');
        btn.className = 'ai-code-copy-btn';
        btn.textContent = '复制';
        btn.addEventListener('click', async () => {
            const code = pre.querySelector('code')?.textContent || pre.textContent;
            try {
                await navigator.clipboard.writeText(code);
                btn.textContent = '已复制';
                btn.classList.add('copied');
                setTimeout(() => {
                    btn.textContent = '复制';
                    btn.classList.remove('copied');
                }, 2000);
            } catch (err) {
                btn.textContent = '失败';
                setTimeout(() => btn.textContent = '复制', 1500);
            }
        });
        pre.appendChild(btn);
    }
}

// AI Writing Assistant for Admin
class AIWritingAssistant {
    constructor() {
        this.enabled = false;
        this.init();
    }

    async init() {
        try {
            const { data } = await api('/ai/status');
            this.enabled = data.enabled;
        } catch (error) {
            this.enabled = false;
        }
    }

    async assist(type, content, context = '') {
        if (!this.enabled) {
            throw new Error('AI 功能未启用');
        }

        const { data } = await api('/admin/ai/writing', {
            method: 'POST',
            body: JSON.stringify({ type, content, context })
        });

        return data.result;
    }

    async generateSummary(content) {
        if (!this.enabled) {
            throw new Error('AI 功能未启用');
        }

        const { data } = await api('/ai/summary', {
            method: 'POST',
            body: JSON.stringify({ content })
        });

        return data.summary;
    }
}

// 初始化
let aiChatWidget = null;
let aiWritingAssistant = null;

document.addEventListener('DOMContentLoaded', () => {
    // 在前台页面初始化聊天组件
    if (document.getElementById('articles-container')) {
        aiChatWidget = new AIChatWidget();
    }
});