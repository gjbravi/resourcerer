// Renders the ```mermaid blocks (<pre class="rz-mermaid">) with zoom, pan and full screen. The theme renders Mermaid
// into a closed shadow root, which nothing can zoom, so mkdocs.yml gives these blocks their own class and this script
// renders them in the page instead. Colors come from CSS (extra.css), so a theme switch needs no new render.
//
// Buttons zoom in, out and reset, and open the diagram full screen. Ctrl/⌘ + wheel zooms (a plain wheel keeps
// scrolling the page), dragging pans once zoomed in, and in full screen the wheel alone zooms.

(() => {
  const MERMAID = 'https://unpkg.com/mermaid@12/dist/mermaid.min.js';
  const MIN = 0.5;
  const MAX = 6;
  const STEP = 1.25;
  const icons = {
    in: '<path d="M11 8v6M8 11h6"/><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
    out: '<path d="M8 11h6"/><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
    reset: '<path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/>',
    full: '<path d="M8 3H5a2 2 0 0 0-2 2v3M21 8V5a2 2 0 0 0-2-2h-3M3 16v3a2 2 0 0 0 2 2h3M16 21h3a2 2 0 0 0 2-2v-3"/>',
  };

  let loading;
  let count = 0;

  function load() {
    if (window.mermaid && !(window.mermaid instanceof Element)) return Promise.resolve(window.mermaid);
    loading ??= new Promise((resolve, reject) => {
      const s = document.createElement('script');
      s.src = MERMAID;
      s.onload = () => {
        window.mermaid.initialize({
          startOnLoad: false,
          theme: 'base',
          themeVariables: { fontFamily: 'Inter, ui-sans-serif, system-ui, sans-serif', fontSize: '15px' },
          flowchart: { htmlLabels: true },
        });
        resolve(window.mermaid);
      };
      s.onerror = reject;
      document.head.append(s);
    });
    return loading;
  }

  function button(name, label) {
    const b = document.createElement('button');
    b.type = 'button';
    b.title = label;
    b.setAttribute('aria-label', label);
    b.innerHTML = `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[name]}</svg>`;
    return b;
  }

  function zoomable(box, viewport, svg) {
    const bar = document.createElement('div');
    bar.className = 'rz-diagram__bar';
    const zoomOut = button('out', 'Zoom out');
    const zoomIn = button('in', 'Zoom in');
    const reset = button('reset', 'Reset zoom');
    const full = button('full', 'Full screen');
    bar.append(zoomOut, zoomIn, reset, full);
    box.append(bar);

    let scale = 1;
    let x = 0;
    let y = 0;
    const apply = () => {
      svg.style.transform = `translate(${x}px, ${y}px) scale(${scale})`;
      box.classList.toggle('is-zoomed', scale !== 1 || x !== 0 || y !== 0);
    };
    const zoomAt = (factor, cx, cy) => {
      const next = Math.min(MAX, Math.max(MIN, scale * factor));
      const r = viewport.getBoundingClientRect();
      // Keep the point under the cursor (or the centre) where it is.
      const px = (cx ?? r.left + r.width / 2) - r.left - r.width / 2;
      const py = (cy ?? r.top + r.height / 2) - r.top - r.height / 2;
      x = px - ((px - x) * next) / scale;
      y = py - ((py - y) * next) / scale;
      scale = next;
      apply();
    };
    const isFull = () => document.fullscreenElement === box;

    zoomIn.addEventListener('click', () => zoomAt(STEP));
    zoomOut.addEventListener('click', () => zoomAt(1 / STEP));
    reset.addEventListener('click', () => {
      scale = 1;
      x = y = 0;
      apply();
    });
    full.addEventListener('click', () => (isFull() ? document.exitFullscreen() : box.requestFullscreen?.()));
    box.addEventListener('fullscreenchange', () => {
      full.title = isFull() ? 'Exit full screen' : 'Full screen';
      full.setAttribute('aria-label', full.title);
    });

    viewport.addEventListener(
      'wheel',
      (e) => {
        if (!(e.ctrlKey || e.metaKey || isFull())) return;
        e.preventDefault();
        zoomAt(e.deltaY < 0 ? 1.1 : 1 / 1.1, e.clientX, e.clientY);
      },
      { passive: false },
    );

    let drag;
    viewport.addEventListener('pointerdown', (e) => {
      if (scale === 1 && !isFull()) return;
      drag = { id: e.pointerId, sx: e.clientX, sy: e.clientY, x, y };
      viewport.setPointerCapture(e.pointerId);
      box.classList.add('is-dragging');
    });
    viewport.addEventListener('pointermove', (e) => {
      if (!drag || drag.id !== e.pointerId) return;
      x = drag.x + e.clientX - drag.sx;
      y = drag.y + e.clientY - drag.sy;
      apply();
    });
    const stop = () => {
      drag = undefined;
      box.classList.remove('is-dragging');
    };
    viewport.addEventListener('pointerup', stop);
    viewport.addEventListener('pointercancel', stop);
  }

  async function render(pre) {
    const source = pre.textContent;
    const box = document.createElement('div');
    box.className = 'rz-diagram';
    const viewport = document.createElement('div');
    viewport.className = 'rz-diagram__viewport';
    box.append(viewport);
    pre.replaceWith(box);
    try {
      const mermaid = await load();
      const { svg, bindFunctions } = await mermaid.render(`rz-mermaid-${count++}`, source);
      viewport.innerHTML = svg;
      bindFunctions?.(viewport);
      zoomable(box, viewport, viewport.querySelector('svg'));
    } catch (err) {
      // Show the source rather than nothing; `make lint-docs` with MERMAID=1 catches diagrams that do not render.
      viewport.append(pre);
      console.error('mermaid', err);
    }
  }

  const scan = () => document.querySelectorAll('pre.rz-mermaid').forEach(render);
  // With instant navigation the theme swaps pages without reloading; document$ emits on each one.
  if (window.document$) window.document$.subscribe(scan);
  else document.addEventListener('DOMContentLoaded', scan);
})();
