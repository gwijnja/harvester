package ftp

import (
	"fmt"
	"log/slog"

	"github.com/jlaffaye/ftp"
)

type FtpConnector struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (c *FtpConnector) connect() (*ftp.ServerConn, error) {
	slog.Info("Connecting to FTP server", slog.String("host", c.Host), slog.Int("port", c.Port))
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, fmt.Errorf("error dialing %s:%d: %s", c.Host, c.Port, err)
	}

	slog.Info("Logging in", slog.String("user", c.Username))
	err = conn.Login(c.Username, c.Password)
	if err != nil {
		return nil, fmt.Errorf("error logging in as %s: %s", c.Username, err)
	}

	return conn, nil
}
