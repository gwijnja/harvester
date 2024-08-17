package sftp

import (
	"regexp"

	"github.com/pkg/sftp"
)

// Connector is a structure that holds the configuration for an SFTP connection.
type Connector struct {
	Name                  string
	IntervalSec           int
	Host                  string
	Port                  int
	FailIfHostKeyChanged  bool
	FailIfHostKeyNotFound bool
	Username              string
	Password              string
	PrivateKeyFile        string
	Passphrase            string
	Root                  string
	ToLoad                string
	Loaded                string
	Regex                 *regexp.Regexp
	Delete                bool
	client                *sftp.Client
}
