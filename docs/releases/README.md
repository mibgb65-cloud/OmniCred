# Release Notes

每次发布必须在此目录保存一份与 Git 标签同名的 Markdown 文件：

```text
docs/releases/v0.1.0.md
docs/releases/v0.2.0.md
```

发布流程：

1. 更新 `wails.json` 中的产品版本。
2. 新建 `docs/releases/vMAJOR.MINOR.PATCH.md`。
3. 完成并推送发布提交。
4. 创建并推送同名标签，例如 `v0.2.0`。
5. GitHub Actions 读取对应 Markdown 作为 GitHub Release Notes。

如果标签对应的发布说明不存在，Release 工作流会失败，避免发布没有变更记录的版本。

发布说明应至少包含：版本摘要、新增功能、修复内容、下载说明、安全提示和已知限制。
