package utils

import (
	"bufio"
	"net"
	"testing"
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
