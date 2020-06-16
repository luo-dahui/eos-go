package main

import (
	"context"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/eoscanada/eos-go/token"
	"log"
	"math"
	"strconv"
	"strings"
)


func getInfo(ctx context.Context){
	api := eos.New("http://192.168.10.128:1888/")
	infoResp, _ := api.GetInfo(ctx)
	// fmt.Println("Permission for infoResp:", infoResp)

	fmt.Println("chainid:", infoResp.ChainID)
	fmt.Println("block_number:", infoResp.HeadBlockNum)

	eosio, _ := api.GetAccount(ctx,"eosio")
	fmt.Println("account:【eosio】 info:", eosio.Permissions[0].RequiredAuth.Keys)

	account1, _ := api.GetAccount(ctx,"account1")
	fmt.Println("account:【account1】 info:", account1.Permissions[0].RequiredAuth.Keys)

	account2, _ := api.GetAccount(ctx,"account2")
	fmt.Println("account:【account2】 info:", account2.Permissions[0].RequiredAuth.Keys)
}

func getTransLen(ctx context.Context, num uint32){
	api := eos.New("http://154.85.35.174:8888/")
	infoResp, _ := api.GetBlockByNum(ctx,num)
	fmt.Println("block number:", infoResp)
}


func splitAsset(input string) (integralPart, decimalPart, symbolPart string, err error) {
	input = strings.Trim(input, " ")
	if len(input) == 0 {
		return "", "", "", fmt.Errorf("input cannot be empty")
	}

	parts := strings.Split(input, " ")
	if len(parts) >= 1 {
		integralPart, decimalPart, err = splitAssetAmount(parts[0])
		if err != nil {
			return
		}
	}

	if len(parts) == 2 {
		symbolPart = parts[1]
		if len(symbolPart) > 7 {
			return "", "", "", fmt.Errorf("invalid asset %q, symbol should have less than 7 characters", input)
		}
	}

	if len(parts) > 2 {
		return "", "", "", fmt.Errorf("invalid asset %q, expecting an amount alone or an amount and a currency symbol", input)
	}

	return
}

func splitAssetAmount(input string) (integralPart, decimalPart string, err error) {
	parts := strings.Split(input, ".")
	switch len(parts) {
	case 1:
		integralPart = parts[0]
	case 2:
		integralPart = parts[0]
		decimalPart = parts[1]

		if len(decimalPart) > math.MaxUint8 {
			err = fmt.Errorf("invalid asset amount precision %q, should have less than %d characters", input, math.MaxUint8)

		}
	default:
		return "", "", fmt.Errorf("invalid asset amount %q, expected amount to have at most a single dot", input)
	}

	return
}
type Int64 int64


func NewFixedSymbolAssetFromString(symbol eos.Symbol, input string) (out eos.Asset, err error) {
	integralPart, decimalPart, symbolPart, err := splitAsset(input)
	if err != nil {
		return out, err
	}

	symbolCode := symbol.MustSymbolCode().String()
	precision := symbol.Precision

	if len(decimalPart) > int(precision) {
		return out, fmt.Errorf("symbol %s precision mismatch: expected %d, got %d", symbol, precision, len(decimalPart))
	}

	if symbolPart != "" && symbolPart != symbolCode {
		return out, fmt.Errorf("symbol %s code mismatch: expected %s, got %s", symbol, symbolCode, symbolPart)
	}

	if len(decimalPart) < int(precision) {
		decimalPart += strings.Repeat("0", int(precision)-len(decimalPart))
	}

	val, err := strconv.ParseInt(integralPart+decimalPart, 10, 64)
	if err != nil {
		return out, err
	}

	return eos.Asset{
		Amount: eos.Int64(val),
		Symbol: eos.Symbol{Precision: precision, Symbol: symbolCode},
	}, nil
}


