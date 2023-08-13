package nknwallet

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

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

// crlfMangledIntro and utf16MangledIntro are the intro lines of the age format
// after mangling by various versions of PowerShell redirection, truncated to
// the length of the correct intro line. See issue 290.
const crlfMangledIntro = "age-encryption.org/v1" + "\r"
const utf16MangledIntro = "\xff\xfe" + "a\x00g\x00e\x00-\x00e\x00n\x00c\x00r\x00y\x00p\x00"

type rejectScryptIdentity struct{}

func (rejectScryptIdentity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	if len(stanzas) != 1 || stanzas[0].Type != "scrypt" {
		return nil, age.ErrIncorrectIdentity
	}
	return nil, fmt.Errorf("file is passphrase-encrypted but identities were specified with -i/--identity or -j",
		"remove all -i/--identity/-j flags to decrypt passphrase-encrypted files")
}

type PasswordEncryptScryptIdentity struct {
	password string
}

var _ age.Identity = &PasswordEncryptScryptIdentity{}

// Unwrap unwraps given password and delegates it to NewScriptIdentity()
func (i *PasswordEncryptScryptIdentity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
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

//	func encryptAccount(account *nkn.Account, password []byte) ([]byte, error) {
//		r, err := age.NewScryptRecipient(string(password))
//		if err != nil {
//			return nil, err
//		}
//
//		in := new(bytes.Buffer)
//		err = json.NewEncoder(in).Encode(account)
//		if err != nil {
//			return nil, err
//		}
//		out := &bytes.Buffer{}
//		encrypt([]age.Recipient{r}, in, out)
//
//		return out.Bytes(), nil
//	}
func decrypt(identities []age.Identity, in io.Reader, out io.Writer) error {
	rr := bufio.NewReader(in)
	if intro, _ := rr.Peek(len(crlfMangledIntro)); string(intro) == crlfMangledIntro ||
		string(intro) == utf16MangledIntro {
		return errors.New("invalid header intro: it looks like this file was corrupted by PowerShell redirection. consider using -o or -a to encrypt files in PowerShell")
	}

	if start, _ := rr.Peek(len(armor.Header)); string(start) == armor.Header {
		in = armor.NewReader(rr)
	} else {
		in = rr
	}

	r, err := age.Decrypt(in, identities...)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func decryptAccountByPassword(walletfile *Wallet) (*nkn.Account, error) {
	in := bytes.NewBuffer([]byte(walletfile.Armor))
	out := &bytes.Buffer{}

	identities := []age.Identity{
		&LazyScryptIdentity{passphrasePromptForDecryption},
	}

	err := decrypt(identities, in, out)
	if err != nil {
		return nil, err
	}

	account := &vault.Account{}
	err = json.NewDecoder(out).Decode(account)
	if err != nil {
		return nil, err
	}
	seed := crypto.GetSeedFromPrivateKey(account.PrivateKey)

	return nkn.NewAccount(seed)
}

func decryptAccountByIdentityFile(walletfile *Wallet, identity string) (*nkn.Account, error) {
	in := bytes.NewBuffer([]byte(walletfile.Armor))
	out := &bytes.Buffer{}

	identities := []age.Identity{rejectScryptIdentity{}}
	ids, err := parseIdentitiesFile(identity)
	if err != nil {
		return nil, err
	}
	identities = append(identities, ids...)

	err = decrypt(identities, in, out)
	if err != nil {
		return nil, err
	}

	account := &vault.Account{}
	err = json.NewDecoder(out).Decode(account)
	if err != nil {
		return nil, err
	}
	seed := crypto.GetSeedFromPrivateKey(account.PrivateKey)

	return nkn.NewAccount(seed)
}

// func decryptAccountByPassword(walletfile *Wallet, password []byte) (*nkn.Account, error) {
// 	identities := []age.Identity{
// 		&PasswordEncryptScryptIdentity{string(password)},
// 	}
//
// 	in := bytes.NewBuffer([]byte(walletfile.Armor))
// 	out := &bytes.Buffer{}
//
// 	err := decrypt(identities, armor.NewReader(in), out)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	account := &vault.Account{}
// 	err = json.NewDecoder(out).Decode(account)
// 	if err != nil {
// 		return nil, err
// 	}
// 	seed := crypto.GetSeedFromPrivateKey(account.PrivateKey)
//
// 	return nkn.NewAccount(seed)
// }

func passphrasePromptForDecryption() (string, error) {
	pass, err := password.GetPassword("Enter passphrase")
	if err != nil {
		return "", fmt.Errorf("could not read passphrase: %v", err)
	}
	return string(pass), nil
}

func passphrasePromptForEncryption() (string, error) {
	pass, err := password.GetPassword("Enter passphrase (leave empty to autogenerate a secure one)")
	if err != nil {
		return "", fmt.Errorf("Could not read passphrase: %v", err)
	}
	p := string(pass)
	if p == "" {
		var words []string
		for i := 0; i < 10; i++ {
			words = append(words, randomWord())
		}
		p = strings.Join(words, "-")
		_, err := fmt.Printf("Using autogenerated passphrase %q\n", p)
		if err != nil {
			return "", fmt.Errorf("Could not print passphrase: %v", err)
		}
	} else {
		confirm, err := password.GetPassword("Confirm passphrase")
		if err != nil {
			return "", fmt.Errorf("Could not read passphrase: %v", err)
		}
		if string(confirm) != p {
			return "", fmt.Errorf("Passphrases didn't match")
		}
	}
	return p, nil
}

func encryptAccount(account *nkn.Account, recipients []age.Recipient) ([]byte, error) {
	in := new(bytes.Buffer)
	err := json.NewEncoder(in).Encode(account)
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}

	err = encrypt(recipients, in, out)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
