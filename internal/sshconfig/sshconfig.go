package sshconfig

import (
    "bufio"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

type HostInfo struct {
    Name     string
    Hostname string
}

func LoadHosts() ([]string, error) {
    infos, err := LoadHostInfos()
    if err != nil { return nil, err }
    hosts := make([]string, 0, len(infos))
    for _, hi := range infos { hosts = append(hosts, hi.Name) }
    sort.Strings(hosts)
    return hosts, nil
}

func LoadHostInfos() ([]HostInfo, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    p := filepath.Join(home, ".ssh", "config")
    f, err := os.Open(p)
    if err != nil {
        if os.IsNotExist(err) {
            return []HostInfo{}, nil
        }
        return nil, err
    }
    defer f.Close()
    infos := map[string]*HostInfo{}
    current := []string{}
    s := bufio.NewScanner(f)
    for s.Scan() {
        raw := s.Text()
        line := strings.TrimSpace(raw)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        lower := strings.ToLower(line)
        if strings.HasPrefix(lower, "host ") {
            rest := strings.TrimSpace(line[5:])
            parts := strings.Fields(rest)
            current = current[:0]
            for _, h := range parts {
                if h == "*" || strings.ContainsAny(h, "*?") {
                    continue
                }
                current = append(current, h)
                if _, ok := infos[h]; !ok {
                    infos[h] = &HostInfo{Name: h}
                }
            }
            continue
        }
        if len(current) == 0 {
            continue
        }
        k, v := parseKV(line)
        if k == "hostname" && v != "" {
            for _, h := range current {
                infos[h].Hostname = v
            }
        }
    }
    if err := s.Err(); err != nil {
        return nil, err
    }
    out := make([]HostInfo, 0, len(infos))
    for _, v := range infos {
        out = append(out, *v)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
    return out, nil
}

func parseKV(line string) (string, string) {
    // support "Key Value" and "Key=Value", case-insensitive key
    if eq := strings.Index(line, "="); eq >= 0 {
        k := strings.ToLower(strings.TrimSpace(line[:eq]))
        v := strings.TrimSpace(line[eq+1:])
        return k, v
    }
    parts := strings.Fields(line)
    if len(parts) == 0 { return "", "" }
    k := strings.ToLower(parts[0])
    v := strings.TrimSpace(strings.TrimPrefix(line, parts[0]))
    v = strings.TrimSpace(v)
    return k, v
}