# fssh
![](images/finger.png)
解决macOS上每次使用公私钥登录目标主机需要输入密码解锁私钥的问题，以及使用`~/.ssh/config`配置登录主机的别名时，时间久了登录机器时需要查看该配置文件的问题。
SSH登录目标机器时，会弹出指纹认证的框要求指纹认证，认证通过后会解密导入的私钥并登录目标主机。同事在fssh的交互式shell中可以快捷的查看,登录`~/.ssh/config`中配置的主机信息。

 - 使用场景
   - 明文 SSH 密钥在本地磁盘上存在风险
   - 加密密钥每次登录都需输入口令，使用起来比较麻烦
   - ~/.ssh/config 文件中配置多个主机时，时间长容易忘记别名,每次连接前都要看一下该配置文件
 - 解决方案
   - 在 macOS 上通过ssh命令连接主机时,自动弹Touch ID要求指纹解锁主密钥，然后使用主密钥解密SSH私钥并登录主机
   - 兼容 OpenSSH的ssh-agent
   - 直接运行fssh命令，会进入一个交互式Shell,然后可以在该shell中使用list,connect等方式查看~/.ssh/config中配置的机器信息。shell中直接输入对应机器的id，host，ip均可使用ssh直接登录目标及其

## 功能截图
SSH登录目标主机时使用指纹解锁私钥
![](images/finger.png)
在交互式shell中查看`~/.ssh/config`配置文件中的主机信息
![](./images/shell.png)
交互式shell中登录目标主机
![](./images/login.png)

## 功能概述
- Touch ID/用户在场验证读取主密钥（存储于 Keychain）
- AES-256-GCM + HKDF 加密私钥（每文件独立 `salt`/`nonce`）
- 私钥导入/导出（PKCS#8 PEM 备份）、列出与状态检查
- ssh-agent：
  - 支持每次签名触发指纹，或配置 TTL 在短时间内免重复验证
  - 兼容 RSA‑SHA2（rsa‑sha2‑256/512）
- 交互式 Shell：解析 `~/.ssh/config` 主机、Tab 补全、默认连接
- 配置生成器：自动生成本地 `~/.ssh/config` 的 `IdentityAgent` 条目


## 安装与配置
1. 构建：`go build ./cmd/fssh`；安装到 `/usr/local/bin/fssh`
2. 执行`fssh init`命令初始化主秘钥
3. 执行`fssh import -alias string -file path/to/private --ask-passphrase`命令导入私钥
4. 启动SSH认证的Agent: `fssh agent --unlock-ttl-seconds 600`
5. 修改`~/.ssh/config`配置文件，以便所有ssh连接都走fssh启动的Agent(也可以使用`export SSH_AUTH_SOCK=~/.fssh/agent.sock`的方式使某个变量走)。
配置OpenSSH的Agent代理,主要是修改`~/.ssh/config`文件，并文件的最开始处添加如下内容:
```
host *
  ServerAliveInterval 30
  AddKeysToAgent yes
  ControlPersist 60
  ControlMaster auto
  IdentityAgent  ~/.fssh/agent.sock
```
6. 如果有需要其他配置可以创建并修改配置文件：`~/.fssh/config.json`，示例：
```
{
    "socket":"~/.fssh/agent.sock",
    "require_touch_id_per_sign":true,
    "unlock_ttl_seconds":600,
    "log_level":"info",
    "log_format":"plain"
}
```
 - socket: ssh Agent Socket的位置
 - require_touch_id_per_sign: 是否在每次 SSH 签名时都要求 Touch ID 验证
    - true: 启用安全模式，每次签名都需 Touch ID（或达到 TTL 时间后重新验证）
    - false: 便捷模式，启动时一次性解密所有密钥并常驻内存
 - unlock_ttl_seconds: Touch ID 解锁后的缓存时间窗口
 - log_level: 控制日志输出级别
    - debug: 显示所有日志（包括缓存命中信息）
    - info: 显示一般信息（默认）
    - warn: 显示警告和错误
    - error: 仅显示错误
- log_format: 控制日志输出格式
    - plain: 人类可读的普通格式
    - json: 结构化 JSON 格式
## 启动开机自启
- 启动 agent：`fssh agent --unlock-ttl-seconds 600`
- 开机自启：`contrib/com.fssh.agent.plist` 到 `~/Library/LaunchAgents/` 并执行命令`launchctl load -w`；修改配置后使用 `launchctl kickstart -k gui/$(id -u)/com.fssh.agent` 重载

## 交互式 Shell
- 启动：`fssh` 或 `fssh shell`
- 命令：
  - `list` 显示 `id\thost(ip)`
  - `search <term>` 按 id/host/ip 过滤
  - `connect <id|host|ip>` 发起连接；非命令输入默认连接
  - Tab 补全覆盖命令与 id/host/ip

## 故障排查
- `incorrect signature type / no mutual signature supported`
  - 确认 agent 运行并设置 `SSH_AUTH_SOCK=~/.fssh/agent.sock`
  - 本地条目包含 `IdentityAgent ~/.fssh/agent.sock`
  - 服务端接受 RSA‑SHA2（使用 `sshd-align` 或手动编辑）
- 连接后输入不可见：使用 `ssh -tt` 并在远端会话期间暂停行编辑
- 日志：配置 `log_out/log_err`、`log_level`、`log_format`；修改后重启 agent

## 安全说明
- 安全模式：每次签名或 TTL 缓存，避免私钥长期解密驻留
- 便捷模式：启动时解密并常驻内存；仅在提示过多场景使用
- 避免明文泄露：不在仓库或日志存储明文密钥/口令


## 致谢
- 本项目由 TRAE AI 软件辅助生成