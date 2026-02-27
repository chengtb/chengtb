/**
 * MindMap – a lightweight mind-map library that supports multiple independent trees.
 *
 * Supported operations
 *   addNode(parentId, data, style?)  → returns new node id
 *   deleteNode(nodeId)
 *   updateNode(nodeId, data, style?) → update text / style of an existing node
 *   getRootId()                      → returns the first root node id (or null)
 *   getRootIds()                     → returns all root node ids
 *
 * Events (register with .on(eventName, handler))
 *   'nodeAdded'    → { node }
 *   'nodeDeleted'  → { nodeIds, nodes }  // nodeIds: ids of deleted node + descendants; nodes: their public data snapshots
 *   'nodeClicked'  → { node }            // fired when the user clicks a node element
 *   'nodeSelected' → { node }            // fired when selection changes; node is null when deselected
 *
 * Layout modes (options.layout)
 *   'directory'  (default) – vertical list with per-level indentation, L-shaped connectors
 *   'tree'                 – top-down tree with horizontal sibling spread, bezier connectors
 *
 * Drag-to-pan
 *   When options.pannable is true (default), drag-to-pan is automatically enabled on the
 *   parent element of the container (the scroll viewport).  Mouse and touch are both supported.
 *
 * Zoom & center
 *   When options.zoomable is true (default), zoom controls (＋／－ and ⊙) are automatically
 *   rendered in the grandparent element of the container (the wrap area).  Keyboard shortcuts
 *   Ctrl/⌘ + Plus/Minus (zoom) and Ctrl/⌘ + 0 (center) are also registered.
 *   Public API: zoomIn(), zoomOut(), setZoom(z), getZoom(), center()
 *
 * Node selection
 *   When options.selectable is true (default), clicking a node selects it (highlighted border)
 *   and clicking the same node again deselects it.  A 'nodeSelected' event is fired on change.
 *   Public API: selectNode(id), deselectNode(), getSelectedId()
 */
class MindMap {
  /**
   * @param {string|HTMLElement} container  CSS selector or DOM element
   * @param {object}             [options]
   * @param {string}  [options.layout='directory']  'directory' | 'tree'
   * @param {number}  [options.nodeWidth=220]
   * @param {number}  [options.nodeHeight=32]
   * @param {number}  [options.indentWidth=28]  horizontal indent per level (directory mode)
   * @param {number}  [options.hSpacing=48]     horizontal gap between sub-trees (tree mode)
   * @param {number}  [options.vSpacing=6]      vertical gap between rows
   * @param {number}  [options.rootChildVSpacing=50]  vertical gap between root and its direct children (mixed layout)
   * @param {number}  [options.rootGroupVSpacing=40]  vertical gap between independent root trees
   * @param {string}  [options.lineColor='#90a4ae']
   * @param {number}  [options.lineWidth=1.5]
   * @param {object}  [options.defaultStyle]  CSS properties applied to every node
   * @param {boolean} [options.pannable=true]  enable drag-to-pan on the parent scroll container
   * @param {boolean} [options.zoomable=true]  render zoom + center controls in the wrap element
   * @param {number}  [options.zoomStep=0.15]  zoom increment per step
   * @param {number}  [options.zoomMin=0.25]   minimum zoom level
   * @param {number}  [options.zoomMax=3.0]    maximum zoom level
   * @param {boolean} [options.selectable=true]   enable click-to-select; selected node gets a highlight border
   * @param {string}  [options.selectColor='#ff9800']  CSS color for the selection outline
   */
  constructor(container, options = {}) {
    this._container =
      typeof container === 'string'
        ? document.querySelector(container)
        : container;

    this._opts = Object.assign(
      {
        layout: 'directory',
        nodeWidth: 220,
        nodeHeight: 32,
        indentWidth: 28,
        hSpacing: 48,
        vSpacing: 6,
        rootChildVSpacing: 50,
        rootGroupVSpacing: 40,
        lineColor: '#90a4ae',
        lineWidth: 1.5,
        pannable: true,
        zoomable: true,
        zoomStep: 0.15,
        zoomMin: 0.25,
        zoomMax: 3.0,
        selectable: true,
        selectColor: '#ff9800',
        defaultStyle: {
          backgroundColor: '#e3f2fd',
          color: '#0d47a1',
          borderRadius: '3px',
          fontSize: '13px',
          fontFamily: 'Consolas, "Courier New", monospace',
          border: '1px solid #90caf9',
          padding: '4px 15px',
          boxSizing: 'border-box',
          textAlign: 'left',
          cursor: 'pointer',
          userSelect: 'none',
          whiteSpace: 'normal',
          wordBreak: 'break-word',
        },
      },
      options,
    );

    /** @type {Map<string, MindMapNode>} */
    this._nodes = new Map();
    this._roots = [];
    this._handlers = {};
    this._zoom = 1.0;
    this._selectedId = null;

    this._initDOM();
  }

