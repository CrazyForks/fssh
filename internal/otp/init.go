package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"fssh/internal/crypt"

	"golang.org/x/crypto/pbkdf2"
)

// InitOptions OTP 初始化选项
type InitOptions struct {
	Password         string // OTP 密码
	SeedUnlockTTL    int    // OTP seed 缓存时间（秒）
	Algorithm        string // TOTP 算法: SHA1, SHA256, SHA512
	Digits           int    // 验证码位数: 6 或 8
	Period           int    // TOTP 时间窗口（秒）
	GenerateRecovery bool   // 是否生成恢复码
}

// DefaultInitOptions 返回默认初始化选项
func DefaultInitOptions() *InitOptions {
	return &InitOptions{
		SeedUnlockTTL:    3600, // 1小时
		Algorithm:        "SHA1",
		Digits:           6,
		Period:           30,
		GenerateRecovery: true,
	}
}

// Initialize 初始化 OTP 配置
// 1. 生成随机 OTP seed
// 2. 使用密码加密 seed
// 3. 生成恢复码
// 4. 保存配置
func Initialize(opts *InitOptions) (seed []byte, recoveryCodes []string, err error) {
	// 验证密码强度
	if err := ValidatePasswordStrength(opts.Password); err != nil {
		return nil, nil, fmt.Errorf("密码强度不足: %w", err)
	}

	// 1. 生成随机 OTP seed (20 字节，标准 TOTP seed 长度)
	seed = make([]byte, 20)
	if _, err := rand.Read(seed); err != nil {
		return nil, nil, fmt.Errorf("生成 OTP seed 失败: %w", err)
	}

	// 2. 生成加密参数
	seedSalt := make([]byte, 32)
	if _, err := rand.Read(seedSalt); err != nil {
		return nil, nil, fmt.Errorf("生成 seed salt 失败: %w", err)
	}

	seedNonce := make([]byte, 12) // AES-GCM nonce 12 字节
	if _, err := rand.Read(seedNonce); err != nil {
		return nil, nil, fmt.Errorf("生成 seed nonce 失败: %w", err)
	}

	masterKeySalt := make([]byte, 32)
	if _, err := rand.Read(masterKeySalt); err != nil {
		return nil, nil, fmt.Errorf("生成 master key salt 失败: %w", err)
	}

	// 3. 派生加密密钥 (PBKDF2)
	encKey := pbkdf2.Key([]byte(opts.Password), seedSalt, 100000, 32, sha256.New)

	// 4. 加密 OTP seed
	encryptedSeed, err := crypt.EncryptAEAD(encKey, seedNonce, seed, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("加密 OTP seed 失败: %w", err)
	}

	// 5. 生成恢复码
	var recoveryCodesHash []string
	if opts.GenerateRecovery {
		recoveryCodes, err = GenerateRecoveryCodes(10)
		if err != nil {
			return nil, nil, fmt.Errorf("生成恢复码失败: %w", err)
		}
		recoveryCodesHash = HashRecoveryCodes(recoveryCodes)
	}

	// 6. 创建配置
	cfg := &Config{
		Version:              "fssh-otp/v1",
		Algorithm:            opts.Algorithm,
		Digits:               opts.Digits,
		Period:               opts.Period,
		EncryptedSeed:        base64.StdEncoding.EncodeToString(encryptedSeed),
		SeedSalt:             base64.StdEncoding.EncodeToString(seedSalt),
		SeedNonce:            base64.StdEncoding.EncodeToString(seedNonce),
		MasterKeySalt:        base64.StdEncoding.EncodeToString(masterKeySalt),
		SeedUnlockTTLSeconds: opts.SeedUnlockTTL,
		RecoveryCodesHash:    recoveryCodesHash,
		CreatedAt:            time.Now().Format(time.RFC3339),
	}

	// 7. 保存配置
	if err := SaveConfig(cfg); err != nil {
		return nil, nil, fmt.Errorf("保存配置失败: %w", err)
	}

	return seed, recoveryCodes, nil
}

// DisplayInitResult 显示初始化结果
func DisplayInitResult(seed []byte, recoveryCodes []string, algorithm string, digits int, period int) error {
	// 获取账户信息
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	account := fmt.Sprintf("%s@%s", user, hostname)

	// 显示 OTP 设置
	secret := base64.StdEncoding.EncodeToString(seed)

	fmt.Println()
	fmt.Println("OTP 认证已初始化")
	fmt.Println("================")
	fmt.Println()

	// 显示配置信息
	fmt.Println("TOTP 配置:")
	fmt.Printf("  发行者: fssh\n")
	fmt.Printf("  账户: %s\n", account)
	fmt.Printf("  密钥: %s\n", secret)
	fmt.Printf("  算法: %s\n", algorithm)
	fmt.Printf("  位数: %d\n", digits)
	fmt.Printf("  间隔: %d秒\n", period)
	fmt.Println()

	fmt.Println("请将以上信息添加到 TOTP 认证器应用（如 Google Authenticator, Authy）")
	fmt.Println()

	// 显示恢复码
	if len(recoveryCodes) > 0 {
		DisplayRecoveryCodes(recoveryCodes)
	}

	// 显示配置文件路径
	fmt.Printf("配置已保存到: %s\n", ConfigPath())
	fmt.Println()

	// 安全提示
	fmt.Println("⚠️  安全提示:")
	fmt.Println("  1. 建议启用 FileVault 全盘加密")
	fmt.Println("  2. 建议在第二台设备也添加此 OTP（记录密钥信息）")
	fmt.Println("  3. 妥善保管恢复码，丢失无法找回")
	fmt.Println()

	// 下一步
	fmt.Println("下一步:")
	fmt.Println("  1. 导入 SSH 私钥: fssh import -alias myserver -file ~/.ssh/id_rsa")
	fmt.Println("  2. 启动 agent: fssh agent")
	fmt.Println()

	return nil
}
