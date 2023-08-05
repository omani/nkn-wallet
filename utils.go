package nknwallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/nknorg/nkn-sdk-go"
	"github.com/nknorg/nkn/v2/crypto"
	"github.com/nknorg/nkn/v2/util/password"
	"github.com/nknorg/nkn/v2/vault"
)

func checkerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type PasswordScryptIdentity struct {
	password string
}

var _ age.Identity = &PasswordScryptIdentity{}

// Unwrap unwraps given password and delegates it to NewScriptIdentity()
func (i *PasswordScryptIdentity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
	for _, s := range stanzas {
		if s.Type == "scrypt" && len(stanzas) != 1 {
			return nil, errors.New("an scrypt recipient must be the only one")
		}
	}
	if len(stanzas) != 1 || stanzas[0].Type != "scrypt" {
		return nil, age.ErrIncorrectIdentity
	}
	ii, err := age.NewScryptIdentity(i.password)
	if err != nil {
		return nil, err
	}
	fileKey, err = ii.Unwrap(stanzas)
	if errors.Is(err, age.ErrIncorrectIdentity) {
		return nil, errors.New("incorrect passphrase")
	}
	return fileKey, err
}

func encrypt(recipients []age.Recipient, in io.Reader, out io.Writer) error {
	a := armor.NewWriter(out)
	defer a.Close()
	out = a
	w, err := age.Encrypt(out, recipients...)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func encryptAccount(account *nkn.Account, password []byte) ([]byte, error) {
	r, err := age.NewScryptRecipient(string(password))
	if err != nil {
		return nil, err
	}

	in := new(bytes.Buffer)
	err = json.NewEncoder(in).Encode(account)
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	encrypt([]age.Recipient{r}, in, out)

	return out.Bytes(), nil
}

func decryptAccount(walletfile *Wallet, password []byte) (*nkn.Account, error) {
	identities := []age.Identity{
		&PasswordScryptIdentity{string(password)},
	}

	in := bytes.NewBuffer([]byte(walletfile.Armor))
	r, err := age.Decrypt(armor.NewReader(in), identities...)
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	account := &vault.Account{}
	if _, err := io.Copy(out, r); err != nil {
		return nil, err
	}

	err = json.NewDecoder(out).Decode(account)
	if err != nil {
		return nil, err
	}
	seed := crypto.GetSeedFromPrivateKey(account.PrivateKey)
	_ = seed

	return nkn.NewAccount(seed)
}

func getConfirmedPassword(passwd string) []byte {
	var tmp []byte
	var err error
	if passwd != "" {
		tmp = []byte(passwd)
	} else {
		tmp, err = password.GetConfirmedPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	return tmp
}
