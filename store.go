package nknwallet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nknorg/nkn-sdk-go"
	"github.com/nknorg/nkn/v2/common"
	"github.com/nknorg/nkn/v2/pb"
	"github.com/nknorg/nkn/v2/program"
	"github.com/nknorg/nkn/v2/signature"
	"github.com/nknorg/nkn/v2/transaction"
)

type Wallet struct {
	ID         int    `json:"id"`
	NKNAddress string `json:"address"`
	Armor      string `json:"armor"`
	Alias      string `json:"alias,omitempty"`

	config  *nkn.WalletConfig
	lock    sync.Mutex
	account *nkn.Account
}

type Store struct {
	wallets []*Wallet
	path    string
}

func NewStore(path string) *Store {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	checkerr(err)
	defer f.Close()

	dat, err := io.ReadAll(f)
	checkerr(err)

	var wallets []*Wallet
	if len(dat) > 0 {
		err = json.Unmarshal(dat, &wallets)
		checkerr(err)
	}

	return &Store{
		wallets: wallets,
		path:    path,
	}
}

func (s *Store) GetWallets() []*Wallet {
	return s.wallets
}

func (s *Store) GetWalletWithAlias(alias string) (*Wallet, error) {
	if len(alias) == 0 {
		return nil, errors.New("Alias is empty")
	}
	for _, w := range s.wallets {
		if w.Alias == alias {
			return w, nil
		}
	}
	return nil, errors.New("Wallet not found")
}

func (s *Store) GetWalletWithIndex(index int) (*Wallet, error) {
	if index == 0 {
		return nil, errors.New("Index is 0")
	}
	for _, w := range s.wallets {
		if w.ID == index {
			return w, nil
		}
	}
	return nil, errors.New("Wallet not found")
}

func (s *Store) RestoreFromSeed(seed, newpassword []byte, alias string) (*Wallet, error) {
	account, err := nkn.NewAccount(seed)
	if err != nil {
		return nil, err
	}
	armor, err := encryptAccount(account, newpassword)
	if err != nil {
		return nil, err
	}

	config, err := nkn.MergeWalletConfig(nil)
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		ID:         s.getNextID(),
		Armor:      string(armor),
		Alias:      alias,
		account:    account,
		NKNAddress: account.WalletAddress(),
		config:     config,
	}
	return w, nil

}

func (s *Store) ListWallets() error {
	if s.wallets == nil {
		return errors.New("No wallet found or wallet has no accounts.")
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = false
	mw := io.MultiWriter(os.Stdout)
	t.SetOutputMirror(mw)
	t.AppendHeader(table.Row{"ID", "Alias", "Address"})

	for _, w := range s.wallets {
		t.AppendRow(table.Row{w.ID, w.Alias, w.Address()})
	}
	t.Render()

	return nil
}

func (s *Store) save() {
	dat, err := json.Marshal(s.wallets)
	if err != nil {
		checkerr(err)
	}
	err = ioutil.WriteFile(s.path, dat, 0644)
	checkerr(err)
}

func (s *Store) DeleteWalletByAlias(alias string) {
	var newwallets []*Wallet
	for _, w := range s.wallets {
		if w.Alias != alias {
			newwallets = append(newwallets, w)
		}
	}
	s.wallets = newwallets
	s.save()
}

func (s *Store) DeleteWalletByIndex(index int) {
	var newwallets []*Wallet
	for _, w := range s.wallets {
		if w.ID != index {
			newwallets = append(newwallets, w)
		}
	}
	s.wallets = newwallets
	s.save()
}

func (s *Store) GetWalletByIndex(index int, password []byte) (*Wallet, error) {
	for _, w := range s.wallets {
		if w.ID == index {
			account, err := decryptAccount(w, password)
			if err != nil {
				return nil, err
			}

			config, err := nkn.MergeWalletConfig(nil)
			if err != nil {
				return nil, err
			}

			w.account = account
			w.config = config
			return w, nil
		}
	}
	return nil, errors.New("Wallet not found.")
}

func (s *Store) GetWalletByAlias(alias string, password []byte) (*Wallet, error) {
	for _, w := range s.wallets {
		if w.Alias == alias {
			account, err := decryptAccount(w, password)
			if err != nil {
				return nil, err
			}

			config, err := nkn.MergeWalletConfig(nil)
			if err != nil {
				return nil, err
			}

			w.account = account
			w.config = config
			return w, nil
		}
	}
	return nil, errors.New("Wallet not found.")
}

func (s *Store) getNextID() int {
	if s.wallets == nil {
		return 1
	}
	return s.wallets[len(s.wallets)-1].ID + 1
}

func (s *Store) NewWallet(password []byte, alias string, config *nkn.WalletConfig) (*Wallet, error) {
	account, err := nkn.NewAccount(nil)
	if err != nil {
		return nil, err
	}

	armor, err := encryptAccount(account, password)
	if err != nil {
		return nil, err
	}

	config, err = nkn.MergeWalletConfig(config)
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		ID:         s.getNextID(),
		Armor:      string(armor),
		Alias:      alias,
		account:    account,
		NKNAddress: account.WalletAddress(),
		config:     config,
	}
	return w, nil
}

