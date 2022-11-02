package evm

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"

	"github.com/chainlydev/evmparser/models"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/google/logger"
	"github.com/influxdata/influxdb/pkg/slices"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionParse struct {
	receipt     *types.Receipt
	transaction *types.Transaction
	chain       int
	msg         *types.Message
	cli         *ethclient.Client
	interacted  []string
}

func initLogger() {
	const logPath = "./contract_parser.log"

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	var verbose = flag.Bool("verbose", false, "print info level logs to stdout")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	flag.Parse()

	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("Listener", *verbose, true, lf).Close()
}

func NewTransactionParse(receipt *types.Receipt, trans *types.Transaction, chain int, cli *ethclient.Client) *TransactionParse {
	initLogger()

	return &TransactionParse{transaction: trans, receipt: receipt, cli: cli}
}

var contracts = make(map[string]*Contract)
var cons = make(map[string]*abi.ABI)
var tokens = make(map[string]*models.TokenInfo)
var tokensCall = make(map[string]*TokenParse)

func (t *TransactionParse) parse_logs(logs []*types.Log) ([]models.Logs, bool, bool) {
	var totalLogs []models.Logs
	//var tokensAct = make(map[string]interface{})
	var all_address []string
	if t.msg != nil {

		all_address = append(all_address, t.msg.To().Hex())
		all_address = append(all_address, t.msg.From().Hex())
	}

	for _, log := range logs {
		if !slices.Exists(all_address, log.Address.Hex()) {
			all_address = append(all_address, log.Address.Hex())
		}
		for _, topic := range log.Topics {
			if len(topic.Hex()) == 42 {
				all_address = append(all_address, topic.Hex())
			}
		}

	}
	// var parsed_logs []models.Logs
	for _, address := range all_address {
		var contract *Contract
		var con *abi.ABI
		address := common.HexToAddress(address)
		resp, err := t.cli.CodeAt(context.Background(), address, nil)
		if err != nil {
			panic(err)
		} else {
			if len(resp) == 0 {

				continue

			}
		}

		if _, ok := contracts[address.Hex()]; ok {

			contract = contracts[address.Hex()]
			con = cons[address.Hex()]
		} else {
			contract = NewContract(address, t.cli, 1)
			con := contract.InitContract()
			contracts[address.Hex()] = contract
			cons[address.Hex()] = con
		}
		var token_data *models.TokenInfo
		if _, ok := tokens[address.Hex()]; ok {

		} else {
			token := NewTokenParse(con, contract, t.chain, false)
			token_data = token.InitToken()
			tokens[address.Hex()] = token_data
			tokensCall[address.Hex()] = token
		}

	}
	var is_swap = false
	var is_nft = false
	for _, log := range logs {
		contract := contracts[log.Address.Hex()]
		tokenCallData := tokensCall[log.Address.Hex()]
		var call string
		var call_type string
		var function *abi.Method
		var event *abi.Event
		var errFunction error
		var errEvent error
		var paramters []interface{}
		outputDataMap := make(map[string]interface{})
		if contract == nil {
			continue
		}
		event, errEvent = contract.evm_contract.EventByID(log.Topics[0])
		if errEvent != nil {
			function, errFunction = contract.evm_contract.MethodById(log.Topics[0].Bytes())
			if errFunction == nil {
				call = function.Name
				call_type = "Function"
			}
		} else {
			call = event.Name
			call_type = "Event"
		}
		if call_type == "Event" {

			err := contract.evm_contract.UnpackIntoMap(outputDataMap, event.Name, log.Data)
			if err != nil {
				panic("unpack error")
			}

		} else if call_type == "Function" {

			err := contract.evm_contract.UnpackIntoMap(outputDataMap, function.Name, log.Data)
			if err != nil {
				fmt.Println(err)
				panic("unpack error")
			}

		}

		for key, val := range outputDataMap {

			if reflect.TypeOf(val).String() == "*big.Int" {
				val, _ = primitive.ParseDecimal128FromBigInt(val.(*big.Int), 0)
			} else if reflect.TypeOf(val).String() == "[]*big.Int" {
				var allvall []primitive.Decimal128
				for _, z := range val.([]*big.Int) {
					v, _ := primitive.ParseDecimal128FromBigInt(z, 0)
					allvall = append(allvall, v)
				}
				val = allvall
			} else if reflect.TypeOf(val).String() == "common.Address" {
				val = val.(common.Address).Hex()
			} else if reflect.TypeOf(val).String() == "uint8" {
				val = int(val.(uint8))
			} else if reflect.TypeOf(val).String() == "[]uint8" {
				var allvall []int
				for _, z := range val.([]uint8) {
					allvall = append(allvall, int(z))
				}
				val = allvall

			} else {
				fmt.Println(reflect.TypeOf(val).Name())
				fmt.Println(reflect.TypeOf(val).String())
			}
			paramters = append(paramters, map[string]any{"key": key, "value": val})
			fmt.Println(key, val)
		}

		var tpk []string
		for _, i := range log.Topics {
			tpk = append(tpk, i.Hex())
		}
		if len(logs) > 1 && strings.ToLower(call) == "swap" {
			is_swap = true
		}
		t.interacted = append(t.interacted, tokenCallData.contractParser.address.Hex())
		logItem := models.Logs{
			Address:    log.Address.Hex(),
			Topics:     tpk,
			Data:       common.Bytes2Hex(log.Data),
			Index:      uint64(log.Index),
			Action:     call,
			ActionType: call_type,
			Parameters: paramters,
			Token:      tokenCallData.contractParser.address.Hex(),
		}
		totalLogs = append(totalLogs, logItem)
	}

	return totalLogs, is_swap, is_nft
}

