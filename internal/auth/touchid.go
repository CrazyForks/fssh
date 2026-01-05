package auth

import (
	"fssh/internal/keychain"
)

// TouchIDProvider Touch ID 认证提供者
// 使用 macOS Keychain 和 Touch ID 进行认证
type TouchIDProvider struct {
	// Touch ID 不需要缓存，Keychain 自己处理
}

// NewTouchIDProvider 创建 Touch ID 认证提供者
func NewTouchIDProvider() *TouchIDProvider {
	return &TouchIDProvider{}
}

// UnlockMasterKey 实现 AuthProvider 接口
// 通过 Touch ID 从 Keychain 加载 master key
func (p *TouchIDProvider) UnlockMasterKey() ([]byte, error) {
	return keychain.LoadMasterKey()
}

// IsAvailable 实现 AuthProvider 接口
// 检查 Keychain 中是否存在 master key
func (p *TouchIDProvider) IsAvailable() bool {
	exists, err := keychain.MasterKeyExists()
	return err == nil && exists
}

// Mode 实现 AuthProvider 接口
func (p *TouchIDProvider) Mode() AuthMode {
	return ModeTouchID
}

// ClearCache 实现 AuthProvider 接口
// Touch ID 不需要清除缓存（Keychain 自己管理）
func (p *TouchIDProvider) ClearCache() {
	// Touch ID 模式不使用内存缓存
	// Keychain 由系统管理，无需手动清理
}