  // ─── DOM initialisation ────────────────────────────────────────────────────

  _initDOM() {
    const c = this._container;
    c.style.position = 'relative';
    c.style.overflow = 'auto';

    // SVG layer (underneath) for connectors
    this._svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    Object.assign(this._svg.style, {
      position: 'absolute',
      top: '0',
      left: '0',
      pointerEvents: 'none',
    });
    c.appendChild(this._svg);

    // HTML layer (on top) for node boxes
    this._layer = document.createElement('div');
    this._layer.style.position = 'relative';
    c.appendChild(this._layer);

    // Click delegation – emit 'nodeClicked' and handle selection
    this._layerClickHandler = (e) => {
      const el = e.target.closest('[data-node-id]');
      if (!el) return;
      const id = el.getAttribute('data-node-id');
      const node = this._nodes.get(id);
      if (node) {
        this._emit('nodeClicked', { node: this._publicNode(node) });
        this._toggleSelection(id);
      }
    };
    this._layer.addEventListener('click', this._layerClickHandler);

    this._initPan();
    this._initZoom();
  }

  /**
   * Attach drag-to-pan (mouse + touch) to the parent scroll container.
   * Skips setup when options.pannable is false or there is no parent element.
   * @private
   */
  _initPan() {
    const scrollEl = this._container.parentElement;
    if (!this._opts.pannable || !scrollEl) return;

    let dragging = false;
    let startX = 0, startY = 0, scrollLeft = 0, scrollTop = 0;

    const beginDrag = (pageX, pageY) => {
      dragging = true;
      startX = pageX;
      startY = pageY;
      scrollLeft = scrollEl.scrollLeft;
      scrollTop  = scrollEl.scrollTop;
      scrollEl.style.cursor = 'grabbing';
    };

    const moveDrag = (pageX, pageY) => {
      if (!dragging) return;
      scrollEl.scrollLeft = scrollLeft - (pageX - startX);
      scrollEl.scrollTop  = scrollTop  - (pageY - startY);
    };

    const endDrag = () => {
      dragging = false;
      scrollEl.style.cursor = 'grab';
    };

    this._panMouseDown = (e) => {
      if (e.target.closest('[data-node-id]')) return;
      e.preventDefault();
      beginDrag(e.pageX, e.pageY);
    };
    this._panMouseMove = (e) => {
      if (!dragging) return;
      e.preventDefault();
      moveDrag(e.pageX, e.pageY);
    };
    this._panMouseUp = endDrag;

    this._panTouchStart = (e) => {
      if (e.target.closest('[data-node-id]')) return;
      e.preventDefault();
      const t = e.touches[0];
      beginDrag(t.pageX, t.pageY);
    };
    this._panTouchMove = (e) => {
      if (!dragging) return;
      e.preventDefault();
      const t = e.touches[0];
      moveDrag(t.pageX, t.pageY);
    };
    this._panTouchEnd = endDrag;

    scrollEl.style.cursor = 'grab';
    scrollEl.addEventListener('mousedown', this._panMouseDown);
    document.addEventListener('mousemove', this._panMouseMove);
    document.addEventListener('mouseup', this._panMouseUp);
    scrollEl.addEventListener('touchstart', this._panTouchStart, { passive: false });
    scrollEl.addEventListener('touchmove', this._panTouchMove, { passive: false });
    scrollEl.addEventListener('touchend', this._panTouchEnd);
  }

