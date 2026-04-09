package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	defaultRouterOSPort    = "8728"
	defaultRouterOSTimeout = 5 * time.Second
)

type RouterOSIPClient struct {
	Addr          string
	Username      string
	Password      string
	InterfaceName string
	Timeout       time.Duration
	Dial          func(network string, address string) (net.Conn, error)
}

type routerOSReply struct {
	kind  string
	attrs map[string]string
}

type routerOSAPIClient struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewRouterOSIPClient(addr string, username string, password string, interfaceName string) *RouterOSIPClient {
	return &RouterOSIPClient{
		Addr:          normalizeRouterOSAddr(addr),
		Username:      username,
		Password:      password,
		InterfaceName: interfaceName,
		Timeout:       defaultRouterOSTimeout,
	}
}

func normalizeRouterOSAddr(addr string) string {
	if addr == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	if strings.Count(addr, ":") > 1 && !strings.HasPrefix(addr, "[") {
		return net.JoinHostPort(addr, defaultRouterOSPort)
	}
	if strings.HasPrefix(addr, "[") && strings.Contains(addr, "]:") {
		return addr
	}
	return net.JoinHostPort(addr, defaultRouterOSPort)
}

func (cli *RouterOSIPClient) GetOuterIP() (string, error) {
	if cli.Addr == "" {
		return "", fmt.Errorf("routeros addr is empty")
	}
	if cli.Username == "" {
		return "", fmt.Errorf("routeros username is empty")
	}
	if cli.Password == "" {
		return "", fmt.Errorf("routeros password is empty")
	}
	if cli.InterfaceName == "" {
		return "", fmt.Errorf("routeros interface is empty")
	}

	conn, err := cli.dial("tcp", normalizeRouterOSAddr(cli.Addr))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	api := &routerOSAPIClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
	if err := api.login(cli.Username, cli.Password); err != nil {
		return "", err
	}

	replies, err := api.run("/ip/address/print", "=.proplist=address,interface", "?interface="+cli.InterfaceName)
	if err != nil {
		return "", err
	}
	for _, reply := range replies {
		if reply.attrs["interface"] != cli.InterfaceName {
			continue
		}
		address := reply.attrs["address"]
		if address == "" {
			continue
		}
		return trimCIDR(address), nil
	}

	return "", fmt.Errorf("can not found address for interface %s", cli.InterfaceName)
}

func (cli *RouterOSIPClient) dial(network string, address string) (net.Conn, error) {
	if cli.Dial != nil {
		return cli.Dial(network, address)
	}
	timeout := cli.Timeout
	if timeout <= 0 {
		timeout = defaultRouterOSTimeout
	}
	return net.DialTimeout(network, address, timeout)
}

func trimCIDR(address string) string {
	parts := strings.SplitN(address, "/", 2)
	return parts[0]
}

func (cli *routerOSAPIClient) login(username string, password string) error {
	replies, err := cli.run("/login", "=name="+username, "=password="+password)
	if err != nil {
		return err
	}
	for _, reply := range replies {
		ret := reply.attrs["ret"]
		if ret == "" {
			continue
		}
		response, err := routerOSChallengeResponse(password, ret)
		if err != nil {
			return err
		}
		_, err = cli.run("/login", "=name="+username, "=response="+response)
		return err
	}
	return nil
}

func routerOSChallengeResponse(password string, challenge string) (string, error) {
	raw, err := hex.DecodeString(challenge)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	hash.Write([]byte{0})
	hash.Write([]byte(password))
	hash.Write(raw)
	return "00" + hex.EncodeToString(hash.Sum(nil)), nil
}

func (cli *routerOSAPIClient) run(words ...string) ([]routerOSReply, error) {
	if err := cli.writeSentence(words...); err != nil {
		return nil, err
	}

	var replies []routerOSReply
	var trapErr error
	for {
		sentence, err := cli.readSentence()
		if err != nil {
			return nil, err
		}
		if len(sentence) == 0 {
			continue
		}
		reply := parseRouterOSReply(sentence)
		switch reply.kind {
		case "!re":
			replies = append(replies, reply)
		case "!trap", "!fatal":
			trapErr = fmt.Errorf(routerOSErrorMessage(reply.attrs))
		case "!done":
			if trapErr != nil {
				return nil, trapErr
			}
			if len(reply.attrs) > 0 {
				replies = append(replies, reply)
			}
			return replies, nil
		}
	}
}

func routerOSErrorMessage(attrs map[string]string) string {
	for _, key := range []string{"message", "category"} {
		if attrs[key] != "" {
			return attrs[key]
		}
	}
	return "routeros api error"
}

func parseRouterOSReply(sentence []string) routerOSReply {
	reply := routerOSReply{
		kind:  sentence[0],
		attrs: make(map[string]string),
	}
	for _, word := range sentence[1:] {
		if strings.HasPrefix(word, "=") {
			parts := strings.SplitN(word[1:], "=", 2)
			if len(parts) == 2 {
				reply.attrs[parts[0]] = parts[1]
			}
		}
	}
	return reply
}

func (cli *routerOSAPIClient) writeSentence(words ...string) error {
	for _, word := range words {
		if err := writeRouterOSWord(cli.writer, word); err != nil {
			return err
		}
	}
	if err := writeRouterOSLength(cli.writer, 0); err != nil {
		return err
	}
	return cli.writer.Flush()
}

func (cli *routerOSAPIClient) readSentence() ([]string, error) {
	var words []string
	for {
		word, err := readRouterOSWord(cli.reader)
		if err != nil {
			return nil, err
		}
		if word == "" {
			return words, nil
		}
		words = append(words, word)
	}
}

func writeRouterOSWord(writer *bufio.Writer, word string) error {
	if err := writeRouterOSLength(writer, len(word)); err != nil {
		return err
	}
	_, err := writer.WriteString(word)
	return err
}

func readRouterOSWord(reader *bufio.Reader) (string, error) {
	length, err := readRouterOSLength(reader)
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	_, err = reader.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeRouterOSLength(writer *bufio.Writer, length int) error {
	buf := encodeRouterOSLength(length)
	_, err := writer.Write(buf)
	return err
}

func encodeRouterOSLength(length int) []byte {
	switch {
	case length < 0x80:
		return []byte{byte(length)}
	case length < 0x4000:
		length |= 0x8000
		return []byte{byte(length >> 8), byte(length)}
	case length < 0x200000:
		length |= 0xC00000
		return []byte{byte(length >> 16), byte(length >> 8), byte(length)}
	case length < 0x10000000:
		length |= 0xE0000000
		return []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
	default:
		return []byte{0xF0, byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
	}
}

func readRouterOSLength(reader *bufio.Reader) (int, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	switch {
	case first&0x80 == 0x00:
		return int(first), nil
	case first&0xC0 == 0x80:
		second, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(uint16(first&^0xC0)<<8 | uint16(second)), nil
	case first&0xE0 == 0xC0:
		buf := make([]byte, 2)
		_, err := reader.Read(buf)
		if err != nil {
			return 0, err
		}
		return int(uint32(first&^0xE0)<<16 | uint32(buf[0])<<8 | uint32(buf[1])), nil
	case first&0xF0 == 0xE0:
		buf := make([]byte, 3)
		_, err := reader.Read(buf)
		if err != nil {
			return 0, err
		}
		return int(uint32(first&^0xF0)<<24 | uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])), nil
	case first == 0xF0:
		buf := make([]byte, 4)
		_, err := reader.Read(buf)
		if err != nil {
			return 0, err
		}
		return int(uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])), nil
	default:
		return 0, fmt.Errorf("invalid routeros length prefix: %d", first)
	}
}
