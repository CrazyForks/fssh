package agentserver

import (
    "crypto/x509"
    "encoding/base64"
    "errors"
    "os"
    "path/filepath"

    "fssh/internal/keychain"
    "fssh/internal/store"
    "golang.org/x/crypto/ssh"
    xagent "golang.org/x/crypto/ssh/agent"
    "encoding/json"
)

type secureAgent struct {
    metas []store.EncryptedFile
}

func newSecureAgent() (*secureAgent, error) {
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
        if err != nil { continue }
        var m store.EncryptedFile
        if err := jsonUnmarshal(b, &m); err != nil { continue }
        metas = append(metas, m)
    }
    return &secureAgent{metas: metas}, nil
}

func (a *secureAgent) List() ([]*xagent.Key, error) {
    var ks []*xagent.Key
    for _, m := range a.metas {
        if m.PubKey == "" { continue }
        pb, err := base64.StdEncoding.DecodeString(m.PubKey)
        if err != nil { continue }
        pk, err := ssh.ParsePublicKey(pb)
        if err != nil { continue }
        ks = append(ks, &xagent.Key{Format: pk.Type(), Blob: pk.Marshal(), Comment: m.Alias})
    }
    return ks, nil
}

func (a *secureAgent) Sign(pubkey ssh.PublicKey, data []byte) (*ssh.Signature, error) {
    fp := ssh.FingerprintSHA256(pubkey)
    var alias string
    for _, m := range a.metas {
        if m.Fingerprint == fp { alias = m.Alias; break }
    }
    if alias == "" {
        return nil, errors.New("key not found")
    }
    mk, err := keychain.LoadMasterKey()
    if err != nil { return nil, err }
    rec, err := store.LoadDecryptedRecord(alias, mk)
    if err != nil { return nil, err }
    priv, err := x509.ParsePKCS8PrivateKey(rec.PKCS8DER)
    if err != nil { return nil, err }
    signer, err := ssh.NewSignerFromKey(priv)
    if err != nil { return nil, err }
    return signer.Sign(nil, data)
}

func (a *secureAgent) Add(key xagent.AddedKey) error { return errors.New("unsupported") }
func (a *secureAgent) Remove(pubkey ssh.PublicKey) error { return errors.New("unsupported") }
func (a *secureAgent) RemoveAll() error { return nil }
func (a *secureAgent) Lock(passphrase []byte) error { return nil }
func (a *secureAgent) Unlock(passphrase []byte) error { return nil }
func (a *secureAgent) Signers() ([]ssh.Signer, error) { return nil, errors.New("unsupported") }

func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }
