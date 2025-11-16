# fssh

 - 使用场景
   - 明文 SSH 密钥在本地磁盘上存在风险
   - 加密密钥每次登录都需输入口令，使用成本高
   - ~/.ssh/config 主机多时，别名易忘，连接前常要查文件
 - 解决方案
   - 在 macOS 上通过 Touch ID 解锁主密钥，解密本地加密的 SSH 私钥用于认证
   - 提供兼容 OpenSSH 的 ssh-agent
   - 交互式 Shell 从 ~/.ssh/config 读取主机，支持查询并直接连接

macOS 安装与开机自启请参阅 `docs/macos.md`（包含 `launchd` 配置示例）。

在 macOS 上通过 Touch ID 解锁主密钥，解密本地加密的 SSH 私钥并用于登录；同时提供兼容 OpenSSH 的 ssh-agent 与交互式 Shell。

## 功能概述
- Touch ID/用户在场验证读取主密钥（存储于 Keychain）
- AES-256-GCM + HKDF 加密私钥（每文件独立 `salt`/`nonce`）
- 私钥导入/导出（PKCS#8 PEM 备份）、列出与状态检查
- ssh-agent：
  - 支持每次签名触发指纹，或配置 TTL 在短时间内免重复验证
  - 兼容 RSA‑SHA2（rsa‑sha2‑256/512）
- 交互式 Shell：解析 `~/.ssh/config` 主机、Tab 补全、默认连接
- 配置生成器：自动生成本地 `~/.ssh/config` 的 `IdentityAgent` 条目

English documentation: see `README.md` for overview and macOS setup.

## 安装与配置
- 构建：`go build ./cmd/fssh`；安装到 `/usr/local/bin/fssh`
- 配置文件：`~/.fssh/config.json`，示例：
  - `{"socket":"~/.fssh/agent.sock","require_touch_id_per_sign":true,"unlock_ttl_seconds":600,"log_level":"info","log_format":"plain"}`

## 启动与自启
- 启动 agent：`fssh agent --unlock-ttl-seconds 600`
- 使用代理：`export SSH_AUTH_SOCK=~/.fssh/agent.sock`
- 自启：复制 `contrib/com.fssh.agent.plist` 到 `~/Library/LaunchAgents/` 并 `launchctl load -w`；修改配置后使用 `launchctl kickstart -k gui/$(id -u)/com.fssh.agent` 重载

## 交互式 Shell
- 启动：`fssh` 或 `fssh shell`
- 命令：
  - `list` 显示 `id\thost(ip)`
  - `search <term>` 按 id/host/ip 过滤
  - `connect <id|host|ip>` 发起连接；非命令输入默认连接
  - Tab 补全覆盖命令与 id/host/ip

## 配置生成器
- 打印：`fssh config-gen --host backuphost --user root`
- 写入：`fssh config-gen --host backuphost --user root --write`
- 覆盖：`fssh config-gen --host backuphost --overwrite --write`
- 全局算法（可选）：`fssh config-gen --global-algos --write` 在 `Host *` 写入 RSA‑SHA2 一次；单条目默认不写算法

## 服务端对齐（可选）
- 一键对齐 RSA‑SHA2：`fssh sshd-align --host backuphost --sudo`
- 修改远端 `/etc/ssh/sshd_config`：
  - `PubkeyAuthentication yes`
  - `PubkeyAcceptedAlgorithms +rsa-sha2-512,rsa-sha2-256`
  - `PubkeyAcceptedKeyTypes +rsa-sha2-512,rsa-sha2-256`（兼容）

## 故障排查
- “incorrect signature type / no mutual signature supported”
  - 确认 agent 运行并设置 `SSH_AUTH_SOCK=~/.fssh/agent.sock`
  - 本地条目包含 `IdentityAgent ~/.fssh/agent.sock`
  - 服务端接受 RSA‑SHA2（使用 `sshd-align` 或手动编辑）
- 连接后输入不可见：使用 `ssh -tt` 并在远端会话期间暂停行编辑
- 日志：配置 `log_out/log_err`、`log_level`、`log_format`；修改后重启 agent

## 安全说明
- 安全模式：每次签名或 TTL 缓存，避免私钥长期解密驻留
- 便捷模式：启动时解密并常驻内存；仅在提示过多场景使用
- 避免明文泄露：不在仓库或日志存储明文密钥/口令

## 与 Secretive 的差异与本产品优势
- 存储模型
  - Secretive：使用 Secure Enclave，密钥不可导出
  - 本产品：以加密 PKCS#8 文件存储于 `~/.fssh/keys`，主密钥存于 Keychain 并受 Touch ID 保护；可导出口令保护的 PEM 以便备份
- 访问控制
  - Secretive：Touch ID/Apple Watch 验证并有访问通知
  - 本产品：签名前或 TTL 内解锁（LocalAuthentication）；暂不提供访问通知
- 硬件支持
  - Secretive：支持智能卡/YubiKey
  - 本产品：暂不支持（列为路线图）
- 代理与算法
  - Secretive：SE 密钥不可导出签名
  - 本产品：OpenSSH 兼容代理，支持 RSA‑SHA2 扩展签名；提供 `sshd-align` 对齐服务端算法
- 工具与体验
  - Secretive：原生应用与 Homebrew 安装
  - 本产品：CLI + 交互式 Shell（主机解析、Tab 补全、默认连接）、`config-gen` 自动生成配置、`launchd` 自启、统一日志
- 平台
  - Secretive：macOS（含 SE）
  - 本产品：当前 macOS，后续计划扩展其他平台

## 致谢
- 本项目由 TRAE AI 软件辅助生成