package main

import (
    "fmt"
    "net"
    "os"
    "os/exec"
    "strconv"
    "strings"

    "fssh/internal/sshconfig"
    "github.com/peterh/liner"
)

func runShell() {
    infos, err := sshconfig.LoadHostInfos()
    if err != nil {
        fatal(err)
    }
    hosts := make([]string, 0, len(infos))
    for _, hi := range infos { hosts = append(hosts, hi.Name) }
    byName := map[string]sshconfig.HostInfo{}
    for _, hi := range infos { byName[hi.Name] = hi }
    byHostname := map[string]sshconfig.HostInfo{}
    ipToName := map[string]string{}
    hostnames := make([]string, 0, len(infos))
    ips := make([]string, 0, len(infos))
    idToName := map[string]string{}
    ids := make([]string, 0, len(infos))
    for _, hi := range infos {
        hn := displayHostname(hi)
        if hn != "" {
            byHostname[hn] = hi
            hostnames = append(hostnames, hn)
        }
    }
    for i, hi := range infos {
        id := strconv.Itoa(i + 1)
        idToName[id] = hi.Name
        ids = append(ids, id)
    }
    for _, hi := range infos {
        hn := displayHostname(hi)
        ip := resolveIPName(hn)
        if ip != "" {
            ips = append(ips, ip)
            if _, ok := ipToName[ip]; !ok {
                ipToName[ip] = hi.Name
            }
        }
    }
    commands := []string{"list", "search", "connect", "help", "exit", "quit"}
    l := setupLiner(commands, hosts, hostnames, ips, ids)
    defer l.Close()
    for {
        line, err := l.Prompt("fssh> ")
        if err != nil {
            return
        }
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        l.AppendHistory(line)
        if line == "exit" || line == "quit" {
            return
        }
        if line == "help" {
            fmt.Println("commands: list | search <term> | connect <host> | help | exit | Tab for completion; non-command defaults to connect")
            continue
        }
        if line == "list" {
            for i, hi := range infos {
                target := hi.Hostname
                if target == "" { target = hi.Name }
                ip := resolveIPName(target)
                if ip == "" { ip = "-" }
                fmt.Printf("%d\t%s(%s)\n", i+1, hi.Name, ip)
            }
            continue
        }
        if strings.HasPrefix(line, "search ") {
            term := strings.TrimSpace(line[7:])
            if term == "" {
                continue
            }
            tl := strings.ToLower(term)
            for i, hi := range infos {
                hn := hi.Hostname
                target := hn
                if target == "" { target = hi.Name }
                ip := resolveIPName(target)
                match := strings.Contains(strings.ToLower(hi.Name), tl) || (hn != "" && strings.Contains(strings.ToLower(hn), tl)) || (ip != "" && strings.Contains(strings.ToLower(ip), tl))
                if match {
                    if ip == "" { ip = "-" }
                    fmt.Printf("%d\t%s(%s)\n", i+1, hi.Name, ip)
                }
            }
            continue
        }
        if strings.HasPrefix(line, "connect ") {
            host := strings.TrimSpace(line[8:])
            if host == "" {
                continue
            }
            if name, ok := idToName[host]; ok {
                host = name
            }
            _, found := byName[host]
            if !found {
                if hi, ok := byHostname[host]; ok {
                    host = hi.Name
                    found = true
                }
            }
            if !found {
                if name, ok := ipToName[host]; ok {
                    host = name
                    found = true
                }
            }
            if !found {
                fmt.Fprintf(os.Stderr, "unknown host: %s\n", host)
                continue
            }
            l.Close()
            cmd := exec.Command("ssh", "-tt", host)
            cmd.Stdin = os.Stdin
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            _ = cmd.Run()
            l = setupLiner(commands, hosts, hostnames, ips, ids)
            continue
        }
        host := line
        if name, ok := idToName[host]; ok {
            host = name
        }
        _, found := byName[host]
        if !found {
            if hi, ok := byHostname[host]; ok {
                host = hi.Name
                found = true
            }
        }
        if !found {
            if name, ok := ipToName[host]; ok {
                host = name
                found = true
            }
        }
        if !found {
            fmt.Fprintf(os.Stderr, "unknown host: %s\n", host)
            continue
        }
        l.Close()
        cmd := exec.Command("ssh", "-tt", host)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        _ = cmd.Run()
        l = setupLiner(commands, hosts, hostnames, ips, ids)
    }
}

func setupLiner(commands, hosts, hostnames, ips, ids []string) *liner.State {
    l := liner.NewLiner()
    l.SetCtrlCAborts(true)
    l.SetCompleter(func(line string) []string {
        line = strings.TrimSpace(line)
        var out []string
        if line == "" {
            out = append(out, commands...)
            out = append(out, hosts...)
            out = append(out, hostnames...)
            out = append(out, ips...)
            out = append(out, ids...)
            return out
        }
        if strings.HasPrefix(line, "connect ") {
            p := strings.TrimSpace(line[8:])
            for _, h := range hosts {
                if strings.HasPrefix(h, p) {
                    out = append(out, "connect "+h)
                }
            }
            for _, h := range hostnames {
                if strings.HasPrefix(h, p) {
                    out = append(out, "connect "+h)
                }
            }
            for _, ip := range ips {
                if strings.HasPrefix(ip, p) {
                    out = append(out, "connect "+ip)
                }
            }
            for _, id := range ids {
                if strings.HasPrefix(id, p) {
                    out = append(out, "connect "+id)
                }
            }
            return out
        }
        for _, c := range commands {
            if strings.HasPrefix(c, line) {
                out = append(out, c)
            }
        }
        for _, h := range hosts {
            if strings.HasPrefix(h, line) {
                out = append(out, h)
            }
        }
        for _, h := range hostnames {
            if strings.HasPrefix(h, line) {
                out = append(out, h)
            }
        }
        for _, ip := range ips {
            if strings.HasPrefix(ip, line) {
                out = append(out, ip)
            }
        }
        for _, id := range ids {
            if strings.HasPrefix(id, line) {
                out = append(out, id)
            }
        }
        return out
    })
    return l
}

func resolveIPName(target string) string {
    if target == "" { return "" }
    ips, err := net.LookupIP(target)
    if err != nil || len(ips) == 0 { return "" }
    for _, ip := range ips {
        if ip.To4() != nil { return ip.String() }
    }
    return ips[0].String()
}

func displayHostname(hi sshconfig.HostInfo) string {
    if hi.Hostname != "" { return hi.Hostname }
    return hi.Name
}