func (s *Store) SetName(index int, alias string) {
	for i, w := range s.wallets {
		if w.ID == index {
			s.wallets[i].Alias = alias
		}
	}
}

func (s *Store) SaveWallet(wallet *Wallet) error {
	for _, w := range s.wallets {
		if w.Address() == wallet.Address() {
			return errors.New("Account already exists in store.")
		}
	}
	s.wallets = append(s.wallets, wallet)
	s.save()

	return nil
}

func (w *Wallet) OpenAPI() *Openapi {
	return newOpenAPI(w)
}

// Account returns the account of the wallet.
func (w *Wallet) Account() *nkn.Account {
	return w.account
}

// Seed returns the secret seed of the wallet. Secret seed can be used to create
// client/wallet with the same key pair and should be kept secret and safe.
func (w *Wallet) Seed() []byte {
	return w.Account().Seed()
}

func (w *Wallet) ShowSeed() string {
	return hex.EncodeToString(w.Account().Seed())
}

// PubKey returns the public key of the wallet.
func (w *Wallet) PubKey() []byte {
	return w.Account().PubKey()
}

// Address returns the NKN wallet address of the wallet.
func (w *Wallet) Address() string {
	return w.NKNAddress
}

// VerifyPassword returns nil if provided password is the correct password of account
func (w *Wallet) VerifyPassword(password []byte) error {
	account, err := decryptAccount(w, password)
	if err != nil {
		return err
	}

	address, err := account.ProgramHash.ToAddress()
	if err != nil {
		return err
	}

	if address != w.Address() {
		return errors.New("wrong password")
	}

	return nil
}

// ProgramHash returns the program hash of this wallet's account.
func (w *Wallet) ProgramHash() common.Uint160 {
	return w.Account().ProgramHash
}

// SignTransaction signs an unsigned transaction using this wallet's key pair.
func (w *Wallet) SignTransaction(tx *transaction.Transaction) error {
	ct, err := program.CreateSignatureProgramContext(w.Account().PublicKey)
	if err != nil {
		return err
	}

	sig, err := signature.SignBySigner(tx, w.Account().Account)
	if err != nil {
		return err
	}

	tx.SetPrograms([]*pb.Program{ct.NewProgram(sig)})
	return nil
}

// NewNanoPay is a shortcut for NewNanoPay using this wallet as sender.
//
// Duration is changed to signed int for gomobile compatibility.
func (w *Wallet) NewNanoPay(recipientAddress, fee string, duration int) (*nkn.NanoPay, error) {
	nknwallet, err := nkn.NewWallet(w.Account(), w.config)
	if err != nil {
		return nil, err
	}
	return nkn.NewNanoPay(w, nknwallet, recipientAddress, fee, duration)
}

// NewNanoPayClaimer is a shortcut for NewNanoPayClaimer using this wallet as
// RPC client.
func (w *Wallet) NewNanoPayClaimer(recipientAddress string, claimIntervalMs, lingerMs int32, minFlushAmount string, onError *nkn.OnError) (*nkn.NanoPayClaimer, error) {
	if len(recipientAddress) == 0 {
		recipientAddress = w.Address()
	}
	return nkn.NewNanoPayClaimer(w, recipientAddress, claimIntervalMs, lingerMs, minFlushAmount, onError)
}

// GetNonce wraps GetNonceContext with background context.
func (w *Wallet) GetNonce(txPool bool) (int64, error) {
	return w.GetNonceContext(context.Background(), txPool)
}

