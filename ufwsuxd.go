package main

import (
	"EX0l0N/ufwsux/v2/netio"
	"EX0l0N/ufwsux/v2/tokens"

	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

const listenAddr = "127.0.0.1:1199"
const protoName = "ufwsux"

func main() {
	http.HandleFunc("/", tunnelHandler)
	srv := &http.Server{
		Addr: listenAddr,
	}
	log.Printf("Tunnel server listening on %s", listenAddr)
	log.Fatal(srv.ListenAndServe())
}

func tunnelHandler(w http.ResponseWriter, r *http.Request) {
	// validate Upgrade
	if r.Header.Get("Connection") != "Upgrade" || r.Header.Get("Upgrade") != protoName {
		http.Error(w, "upgrade required", http.StatusBadRequest)
		return
	}

	targetHost := r.Header.Get("X-SSH-Host")
	targetPort := r.Header.Get("X-SSH-Port")
	token := r.Header.Get("X-SSH-Auth")
	if targetHost == "" || targetPort == "" || token == "" {
		http.Error(w, "missing headers", http.StatusBadRequest)
		return
	}

	if !tokens.ValidateToken(token, targetHost, targetPort, time.Now()) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// hijack connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack unsupported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Printf("hijack error: %v", err)
		return
	}
	defer clientConn.Close()

	_, _ = clientConn.Write([]byte(fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\nUpgrade: %s\r\nConnection: Upgrade\r\n\r\n",
		protoName,
	)))

	remoteAddr := net.JoinHostPort(targetHost, targetPort)
	remote, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("dial %s failed: %v", remoteAddr, err)
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n"))
		return
	}
	defer remote.Close()

	log.Printf("Tunnel established: %v -> %v", clientConn.RemoteAddr(), remoteAddr)
	netio.BiCopy(clientConn, remote)
	log.Printf("Tunnel closed: %v", clientConn.RemoteAddr())
}
