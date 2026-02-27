/**
 * MindMap – a lightweight tree-shaped mind-map library.
 *
 * Supported operations
 *   addNode(parentId, data, style?)  → returns new node id
 *   deleteNode(nodeId)
 *   updateNode(nodeId, data, style?) → update text / style of an existing node
 *   getRootId()                      → returns the current root node id (or null)
 *
 * Events (register with .on(eventName, handler))
 *   'nodeAdded'   → { node }
 *   'nodeDeleted' → { nodeIds }   // the deleted node and all its descendants
 *
 * Layout modes (options.layout)
 *   'directory'  (default) – vertical list with per-level indentation, L-shaped connectors
 *   'tree'                 – top-down tree with horizontal sibling spread, bezier connectors
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
   * @param {string}  [options.lineColor='#90a4ae']
   * @param {number}  [options.lineWidth=1.5]
   * @param {object}  [options.defaultStyle]  CSS properties applied to every node
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
        lineColor: '#90a4ae',
        lineWidth: 1.5,
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
    this._rootId = null;
    this._handlers = {};

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
  }

  // ─── Public API ────────────────────────────────────────────────────────────

  /**
   * Add a node to the tree.
   * @param {string|null} parentId  null → becomes the root
   * @param {object}      data      { id?, text }
   * @param {object}      [style]   CSS overrides for this node
   * @returns {string}  The new node's id
   */
  addNode(parentId, data = {}, style = {}) {
    if (parentId === null && this._rootId !== null) {
      throw new Error('A root node already exists. Provide a parentId.');
    }
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
      this._rootId = id;
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
      // Deleting root
      this._rootId = null;
    }

    // Collect and remove sub-tree
    const removed = [];
    this._removeSubtree(nodeId, removed);

    this._render();
    this._emit('nodeDeleted', { nodeIds: removed });
  }

  /**
   * Register an event handler.
   * @param {'nodeAdded'|'nodeDeleted'} event
   * @param {function} handler
   * @returns {MindMap} this (chainable)
   */
  on(event, handler) {
    (this._handlers[event] = this._handlers[event] || []).push(handler);
    return this;
  }

  /**
   * Unregister an event handler.
   * @param {'nodeAdded'|'nodeDeleted'} event
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
   * Return the id of the current root node, or null if the tree is empty.
   * @returns {string|null}
   */
  getRootId() {
    return this._rootId;
  }

  /**
   * Update a node's text and/or style.
   * @param {string}  nodeId
   * @param {object}  [data]   { text? }
   * @param {object}  [style]  CSS overrides (merged with existing)
   */
  updateNode(nodeId, data = {}, style = {}) {
    const node = this._nodes.get(nodeId);
    if (!node) throw new Error(`Node "${nodeId}" does not exist.`);
    if (data.text != null) node.text = String(data.text);
    if (style) node.style = Object.assign({}, node.style, style);
    this._render();
  }

  // ─── Internal helpers ──────────────────────────────────────────────────────

  _emit(event, payload) {
    (this._handlers[event] || []).forEach((fn) => fn(payload));
  }

  _publicNode(node) {
    return {
      id: node.id,
      text: node.text,
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
   * Compute node positions. Dispatches to the appropriate layout algorithm.
   */
  _computeLayout() {
    if (!this._rootId) return;
    const rootNode = this._nodes.get(this._rootId);
    if (rootNode && rootNode.layoutType === '组织结构') {
      this._computeLayoutMixed();
    } else if (this._opts.layout === 'directory') {
      this._computeLayoutDirectory();
    } else {
      this._computeLayoutTree();
    }
  }

  /**
   * Directory layout: depth-first traversal assigns each node a sequential
   * row (y) and an indented column (x = depth × indentWidth).
   */
  _computeLayoutDirectory() {
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

    traverse(this._rootId, 0);
  }

  /**
   * Tree layout: compute sub-tree widths (bottom-up) then assign positions
   * (top-down) so children fan out horizontally below their parent.
   */
  _computeLayoutTree() {
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

    measureWidth(this._rootId);

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

    position(this._rootId, subtreeW.get(this._rootId) / 2, 0);
  }

  /**
   * Mixed layout: the root node's children fan out horizontally (org-chart style,
   * with bezier connectors), while every non-root node lays its own children out
   * vertically in directory style (indented list, L-shaped connectors).
   */
  _computeLayoutMixed() {
    const { nodeHeight, hSpacing, vSpacing } = this._opts;
    const indentWidth = Math.max(1, this._opts.indentWidth);
    const rootNode = this._nodes.get(this._rootId);

    if (!rootNode || !rootNode.children.length) {
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
    if (!this._rootId) {
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
      el.style.alignItems = 'center';
      el.style.justifyContent = 'center';

      el.textContent = node.text;
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
    const rootNode = this._rootId ? this._nodes.get(this._rootId) : null;
    const isMixed = !!(rootNode && rootNode.layoutType === '组织结构');
    const isDirectory = this._opts.layout === 'directory' && !isMixed;
    const indentWidth = Math.max(1, this._opts.indentWidth);

    this._nodes.forEach((node) => {
      if (node.parentId === null) return;
      const parent = this._nodes.get(node.parentId);
      if (!parent) return;

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
  }
}

// CommonJS / ES-module compatibility shim
if (typeof module !== 'undefined' && module.exports) {
  module.exports = MindMap;
}
