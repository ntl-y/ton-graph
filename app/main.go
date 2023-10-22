package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

const configUrl = "https://ton.org/global.config.json"

type Transaction struct {
	From   string
	To     string
	Amount float64
	Hash   string
}

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("static/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/graph", func(c *gin.Context) {

		walletAdress := c.PostForm("addressInput")
		var depth int8 = 5
		var transAmount int8 = 5

		hashMap := make(map[string]bool)
		uniqueTransactions := make([]Transaction, 0, transAmount)
		api, ctx, b := client()

		getTransactions(api, ctx, b, &uniqueTransactions, hashMap, walletAdress, depth, transAmount)

		c.JSON(http.StatusOK, uniqueTransactions)
	})
	r.Run(":3000")
}

func client() (ton.APIClientWrapped, context.Context, *tlb.BlockInfo) {
	client := liteclient.NewConnectionPool()

	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		log.Fatal(err)
	}
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	ctx := client.StickyContext(context.Background())

	// we need fresh block info to run get methods
	b, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return api, ctx, b

}

func getAddressTransactions(api ton.APIClientWrapped, ctx context.Context, b *tlb.BlockInfo, addressToken string, transAmount int8) ([]Transaction, error) {
	if addressToken == "" {
		fmt.Println("Empty address")
		return []Transaction{}, nil
	}
	// TON Foundation account
	addr := address.MustParseAddr(addressToken)

	account, err := api.GetAccount(ctx, b, addr)
	if err != nil {
		fmt.Println(err)
		return []Transaction{}, err
	}

	// load last transAmount transactions
	list, err := api.ListTransactions(ctx, addr, uint32(transAmount), account.LastTxLT, account.LastTxHash)
	if err != nil {
		fmt.Println(err)
		return []Transaction{}, err
	}
	transactions := make([]Transaction, 0, transAmount)

	for _, t := range list {
		var amount *big.Int
		var sender, destination string
		hash := base64.StdEncoding.EncodeToString(t.Hash)

		if t.IO.Out != nil {
			listOut, err := t.IO.Out.ToSlice()
			if err != nil {
				fmt.Println(err)
				continue
			}

			destination = listOut[0].Msg.DestAddr().String()
			if listOut[0].MsgType == tlb.MsgTypeInternal {
				amount = listOut[0].AsInternal().Amount.Nano()
			}
		}
		switch t.Description.Description.(type) {
		default:
			continue
		case tlb.TransactionDescriptionOrdinary:
		}
		if t.IO.In != nil {
			if t.IO.In.MsgType == tlb.MsgTypeInternal {
				amount = t.IO.In.AsInternal().Amount.Nano()
				if amount.Cmp(big.NewInt(0)) != 0 {
					sender = t.IO.In.AsInternal().SrcAddr.String()
				}
			}
		}

		switch sender {
		case "":
			sender = addressToken
		default:
			destination = addressToken
		}

		if amount == nil {
			fmt.Println("Invalid amount")
			continue
		}
		intValue := amount.Int64()
		floatValue := float64(intValue) / 1e9
		transactions = append(transactions, Transaction{
			Amount: floatValue,
			From:   sender,
			To:     destination,
			Hash:   hash,
		})
	}
	return transactions, nil
}

func getTransactions(api ton.APIClientWrapped, ctx context.Context, b *tlb.BlockInfo, uniqueTransactions *[]Transaction, hashMap map[string]bool, address string, depth int8, transAmount int8) {
	transactions, err := getAddressTransactions(api, ctx, b, address, transAmount)
	if depth == 4 {
		fmt.Println(transactions)
	}
	if err != nil {
		fmt.Println(err)
		transactions = []Transaction{}
	}
	for i := range transactions {
		_, found := hashMap[transactions[i].Hash]
		if !found {
			*uniqueTransactions = append(*uniqueTransactions, transactions[i])
			hashMap[transactions[i].Hash] = true

			if depth > 1 {
				var nextAddress string
				if transactions[i].From == address {
					nextAddress = transactions[i].To
				} else {
					nextAddress = transactions[i].From
				}
				getTransactions(api, ctx, b, uniqueTransactions, hashMap, nextAddress, depth-1, transAmount)
			}
		}
	}

}