// GetNonceContext is the same as package level GetNonceContext, but using this
// wallet's SeedRPCServerAddr.
func (w *Wallet) GetNonceContext(ctx context.Context, txPool bool) (int64, error) {
	return w.GetNonceByAddressContext(ctx, w.Address(), txPool)
}

// GetNonceByAddress wraps GetNonceByAddressContext with background context.
func (w *Wallet) GetNonceByAddress(address string, txPool bool) (int64, error) {
	return w.GetNonceByAddressContext(context.Background(), address, txPool)
}

// GetNonceByAddressContext is the same as package level GetNonceContext, but
// using this wallet's SeedRPCServerAddr.
func (w *Wallet) GetNonceByAddressContext(ctx context.Context, address string, txPool bool) (int64, error) {
	return nkn.GetNonceContext(ctx, address, txPool, w.config)
}

// GetHeight wraps GetHeightContext with background context.
func (w *Wallet) GetHeight() (int32, error) {
	return w.GetHeightContext(context.Background())
}

// GetHeightContext is the same as package level GetHeightContext, but using
// this wallet's SeedRPCServerAddr.
func (w *Wallet) GetHeightContext(ctx context.Context) (int32, error) {
	return nkn.GetHeightContext(ctx, w.config)
}

// Balance wraps BalanceContext with background context.
func (w *Wallet) Balance() (*nkn.Amount, error) {
	return w.BalanceContext(context.Background())
}

// BalanceContext is the same as package level GetBalanceContext, but using this
// wallet's SeedRPCServerAddr.
func (w *Wallet) BalanceContext(ctx context.Context) (*nkn.Amount, error) {
	return w.BalanceByAddressContext(ctx, w.Address())
}

// BalanceByAddress wraps BalanceByAddressContext with background context.
func (w *Wallet) BalanceByAddress(address string) (*nkn.Amount, error) {
	return w.BalanceByAddressContext(context.Background(), address)
}

// BalanceByAddressContext is the same as package level GetBalanceContext, but
// using this wallet's SeedRPCServerAddr.
func (w *Wallet) BalanceByAddressContext(ctx context.Context, address string) (*nkn.Amount, error) {
	return nkn.GetBalanceContext(ctx, address, w.config)
}

// GetSubscribers wraps GetSubscribersContext with background context.
func (w *Wallet) GetSubscribers(topic string, offset, limit int, meta, txPool bool, subscriberHashPrefix []byte) (*nkn.Subscribers, error) {
	return w.GetSubscribersContext(context.Background(), topic, offset, limit, meta, txPool, subscriberHashPrefix)
}

// GetSubscribersContext is the same as package level GetSubscribersContext, but
// using this wallet's SeedRPCServerAddr.
func (w *Wallet) GetSubscribersContext(ctx context.Context, topic string, offset, limit int, meta, txPool bool, subscriberHashPrefix []byte) (*nkn.Subscribers, error) {
	return nkn.GetSubscribersContext(ctx, topic, offset, limit, meta, txPool, subscriberHashPrefix, w.config)
}

// GetSubscription wraps GetSubscriptionContext with background context.
func (w *Wallet) GetSubscription(topic string, subscriber string) (*nkn.Subscription, error) {
	return w.GetSubscriptionContext(context.Background(), topic, subscriber)
}

// GetSubscriptionContext is the same as package level GetSubscriptionContext,
// but using this wallet's SeedRPCServerAddr.
func (w *Wallet) GetSubscriptionContext(ctx context.Context, topic string, subscriber string) (*nkn.Subscription, error) {
	return nkn.GetSubscriptionContext(ctx, topic, subscriber, w.config)
}

// GetSubscribersCount wraps GetSubscribersCountContext with background context.
func (w *Wallet) GetSubscribersCount(topic string, subscriberHashPrefix []byte) (int, error) {
	return w.GetSubscribersCountContext(context.Background(), topic, subscriberHashPrefix)
}

// GetSubscribersCountContext is the same as package level
// GetSubscribersCountContext, but this wallet's SeedRPCServerAddr.
func (w *Wallet) GetSubscribersCountContext(ctx context.Context, topic string, subscriberHashPrefix []byte) (int, error) {
	return nkn.GetSubscribersCountContext(ctx, topic, subscriberHashPrefix, w.config)
}

// GetRegistrant wraps GetRegistrantContext with background context.
func (w *Wallet) GetRegistrant(name string) (*nkn.Registrant, error) {
	return w.GetRegistrantContext(context.Background(), name)
}

