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

## 应用内更新发布约定

发布工作流会在安装器构建完成后生成并上传 `update-manifest.json`，其中包含协议版本 `1`、Git 标签、目标系统架构、安装器名称、文件大小与 SHA-256。清单只允许指向同一次 Release 的 `OmniCred-amd64-installer.exe` 或 `OmniCred-arm64-installer.exe`。

不要为旧安装器补发这份清单：协议 1 依赖新的 NSIS 等待进程退出与失败回滚逻辑。旧发布缺少清单时，应用仍能提示版本，但会引导用户手动安装。修改或重新签名安装器后，必须重新生成清单及 `SHA256SUMS.txt`。

当前已发布的旧版客户端没有下载和安装逻辑，需要先手动安装一次包含本功能的版本；之后的兼容版本才可使用应用内更新。发布前保留一次真实 Windows 安装版的升级验收，检查 UAC 取消、原用户权限重启和实际数据目录。

流程与测试说明见 [应用内更新](../updating.md)。
