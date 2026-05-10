# Go 图片模板生成 PDF

使用 Go 将图片作为模板生成 PDF，要求输入必须参数和图片路径，输出 PDF 文件。

## 功能说明

- 必填参数：
  - `-image`：模板图片文件路径（支持 jpeg/png/gif 等）
  - `-output`：输出 PDF 文件路径
  - `-param`：业务参数，格式 `key=value`，可重复传入，至少一个
- 输出：生成包含模板图片和参数信息的 PDF 文件

## 运行方式

```bash
go run . \
  -image ./template.png \
  -output ./result.pdf \
  -param name=张三 \
  -param order_no=A10001 \
  -param date=2026-05-10
```

## 构建

```bash
go build -o pdfgen .
```
