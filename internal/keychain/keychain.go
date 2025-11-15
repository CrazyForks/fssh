package keychain

import (
    "errors"
    "fmt"

    kc "github.com/keybase/go-keychain"
    "fssh/internal/macos"
)

const (
    serviceNew = "fssh"
    serviceOld = "fingerpass"
    account    = "master_key_v1"
)

func MasterKeyExists() (bool, error) {
    exists, err := masterKeyExistsForService(serviceNew)
    if err != nil {
        return false, err
    }
    if exists {
        return true, nil
    }
    return masterKeyExistsForService(serviceOld)
}

func masterKeyExistsForService(svc string) (bool, error) {
    q := kc.NewItem()
    q.SetSecClass(kc.SecClassGenericPassword)
    q.SetService(svc)
    q.SetAccount(account)
    q.SetMatchLimit(kc.MatchLimitOne)
    q.SetReturnData(true)
    res, err := kc.QueryItem(q)
    if err != nil {
        if errors.Is(err, kc.ErrorItemNotFound) {
            return false, nil
        }
        return false, err
    }
    return len(res) > 0, nil
}

func StoreMasterKey(key []byte, overwrite bool) error {
    exists, err := MasterKeyExists()
    if err != nil {
        return err
    }
    if exists && !overwrite {
        return nil
    }
    if exists && overwrite {
        if err := DeleteMasterKey(); err != nil {
            return err
        }
    }
    it := kc.NewItem()
    it.SetSecClass(kc.SecClassGenericPassword)
    it.SetService(serviceNew)
    it.SetAccount(account)
    it.SetAccessible(kc.AccessibleWhenUnlocked)
    it.SetData(key)
    return kc.AddItem(it)
}

func LoadMasterKey() ([]byte, error) {
    // Gate access behind biometry prompt
    if err := macos.RequireBiometry("解锁指纹受保护的主密钥以使用 SSH 私钥"); err != nil {
        return nil, err
    }
    res, err := queryMasterKey(serviceNew)
    if err != nil {
        return nil, err
    }
    if len(res) == 0 {
        // try old service for backward compatibility
        res, err = queryMasterKey(serviceOld)
        if err != nil {
            return nil, err
        }
        if len(res) == 0 {
            return nil, fmt.Errorf("master key not initialized")
        }
    }
    return res[0].Data, nil
}

func queryMasterKey(svc string) ([]kc.QueryResult, error) {
    q := kc.NewItem()
    q.SetSecClass(kc.SecClassGenericPassword)
    q.SetService(svc)
    q.SetAccount(account)
    q.SetMatchLimit(kc.MatchLimitOne)
    q.SetReturnData(true)
    return kc.QueryItem(q)
}

func DeleteMasterKey() error {
    it := kc.NewItem()
    it.SetSecClass(kc.SecClassGenericPassword)
    it.SetService(serviceNew)
    it.SetAccount(account)
    _ = kc.DeleteItem(it)
    it2 := kc.NewItem()
    it2.SetSecClass(kc.SecClassGenericPassword)
    it2.SetService(serviceOld)
    it2.SetAccount(account)
    return kc.DeleteItem(it2)
}