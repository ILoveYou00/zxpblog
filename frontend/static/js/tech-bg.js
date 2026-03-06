/**
 * Dynamic Tech Background
 * Aurora + Cyberpunk + Particles
 */

(function() {
    'use strict';

    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Create background container
    function createTechBackground() {
        const bg = document.createElement('div');
        bg.className = 'tech-bg';
        bg.innerHTML = `
            <div class="aurora"></div>
            <div class="grid"></div>
            <canvas class="particles"></canvas>
            <div class="orb orb-1"></div>
            <div class="orb orb-2"></div>
            <div class="orb orb-3"></div>
            <div class="orb orb-4"></div>
            <div class="scanline"></div>
            <div class="noise"></div>
            <div class="vignette"></div>
        `;

        document.body.insertBefore(bg, document.body.firstChild);
        document.body.classList.add('has-tech-bg');

        return bg;
    }

    // Particle System
    class ParticleSystem {
        constructor(canvas) {
            this.canvas = canvas;
            this.ctx = canvas.getContext('2d');
            this.particles = [];
            this.connections = [];
            this.mouse = { x: null, y: null, radius: 150 };
            this.isDark = document.documentElement.getAttribute('data-theme') === 'dark';

            this.resize();
            this.init();
            this.bindEvents();
            this.animate();
        }

        resize() {
            this.canvas.width = window.innerWidth;
            this.canvas.height = window.innerHeight;
        }

        init() {
            const particleCount = Math.min(
                Math.floor((this.canvas.width * this.canvas.height) / 15000),
                80
            );

            this.particles = [];
            for (let i = 0; i < particleCount; i++) {
                this.particles.push(new Particle(this));
            }
        }

        bindEvents() {
            window.addEventListener('resize', () => {
                this.resize();
                this.init();
            });

            window.addEventListener('mousemove', (e) => {
                this.mouse.x = e.clientX;
                this.mouse.y = e.clientY;
            });

            window.addEventListener('mouseout', () => {
                this.mouse.x = null;
                this.mouse.y = null;
            });

            // Theme change observer
            const observer = new MutationObserver(() => {
                this.isDark = document.documentElement.getAttribute('data-theme') === 'dark';
            });
            observer.observe(document.documentElement, {
                attributes: true,
                attributeFilter: ['data-theme']
            });
        }

        animate() {
            if (prefersReducedMotion) return;

            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

            // Update and draw particles
            this.particles.forEach(particle => {
                particle.update();
                particle.draw();
            });

            // Draw connections
            this.drawConnections();

            requestAnimationFrame(() => this.animate());
        }

        drawConnections() {
            const maxDistance = 120;

            for (let i = 0; i < this.particles.length; i++) {
                for (let j = i + 1; j < this.particles.length; j++) {
                    const dx = this.particles[i].x - this.particles[j].x;
                    const dy = this.particles[i].y - this.particles[j].y;
                    const distance = Math.sqrt(dx * dx + dy * dy);

                    if (distance < maxDistance) {
                        const opacity = (1 - distance / maxDistance) * 0.3;
                        this.ctx.beginPath();
                        this.ctx.strokeStyle = this.isDark
                            ? `rgba(99, 102, 241, ${opacity})`
                            : `rgba(99, 102, 241, ${opacity * 0.5})`;
                        this.ctx.lineWidth = 0.5;
                        this.ctx.moveTo(this.particles[i].x, this.particles[i].y);
                        this.ctx.lineTo(this.particles[j].x, this.particles[j].y);
                        this.ctx.stroke();
                    }
                }
            }
        }
    }

    class Particle {
        constructor(system) {
            this.system = system;
            this.canvas = system.canvas;
            this.ctx = system.ctx;

            this.x = Math.random() * this.canvas.width;
            this.y = Math.random() * this.canvas.height;
            this.size = Math.random() * 2 + 0.5;
            this.baseSize = this.size;

            this.speedX = (Math.random() - 0.5) * 0.5;
            this.speedY = (Math.random() - 0.5) * 0.5;

            // Color variations
            const colors = [
                { r: 99, g: 102, b: 241 },   // Primary purple
                { r: 139, g: 92, b: 246 },   // Secondary purple
                { r: 6, g: 182, b: 212 },    // Cyan
                { r: 236, g: 72, b: 153 },   // Pink
            ];
            this.color = colors[Math.floor(Math.random() * colors.length)];
        }

        update() {
            // Mouse interaction
            if (this.system.mouse.x !== null) {
                const dx = this.x - this.system.mouse.x;
                const dy = this.y - this.system.mouse.y;
                const distance = Math.sqrt(dx * dx + dy * dy);

                if (distance < this.system.mouse.radius) {
                    const force = (this.system.mouse.radius - distance) / this.system.mouse.radius;
                    const angle = Math.atan2(dy, dx);
                    this.x += Math.cos(angle) * force * 2;
                    this.y += Math.sin(angle) * force * 2;
                    this.size = this.baseSize + force * 2;
                } else {
                    this.size = this.baseSize;
                }
            }

            // Movement
            this.x += this.speedX;
            this.y += this.speedY;

            // Boundary wrap
            if (this.x < 0) this.x = this.canvas.width;
            if (this.x > this.canvas.width) this.x = 0;
            if (this.y < 0) this.y = this.canvas.height;
            if (this.y > this.canvas.height) this.y = 0;
        }

        draw() {
            const opacity = this.system.isDark ? 0.8 : 0.5;

            // Glow effect
            this.ctx.beginPath();
            this.ctx.arc(this.x, this.y, this.size * 2, 0, Math.PI * 2);
            this.ctx.fillStyle = `rgba(${this.color.r}, ${this.color.g}, ${this.color.b}, ${opacity * 0.2})`;
            this.ctx.fill();

            // Core
            this.ctx.beginPath();
            this.ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2);
            this.ctx.fillStyle = `rgba(${this.color.r}, ${this.color.g}, ${this.color.b}, ${opacity})`;
            this.ctx.fill();
        }
    }

    // Matrix Rain Effect (optional, for dark mode)
    class MatrixRain {
        constructor(canvas) {
            this.canvas = canvas;
            this.ctx = canvas.getContext('2d');
            this.columns = [];
            this.fontSize = 14;
            this.chars = 'アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン0123456789';

            this.resize();
            this.init();
            this.animate();
        }

        resize() {
            this.canvas.width = window.innerWidth;
            this.canvas.height = window.innerHeight;
        }

        init() {
            const columnCount = Math.floor(this.canvas.width / this.fontSize);
            this.columns = [];
            for (let i = 0; i < columnCount; i++) {
                this.columns.push({
                    x: i * this.fontSize,
                    y: Math.random() * this.canvas.height,
                    speed: Math.random() * 2 + 1,
                    chars: []
                });
            }
        }

        animate() {
            if (prefersReducedMotion) return;

            this.ctx.fillStyle = 'rgba(10, 10, 15, 0.05)';
            this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

            this.ctx.fillStyle = 'rgba(99, 102, 241, 0.3)';
            this.ctx.font = `${this.fontSize}px monospace`;

            this.columns.forEach(column => {
                const char = this.chars[Math.floor(Math.random() * this.chars.length)];
                this.ctx.fillText(char, column.x, column.y);

                column.y += column.speed * this.fontSize;

                if (column.y > this.canvas.height && Math.random() > 0.975) {
                    column.y = 0;
                }
            });

            requestAnimationFrame(() => this.animate());
        }
    }

    // Initialize
    function init() {
        const bg = createTechBackground();
        const particlesCanvas = bg.querySelector('.particles');

        if (!prefersReducedMotion && particlesCanvas) {
            new ParticleSystem(particlesCanvas);
        }
    }

    // Start when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();