package agentserver

import (
    "crypto/x509"
    "fmt"
    "net"
    "os"
    "path/filepath"

    "fssh/internal/auth"
    "fssh/internal/log"
    "fssh/internal/store"
    xagent "golang.org/x/crypto/ssh/agent"
)

func defaultSocket() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".fssh", "agent.sock")
}

func Start(socketPath string) error { return StartWithOptions(socketPath, true, 0) }

func StartWithOptions(socketPath string, requireTouchPerSign bool, ttlSeconds int) error {
    log.Info("启动 fssh SSH 认证代理", nil)

    if socketPath == "" {
        socketPath = defaultSocket()
    }

    // 获取认证提供者（用于显示认证模式）
    provider, err := auth.GetAuthProvider(ttlSeconds)
    if err != nil {
        return fmt.Errorf("初始化认证提供者失败: %w", err)
    }

    log.Info("认证模式", map[string]interface{}{
        "mode": provider.Mode(),
    })

    // OTP 模式：启动时预先解锁
    if provider.Mode() == auth.ModeOTP && requireTouchPerSign {
        if err := preUnlockOTP(); err != nil {
            return err
        }
    }

    _ = os.Remove(socketPath)
    if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
        return err
    }
    ln, err := net.Listen("unix", socketPath)
    if err != nil {
        return err
    }

    var ag xagent.Agent
    if requireTouchPerSign {
        sa, err := newSecureAgentWithTTL(ttlSeconds)
        if err != nil { ln.Close(); return err }
        ag = sa
        log.Info("安全模式: 每次签名需要认证", map[string]interface{}{
            "ttl_seconds": ttlSeconds,
        })
    } else {
        // 便利模式：启动时解密所有私钥
        mk, err := provider.UnlockMasterKey()
        if err != nil { ln.Close(); return err }
        keyring := xagent.NewKeyring()
        dir := store.KeysDir()
        entries, err := os.ReadDir(dir)
        if err == nil {
            for _, e := range entries {
                if e.IsDir() || filepath.Ext(e.Name()) != ".enc" { continue }
                alias := e.Name()[:len(e.Name())-4]
                rec, err := store.LoadDecryptedRecord(alias, mk)
                if err != nil { continue }
                pk, err := x509.ParsePKCS8PrivateKey(rec.PKCS8DER)
                if err != nil { continue }
                _ = keyring.Add(xagent.AddedKey{PrivateKey: pk, Comment: rec.Alias})
            }
        }
        ag = keyring
        log.Info("便利模式: 启动时解密所有私钥", nil)
    }

    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                return
            }
            go func(c net.Conn) {
                _ = xagent.ServeAgent(ag, c)
                c.Close()
            }(conn)
        }
    }()

    fmt.Println()
    fmt.Println("✓ Agent 已启动")
    fmt.Printf("Socket: %s\n", socketPath)
    fmt.Println()
    fmt.Println("请设置环境变量:")
    fmt.Printf("  export SSH_AUTH_SOCK=%s\n", socketPath)
    fmt.Println()

    // Block until interrupted
    select {}
}

// preUnlockOTP OTP 模式启动时预先解锁
// 提示用户输入密码和验证码，避免首次 SSH 连接时等待
func preUnlockOTP() error {
    fmt.Println()
    fmt.Println("OTP 认证初始化")
    fmt.Println("==============")

    // 创建临时 provider 进行预热
    provider, err := auth.NewOTPProvider(0) // TTL=0，不影响后续的缓存配置
    if err != nil {
        return err
    }

    // 调用 UnlockMasterKey，会提示输入密码和验证码
    _, err = provider.UnlockMasterKey()
    if err != nil {
        return fmt.Errorf("OTP 认证失败: %w", err)
    }

    fmt.Println()
    fmt.Println("✓ OTP 认证成功")

    return nil
}