package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const recoveryCodeLength = 16 // XXXX-XXXX-XXXX-XXXX

// GenerateRecoveryCodes 生成恢复码
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code, err := generateSingleRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}

	return codes, nil
}

// generateSingleRecoveryCode 生成单个恢复码
// 格式: XXXX-XXXX-XXXX-XXXX（去除易混淆字符）
func generateSingleRecoveryCode() (string, error) {
	// 字符集：大写字母和数字（去除易混淆字符 0/O, 1/I/L）
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	bytes := make([]byte, recoveryCodeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		part := ""
		for j := 0; j < 4; j++ {
			idx := int(bytes[i*4+j]) % len(charset)
			part += string(charset[idx])
		}
		parts[i] = part
	}

	return strings.Join(parts, "-"), nil
}

// HashRecoveryCodes 计算恢复码的哈希
// 存储哈希而非明文，提高安全性
func HashRecoveryCodes(codes []string) []string {
	hashes := make([]string, len(codes))

	for i, code := range codes {
		hash := sha256.Sum256([]byte(code))
		hashes[i] = hex.EncodeToString(hash[:])
	}

	return hashes
}

// VerifyRecoveryCode 验证恢复码
// 返回是否有效和在哈希列表中的索引
func VerifyRecoveryCode(code string, hashes []string) (bool, int) {
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	for i, h := range hashes {
		if h == codeHash {
			return true, i
		}
	}

	return false, -1
}

// DisplayRecoveryCodes 显示恢复码
func DisplayRecoveryCodes(codes []string) {
	fmt.Println("恢复码（请安全保存，打印或存入密码管理器）：")
	fmt.Println()

	for i, code := range codes {
		fmt.Printf("  %2d. %s\n", i+1, code)
	}

	fmt.Println()
	fmt.Println("⚠️  重要提示：")
	fmt.Println("  - 每个恢复码仅可使用一次")
	fmt.Println("  - 使用后该恢复码将立即失效")
	fmt.Println("  - 请妥善保管，丢失无法找回")
	fmt.Println()
}
