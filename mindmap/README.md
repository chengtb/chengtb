# MindMap 技术文档

轻量级、零依赖的思维导图 / 目录组织图库，纯原生 JavaScript 实现，支持多棵独立树、多种布局模式及完全自定义节点样式。

---

## 目录

1. [快速开始](#1-快速开始)
2. [构造函数与选项](#2-构造函数与选项)
3. [公共 API](#3-公共-api)
   - [addNode](#addnode)
   - [deleteNode](#deletenode)
   - [updateNode](#updatenode)
   - [getNode](#getnode)
   - [getRootId](#getrootid)
   - [getRootIds](#getrootids)
   - [on / off](#on--off)
   - [zoomIn / zoomOut / setZoom / getZoom / center](#zoomin--zoomout--setzoom--getzoom--center)
4. [节点数据字段](#4-节点数据字段)
5. [布局模式](#5-布局模式)
   - [directory（默认）](#directory默认)
   - [tree](#tree)
   - [组织结构（混合）](#组织结构混合)
6. [状态图标](#6-状态图标)
7. [事件系统](#7-事件系统)
8. [样式定制](#8-样式定制)
9. [缩放与居中（Demo 页）](#9-缩放与居中demo-页)
10. [完整示例](#10-完整示例)

---

## 1. 快速开始

```html
<!-- 1. 引入库 -->
<script src="mindmap.js"></script>

<!-- 2. 准备容器 -->
<div id="map" style="width:100%;height:600px;"></div>

<script>
  // 3. 创建实例
  const mm = new MindMap('#map', { layout: 'directory' });

  // 4. 添加节点
  const rootId = mm.addNode(null, { text: '项目根目录' });
  const srcId  = mm.addNode(rootId, { text: 'src' });
  mm.addNode(srcId, { text: 'index.js', state: 'success' });
  mm.addNode(srcId, { text: 'utils.js', state: 'warn' });
</script>
```

---

## 2. 构造函数与选项

```js
new MindMap(container, options?)
```

| 参数 | 类型 | 说明 |
|---|---|---|
| `container` | `string \| HTMLElement` | CSS 选择器字符串或 DOM 元素 |
| `options` | `object` | 可选配置对象（见下表） |

### options 配置项

| 属性 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `layout` | `string` | `'directory'` | 全局布局模式：`'directory'` \| `'tree'`。若根节点的 `layoutType` 设为 `'组织结构'`，则该根树强制使用混合布局（优先级高于此全局设置） |
| `nodeWidth` | `number` | `220` | 节点默认宽度（px）|
| `nodeHeight` | `number` | `32` | 节点最小高度（px）；实际高度随内容自动伸展 |
| `indentWidth` | `number` | `28` | directory 模式下每级的水平缩进量（px）|
| `hSpacing` | `number` | `48` | tree 模式下子树间的水平间距（px）|
| `vSpacing` | `number` | `6` | 相邻节点行之间的垂直间距（px）|
| `rootChildVSpacing` | `number` | `50` | 组织结构模式下根节点与第一层子节点的垂直间距（px）|
| `rootGroupVSpacing` | `number` | `40` | 多棵独立根树之间的垂直间距（px）|
| `lineColor` | `string` | `'#90a4ae'` | 连接线颜色 |
| `lineWidth` | `number` | `1.5` | 连接线宽度（px）|
| `pannable`  | `boolean` | `true` | 在容器父元素（滚动视口）上启用拖动平移（鼠标 + 触摸）。设为 `false` 可禁用。|
| `zoomable`  | `boolean` | `true` | 自动在容器祖父元素（包裹区域）内渲染放大缩小控制按钮（＋／－ 及 ⊙ 居中），并注册键盘快捷键（Ctrl/⌘ + ±/0）。设为 `false` 可禁用。|
| `zoomStep`  | `number`  | `0.15` | 每次缩放的步长 |
| `zoomMin`   | `number`  | `0.25` | 最小缩放倍率 |
| `zoomMax`   | `number`  | `3.0`  | 最大缩放倍率 |
| `defaultStyle` | `object` | 见下方 | 所有节点共用的默认 CSS 属性 |

#### defaultStyle 默认值

```js
{
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
}
```

---

## 3. 公共 API

### addNode

```js
mm.addNode(parentId, data, style?) → string
```

添加一个节点，返回新节点的 `id`。

| 参数 | 类型 | 说明 |
|---|---|---|
| `parentId` | `string \| null` | 父节点 ID；传 `null` 创建根节点 |
| `data` | `object` | 节点数据，见[节点数据字段](#4-节点数据字段) |
| `style` | `object` | （可选）CSS 属性覆盖，合并到 `defaultStyle` |

**抛出错误**：
- 父节点 ID 不存在时
- 指定的节点 ID 已存在时

```js
const id = mm.addNode(null, {
  id: 'root',          // 可选，留空则自动生成
  text: '根节点',
  tags: ['v1.0'],
  state: 'success',
});
```

---

### deleteNode

```js
mm.deleteNode(nodeId)
```

删除指定节点及其**所有后代节点**，并触发 `nodeDeleted` 事件。

**抛出错误**：节点不存在时

```js
mm.deleteNode('node-abc123');
```

---

### updateNode

```js
mm.updateNode(nodeId, data?, style?)
```

更新已有节点的文字、标签、状态和/或样式。

| 参数 | 类型 | 说明 |
|---|---|---|
| `nodeId` | `string` | 目标节点 ID |
| `data` | `object` | `{ text?, tags?, state? }` 中的任意字段 |
| `style` | `object` | CSS 属性（与现有样式**合并**，不替换） |

**抛出错误**：节点不存在时

```js
mm.updateNode('node-abc123', { text: '已完成', state: 'success' }, { backgroundColor: '#c8e6c9' });
```

---

### getNode

```js
mm.getNode(nodeId) → object | null
```

返回节点的公开快照（纯数据对象，修改不影响内部状态），节点不存在时返回 `null`。

返回对象结构：

```js
{
  id: string,
  text: string,
  tags: string[],
  state: string,        // '' | 'success' | 'warn'
  layoutType: string | null,
  style: object,        // 当前生效的 CSS 样式
  parentId: string | null,
  children: string[],   // 直接子节点 ID 数组
}
```

---

### getRootId

```js
mm.getRootId() → string | null
```

返回第一棵根树的根节点 ID；图谱为空时返回 `null`。

---

### getRootIds

```js
mm.getRootIds() → string[]
```

返回所有根节点 ID 的数组（按添加顺序排列）。可用于遍历多棵独立树。

```js
mm.getRootIds().forEach((rootId) => {
  console.log(mm.getNode(rootId));
});
```

---

### on / off

```js
mm.on(event, handler) → MindMap   // 注册事件监听，可链式调用
mm.off(event, handler) → MindMap  // 注销事件监听，可链式调用
```

| 参数 | 类型 | 说明 |
|---|---|---|
| `event` | `string` | 事件名，见[事件系统](#7-事件系统) |
| `handler` | `function` | 回调函数 |

```js
const onAdded = ({ node }) => console.log('新增节点:', node.text);
mm.on('nodeAdded', onAdded);

// 稍后取消监听
mm.off('nodeAdded', onAdded);
```

---

### zoomIn / zoomOut / setZoom / getZoom / center

当 `options.zoomable` 为 `true`（默认）时，以下公共方法可通过代码控制视图缩放：

```js
mm.zoomIn()         // 放大一步（+zoomStep），可链式调用
mm.zoomOut()        // 缩小一步（-zoomStep），可链式调用
mm.setZoom(1.5)     // 设置绝对缩放倍率（自动夹紧到 zoomMin..zoomMax），可链式调用
mm.getZoom()        // 返回当前缩放倍率（number）
mm.center()         // 将内容居中到视口，可链式调用
```

**键盘快捷键**（同时注册）：

| 快捷键 | 效果 |
|--------|------|
| Ctrl / ⌘ + `+` 或 `=` | 放大 |
| Ctrl / ⌘ + `-` | 缩小 |
| Ctrl / ⌘ + `0` | 居中 |

---

## 4. 节点数据字段

`addNode` 的 `data` 参数支持以下字段：

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `id` | `string \| number` | 自动生成 | 节点唯一标识符；若不指定则自动生成 `node-{timestamp}-{random}` |
| `text` | `string` | `'New Node'` | 节点显示文字 |
| `tags` | `string[]` | `[]` | 标签列表，渲染在文字下方的小徽章 |
| `state` | `string` | `''` | 节点状态：`''`（无图标）\| `'success'` \| `'warn'` |
| `layoutType` | `string \| null` | `null` | 覆盖根节点的布局类型；目前支持 `'组织结构'` |

---

## 5. 布局模式

### directory（默认）

**垂直列表 + 分级缩进**，类似文件资源管理器的树形目录。

- 节点按深度优先顺序从上到下排列
- 每深一级，X 轴向右缩进 `indentWidth` px
- 父子之间使用 **L 形折线**（orthogonal）连接

```js
const mm = new MindMap('#map', { layout: 'directory', indentWidth: 28 });
```

```
根节点
├─ 子节点 A
│   ├─ 子子节点 1
│   └─ 子子节点 2
└─ 子节点 B
```

---

### tree

**自顶向下扇形展开**，类似传统思维导图。

- 根节点位于顶部，子节点水平铺开在下方
- 子树宽度自底向上测量，保证各子树不重叠
- 父子之间使用 **贝塞尔曲线**连接

```js
const mm = new MindMap('#map', { layout: 'tree', hSpacing: 48 });
```

---

### 组织结构（混合）

当根节点的 `layoutType` 设为 `'组织结构'` 时启用。

- **根节点 → 第一层子节点**：水平铺开（org-chart 风格）+ 贝塞尔曲线连接
- **第一层子节点及其后代**：各自独立采用 directory 缩进布局 + L 形连接线

适合展示组织架构图、项目模块划分等场景。

```js
mm.addNode(null, {
  id: 'root',
  text: '总目录',
  layoutType: '组织结构',
});
```

---

## 6. 状态图标

每个节点可通过 `state` 字段显示一个圆形状态徽章，渲染在节点文字的左侧。

| state 值 | 图标 | 背景色 | 含义 |
|---|---|---|---|
| `''`（空字符串）| 无 | — | 无状态 |
| `'success'` | ✓ | `#43a047`（绿色）| 成功 / 完成 |
| `'warn'` | ! | `#fb8c00`（橙色）| 警告 / 需注意 |

徽章样式：18 × 18 px 圆形，白色符号，无论节点背景色如何均清晰可见。

```js
// 添加时指定
mm.addNode(parentId, { text: '构建', state: 'success' });

// 或事后修改
mm.updateNode(nodeId, { state: 'warn' });

// 清除状态图标
mm.updateNode(nodeId, { state: '' });
```

---

## 7. 事件系统

### 事件一览

| 事件名 | 触发时机 | 回调参数 |
|---|---|---|
| `nodeAdded` | 节点成功添加后 | `{ node }` |
| `nodeDeleted` | 节点（及其后代）被删除后 | `{ nodeIds, nodes }` |
| `nodeClicked` | 用户点击某个节点元素时 | `{ node }` |

所有回调中的 `node` / `nodes` 均为公开快照对象（与 `getNode()` 返回值结构相同）。与库内部状态互相独立，修改不会影响原节点。

### nodeAdded

节点成功添加后触发。

```js
mm.on('nodeAdded', ({ node }) => {
  // node 为节点的公开快照对象（同 getNode 返回值）
  console.log(`添加节点 "${node.text}"，父节点: ${node.parentId}`);
});
```

### nodeDeleted

节点（及其所有后代）被删除后触发。回调参数同时包含被删除节点的 ID 列表和完整数据快照，方便记录日志或撤销操作。

```js
mm.on('nodeDeleted', ({ nodeIds, nodes }) => {
  // nodeIds: string[]     — 被删除的节点 ID 数组（含后代，BFS 顺序）
  // nodes:   object[]     — 被删除节点的完整数据快照数组（BFS 顺序）
  console.log(`已删除 ${nodeIds.length} 个节点`);
  nodes.forEach((n) => console.log(`  • "${n.text}" (${n.id})`));
});
```

### nodeClicked

用户点击任意节点元素时触发（通过事件委托实现，单个监听器覆盖全部节点，无需在每个节点上单独绑定）。

```js
mm.on('nodeClicked', ({ node }) => {
  // node 为被点击节点的公开快照
  console.log(`点击了节点 "${node.text}"，标签: [${node.tags.join(', ')}]`);
});
```

---

## 8. 样式定制

### 全局默认样式

通过构造函数的 `options.defaultStyle` 设置所有节点的基础样式：

```js
const mm = new MindMap('#map', {
  defaultStyle: {
    backgroundColor: '#fce4ec',
    color: '#880e4f',
    borderRadius: '8px',
    fontSize: '14px',
    border: '1px solid #f48fb1',
    padding: '6px 12px',
  },
});
```

### 节点级样式覆盖

`addNode` 的第三参数 `style` 会与 `defaultStyle` **合并**，节点级 `style` 中的属性**优先于** `defaultStyle`，可对单个节点定制外观：

```js
const rootStyle = {
  backgroundColor: '#1565c0',
  color: '#ffffff',
  fontWeight: 'bold',
  fontSize: '16px',
};
mm.addNode(null, { text: '根节点' }, rootStyle);
```

### 选中高亮

MindMap 不内置选中状态，示例页面通过 CSS 类实现：

```css
.mindmap-node-selected {
  outline: 3px solid #ff9800 !important;
  outline-offset: 2px;
}
```

```js
document.getElementById('map-inner').addEventListener('click', (e) => {
  const el = e.target.closest('[data-node-id]');
  if (el) el.classList.toggle('mindmap-node-selected');
});
```

### 标签（Tags）样式

标签以内联徽章形式渲染在节点文字下方，采用半透明背景：

```js
mm.addNode(parentId, { text: '版本', tags: ['v2.0', 'stable'] });
```

---

## 9. 缩放与居中（Demo 页）

`index.html` 在画布右下角提供了一组悬浮缩放控件，无需修改 `mindmap.js` 核心库即可使用。

### 控件说明

| 控件 | 元素 ID | 功能 |
|---|---|---|
| `－` 按钮 | `btn-zoom-out` | 缩小 15 %（最小 25 %）|
| 百分比显示 | `zoom-level` | 显示当前缩放比例 |
| `＋` 按钮 | `btn-zoom-in` | 放大 15 %（最大 300 %）|
| `⊙` 按钮 | `btn-center` | 将内容居中显示 |

### 实现原理

缩放通过对 `#map-inner` 元素设置 CSS `zoom` 属性实现。与 `transform: scale()` 不同，`zoom` 属性会影响元素的布局尺寸，滚动容器的可滚动范围会随缩放比例自动同步扩展，无需额外的"占位尺寸"技巧。

```js
// 缩放核心逻辑
function applyZoom(z) {
  const container = document.getElementById('map-container');
  // 记住缩放前视口中心对应的内容坐标（保持视觉中心不变）
  const cx = (container.scrollLeft + container.clientWidth  / 2) / currentZoom;
  const cy = (container.scrollTop  + container.clientHeight / 2) / currentZoom;

  currentZoom = Math.min(3.0, Math.max(0.25, z));
  document.getElementById('map-inner').style.zoom = currentZoom;
  document.getElementById('zoom-level').textContent = Math.round(currentZoom * 100) + '%';

  // 缩放后恢复视口中心
  requestAnimationFrame(() => {
    container.scrollLeft = cx * currentZoom - container.clientWidth  / 2;
    container.scrollTop  = cy * currentZoom - container.clientHeight / 2;
  });
}

// 居中逻辑（页面加载后自动调用一次）
function centerContent() {
  const container = document.getElementById('map-container');
  const inner     = document.getElementById('map-inner');
  container.scrollLeft = Math.max(0, (inner.offsetWidth  - container.clientWidth)  / 2);
  container.scrollTop  = Math.max(0, (inner.offsetHeight - container.clientHeight) / 2);
}
```

### 自动居中

页面加载完成、所有初始节点渲染后，通过 `requestAnimationFrame(centerContent)` 自动将内容居中，确保用户无需手动滚动即可看到完整的思维导图。

---

## 10. 完整示例

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <style>
    #map { width: 100%; height: 500px; border: 1px solid #ddd; }
    .selected { outline: 3px solid #ff9800 !important; }
  </style>
</head>
<body>
  <div id="map"></div>
  <script src="mindmap.js"></script>
  <script>
    const mm = new MindMap('#map', {
      layout: 'directory',
      lineColor: '#78909c',
      lineWidth: 1.5,
      defaultStyle: {
        backgroundColor: '#e8f5e9',
        color: '#1b5e20',
        border: '1px solid #a5d6a7',
        borderRadius: '4px',
        fontSize: '13px',
        padding: '4px 12px',
        cursor: 'pointer',
        userSelect: 'none',
        whiteSpace: 'normal',
        wordBreak: 'break-word',
        boxSizing: 'border-box',
      },
    });

    // 事件监听
    mm.on('nodeAdded',   ({ node }) => console.log('+ 新增', node.text));
    mm.on('nodeDeleted', ({ nodeIds }) => console.log('- 删除', nodeIds));

    // 构建树
    const root = mm.addNode(null, { text: '项目', tags: ['v1.0'] },
      { backgroundColor: '#1b5e20', color: '#fff', fontWeight: 'bold' });

    const src  = mm.addNode(root, { text: 'src', state: 'success' });
    mm.addNode(src, { text: 'index.js', state: 'success' });
    mm.addNode(src, { text: 'utils.js', state: 'warn'    });

    const docs = mm.addNode(root, { text: 'docs' });
    mm.addNode(docs, { text: 'README.md', state: 'success' });

    // 节点点击选中
    document.getElementById('map').addEventListener('click', (e) => {
      const el = e.target.closest('[data-node-id]');
      document.querySelectorAll('[data-node-id]').forEach((n) =>
        n.classList.remove('selected'));
      if (el) el.classList.add('selected');
    });

    // 动态更新
    setTimeout(() => {
      mm.updateNode(src, { state: 'warn' }, { backgroundColor: '#fff3e0' });
    }, 3000);
  </script>
</body>
</html>
```

---

## 内部架构速览

```
MindMap
├── _initDOM()          初始化 SVG 层（连线）+ HTML 层（节点盒）
├── addNode()           写入节点Map → _render()
├── deleteNode()        _removeSubtree() → _render()
├── updateNode()        修改节点数据 → _render()
│
├── _computeLayout()    按根树依次调用对应布局算法
│   ├── _computeLayoutDirectory()   目录缩进布局
│   ├── _computeLayoutTree()        自顶向下扇形布局
│   └── _computeLayoutMixed()       混合组织结构布局
│
└── _render()           三阶段渲染
    ├── Pass 1: 创建/更新 DOM 节点（含状态图标 + 标签徽章）
    ├── Pass 2: 读取实际宽高 → 驱动精确布局
    └── Pass 3: 绘制 SVG 连线 + 定位节点盒
```

渲染流程保证先完成 DOM 插入并读取真实尺寸，再计算位置，避免因字体、换行等导致的位置偏差。
