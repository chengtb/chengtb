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
 */
class MindMap {
  /**
   * @param {string|HTMLElement} container  CSS selector or DOM element
   * @param {object}             [options]
   * @param {number}  [options.nodeWidth=140]
   * @param {number}  [options.nodeHeight=44]
   * @param {number}  [options.hSpacing=48]   horizontal gap between sibling sub-trees
   * @param {number}  [options.vSpacing=64]   vertical gap between parent and children
   * @param {string}  [options.lineColor='#b0bec5']
   * @param {number}  [options.lineWidth=2]
   * @param {object}  [options.defaultStyle]  CSS properties applied to every node
   */
  constructor(container, options = {}) {
    this._container =
      typeof container === 'string'
        ? document.querySelector(container)
        : container;

    this._opts = Object.assign(
      {
        nodeWidth: 140,
        nodeHeight: 44,
        hSpacing: 48,
        vSpacing: 64,
        lineColor: '#b0bec5',
        lineWidth: 2,
        defaultStyle: {
          backgroundColor: '#4285f4',
          color: '#ffffff',
          borderRadius: '6px',
          fontSize: '14px',
          fontFamily: 'Arial, sans-serif',
          border: '2px solid #2b5fc7',
          padding: '6px 12px',
          boxSizing: 'border-box',
          textAlign: 'center',
          cursor: 'pointer',
          userSelect: 'none',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
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
   * Compute sub-tree widths (bottom-up) and node positions (top-down).
   */
  _computeLayout() {
    if (!this._rootId) return;

    const { nodeWidth, nodeHeight, hSpacing, vSpacing } = this._opts;
    const subtreeW = new Map();

    const measureWidth = (id) => {
      const node = this._nodes.get(id);
      if (!node.children.length) {
        subtreeW.set(id, nodeWidth);
        return nodeWidth;
      }
      const childrenTotal =
        node.children.reduce((s, cid) => s + measureWidth(cid), 0) +
        hSpacing * (node.children.length - 1);
      const w = Math.max(childrenTotal, nodeWidth);
      subtreeW.set(id, w);
      return w;
    };

    measureWidth(this._rootId);

    const position = (id, cx, y) => {
      const node = this._nodes.get(id);
      node.x = cx - nodeWidth / 2;
      node.y = y;
      if (!node.children.length) return;

      const totalW =
        node.children.reduce((s, cid) => s + subtreeW.get(cid), 0) +
        hSpacing * (node.children.length - 1);

      let left = cx - totalW / 2;
      const childY = y + nodeHeight + vSpacing;
      node.children.forEach((cid) => {
        const cw = subtreeW.get(cid);
        position(cid, left + cw / 2, childY);
        left += cw + hSpacing;
      });
    };

    position(this._rootId, subtreeW.get(this._rootId) / 2, 0);
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

    this._computeLayout();

    const PADDING = 24;
    const { nodeWidth, nodeHeight, lineColor, lineWidth } = this._opts;

    // Compute canvas bounds
    let minX = Infinity,
      maxX = -Infinity,
      maxY = -Infinity;
    this._nodes.forEach((n) => {
      minX = Math.min(minX, n.x);
      maxX = Math.max(maxX, n.x + nodeWidth);
      maxY = Math.max(maxY, n.y + nodeHeight);
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

    // ── Draw connectors ──────────────────────────────────────────────────────
    this._svg.innerHTML = '';
    this._nodes.forEach((node) => {
      if (node.parentId === null) return;
      const parent = this._nodes.get(node.parentId);
      if (!parent) return;

      const x1 = parent.x + nodeWidth / 2 + offsetX;
      const y1 = parent.y + nodeHeight + offsetY;
      const x2 = node.x + nodeWidth / 2 + offsetX;
      const y2 = node.y + offsetY;
      const midY = (y1 + y2) / 2;

      const path = document.createElementNS(
        'http://www.w3.org/2000/svg',
        'path',
      );
      path.setAttribute(
        'd',
        `M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`,
      );
      path.setAttribute('fill', 'none');
      path.setAttribute('stroke', lineColor);
      path.setAttribute('stroke-width', lineWidth);
      this._svg.appendChild(path);
    });

    // ── Draw / update node boxes ─────────────────────────────────────────────

    // Remove stale DOM nodes
    [...this._layer.querySelectorAll('[data-node-id]')].forEach((el) => {
      if (!this._nodes.has(el.getAttribute('data-node-id'))) el.remove();
    });

    this._nodes.forEach((node) => {
      let el = this._layer.querySelector(`[data-node-id="${node.id}"]`);
      if (!el) {
        el = document.createElement('div');
        el.setAttribute('data-node-id', node.id);
        el.style.position = 'absolute';
        el.style.display = 'flex';
        el.style.alignItems = 'center';
        el.style.justifyContent = 'center';
        this._layer.appendChild(el);
        node.el = el;
      }

      // Apply custom style, then enforce layout geometry
      Object.assign(el.style, node.style);
      el.style.position = 'absolute';
      el.style.left = node.x + offsetX + 'px';
      el.style.top = node.y + offsetY + 'px';
      el.style.width = nodeWidth + 'px';
      el.style.height = nodeHeight + 'px';
      el.style.display = 'flex';
      el.style.alignItems = 'center';
      el.style.justifyContent = 'center';

      el.textContent = node.text;
    });
  }
}

// CommonJS / ES-module compatibility shim
if (typeof module !== 'undefined' && module.exports) {
  module.exports = MindMap;
}
