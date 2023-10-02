package main

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("static/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/graph", func(c *gin.Context) {
		//jsonData := getTransactions("EQBYgo6YoU4-Tfbwxk3bLDADmYvoGBkih8pGcEA9PfBz2JJ8")
		walletAdress := c.PostForm("addressInput")
		jsonData := getTransactions(walletAdress)
		fmt.Println("!!!TRANSACTIONS:", jsonData)
		c.JSON(http.StatusOK, jsonData)
	})
	r.Run(":3000")
}

func getTransactions(adressToken string) []Transaction {
	client := liteclient.NewConnectionPool()

	configUrl := "https://ton.org/global.config.json"
	// connect to mainnet lite server
	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		panic(err)
	}

	// initialize ton api lite connection wrapper
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()

	// if we want to route all requests to the same node, we can use it
	ctx := client.StickyContext(context.Background())

	// we need fresh block info to run get methods
	b, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		panic(err)
	}

	// TON Foundation account
	addr := address.MustParseAddr(adressToken)

	account, err := api.GetAccount(context.Background(), b, addr)
	if err != nil {
		panic(err)
	}

	// load last 30 transactions
	list, err := api.ListTransactions(context.Background(), addr, 50, account.LastTxLT, account.LastTxHash)
	if err != nil {
		panic(err)
	}
	transactions := make([]Transaction, 0, 10)

	for _, t := range list {
		var amount *big.Int
		var sender, destination string
		if t.IO.Out != nil {
			listOut, err := t.IO.Out.ToSlice()
			if err != nil {
				panic(err)
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
			sender = adressToken
		default:
			destination = adressToken
		}

		intValue := amount.Int64()
		floatValue := float64(intValue) / 1e9
		transactions = append(transactions, Transaction{
			Amount: floatValue,
			From:   sender,
			To:     destination,
		})
	}
	return transactions
}

type Transaction struct {
	From   string
	To     string
	Amount float64
}
