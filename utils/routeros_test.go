package utils

import (
	"bufio"
	"net"
	"testing"
	"time"
)

func TestRouterOSIPClient_GetOuterIP(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer serverConn.Close()

		api := &routerOSAPIClient{
			conn:   serverConn,
			reader: bufio.NewReader(serverConn),
			writer: bufio.NewWriter(serverConn),
		}

		sentence, err := api.readSentence()
		if err != nil {
			t.Error(err)
			return
		}
		if len(sentence) < 3 || sentence[0] != "/login" {
			t.Errorf("unexpected login sentence: %#v", sentence)
			return
		}
		if err := api.writeSentence("!done"); err != nil {
			t.Error(err)
			return
		}

		sentence, err = api.readSentence()
		if err != nil {
			t.Error(err)
			return
		}
		if len(sentence) < 3 || sentence[0] != "/ip/address/print" {
			t.Errorf("unexpected print sentence: %#v", sentence)
			return
		}
		if err := api.writeSentence("!re", "=address=118.114.31.18/32", "=interface=pppoe-out1"); err != nil {
			t.Error(err)
			return
		}
		if err := api.writeSentence("!done"); err != nil {
			t.Error(err)
			return
		}
	}()

	cli := NewRouterOSIPClient("192.168.88.1", "admin", "secret", "pppoe-out1")
	cli.Dial = func(network string, address string) (net.Conn, error) {
		return clientConn, nil
	}
	got, err := cli.GetOuterIP()
	if err != nil {
		t.Fatal(err)
	}
	if got != "118.114.31.18" {
		t.Fatalf("got %s", got)
	}
	<-done
}

func TestRouterOSChallengeResponse(t *testing.T) {
	got, err := routerOSChallengeResponse("admin", "0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if got != "00a284acbc572682d3147ea4b011c005d3" {
		t.Fatalf("got %s", got)
	}
}

func TestEncodeRouterOSLength(t *testing.T) {
	tests := []struct {
		input int
		want  []byte
	}{
		{input: 127, want: []byte{0x7f}},
		{input: 128, want: []byte{0x80, 0x80}},
		{input: 16384, want: []byte{0xc0, 0x40, 0x00}},
	}
	for _, tt := range tests {
		got := encodeRouterOSLength(tt.input)
		if string(got) != string(tt.want) {
			t.Fatalf("input %d got %#v want %#v", tt.input, got, tt.want)
		}
	}
}

// shortReadConn 每次 Read 最多只交付 1 字节, 用于复现 TCP 分段到达导致的短读。
type shortReadConn struct {
	net.Conn
}

func (c shortReadConn) Read(b []byte) (int, error) {
	if len(b) > 1 {
		b = b[:1]
	}
	return c.Conn.Read(b)
}

// 对端短读时不得丢帧: 曾经用单次 Read 读 word, 短读会破坏帧边界导致错乱/卡死。
func TestRouterOSIPClient_GetOuterIP_PartialRead(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	// 出错时快速失败, 而不是永久阻塞。
	_ = clientConn.SetDeadline(time.Now().Add(10 * time.Second))

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer serverConn.Close()

		api := &routerOSAPIClient{
			conn:   serverConn,
			reader: bufio.NewReader(serverConn),
			writer: bufio.NewWriter(serverConn),
		}
		if _, err := api.readSentence(); err != nil {
			t.Error(err)
			return
		}
		if err := api.writeSentence("!done"); err != nil {
			t.Error(err)
			return
		}
		if _, err := api.readSentence(); err != nil {
			t.Error(err)
			return
		}
		if err := api.writeSentence("!re", "=address=118.114.31.18/32", "=interface=pppoe-out1"); err != nil {
			t.Error(err)
			return
		}
		if err := api.writeSentence("!done"); err != nil {
			t.Error(err)
			return
		}
	}()

	cli := NewRouterOSIPClient("192.168.88.1", "admin", "secret", "pppoe-out1")
	cli.Dial = func(network string, address string) (net.Conn, error) {
		return shortReadConn{clientConn}, nil
	}
	got, err := cli.GetOuterIP()
	if err != nil {
		t.Fatal(err)
	}
	if got != "118.114.31.18" {
		t.Fatalf("got %q, want 118.114.31.18", got)
	}
	<-done
}

// 对端 accept 后不回包时必须超时返回, 否则监控循环会永久卡死。
func TestRouterOSIPClient_GetOuterIP_ServerNoReply(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	held := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		held <- conn // 接受连接但不回任何数据
	}()

	cli := NewRouterOSIPClient(ln.Addr().String(), "admin", "secret", "pppoe-out1")
	cli.Timeout = 300 * time.Millisecond

	type result struct {
		ip  string
		err error
	}
	resCh := make(chan result, 1)
	go func() {
		ip, err := cli.GetOuterIP()
		resCh <- result{ip, err}
	}()

	select {
	case res := <-resCh:
		if res.err == nil {
			t.Fatalf("对端不回包时应返回错误, got ip=%q", res.ip)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("对端不回包时 GetOuterIP 未返回, 读操作缺少超时")
	}
}
