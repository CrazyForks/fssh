# SSH Agent 调试指南

## 基础检查

### 1. 确认 agent 正常运行
```bash
./fssh status                    # 检查 master key 状态
ps aux | grep "fssh agent"       # 检查进程
ls -la ~/.fssh/agent.sock        # 确认 socket 存在
ssh-add -l                       # 列出 agent 中的 key
```

### 2. 使用 SSH verbose 模式

```bash
# -v: 基础调试信息
ssh -v user@host

# -vv: 更详细的信息
ssh -vv user@host

# -vvv: 最详细的调试信息
ssh -vvv user@host
```

### 关键日志标识

在 verbose 输出中查找：

```
debug1: Connecting to <host> ...              # 连接阶段
debug1: Offering public key: ...              # agent 提供 key
debug1: Server accepts key: ...               # 服务器接受 key
debug1: Authentication succeeded (publickey)  # 认证成功
```

如果看到 `Offering public key` 但没有 agent 提示，说明：
- SSH 找到了 agent
- 但 key 不匹配或服务器配置问题

如果没有看到 `Offering public key`，说明：
- SSH 没有使用 agent
- 检查 SSH_AUTH_SOCK 是否在当前 shell 中设置

### 3. 查看 agent 实时日志

```bash
# 启动 agent 时前台运行以查看日志
./fssh agent --unlock-ttl-seconds 600

# 在另一个终端测试 SSH 连接
export SSH_AUTH_SOCK=~/.fssh/agent.sock
ssh -v user@host
```

agent 日志会显示：
- `创建安全 agent key_count=N` - 加载了多少个 key
- 每次认证请求时的日志

### 4. 常见问题排查

#### 问题：ssh-add -l 显示 "no identities"

**可能原因：**
- Agent 进程启动时未能加载 key
- Key 文件权限问题
- Key 文件损坏

**解决方法：**
```bash
./fssh list                               # 确认 key 已导入
ls -la ~/.fssh/keys/                      # 检查文件权限
pkill -f "fssh agent" && ./fssh agent     # 重启 agent
```

#### 问题：SSH 不使用 agent

**可能原因：**
- SSH_AUTH_SOCK 未在当前 shell 中设置
- SSH 配置指定了其他认证方式
- IdentitiesOnly yes 限制了 key 使用

**解决方法：**
```bash
# 在当前 shell 确认环境变量
echo $SSH_AUTH_SOCK

# 如果未设置，重新设置
export SSH_AUTH_SOCK=~/.fssh/agent.sock

# 检查 SSH 配置
grep -i "IdentitiesOnly\|IdentityFile" ~/.ssh/config

# 强制使用 agent（临时）
ssh -o IdentitiesOnly=no user@host
```

#### 问题：服务器拒绝公钥

**可能原因：**
- 服务器上的 authorized_keys 没有对应的公钥
- 服务器 SSH 配置禁用了公钥认证
- 服务器上文件权限不正确

**解决方法：**
```bash
# 导出公钥
ssh-add -L > /tmp/fssh_pubkey.txt

# 将公钥添加到服务器
ssh-copy-id -i /tmp/fssh_pubkey.txt user@host

# 或手动添加到服务器的 ~/.ssh/authorized_keys
```

### 5. 完整调试流程示例

```bash
# 终端 1: 启动 agent（前台查看日志）
./fssh agent --unlock-ttl-seconds 600

# 终端 2: 测试连接
export SSH_AUTH_SOCK=~/.fssh/agent.sock

# 确认 agent 可访问
ssh-add -l

# verbose 模式连接
ssh -vvv user@host 2>&1 | tee /tmp/ssh-debug.log

# 分析日志
grep "agent" /tmp/ssh-debug.log
grep "Offering public key" /tmp/ssh-debug.log
grep "Authentication" /tmp/ssh-debug.log
```

### 6. 快速诊断命令

```bash
# 一行命令检查所有关键点
(echo "=== Agent Status ===" && ./fssh status) && \
(echo -e "\n=== Agent Process ===" && ps aux | grep "[f]ssh agent") && \
(echo -e "\n=== Environment ===" && echo "SSH_AUTH_SOCK=$SSH_AUTH_SOCK") && \
(echo -e "\n=== Socket File ===" && ls -la $SSH_AUTH_SOCK 2>/dev/null || echo "Not found") && \
(echo -e "\n=== Keys in Agent ===" && ssh-add -l 2>&1)
```

### 7. 日志级别控制

修改 ~/.fssh/config.json 增加调试日志：

```json
{
  "socket": "~/.fssh/agent.sock",
  "require_touch_id_per_sign": true,
  "unlock_ttl_seconds": 600,
  "log_level": "debug",
  "log_format": "plain"
}
```

重启 agent 后会输出更详细的日志。
