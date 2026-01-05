package otp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptPassword 安全地提示输入密码（不回显）
func PromptPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// 检查是否是终端
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// 不是终端，从标准输入读取
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		return scanner.Text(), scanner.Err()
	}

	// 终端环境，使用不回显的密码输入
	password, err := term.ReadPassword(fd)
	fmt.Println() // 换行
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}

	return string(password), nil
}

// PromptPasswordWithConfirm 提示输入密码并确认
func PromptPasswordWithConfirm(prompt, confirmPrompt string) (string, error) {
	password1, err := PromptPassword(prompt)
	if err != nil {
		return "", err
	}

	password2, err := PromptPassword(confirmPrompt)
	if err != nil {
		return "", err
	}

	if password1 != password2 {
		return "", fmt.Errorf("两次输入的密码不一致")
	}

	return password1, nil
}

// PromptCode 提示输入 TOTP 验证码
func PromptCode(prompt string) (string, error) {
	fmt.Print(prompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取验证码失败: %w", err)
	}

	code := strings.TrimSpace(scanner.Text())

	// 验证格式（6或8位数字）
	if len(code) != 6 && len(code) != 8 {
		return "", fmt.Errorf("验证码必须是6位或8位数字")
	}

	for _, c := range code {
		if c < '0' || c > '9' {
			return "", fmt.Errorf("验证码只能包含数字")
		}
	}

	return code, nil
}

// PromptConfirm 提示用户确认（y/n）
func PromptConfirm(prompt string) bool {
	fmt.Printf("%s (y/n): ", prompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))

	return answer == "y" || answer == "yes"
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	// 最小长度检查
	if len(password) < 12 {
		return fmt.Errorf("密码至少需要12位字符")
	}

	// 可选：添加更多密码强度检查
	// 例如：大小写字母、数字、特殊字符等

	return nil
}