func (t *TransactionParse) Parse() models.Transaction {

	config := &params.ChainConfig{
		ChainID:             big.NewInt(int64(1)),
		HomesteadBlock:      big.NewInt(1150000),
		DAOForkBlock:        big.NewInt(1920000),
		EIP150Block:         big.NewInt(2463000),
		EIP155Block:         big.NewInt(2675000),
		EIP158Block:         big.NewInt(2675000),
		ByzantiumBlock:      big.NewInt(4370000),
		ConstantinopleBlock: big.NewInt(7280000),
		PetersburgBlock:     big.NewInt(7280000),
		IstanbulBlock:       big.NewInt(9069000),
		MuirGlacierBlock:    big.NewInt(9200000),
		BerlinBlock:         big.NewInt(12244000),
		LondonBlock:         big.NewInt(12965000),
		ArrowGlacierBlock:   big.NewInt(13773000),
		GrayGlacierBlock:    big.NewInt(15050000),
	}
	signer := types.MakeSigner(config, t.receipt.BlockNumber)
	msg, _ := t.transaction.AsMessage(signer, t.transaction.GasFeeCap())
	t.msg = &msg
	s, err := signer.Sender(t.transaction)
	logs, is_swap, is_nft := t.parse_logs(t.receipt.Logs)

	value, _ := primitive.ParseDecimal128FromBigInt(msg.Value(), 0)
	if msg.Value().Uint64() > 0 {
		token := tokensCall["0x0"]
		if token == nil {
			token = NewTokenParse(nil, nil, t.chain, true)
			tokensCall["0x0"] = token
		}
	}

	//var nft

	transaction := models.Transaction{
		Logs:             logs,
		To:               msg.To().Hex(),
		From:             msg.From().Hex(),
		Index:            int(t.receipt.TransactionIndex),
		Value:            value,
		RawTransaction:   t.transaction,
		RawReciept:       t.receipt,
		Hash:             t.receipt.TxHash.Hex(),
		Block:            t.receipt.BlockNumber.Int64(),
		IsNFT:            is_nft,
		IsSwap:           is_swap,
		InteractedTokens: t.interacted,
	}
	return transaction
}