// GetRegistrantContext is the same as package level GetRegistrantContext, but
// this wallet's SeedRPCServerAddr.
func (w *Wallet) GetRegistrantContext(ctx context.Context, name string) (*nkn.Registrant, error) {
	return nkn.GetRegistrantContext(ctx, name, w.config)
}

// SendRawTransaction wraps SendRawTransactionContext with background context.
func (w *Wallet) SendRawTransaction(txn *transaction.Transaction) (string, error) {
	return w.SendRawTransactionContext(context.Background(), txn)
}

// SendRawTransactionContext is the same as package level
// SendRawTransactionContext, but using this wallet's SeedRPCServerAddr.
func (w *Wallet) SendRawTransactionContext(ctx context.Context, txn *transaction.Transaction) (string, error) {
	return nkn.SendRawTransactionContext(ctx, txn, w.config)
}

// Transfer wraps TransferContext with background context.
func (w *Wallet) Transfer(address, amount string, config *nkn.TransactionConfig) (string, error) {
	return w.TransferContext(context.Background(), address, amount, config)
}

// TransferContext is a shortcut for TransferContext using this wallet as
// SignerRPCClient.
func (w *Wallet) TransferContext(ctx context.Context, address, amount string, config *nkn.TransactionConfig) (string, error) {
	return nkn.TransferContext(ctx, w, address, amount, config)
}

// RegisterName wraps RegisterNameContext with background context.
func (w *Wallet) RegisterName(name string, config *nkn.TransactionConfig) (string, error) {
	return w.RegisterNameContext(context.Background(), name, config)
}

// RegisterNameContext is a shortcut for RegisterNameContext using this wallet
// as SignerRPCClient.
func (w *Wallet) RegisterNameContext(ctx context.Context, name string, config *nkn.TransactionConfig) (string, error) {
	return nkn.RegisterNameContext(ctx, w, name, config)
}

// TransferName wraps TransferNameContext with background context.
func (w *Wallet) TransferName(name string, recipientPubKey []byte, config *nkn.TransactionConfig) (string, error) {
	return w.TransferNameContext(context.Background(), name, recipientPubKey, config)
}

// TransferNameContext is a shortcut for TransferNameContext using this wallet
// as SignerRPCClient.
func (w *Wallet) TransferNameContext(ctx context.Context, name string, recipientPubKey []byte, config *nkn.TransactionConfig) (string, error) {
	return nkn.TransferNameContext(ctx, w, name, recipientPubKey, config)
}

// DeleteName wraps DeleteNameContext with background context.
func (w *Wallet) DeleteName(name string, config *nkn.TransactionConfig) (string, error) {
	return w.DeleteNameContext(context.Background(), name, config)
}

// DeleteNameContext is a shortcut for DeleteNameContext using this wallet as
// SignerRPCClient.
func (w *Wallet) DeleteNameContext(ctx context.Context, name string, config *nkn.TransactionConfig) (string, error) {
	return nkn.DeleteNameContext(ctx, w, name, config)
}

// Subscribe wraps SubscribeContext with background context.
func (w *Wallet) Subscribe(identifier, topic string, duration int, meta string, config *nkn.TransactionConfig) (string, error) {
	return w.SubscribeContext(context.Background(), identifier, topic, duration, meta, config)
}

// SubscribeContext is a shortcut for SubscribeContext using this wallet as
// SignerRPCClient.
//
// Duration is changed to signed int for gomobile compatibility.
func (w *Wallet) SubscribeContext(ctx context.Context, identifier, topic string, duration int, meta string, config *nkn.TransactionConfig) (string, error) {
	return nkn.SubscribeContext(ctx, w, identifier, topic, duration, meta, config)
}

// Unsubscribe wraps UnsubscribeContext with background context.
func (w *Wallet) Unsubscribe(identifier, topic string, config *nkn.TransactionConfig) (string, error) {
	return w.UnsubscribeContext(context.Background(), identifier, topic, config)
}

// UnsubscribeContext is a shortcut for UnsubscribeContext using this wallet as
// SignerRPCClient.
func (w *Wallet) UnsubscribeContext(ctx context.Context, identifier, topic string, config *nkn.TransactionConfig) (string, error) {
	return nkn.UnsubscribeContext(ctx, w, identifier, topic, config)
}
