package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/chainlydev/evmparser/lib"
	"github.com/chainlydev/evmparser/models"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/google/logger"
	"github.com/influxdata/influxdb/pkg/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var TokenLists = make(map[string]models.TokenInfo)
var MemoryToken = make(map[string]models.TokenData)

type TokenParse struct {
	contract       *eth.Contract
	contractParser *Contract
	is_proxy       bool
	is_base        bool
	chain          int
}

func NewTokenParse(contract *eth.Contract, contractParser *Contract, chain int, is_base bool) *TokenParse {
	return &TokenParse{contract: contract, contractParser: contractParser, chain: chain, is_base: is_base}
}
func (tk *TokenParse) ParseWithPrice() *models.TokenData {
	if tk.is_base {
		return &models.TokenData{
			Name:       "Ethereum",
			Symbol:     "Eth",
			Decimal:    18,
			TimedPrice: -1,
		}
	}
	var tkoninf *models.TokenInfo
	keys := make([]string, 0, len(TokenLists))
	for k := range TokenLists {
		keys = append(keys, k)
	}
	if tk == nil {
		return nil
	}
	if tk.contract == nil {
		return nil
	}
	if slices.Exists(keys, tk.contract.Address().Hex()) {
		tinfo := TokenLists[tk.contract.Address().Hex()]
		tkoninf = &tinfo

	} else {
		tkoninf = tk.InitToken()
	}
	if tkoninf == nil {
		return nil
	}
	return &models.TokenData{
		Name:        tkoninf.Name,
		Symbol:      tkoninf.Symbol,
		Address:     tkoninf.Address,
		TotalVolume: tkoninf.TotalSupply,
		Decimal:     tkoninf.Decimal,
	}
}
func (tk *TokenParse) get_db_token() *models.TokenInfo {
	mongo := lib.NewMongo()
	if tk.contract == nil {
		return nil
	}
	result := mongo.Collection("token").FindOne(context.Background(), bson.M{"address": tk.contract.Address().Hex()})
	var token models.TokenData
	err := result.Decode(&token)
	if err != nil {
		return nil
	}
	return &models.TokenInfo{
		Address:     token.Address,
		Name:        token.Name,
		Symbol:      token.Symbol,
		Decimal:     token.Decimal,
		TotalSupply: token.TotalVolume,
	}
}

var scan = NewScanParse()
var inserted []string

func (tk *TokenParse) InitToken() *models.TokenInfo {
	if tk.contract == nil {
		logger.Error("Contract is nil")
		return nil
	}
	if _, ok := TokenLists[tk.contract.Address().Hex()]; ok {
		token := TokenLists[tk.contract.Address().Hex()]
		logger.Info("Contract already fetched")
		return &token
	}
	item := tk.get_db_token()
	if item != nil {
		logger.Info("Contract already fetched in db")
		return item
	}
	logger.Info("Contract Fetching Now")
	cType := tk.contractParser.GetType()
	logger.Info("Contract Type", cType)

	if slices.ExistsIgnoreCase([]string{"ERC20", "ERC721", "ERC777", "ERC1155"}, cType) {

		symbol, _ := tk.contract.Call("symbol")
		name, _ := tk.contract.Call("name")
		decimal, _ := tk.contract.Call("decimals")
		total_supply, _ := tk.contract.Call("totalSupply")
		if name == nil {
			name = ""
		}
		if symbol == nil {
			symbol = ""
		}
		if decimal == nil {
			decimal = 0
		}
		if total_supply == nil {
			total_supply = big.NewInt(0)
		}
		defer func() {
			if r := recover(); r != nil {
				defer func() {
					if r := recover(); r != nil {
						defer func() {
							if r := recover(); r != nil {
								decimal = decimal.(uint8)
							}
						}()
						decimal = uint8(decimal.(int))
					}
				}()
				decimal = uint8(decimal.(*big.Int).Uint64())
			}
		}()
		decimal = decimal.(uint8)
		fmt.Println("---------------------", len(MemoryToken), len(TokenLists), "------------------")
		tsup, _ := primitive.ParseDecimal128(fmt.Sprint(total_supply))

		tokeninf := models.TokenInfo{
			Address:     tk.contract.Address().Hex(),
			Name:        name.(string),
			Symbol:      symbol.(string),
			Decimal:     decimal,
			TotalSupply: tsup,
		}
		TokenLists[tk.contract.Address().Hex()] = tokeninf
		var proxy_address string
		if tk.contractParser.proxy_address != nil {
			proxy_address = tk.contractParser.proxy_address.Hex()
		}
		token := models.TokenData{
			Address:       tk.contract.Address().Hex(),
			Name:          name.(string),
			Symbol:        symbol.(string),
			Decimal:       decimal,
			TotalVolume:   tsup,
			ERCType:       cType,
			ProxyContract: tk.is_proxy,
			ProxyAddress:  proxy_address,
		}
		MemoryToken[tk.contract.Address().Hex()] = token
		mongo := lib.NewMongo()
		if len(MemoryToken) > 0 {
			var items []interface{}
			for _, tokenMemory := range MemoryToken {
				finder, _ := mongo.Collection("token").Find(context.Background(), bson.M{"address": tokenMemory.Address})
				logger.Info("Total Length", finder.RemainingBatchLength())
				if finder.RemainingBatchLength() == 0 {
					if !slices.ExistsIgnoreCase(inserted, tokenMemory.Address) {

						items = append(items, tokenMemory)
						inserted = append(inserted, tokenMemory.Address)
					}
				}
			}
			_, err := mongo.Collection("token").InsertMany(context.Background(), items)
			if err != nil {
				logger.Info(err)
				panic("Error Mongo")
			}

			for k := range MemoryToken {
				address := MemoryToken[k].Address
				scan.GetAddress(address)
				delete(MemoryToken, k)
			}

		}

		return &tokeninf

	}
	return nil

}
