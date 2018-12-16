/*
	smtp-dkim-signer - SMTP-proxy that DKIM-signs e-mails before submission.
	Copyright (C) 2018  Marc Hoersken <info@marc-hoersken.de>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	dkim "github.com/emersion/go-dkim"
	smtp "github.com/emersion/go-smtp"
)

func readFile(filepath string) (*bytes.Buffer, error) {
	var b bytes.Buffer
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	return &b, err
}

func loadPrivKey(privkeypath string) (*rsa.PrivateKey, error) {
	b, err := readFile(privkeypath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b.Bytes())
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, err
}

func makeOptions(cfg *config, cfgvh *configVHost) (*dkim.SignOptions, error) {
	if cfg == nil || cfgvh == nil {
		return nil, fmt.Errorf("this should never happen")
	}
	if cfgvh.Domain == "" {
		return nil, fmt.Errorf("no VirtualHost.Domain specified")
	}
	if cfgvh.Selector == "" {
		return nil, fmt.Errorf("no VirtualHost.Selector specified")
	}
	if cfgvh.PrivKeyPath == "" {
		return nil, fmt.Errorf("no VirtualHost.PrivKeyPath specified")
	}

	if len(cfgvh.HeaderKeys) == 0 {
		cfgvh.HeaderKeys = cfg.HeaderKeys
	}

	privkey, err := loadPrivKey(cfgvh.PrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load VirtualHost.PrivKeyPath due to: %s", err)
	}

	dkimopt := &dkim.SignOptions{
		Domain:                 cfgvh.Domain,
		Selector:               cfgvh.Selector,
		Signer:                 privkey,
		Hash:                   crypto.SHA256,
		HeaderCanonicalization: cfgvh.HeaderCan,
		BodyCanonicalization:   cfgvh.BodyCan,
		HeaderKeys:             cfgvh.HeaderKeys,
	}
	return dkimopt, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	be, err := makeBackend(cfg)
	if err != nil {
		log.Fatal(err)
	}

	s := smtp.NewServer(be)
	s.Addr = cfg.Address
	s.Domain = cfg.Domain
	s.MaxIdleSeconds = cfg.MaxIdleSeconds
	s.MaxMessageBytes = cfg.MaxMessageBytes
	s.MaxRecipients = cfg.MaxRecipients
	s.AllowInsecureAuth = true

	log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
