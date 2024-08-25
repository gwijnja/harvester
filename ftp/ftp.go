package ftp

import (
	"fmt"
	"log/slog"

	"github.com/jlaffaye/ftp"
)

type Connector struct {
	Host     string
	Port     int
	Username string
	Password string
}

// connect connects to the FTP server
func (c *Connector) connect() (*ftp.ServerConn, error) {

	// Dial
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, fmt.Errorf("ftp: Failed to dial %s:%d: %s", c.Host, c.Port, err)
	}
	slog.Info("ftp: Connected", slog.String("host", c.Host), slog.Int("port", c.Port))

	// Login
	err = conn.Login(c.Username, c.Password)
	if err != nil {
		return nil, fmt.Errorf("ftp: Failed to login as %s: %s", c.Username, err)
	}
	slog.Info("ftp: Logged in", slog.String("username", c.Username))

	return conn, nil
}
