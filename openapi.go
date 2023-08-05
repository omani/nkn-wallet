package nknwallet

import (
	"github.com/nknorg/nkn/v2/common"
	client "github.com/omani/nkn-openapi-client"
)

type Openapi struct {
	client client.Client
	wallet *Wallet
}

func New(w *Wallet) *Openapi {
	c := client.New()
	c.SetAddress("https://openapi.nkn.org/api/v1")

	return &Openapi{
		wallet: w,
		client: c,
	}
}

func (o *Openapi) GetTransactions() (*client.ResponseGetAddressTransaction, error) {
	tx, err := o.client.GetAddressTransactions(o.wallet.Address())
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (o *Openapi) GetBalance() (common.Fixed64, error) {
	resp, err := o.client.GetSingleAddress(o.wallet.Address())
	if err != nil {
		return 0, err
	}

	return common.Fixed64(resp.Balance), nil
}