  /**
   * Build and inject zoom + center controls into the wrap element (grandparent of the
   * container), then register keyboard shortcuts.  Skips setup when options.zoomable is
   * false or the required ancestor elements do not exist.
   * @private
   */
  _initZoom() {
    const scrollEl = this._container.parentElement;
    if (!this._opts.zoomable || !scrollEl) return;
    const wrapEl = scrollEl.parentElement;
    if (!wrapEl) return;

    // Ensure the wrap can host an absolutely-positioned overlay
    if (getComputedStyle(wrapEl).position === 'static') {
      wrapEl.style.position = 'relative';
    }

    // ── Build controls overlay ────────────────────────────────────────────────
    const controls = document.createElement('div');
    Object.assign(controls.style, {
      position: 'absolute',
      bottom: '20px',
      right: '20px',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      gap: '6px',
      zIndex: '10',
      pointerEvents: 'none',
    });

    // Zoom bar
    const bar = document.createElement('div');
    Object.assign(bar.style, {
      display: 'flex',
      alignItems: 'center',
      gap: '4px',
      background: 'rgba(255,255,255,0.93)',
      border: '1px solid #bbb',
      borderRadius: '24px',
      padding: '4px 8px',
      boxShadow: '0 2px 8px rgba(0,0,0,0.18)',
      pointerEvents: 'auto',
    });

    const mkZoomBtn = (label, title) => {
      const btn = document.createElement('button');
      Object.assign(btn.style, {
        width: '28px',
        height: '28px',
        border: 'none',
        borderRadius: '50%',
        background: '#1a237e',
        color: '#fff',
        fontSize: '18px',
        fontWeight: 'bold',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '0',
        lineHeight: '1',
        flexShrink: '0',
        transition: 'background .15s',
      });
      btn.textContent = label;
      btn.title = title;
      btn.setAttribute('aria-label', title);
      btn.addEventListener('mouseover', () => { btn.style.background = '#283593'; });
      btn.addEventListener('mouseout',  () => { btn.style.background = '#1a237e'; });
      return btn;
    };

    const btnOut = mkZoomBtn('－', '缩小');
    this._zoomLevelEl = document.createElement('span');
    Object.assign(this._zoomLevelEl.style, {
      minWidth: '44px',
      textAlign: 'center',
      fontSize: '13px',
      color: '#333',
      fontWeight: '600',
      userSelect: 'none',
      fontFamily: 'Arial, sans-serif',
    });
    this._zoomLevelEl.textContent = '100%';
    const btnIn = mkZoomBtn('＋', '放大');

    bar.appendChild(btnOut);
    bar.appendChild(this._zoomLevelEl);
    bar.appendChild(btnIn);

    // Center button
    const btnCenter = document.createElement('button');
    Object.assign(btnCenter.style, {
      background: 'rgba(255,255,255,0.93)',
      border: '1px solid #bbb',
      borderRadius: '50%',
      width: '36px',
      height: '36px',
      fontSize: '18px',
      boxShadow: '0 2px 8px rgba(0,0,0,0.18)',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: '#1a237e',
      padding: '0',
      transition: 'background .15s',
      pointerEvents: 'auto',
    });
    btnCenter.textContent = '⊙';
    btnCenter.title = '内容居中';
    btnCenter.setAttribute('aria-label', '内容居中');
    btnCenter.addEventListener('mouseover', () => { btnCenter.style.background = '#e8eaf6'; });
    btnCenter.addEventListener('mouseout',  () => { btnCenter.style.background = 'rgba(255,255,255,0.93)'; });

    controls.appendChild(bar);
    controls.appendChild(btnCenter);
    wrapEl.appendChild(controls);
    this._zoomControls = controls;

    // ── Zoom / center logic ───────────────────────────────────────────────────
    const applyZoom = (z) => {
      const { zoomMin, zoomMax } = this._opts;
      const contentCenterX = (scrollEl.scrollLeft + scrollEl.clientWidth  / 2) / this._zoom;
      const contentCenterY = (scrollEl.scrollTop  + scrollEl.clientHeight / 2) / this._zoom;

      this._zoom = Math.min(zoomMax, Math.max(zoomMin, z));
      this._container.style.zoom = this._zoom;
      this._zoomLevelEl.textContent = Math.round(this._zoom * 100) + '%';

      requestAnimationFrame(() => {
        scrollEl.scrollLeft = contentCenterX * this._zoom - scrollEl.clientWidth  / 2;
        scrollEl.scrollTop  = contentCenterY * this._zoom - scrollEl.clientHeight / 2;
      });
    };

    const centerContent = () => {
      scrollEl.scrollLeft = Math.max(0, (this._container.offsetWidth  - scrollEl.clientWidth)  / 2);
      scrollEl.scrollTop  = Math.max(0, (this._container.offsetHeight - scrollEl.clientHeight) / 2);
    };

    // Store for public API
    this._applyZoom    = applyZoom;
    this._centerContent = centerContent;

    // Button listeners
    const { zoomStep } = this._opts;
    btnIn    .addEventListener('click', () => applyZoom(this._zoom + zoomStep));
    btnOut   .addEventListener('click', () => applyZoom(this._zoom - zoomStep));
    btnCenter.addEventListener('click', centerContent);

    // Keyboard shortcuts: Ctrl/⌘ + Plus/Minus to zoom, Ctrl/⌘ + 0 to center
    this._zoomKeyHandler = (e) => {
      if (!e.ctrlKey && !e.metaKey) return;
      if (e.key === '+' || e.key === '=') { e.preventDefault(); applyZoom(this._zoom + zoomStep); }
      if (e.key === '-')                   { e.preventDefault(); applyZoom(this._zoom - zoomStep); }
      if (e.key === '0')                   { e.preventDefault(); centerContent(); }
    };
    document.addEventListener('keydown', this._zoomKeyHandler);

    // Auto-center on initial load
    requestAnimationFrame(centerContent);
  }

