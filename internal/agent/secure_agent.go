package agentserver

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"encoding/json"
	"fssh/internal/auth"
	"fssh/internal/log"
	"fssh/internal/store"

	"golang.org/x/crypto/ssh"
	xagent "golang.org/x/crypto/ssh/agent"
)

type secureAgent struct {
	authProvider auth.AuthProvider
}

func newSecureAgentWithTTL(ttlSeconds int) (*secureAgent, error) {
	// 获取认证提供者
	provider, err := auth.GetAuthProvider(ttlSeconds)
	if err != nil {
		return nil, err
	}

	agent := &secureAgent{
		authProvider: provider,
	}

	// 加载密钥计数用于日志
	metas, _ := agent.loadMetas()
	log.Info("创建安全 agent", map[string]interface{}{
		"auth_mode": provider.Mode(),
		"key_count": len(metas),
	})

	return agent, nil
}

// loadMetas 动态加载所有加密私钥的元数据
func (a *secureAgent) loadMetas() ([]store.EncryptedFile, error) {
	dir := store.KeysDir()
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var metas []store.EncryptedFile
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".enc" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var m store.EncryptedFile
		if err := jsonUnmarshal(b, &m); err != nil {
			continue
		}
		metas = append(metas, m)
	}
	return metas, nil
}

func (a *secureAgent) List() ([]*xagent.Key, error) {
	// 动态加载最新的密钥列表
	metas, err := a.loadMetas()
	if err != nil {
		return nil, err
	}

	var ks []*xagent.Key
	for _, m := range metas {
		if m.PubKey == "" {
			continue
		}
		pb, err := base64.StdEncoding.DecodeString(m.PubKey)
		if err != nil {
			continue
		}
		pk, err := ssh.ParsePublicKey(pb)
		if err != nil {
			continue
		}
		ks = append(ks, &xagent.Key{Format: pk.Type(), Blob: pk.Marshal(), Comment: m.Alias})
	}
	return ks, nil
}

func (a *secureAgent) Sign(pubkey ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	fp := ssh.FingerprintSHA256(pubkey)

	// 动态加载最新的密钥列表
	metas, err := a.loadMetas()
	if err != nil {
		return nil, err
	}

	var alias string
	for _, m := range metas {
		if m.Fingerprint == fp {
			alias = m.Alias
			break
		}
	}
	if alias == "" {
		return nil, errors.New("key not found")
	}

	// 使用 AuthProvider 解锁 master key
	mk, err := a.authProvider.UnlockMasterKey()
	if err != nil {
		return nil, fmt.Errorf("认证失败: %w", err)
	}

	rec, err := store.LoadDecryptedRecord(alias, mk)
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParsePKCS8PrivateKey(rec.PKCS8DER)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, err
	}

	log.Debug("SSH 签名", map[string]interface{}{
		"fingerprint": fp,
		"key_type":    signer.PublicKey().Type(),
	})

	return signer.Sign(nil, data)
}

// Support RSA-SHA2 algorithms when requested by the client.
func (a *secureAgent) SignWithFlags(pubkey ssh.PublicKey, data []byte, flags xagent.SignatureFlags) (*ssh.Signature, error) {
	fp := ssh.FingerprintSHA256(pubkey)

	// 动态加载最新的密钥列表
	metas, err := a.loadMetas()
	if err != nil {
		return nil, err
	}

	var alias string
	for _, m := range metas {
		if m.Fingerprint == fp {
			alias = m.Alias
			break
		}
	}
	if alias == "" {
		return nil, errors.New("key not found")
	}

	// 使用 AuthProvider 解锁 master key
	mk, err := a.authProvider.UnlockMasterKey()
	if err != nil {
		return nil, fmt.Errorf("认证失败: %w", err)
	}

	rec, err := store.LoadDecryptedRecord(alias, mk)
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParsePKCS8PrivateKey(rec.PKCS8DER)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, err
	}

	if algSigner, ok := signer.(ssh.AlgorithmSigner); ok {
		algo := ""
		if (flags & xagent.SignatureFlagRsaSha512) != 0 {
			algo = "rsa-sha2-512"
		} else if (flags & xagent.SignatureFlagRsaSha256) != 0 {
			algo = "rsa-sha2-256"
		}
		if algo != "" {
			log.Debug("SSH 签名 (RSA-SHA2)", map[string]interface{}{
				"fingerprint": fp,
				"algorithm":   algo,
				"flags":       flags,
			})
			return algSigner.SignWithAlgorithm(rand.Reader, data, algo)
		}
	}

	log.Debug("SSH 签名 (fallback)", map[string]interface{}{
		"fingerprint": fp,
		"flags":       flags,
	})

	return signer.Sign(nil, data)
}

func (a *secureAgent) Extension(extensionType string, contents []byte) ([]byte, error) {
	if extensionType == "ext-info-c" {
		// Advertise support for RSA SHA-256 and SHA-512 signature flags as a uint32 bitmask (big-endian)
		flags := uint32(xagent.SignatureFlagRsaSha256 | xagent.SignatureFlagRsaSha512)
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, flags)
		log.Debug("agent extension ext-info-c", map[string]interface{}{
			"flags": flags,
		})
		return b, nil
	}
	log.Debug("agent extension unsupported", map[string]interface{}{
		"type": extensionType,
	})
	return nil, errors.New("unsupported extension")
}

func (a *secureAgent) Add(key xagent.AddedKey) error     { return errors.New("unsupported") }
func (a *secureAgent) Remove(pubkey ssh.PublicKey) error { return errors.New("unsupported") }
func (a *secureAgent) RemoveAll() error                  { return nil }
func (a *secureAgent) Lock(passphrase []byte) error      { return nil }
func (a *secureAgent) Unlock(passphrase []byte) error    { return nil }
func (a *secureAgent) Signers() ([]ssh.Signer, error)    { return nil, errors.New("unsupported") }

func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }
