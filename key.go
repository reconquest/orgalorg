package main

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ScaleFT/sshkeys"
	"github.com/reconquest/hierr-go"
	"github.com/youmark/pkcs8"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

type sshKey struct {
	raw        []byte
	block      *pem.Block
	extra      []byte
	private    interface{}
	passphrase []byte
}

func (key *sshKey) validate() error {
	if len(key.extra) != 0 {
		return hierr.Errorf(
			errors.New(string(key.extra)),
			`extra data found in the SSH key`,
		)
	}

	return nil
}

func (key *sshKey) isOpenSSH() bool {
	return key.block.Type == "OPENSSH PRIVATE KEY"
}

func (key *sshKey) isPKCS8() bool {
	return key.block.Type == "ENCRYPTED PRIVATE KEY" ||
		key.block.Type == "PRIVATE KEY"
}

func (key *sshKey) isEncrypted() bool {
	if key.block.Type == "ENCRYPTED PRIVATE KEY" {
		return true
	}

	if strings.Contains(key.block.Headers["Proc-Type"], "ENCRYPTED") {
		return true
	}

	if key.isOpenSSH() {
		_, err := ssh.ParseRawPrivateKey([]byte(key.raw))
		return err != nil
	}

	return false
}

func (key *sshKey) parse() error {
	var err error
	switch {
	case key.isOpenSSH() && key.isEncrypted():
		key.private, err = sshkeys.ParseEncryptedRawPrivateKey(
			[]byte(key.raw),
			key.passphrase,
		)

	case key.isPKCS8() && key.isEncrypted():
		key.private, err = pkcs8.ParsePKCS8PrivateKey(
			key.block.Bytes,
			key.passphrase,
		)

	case key.isPKCS8():
		key.private, err = pkcs8.ParsePKCS8PrivateKey(
			key.block.Bytes,
			nil,
		)

	case key.isEncrypted():
		key.private, err = ssh.ParseRawPrivateKeyWithPassphrase(
			[]byte(key.raw),
			key.passphrase,
		)

	default:
		key.private, err = ssh.ParseRawPrivateKey(
			[]byte(key.raw),
		)
	}
	return err
}

func readSSHKey(keyring agent.Agent, path string) error {
	var key sshKey
	var err error

	key.raw, err = ioutil.ReadFile(path)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't read SSH key from file`,
		)
	}

	key.block, key.extra = pem.Decode(key.raw)

	err = key.validate()
	if err != nil {
		return err
	}

	if key.isEncrypted() {
		key.passphrase, err = readPassword(sshPassphrasePrompt)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't read key passphrase`,
			)
		}
	}

	err = key.parse()
	if err != nil {
		if err == sshkeys.ErrIncorrectPassword {
			err = errors.New("invalid passphrase for private key specified")
		}

		return hierr.Errorf(
			err,
			"unable to parse ssh key",
		)
	}

	keyring.Add(agent.AddedKey{
		PrivateKey: key.private,
		Comment:    "passed by orgalorg",
	})

	return nil
}

func readPassword(prompt string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, prompt)

	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`TTY is required for reading password, `+
				`but /dev/tty can't be opened`,
		)
	}

	password, err := terminal.ReadPassword(int(tty.Fd()))
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't read password`,
		)
	}

	if prompt != "" {
		fmt.Fprintln(os.Stderr)
	}

	return password, nil
}
