// Copyright 2019 The age Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nknwallet

import (
	"bytes"
	"errors"
	"fmt"

	"filippo.io/age"
)

// LazyScryptIdentity is an age.Identity that requests a passphrase only if it
// encounters an scrypt stanza. After obtaining a passphrase, it delegates to
// ScryptIdentity.
type LazyScryptIdentity struct {
	Passphrase func() (string, error)
}

var _ age.Identity = &LazyScryptIdentity{}

func (i *LazyScryptIdentity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
	if len(stanzas) != 1 || stanzas[0].Type != "scrypt" {
		return nil, age.ErrIncorrectIdentity
	}
	pass, err := i.Passphrase()
	if err != nil {
		return nil, fmt.Errorf("could not read passphrase: %v", err)
	}
	ii, err := age.NewScryptIdentity(pass)
	if err != nil {
		return nil, err
	}
	fileKey, err = ii.Unwrap(stanzas)
	if errors.Is(err, age.ErrIncorrectIdentity) {
		return nil, fmt.Errorf("incorrect passphrase")
	}
	return fileKey, err
}

type EncryptedIdentity struct {
	Contents       []byte
	Passphrase     func() (string, error)
	NoMatchWarning func()

	identities []age.Identity
}

var _ age.Identity = &EncryptedIdentity{}

func (i *EncryptedIdentity) Recipients() ([]age.Recipient, error) {
	if len(i.identities) == 0 {
		if err := i.decrypt(); err != nil {
			return nil, err
		}
	}

	return identitiesToRecipients(i.identities)
}

func (i *EncryptedIdentity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
	if len(i.identities) == 0 {
		if err := i.decrypt(); err != nil {
			return nil, err
		}
	}

	for _, id := range i.identities {
		fileKey, err = id.Unwrap(stanzas)
		if errors.Is(err, age.ErrIncorrectIdentity) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return fileKey, nil
	}
	i.NoMatchWarning()
	return nil, age.ErrIncorrectIdentity
}

func (i *EncryptedIdentity) decrypt() error {
	d, err := age.Decrypt(bytes.NewReader(i.Contents), &LazyScryptIdentity{i.Passphrase})
	if e := new(age.NoIdentityMatchError); errors.As(err, &e) {
		return fmt.Errorf("identity file is encrypted with age but not with a passphrase")
	}
	if err != nil {
		return fmt.Errorf("failed to decrypt identity file: %v", err)
	}
	i.identities, err = age.ParseIdentities(d)
	return err
}