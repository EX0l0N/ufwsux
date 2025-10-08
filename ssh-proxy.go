package main

import (
	"EX0l0N/ufwsux/v2/netio"
	"EX0l0N/ufwsux/v2/tokens"

	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const protoName = "ufwsux"
const httpPort = "80"

type stdioRW struct {
	in  *os.File
	out *os.File
}

func (s *stdioRW) Read(p []byte) (int, error)  { return s.in.Read(p) }
func (s *stdioRW) Write(p []byte) (int, error) { return s.out.Write(p) }
func (s *stdioRW) Close() error                { return nil }

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <http-tunnel-host> <ssh-host> <ssh-port>", os.Args[0])
	}
	httpHost := os.Args[1]
	sshHost := os.Args[2]
	sshPort := os.Args[3]

	addr := net.JoinHostPort(httpHost, httpPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("dial %s failed: %v", addr, err)
	}
	defer conn.Close()

	token := tokens.GenerateToken(sshHost, sshPort, time.Now())
	reqHeaders := []string{
		fmt.Sprintf("GET /ufwsux HTTP/1.1"),
		fmt.Sprintf("Host: %s", httpHost),
		"Connection: Upgrade",
		fmt.Sprintf("Upgrade: %s", protoName),
		fmt.Sprintf("X-SSH-Host: %s", sshHost),
		fmt.Sprintf("X-SSH-Port: %s", sshPort),
		fmt.Sprintf("X-SSH-Auth: %s", token),
		"", "",
	}
	req := strings.Join(reqHeaders, "\r\n")
	if _, err := conn.Write([]byte(req)); err != nil {
		log.Fatalf("write request failed: %v", err)
	}

	br := bufio.NewReader(conn)
	var statusLine string
	statusLine, err = br.ReadString('\n')
	if err != nil {
		log.Fatalf("reading status line failed: %v", err)
	}
	statusLine = strings.TrimRight(statusLine, "\r\n")
	if !strings.HasPrefix(statusLine, "HTTP/1.1 101") {
		rest, _ := br.ReadString('\n')
		log.Fatalf("upgrade failed: %s -- %s", statusLine, rest)
	}
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			log.Fatalf("reading response headers failed: %v", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	remaining := br.Buffered()
	var pre []byte
	if remaining > 0 {
		pre = make([]byte, remaining)
		_, _ = br.Read(pre)
	}

	if len(pre) > 0 {
		if _, err := os.Stdout.Write(pre); err != nil {
			log.Printf("warning: writing leftover bytes to stdout: %v", err)
		}
	}

	log.Println("Upgrade successful, starting bidirectional copy")

	std := &stdioRW{in: os.Stdin, out: os.Stdout}

	netio.BiCopy(conn, std)

	log.Println("connection closed, exiting")
}
