package dot

import (
	"crypto/tls"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
)

type trackedConn struct {
	net.Conn
	serverName  string
	cleanupFunc func()
}

func (c *trackedConn) Close() error {
	c.cleanupFunc()
	return c.Conn.Close()
}

type Listener struct {
	net.Listener
	connections sync.Map
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		slog.Error("error while accepting connection")
		return nil, err
	}

	tlsConn := conn.(*tls.Conn)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		slog.Error("error during tls handshake")
		return nil, err
	}

	state := tlsConn.ConnectionState()
	serverName := state.ServerName
	key := conn.RemoteAddr().String()

	wrapped := &trackedConn{
		Conn:       conn,
		serverName: serverName,
		cleanupFunc: func() {
			l.connections.Delete(key)
		},
	}
	hostname := "." + os.Getenv("hostname")
	serverName = strings.Replace(serverName, hostname, "", -1)
	l.connections.Store(key, serverName)
	return wrapped, nil
}

func (l *Listener) GetServerName(remoteAddr string) (string, bool) {
	val, ok := l.connections.Load(remoteAddr)
	if !ok {
		return "", false
	}
	return val.(string), true
}
