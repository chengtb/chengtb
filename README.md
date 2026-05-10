# Recipe Card PDF Generator (Golang)

使用 Go 生成菜谱卡片 PDF，输入文本字段和图片 URL，输出一张接近模板布局的菜谱 PDF。

## 功能

- 输入 JSON（菜谱名称、编号、研发时间、机型、描述、步骤、A/B/C/D 仓格信息）
- 下载并嵌入两张远程图片（原材料图、成品图）
- 输出单页横向 A4 PDF
- 对输入字段、URL、图片格式和图片大小做基础校验

## 快速开始

```bash
go run . -input sample-input.json -output recipe-card.pdf
```

生成后文件位于：

`recipe-card.pdf`

## 输入 JSON 格式

```json
{
  "title": "Garlic-Fragrant Romaine Lettuce",
  "recipe_no": "00015491",
  "development_date": "January 1, 2024",
  "machine": "Cheetah (H22)",
  "description": "...",
  "preparation_steps": ["..."],
  "compartments": {
    "A": ["..."],
    "B": ["..."],
    "C": ["..."],
    "D": ["..."]
  },
  "ingredients_image_url": "https://...",
  "dish_image_url": "https://..."
}
```

## 测试

```bash
go test ./...
```
