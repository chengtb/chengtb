# Recipe Card PDF Generator (Golang)

一个基于 Go 的菜谱卡片 PDF 生成程序，输入文本字段和图片链接，输出横向 A4 的菜谱 PDF 文件。

## 功能

- 横向 A4（297mm x 210mm）页面
- 菜谱基础信息：名称、编号、研发时间、机型
- 菜谱描述与准备步骤
- A/B/C/D 仓格食材信息
- 原材料图片与成品图片
- 支持 HTTP/HTTPS 图片链接（也支持本地图片路径）

## 使用方式

```bash
go run . -input /absolute/path/to/recipe.json -output /absolute/path/to/recipe-card.pdf
```

如需渲染中文字符，请额外提供 UTF-8 TTF 字体文件：

```bash
go run . -input /absolute/path/to/recipe.json -output /absolute/path/to/recipe-card.pdf -font /absolute/path/to/font.ttf
```

## 输入 JSON 示例

`sample.recipe.json`:

```json
{
  "recipeName": "宫保鸡丁",
  "recipeNumber": "RCP-2026-001",
  "developmentTime": "2026-05-10",
  "machineType": "X1000",
  "description": "经典川味下饭菜，口感香辣微甜。",
  "preparationSteps": [
    "鸡胸肉切丁，加入料酒、生抽、淀粉腌制 15 分钟。",
    "准备花生米、干辣椒、葱姜蒜。",
    "按顺序下锅翻炒并收汁。"
  ],
  "bins": {
    "A": { "ingredient": "鸡丁", "amount": "300g", "remark": "去筋膜" },
    "B": { "ingredient": "花生米", "amount": "80g", "remark": "提前炸香" },
    "C": { "ingredient": "干辣椒", "amount": "20g", "remark": "剪段" },
    "D": { "ingredient": "葱姜蒜", "amount": "50g", "remark": "切末" }
  },
  "rawMaterialImage": "https://picsum.photos/seed/raw-food/800/600",
  "finishedImage": "https://picsum.photos/seed/finished-food/800/600"
}
```
