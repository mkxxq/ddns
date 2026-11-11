package utils

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// fakeDNS 起一个只回 A/NXDOMAIN 的最小 UDP DNS 服务器, 用于验证"显式指定服务器"
// 与"按顺序回退"两条路径, 不依赖外网。
type fakeDNS struct {
	conn    net.PacketConn
	addr    string
	ip      net.IP
	nx      bool
	silent  bool
	queries chan string
}

func startFakeDNS(t *testing.T, ip string, nx bool, silent bool) *fakeDNS {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	f := &fakeDNS{conn: conn, addr: conn.LocalAddr().String(), nx: nx, silent: silent, queries: make(chan string, 8)}
	if ip != "" {
		f.ip = net.ParseIP(ip)
	}
	go f.serve()
	t.Cleanup(func() { conn.Close() })
	return f
}

func (f *fakeDNS) serve() {
	buf := make([]byte, 512)
	for {
		n, from, err := f.conn.ReadFrom(buf)
		if err != nil {
			return
		}
		query := append([]byte(nil), buf[:n]...)
		if f.silent {
			continue
		}
		select {
		case f.queries <- decodeQName(query):
		default:
		}
		f.conn.WriteTo(f.reply(query), from)
	}
}

func decodeQName(msg []byte) string {
	if len(msg) < 13 {
		return ""
	}
	var name []byte
	i := 12
	for i < len(msg) && msg[i] != 0 {
		l := int(msg[i])
		i++
		if i+l > len(msg) {
			return ""
		}
		name = append(name, msg[i:i+l]...)
		name = append(name, '.')
		i += l
	}
	return string(name)
}

func (f *fakeDNS) reply(query []byte) []byte {
	// 问题段整体回显, 之后追加一条 A 记录或置 NXDOMAIN。
	i := 12
	for i < len(query) && query[i] != 0 {
		i += 1 + int(query[i])
	}
	questionEnd := i + 5 // 根标签 + QTYPE + QCLASS
	if questionEnd > len(query) {
		return query
	}

	resp := append([]byte(nil), query[:questionEnd]...)
	resp[2] |= 0x80 // QR
	if f.nx {
		resp[3] = 0x03 // NXDOMAIN
		binary.BigEndian.PutUint16(resp[6:8], 0)
		return resp
	}
	resp[3] = 0x00
	binary.BigEndian.PutUint16(resp[6:8], 1) // ANCOUNT
	resp = append(resp, 0xC0, 0x0C)          // 指向问题中的域名
	resp = append(resp, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x3C, 0x00, 0x04)
	resp = append(resp, f.ip.To4()...)
	return resp
}

// 必须只向配置的服务器发查询: 这台机器的系统 resolver 指向旁路由,
// 用它比对出口 IP 会误判记录已过期。
func TestLookupDomainIPWithDNSServers_QueriesPinnedServer(t *testing.T) {
	srv := startFakeDNS(t, "203.0.113.9", false, false)

	ip, err := LookupDomainIPWithDNSServers("cloud.example.com", []string{srv.addr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "203.0.113.9" {
		t.Fatalf("ip = %q, want 203.0.113.9", ip)
	}
	select {
	case q := <-srv.queries:
		if q != "cloud.example.com." {
			t.Fatalf("query name = %q", q)
		}
	default:
		t.Fatal("pinned server received no query")
	}
}

// 第一个解析器失败时必须回退到下一个, 而不是放弃这一轮。
func TestLookupDomainIPWithDNSServers_FallsBack(t *testing.T) {
	bad := startFakeDNS(t, "", true, false)
	good := startFakeDNS(t, "198.51.100.7", false, false)

	ip, err := LookupDomainIPWithDNSServers("cloud.example.com", []string{bad.addr, good.addr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "198.51.100.7" {
		t.Fatalf("ip = %q, want 198.51.100.7", ip)
	}
}

// 全部解析器失败时返回组合错误, 上层据此保留缓存而不是清空。
func TestLookupDomainIPWithDNSServers_AllFail(t *testing.T) {
	bad := startFakeDNS(t, "", true, false)

	_, err := LookupDomainIPWithDNSServers("cloud.example.com", []string{bad.addr})
	if err == nil {
		t.Fatal("expected error when all servers fail")
	}
}

// 单个解析器不回包时必须有上界: Go 的 resolver 默认会挂 20s, 期间循环停摆。
func TestLookupDomainIPWithDNSServers_TimeoutBound(t *testing.T) {
	silent := startFakeDNS(t, "", false, true)

	start := time.Now()
	_, err := LookupDomainIPWithDNSServers("cloud.example.com", []string{silent.addr})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 3*lookupTimeout {
		t.Fatalf("lookup took %v, want bounded by %v", elapsed, 3*lookupTimeout)
	}
}

// 空配置必须报错, 避免静默走系统 resolver。
func TestLookupDomainIPWithDNSServers_NoServer(t *testing.T) {
	for _, servers := range [][]string{nil, {" "}, {""}} {
		if _, err := LookupDomainIPWithDNSServers("cloud.example.com", servers); err == nil {
			t.Fatalf("servers %#v: expected error", servers)
		}
	}
}