func createAccount() {

	api := eos.New("http://154.85.35.174:8888/")
	// infoResp, _ := api.GetInfo((context.Background()))

	// fmt.Println("chainid:", infoResp.ChainID)

	eosio_pri_key := "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"

	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(context.Background(), eosio_pri_key)
	if err != nil {
		panic(fmt.Errorf("import private key: %s", err))
	}
	api.SetSigner(keyBag)

	//txOpts := &eos.TxOptions{ChainID:infoResp.ChainID}

	pub_key := "EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV"
	pubkey, err := ecc.NewPublicKey(pub_key)
	if err != nil {
		log.Fatalln("invalid public key:", err)
	}

	var i = 1
	for i = 1; i <= 1; i++ {

		txOpts := &eos.TxOptions{}
		if err := txOpts.FillFromChain(context.Background(), api); err != nil {
			panic(fmt.Errorf("filling tx opts: %s", err))
		}

		tx := eos.NewTransaction([]*eos.Action{system.NewNewAccount("eosio", "test1", pubkey)}, txOpts)
		_, packedTx, err := api.SignTransaction(context.Background(), tx, txOpts.ChainID, eos.CompressionNone)
		if err != nil {
			panic(fmt.Errorf("sign transaction: %s", err))
		}

		// content, err := json.MarshalIndent(signedTx, "", "  ")
		if err != nil {
			panic(fmt.Errorf("json marshalling transaction: %s", err))
		}

		//fmt.Println(string(content))
		//fmt.Println()

		response, err := api.PushTransaction(context.Background(), packedTx)
		if err != nil {
			i = i -1
			//	panic(fmt.Errorf("push transaction: %s", err))
			continue
		}
		fmt.Printf("i:{%d}=========", i)

		//fmt.Printf("Transaction [%s] submitted to the network succesfully.\n", hex.EncodeToString(response.Processed.ID))
		fmt.Printf("Transaction ID: [%s].\n", response.TransactionID)

		infoResp, _ := api.GetInfo((context.Background()))
		fmt.Println("block_number:", infoResp.HeadBlockNum)
	}

}

func ExampleAPI_PushTransaction_transfer_EOS() {
	api := eos.New("http://154.85.35.174:8888/")
	// infoResp, _ := api.GetInfo((context.Background()))

	// fmt.Println("chainid:", infoResp.ChainID)

	eosio_pri_key := "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"

	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(context.Background(), eosio_pri_key)
	if err != nil {
		panic(fmt.Errorf("import private key: %s", err))
	}
	api.SetSigner(keyBag)

	from := eos.AccountName("eosio")
	to := eos.AccountName("tfacc2")
	//quantity, err := eos.NewSYSAssetFromString("0.1000 SYS")
	var SYSSymbol = eos.Symbol{Precision: 4, Symbol: "SYS"}
	quantity, err := NewFixedSymbolAssetFromString(SYSSymbol, "100 SYS")
	memo := ""

	if err != nil {
		panic(fmt.Errorf("invalid quantity: %s", err))
	}

	//txOpts := &eos.TxOptions{ChainID:infoResp.ChainID}

	var i = 1
	for i = 1; i <= 10; i++ {

		txOpts := &eos.TxOptions{}
		if err := txOpts.FillFromChain(context.Background(), api); err != nil {
			panic(fmt.Errorf("filling tx opts: %s", err))
		}

		tx := eos.NewTransaction([]*eos.Action{token.NewTransfer(from, to, quantity, memo)}, txOpts)
		_, packedTx, err := api.SignTransaction(context.Background(), tx, txOpts.ChainID, eos.CompressionNone)
		if err != nil {
			panic(fmt.Errorf("sign transaction: %s", err))
		}

		// content, err := json.MarshalIndent(signedTx, "", "  ")
		if err != nil {
			panic(fmt.Errorf("json marshalling transaction: %s", err))
		}

		//fmt.Println(string(content))
		//fmt.Println()

		response, err := api.PushTransaction(context.Background(), packedTx)
		if err != nil {
			i = i -1
		//	panic(fmt.Errorf("push transaction: %s", err))
			continue
		}
		fmt.Printf("i:{%d}=========", i)

		//fmt.Printf("Transaction [%s] submitted to the network succesfully.\n", hex.EncodeToString(response.Processed.ID))
		fmt.Printf("Transaction ID: [%s].\n", response.TransactionID)

		infoResp, _ := api.GetInfo((context.Background()))
		fmt.Println("block_number:", infoResp.HeadBlockNum)
	}

}

func main() {
	//ExampleAPI_PushTransaction_transfer_EOS()
	// getInfo(context.Background())
	//getTransLen(context.Background(),491954)
	createAccount()
}