  // ─── Public API ────────────────────────────────────────────────────────────

  /**
   * Add a node to the tree.
   * @param {string|null} parentId  null → creates a new root node
   * @param {object}      data      { id?, text }
   * @param {object}      [style]   CSS overrides for this node
   * @returns {string}  The new node's id
   */
  addNode(parentId, data = {}, style = {}) {
    if (parentId !== null && !this._nodes.has(parentId)) {
      throw new Error(`Parent node "${parentId}" does not exist.`);
    }

    const id =
      data.id != null
        ? String(data.id)
        : `node-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

    if (this._nodes.has(id)) {
      throw new Error(`A node with id "${id}" already exists.`);
    }

    /** @type {MindMapNode} */
    const node = {
      id,
      text: data.text != null ? String(data.text) : 'New Node',
      tags: Array.isArray(data.tags) ? data.tags.map(String) : [],
      state: data.state != null ? String(data.state) : '',
      layoutType: data.layoutType != null ? String(data.layoutType) : null,
      style: Object.assign({}, this._opts.defaultStyle, style),
      parentId: parentId,
      children: [],
      el: null,
      x: 0,
      y: 0,
    };

    this._nodes.set(id, node);

    if (parentId === null) {
      this._roots.push(id);
    } else {
      this._nodes.get(parentId).children.push(id);
    }

    this._render();
    this._emit('nodeAdded', { node: this._publicNode(node) });
    return id;
  }

  /**
   * Delete a node and all its descendants.
   * @param {string} nodeId
   */
  deleteNode(nodeId) {
    if (!this._nodes.has(nodeId)) {
      throw new Error(`Node "${nodeId}" does not exist.`);
    }

    const node = this._nodes.get(nodeId);

    // Detach from parent's children list
    if (node.parentId !== null) {
      const parent = this._nodes.get(node.parentId);
      parent.children = parent.children.filter((id) => id !== nodeId);
    } else {
      // Deleting a root
      this._roots = this._roots.filter((id) => id !== nodeId);
    }

    // Snapshot public data of every node in the subtree before removal
    const deletedNodes = [];
    const snapshotQueue = [nodeId];
    while (snapshotQueue.length) {
      const qid = snapshotQueue.shift();
      const qn = this._nodes.get(qid);
      if (!qn) continue;
      deletedNodes.push(this._publicNode(qn));
      qn.children.forEach((cid) => snapshotQueue.push(cid));
    }

    // Collect and remove sub-tree
    const removed = [];
    this._removeSubtree(nodeId, removed);

    // Auto-deselect if the selected node was part of the deleted subtree
    if (this._selectedId && removed.includes(this._selectedId)) {
      this._selectedId = null;
      this._emit('nodeSelected', { node: null });
    }

    this._render();
    this._emit('nodeDeleted', { nodeIds: removed, nodes: deletedNodes });
  }

  /**
   * Register an event handler.
   * @param {'nodeAdded'|'nodeDeleted'|'nodeClicked'|'nodeSelected'} event
   * @param {function} handler
   * @returns {MindMap} this (chainable)
   */
  on(event, handler) {
    (this._handlers[event] = this._handlers[event] || []).push(handler);
    return this;
  }

  /**
   * Unregister an event handler.
   * @param {'nodeAdded'|'nodeDeleted'|'nodeClicked'|'nodeSelected'} event
   * @param {function} handler
   * @returns {MindMap} this (chainable)
   */
  off(event, handler) {
    if (this._handlers[event]) {
      this._handlers[event] = this._handlers[event].filter(
        (h) => h !== handler,
      );
    }
    return this;
  }

  /**
   * Retrieve a snapshot of a node's public data.
   * @param {string} nodeId
   * @returns {object|null}
   */
  getNode(nodeId) {
    const n = this._nodes.get(nodeId);
    return n ? this._publicNode(n) : null;
  }

  /**
   * Return the id of the first root node, or null if the map is empty.
   * @returns {string|null}
   */
  getRootId() {
    return this._roots[0] || null;
  }

  /**
   * Return the ids of all root nodes (independent trees).
   * @returns {string[]}
   */
  getRootIds() {
    return this._roots.slice();
  }

  /**
   * Update a node's text, tags, state and/or style.
   * @param {string}  nodeId
   * @param {object}  [data]   { text?, tags?, state? }
   * @param {object}  [style]  CSS overrides (merged with existing)
   */
  updateNode(nodeId, data = {}, style = {}) {
    const node = this._nodes.get(nodeId);
    if (!node) throw new Error(`Node "${nodeId}" does not exist.`);
    if (data.text != null) node.text = String(data.text);
    if (data.tags != null) node.tags = Array.isArray(data.tags) ? data.tags.map(String) : [];
    if (data.state != null) node.state = String(data.state);
    if (style) node.style = Object.assign({}, node.style, style);
    this._render();
  }

  /**
   * Zoom in by one step.
   * @returns {MindMap} this (chainable)
   */
  zoomIn() {
    if (this._applyZoom) this._applyZoom(this._zoom + this._opts.zoomStep);
    return this;
  }

  /**
   * Zoom out by one step.
   * @returns {MindMap} this (chainable)
   */
  zoomOut() {
    if (this._applyZoom) this._applyZoom(this._zoom - this._opts.zoomStep);
    return this;
  }

  /**
   * Set the zoom level to an absolute value (clamped to zoomMin .. zoomMax).
   * @param {number} z
   * @returns {MindMap} this (chainable)
   */
  setZoom(z) {
    if (this._applyZoom) this._applyZoom(z);
    return this;
  }

  /**
   * Return the current zoom level.
   * @returns {number}
   */
  getZoom() {
    return this._zoom;
  }

  /**
   * Scroll the viewport so that the content is centered.
   * @returns {MindMap} this (chainable)
   */
  center() {
    if (this._centerContent) this._centerContent();
    return this;
  }

  /**
   * Programmatically select a node.  Fires the 'nodeSelected' event.
   * @param {string} nodeId
   * @returns {MindMap} this (chainable)
   */
  selectNode(nodeId) {
    if (!this._nodes.has(nodeId)) {
      throw new Error(`Node "${nodeId}" does not exist.`);
    }
    this._clearSelectionStyle();
    this._selectedId = nodeId;
    const node = this._nodes.get(nodeId);
    this._applySelectionStyle(node);
    this._emit('nodeSelected', { node: this._publicNode(node) });
    return this;
  }

  /**
   * Deselect the currently selected node.  Fires the 'nodeSelected' event with node = null.
   * @returns {MindMap} this (chainable)
   */
  deselectNode() {
    if (this._selectedId) {
      this._clearSelectionStyle();
      this._selectedId = null;
      this._emit('nodeSelected', { node: null });
    }
    return this;
  }

  /**
   * Return the id of the currently selected node, or null if nothing is selected.
   * @returns {string|null}
   */
  getSelectedId() {
    return this._selectedId;
  }

  // ─── Internal helpers ──────────────────────────────────────────────────────

  _emit(event, payload) {
    (this._handlers[event] || []).forEach((fn) => fn(payload));
  }

  /** Remove the selection outline from the currently selected node element. @private */
  _clearSelectionStyle() {
    if (!this._selectedId) return;
    const node = this._nodes.get(this._selectedId);
    if (node && node.el) {
      node.el.style.outline = '';
      node.el.style.outlineOffset = '';
    }
  }

  /** Apply the selection outline to a node element. @private */
  _applySelectionStyle(node) {
    if (!node || !node.el) return;
    node.el.style.outline = `3px solid ${this._opts.selectColor}`;
    node.el.style.outlineOffset = '2px';
  }

  /**
   * Toggle the selection for the given node id (called on click).
   * Clicking the already-selected node deselects it.
   * @private
   */
  _toggleSelection(id) {
    if (!this._opts.selectable) return;
    if (id === this._selectedId) {
      this._clearSelectionStyle();
      this._selectedId = null;
      this._emit('nodeSelected', { node: null });
    } else {
      this._clearSelectionStyle();
      this._selectedId = id;
      const node = this._nodes.get(id);
      this._applySelectionStyle(node);
      this._emit('nodeSelected', { node: this._publicNode(node) });
    }
  }

  _publicNode(node) {
    return {
      id: node.id,
      text: node.text,
      tags: node.tags.slice(),
      state: node.state,
      layoutType: node.layoutType,
      style: Object.assign({}, node.style),
      parentId: node.parentId,
      children: node.children.slice(),
    };
  }

  _removeSubtree(nodeId, collected) {
    const node = this._nodes.get(nodeId);
    if (!node) return;
    [...node.children].forEach((cid) => this._removeSubtree(cid, collected));
    if (node.el && node.el.parentNode) node.el.parentNode.removeChild(node.el);
    collected.push(nodeId);
    this._nodes.delete(nodeId);
  }

  // ─── Layout ───────────────────────────────────────────────────────────────

  /**
   * Compute node positions for all root trees, stacked vertically.
   */
  _computeLayout() {
    if (!this._roots.length) return;
    const rootGroupVSpacing = this._opts.rootGroupVSpacing;
    let groupOffsetY = 0;

    this._roots.forEach((rootId) => {
      const rootNode = this._nodes.get(rootId);
      // Run layout for this subtree (y-values start at 0)
      if (rootNode && rootNode.layoutType === '组织结构') {
        this._computeLayoutMixed(rootId);
      } else if (this._opts.layout === 'directory') {
        this._computeLayoutDirectory(rootId);
      } else {
        this._computeLayoutTree(rootId);
      }

      // Shift all nodes in this subtree down by groupOffsetY and find new bottom
      let maxBottom = 0;
      this._visitSubtree(rootId, (n) => {
        n.y += groupOffsetY;
        maxBottom = Math.max(maxBottom, n.y + (n.h || this._opts.nodeHeight));
      });
      groupOffsetY = maxBottom + rootGroupVSpacing;
    });
  }

  /** Depth-first traversal of a subtree rooted at rootId. */
  _visitSubtree(rootId, fn) {
    const node = this._nodes.get(rootId);
    if (!node) return;
    fn(node);
    node.children.forEach((cid) => this._visitSubtree(cid, fn));
  }

  /**
   * Directory layout: depth-first traversal assigns each node a sequential
   * row (y) and an indented column (x = depth × indentWidth).
   */
  _computeLayoutDirectory(rootId) {
    const { vSpacing } = this._opts;
    const indentWidth = Math.max(1, this._opts.indentWidth);
    let curY = 0;

    const traverse = (id, depth) => {
      const node = this._nodes.get(id);
      node.x = depth * indentWidth;
      node.y = curY;
      curY += (node.h || this._opts.nodeHeight) + vSpacing;
      node.children.forEach((cid) => traverse(cid, depth + 1));
    };

    traverse(rootId, 0);
  }

  /**
   * Tree layout: compute sub-tree widths (bottom-up) then assign positions
   * (top-down) so children fan out horizontally below their parent.
   */
  _computeLayoutTree(rootId) {
    const { nodeHeight, hSpacing, vSpacing } = this._opts;
    const subtreeW = new Map();

    const measureWidth = (id) => {
      const node = this._nodes.get(id);
      const nw = node.w || this._opts.nodeWidth;
      if (!node.children.length) {
        subtreeW.set(id, nw);
        return nw;
      }
      const childrenTotal =
        node.children.reduce((s, cid) => s + measureWidth(cid), 0) +
        hSpacing * (node.children.length - 1);
      const w = Math.max(childrenTotal, nw);
      subtreeW.set(id, w);
      return w;
    };

    measureWidth(rootId);

    const position = (id, cx, y) => {
      const node = this._nodes.get(id);
      const nw = node.w || this._opts.nodeWidth;
      node.x = cx - nw / 2;
      node.y = y;
      if (!node.children.length) return;

      const totalW =
        node.children.reduce((s, cid) => s + subtreeW.get(cid), 0) +
        hSpacing * (node.children.length - 1);

      let left = cx - totalW / 2;
      const childY = y + (node.h || nodeHeight) + vSpacing;
      node.children.forEach((cid) => {
        const cw = subtreeW.get(cid);
        position(cid, left + cw / 2, childY);
        left += cw + hSpacing;
      });
    };

    position(rootId, subtreeW.get(rootId) / 2, 0);
  }

  /**
   * Mixed layout: the root node's children fan out horizontally (org-chart style,
   * with bezier connectors), while every non-root node lays its own children out
   * vertically in directory style (indented list, L-shaped connectors).
   */
  _computeLayoutMixed(rootId) {
    const { nodeHeight, hSpacing, vSpacing } = this._opts;
    const indentWidth = Math.max(1, this._opts.indentWidth);
    const rootNode = this._nodes.get(rootId);

    if (!rootNode) return;
    if (!rootNode.children.length) {
      rootNode.x = 0;
      rootNode.y = 0;
      return;
    }

    // Maximum horizontal extent of a directory sub-tree when its root is at the
    // given depth (0 = column origin).
    const subtreeColWidth = (id, depth) => {
      const node = this._nodes.get(id);
      if (!node) return depth * indentWidth + this._opts.nodeWidth;
      const nw = node.w || this._opts.nodeWidth;
      const w = depth * indentWidth + nw;
      return node.children.reduce(
        (max, cid) => Math.max(max, subtreeColWidth(cid, depth + 1)),
        w,
      );
    };

    // Column widths for each direct child of root.
    const childWidths = rootNode.children.map((cid) => subtreeColWidth(cid, 0));
    const totalW =
      childWidths.reduce((s, w) => s + w, 0) +
      hSpacing * (rootNode.children.length - 1);

    // Root is horizontally centered above all child columns.
    rootNode.x = totalW / 2 - (rootNode.w || this._opts.nodeWidth) / 2;
    rootNode.y = 0;

    const childStartY = (rootNode.h || nodeHeight) + this._opts.rootChildVSpacing;

    // Lay out each child's directory sub-tree inside its allocated column.
    let curX = 0;
    rootNode.children.forEach((cid, i) => {
      if (i > 0) curX += hSpacing;
      let curY = childStartY;
      const layoutDir = (id, colX, depth) => {
        const node = this._nodes.get(id);
        if (!node) return;
        node.x = colX + depth * indentWidth;
        node.y = curY;
        curY += (node.h || nodeHeight) + vSpacing;
        node.children.forEach((childId) => layoutDir(childId, colX, depth + 1));
      };
      layoutDir(cid, curX, 0);
      curX += childWidths[i];
    });
  }

  // ─── Rendering ────────────────────────────────────────────────────────────

  _render() {
    if (!this._roots.length) {
      this._svg.innerHTML = '';
      // Remove all node elements
      [...this._layer.querySelectorAll('[data-node-id]')].forEach((el) =>
        el.remove(),
      );
      return;
    }

    const PADDING = 24;
    const { nodeHeight, lineColor, lineWidth } = this._opts;
    const nodeMaxWidth = 200;

    // ── Pass 1: create / update node DOM elements (no positioning yet) ─────────
    // Remove stale DOM nodes
    [...this._layer.querySelectorAll('[data-node-id]')].forEach((el) => {
      if (!this._nodes.has(el.getAttribute('data-node-id'))) el.remove();
    });

    this._nodes.forEach((node) => {
      let el = this._layer.querySelector(`[data-node-id="${node.id}"]`);
      if (!el) {
        el = document.createElement('div');
        el.setAttribute('data-node-id', node.id);
        this._layer.appendChild(el);
        node.el = el;
      }

      // Apply node style, then enforce max-width and auto height (no fixed width/height)
      Object.assign(el.style, node.style);
      el.style.position = 'absolute';
      el.style.maxWidth = nodeMaxWidth + 'px';
      el.style.height = 'auto';
      el.style.minHeight = nodeHeight + 'px';
      el.style.display = 'flex';
      el.style.flexDirection = 'column';
      el.style.alignItems = 'flex-start';
      el.style.justifyContent = 'center';

      // Text line (flex row: optional state icon + text)
      el.textContent = '';
      const textEl = document.createElement('div');
      textEl.style.cssText = 'display:flex;align-items:center;gap:6px;';
      const stateIcons = {
        success: { char: '✓', color: '#fff', bg: '#43a047' },
        warn: { char: '!', color: '#fff', bg: '#fb8c00' },
      };
      const iconDef = stateIcons[node.state];
      if (iconDef) {
        const iconEl = document.createElement('span');
        iconEl.textContent = iconDef.char;
        iconEl.style.cssText =
          `display:inline-flex;align-items:center;justify-content:center;` +
          `width:18px;height:18px;border-radius:50%;` +
          `background:${iconDef.bg};color:${iconDef.color};` +
          `font-size:13px;font-weight:bold;flex-shrink:0;line-height:1;`;
        textEl.appendChild(iconEl);
      }
      const textSpan = document.createElement('span');
      textSpan.textContent = node.text;
      textEl.appendChild(textSpan);
      el.appendChild(textEl);

      // Tags row (rendered below the text if any tags exist)
      if (node.tags && node.tags.length) {
        const tagsEl = document.createElement('div');
        tagsEl.style.cssText = 'display:flex;flex-wrap:wrap;gap:3px;margin-top:3px;';
        node.tags.forEach((tag) => {
          const span = document.createElement('span');
          span.textContent = tag;
          span.style.cssText =
            'display:inline-block;padding:1px 6px;border-radius:10px;' +
            'font-size:11px;background:rgba(0,0,0,0.12);line-height:1.4;';
          tagsEl.appendChild(span);
        });
        el.appendChild(tagsEl);
      }
    });

    // ── Measure actual rendered widths and heights ──────────────────────────────
    this._nodes.forEach((node) => {
      const el = node.el || this._layer.querySelector(`[data-node-id="${node.id}"]`);
      node.w = el ? el.offsetWidth : nodeMaxWidth;
      node.h = el ? el.offsetHeight : nodeHeight;
    });

    // ── Compute layout (uses node.w) ────────────────────────────────────────────
    this._computeLayout();

    // ── Compute canvas bounds ───────────────────────────────────────────────────
    let minX = Infinity,
      maxX = -Infinity,
      maxY = -Infinity;
    this._nodes.forEach((n) => {
      minX = Math.min(minX, n.x);
      maxX = Math.max(maxX, n.x + n.w);
      maxY = Math.max(maxY, n.y + n.h);
    });

    const W = maxX - minX + PADDING * 2;
    const H = maxY + PADDING * 2;
    const offsetX = -minX + PADDING;
    const offsetY = PADDING;

    // Resize layers
    this._svg.setAttribute('width', W);
    this._svg.setAttribute('height', H);
    this._svg.setAttribute('viewBox', `0 0 ${W} ${H}`);
    this._layer.style.width = W + 'px';
    this._layer.style.height = H + 'px';

    // ── Draw connectors ──────────────────────────────────────────────────────────
    this._svg.innerHTML = '';
    // Precompute nodeId → rootNode mapping for per-connection connector style
    const nodeToRoot = new Map();
    this._roots.forEach((rootId) => {
      this._visitSubtree(rootId, (n) => nodeToRoot.set(n.id, this._nodes.get(rootId)));
    });
    const indentWidth = Math.max(1, this._opts.indentWidth);

    this._nodes.forEach((node) => {
      if (node.parentId === null) return;
      const parent = this._nodes.get(node.parentId);
      if (!parent) return;

      const rootNode = nodeToRoot.get(node.id);
      const isMixed = !!(rootNode && rootNode.layoutType === '组织结构');
      const isDirectory = this._opts.layout === 'directory' && !isMixed;

      const path = document.createElementNS(
        'http://www.w3.org/2000/svg',
        'path',
      );

      if (isMixed && parent.parentId === null) {
        // Root → direct child: smooth bezier (org-chart style)
        const x1 = parent.x + parent.w / 2 + offsetX;
        const y1 = parent.y + parent.h + offsetY;
        const x2 = node.x + node.w / 2 + offsetX;
        const y2 = node.y + offsetY;
        const midY = (y1 + y2) / 2;
        path.setAttribute(
          'd',
          `M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`,
        );
      } else if (isMixed || isDirectory) {
        // Orthogonal L-shaped connector for directory mode:
        // vertical spine goes from parent's bottom-center down to child's row,
        // then a horizontal stub runs right to the child's left edge.
        const xSpine = node.x - indentWidth / 2 + offsetX;
        const yTop = parent.y + parent.h / 2 + offsetY;
        const yBot = node.y + node.h / 2 + offsetY;
        const xEnd = node.x + offsetX;
        path.setAttribute('d', `M ${xSpine} ${yTop} V ${yBot} H ${xEnd}`);
      } else {
        // Smooth bezier connector for tree mode
        const x1 = parent.x + parent.w / 2 + offsetX;
        const y1 = parent.y + parent.h + offsetY;
        const x2 = node.x + node.w / 2 + offsetX;
        const y2 = node.y + offsetY;
        const midY = (y1 + y2) / 2;
        path.setAttribute(
          'd',
          `M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`,
        );
      }

      path.setAttribute('fill', 'none');
      path.setAttribute('stroke', lineColor);
      path.setAttribute('stroke-width', lineWidth);
      this._svg.appendChild(path);
    });

    // ── Position node boxes ──────────────────────────────────────────────────────
    this._nodes.forEach((node) => {
      const el = node.el || this._layer.querySelector(`[data-node-id="${node.id}"]`);
      if (!el) return;
      el.style.left = node.x + offsetX + 'px';
      el.style.top = node.y + offsetY + 'px';
    });

    // ── Re-apply selection outline (overridden by node style reset above) ────────
    if (this._opts.selectable && this._selectedId) {
      const selNode = this._nodes.get(this._selectedId);
      this._applySelectionStyle(selNode);
    }
  }
}

// CommonJS / ES-module compatibility shim
if (typeof module !== 'undefined' && module.exports) {
  module.exports = MindMap;
}
