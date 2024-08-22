package ftp

import (
	"fmt"
	"log/slog"

	"github.com/jlaffaye/ftp"
)

type FtpConnector struct {
	Host       string
	Port       int
	Username   string
	Password   string
	connection *ftp.ServerConn
}

func (c *FtpConnector) connect() error {
	var err error
	slog.Info("Connecting to FTP server", slog.String("host", c.Host), slog.Int("port", c.Port))
	c.connection, err = ftp.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return fmt.Errorf("error dialing %s:%d: %s", c.Host, c.Port, err)
	}

	slog.Info("Logging in", slog.String("user", c.Username))
	err = c.connection.Login(c.Username, c.Password)
	if err != nil {
		return fmt.Errorf("error logging in as %s: %s", c.Username, err)
	}

	slog.Info("Connected to FTP server", slog.String("host", c.Host), slog.Int("port", c.Port), slog.String("user", c.Username))
	return nil
}
