package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"fssh/internal/auth"
	"fssh/internal/crypt"
	"fssh/internal/keychain"
	"fssh/internal/otp"
)

// initOTPMode 初始化 OTP 认证模式
func initOTPMode(force bool, seedTTL int, algorithm string, digits int) {
	// 检查是否已存在 OTP 配置
	if otp.ConfigExists() && !force {
		fmt.Println("OTP 配置已存在，使用 --force 覆盖")
		return
	}

	// 提示设置密码
	fmt.Println("初始化 OTP 认证模式")
	fmt.Println()

	password, err := otp.PromptPasswordWithConfirm(
		"请设置 OTP 密码（至少12位）: ",
		"确认密码: ",
	)
	if err != nil {
		fatal(err)
	}

	// 初始化选项
	opts := &otp.InitOptions{
		Password:         password,
		SeedUnlockTTL:    seedTTL,
		Algorithm:        algorithm,
		Digits:           digits,
		Period:           30,
		GenerateRecovery: true,
	}

	// 执行初始化
	seed, recoveryCodes, err := otp.Initialize(opts)
	if err != nil {
		fatal(err)
	}

	// 生成 master key（从 OTP seed 派生）
	masterKey, err := deriveMasterKeyFromSeed(seed, opts)
	if err != nil {
		fatal(err)
	}

	// 保存 master key 到 Keychain（用于 import/export 等命令）
	if err := keychain.StoreMasterKey(masterKey, force); err != nil {
		fatal(err)
	}

	// 显示结果
	if err := otp.DisplayInitResult(seed, recoveryCodes, algorithm, digits, 30); err != nil {
		fatal(err)
	}

	// 保存认证模式
	if err := auth.SaveMode(auth.ModeOTP); err != nil {
		fatal(fmt.Errorf("保存认证模式失败: %w", err))
	}
}

// initTouchIDMode 初始化 Touch ID 认证模式
func initTouchIDMode(force bool) {
	exists, err := keychain.MasterKeyExists()
	if err != nil {
		fatal(err)
	}
	if exists && !force {
		fmt.Println("master key already exists")
		return
	}
	mk := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, mk); err != nil {
		fatal(err)
	}
	if err := keychain.StoreMasterKey(mk, force); err != nil {
		fatal(err)
	}

	// 保存认证模式
	if err := auth.SaveMode(auth.ModeTouchID); err != nil {
		fmt.Printf("警告: 保存认证模式失败: %v\n", err)
	}

	fmt.Println("initialized master key with Touch ID protection")
}

// deriveMasterKeyFromSeed 从 OTP seed 派生 master key
// 使用与 OTPProvider 相同的方法，确保一致性
func deriveMasterKeyFromSeed(seed []byte, opts *otp.InitOptions) ([]byte, error) {
	// 从配置中读取 master key salt
	cfg, err := otp.LoadConfig(otp.ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	masterKeySalt, err := base64.StdEncoding.DecodeString(cfg.MasterKeySalt)
	if err != nil {
		return nil, fmt.Errorf("解码 master key salt 失败: %w", err)
	}

	// 使用 HKDF 派生 master key
	masterKey := crypt.HKDF(seed, masterKeySalt, []byte("fssh-master-key-v1"), 32)
	return masterKey, nil
}
