package evm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/chainlydev/evmparser/lib"
	"github.com/chainlydev/evmparser/models"
	"github.com/ethereum/go-ethereum/common"
)

type AbiParser struct {
	client lib.MongoConnect
}

var abiParser *AbiParser

func NewAbiParser(client lib.MongoConnect) *AbiParser {
	if abiParser != nil {
		return abiParser
	} else {
		abiParser = &AbiParser{client: client}
		return abiParser
	}
}

func (abi AbiParser) GetAbiEth(addr common.Address) *models.AbiResponse {
	uri := fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s", addr.Hex(), os.Getenv("ETHERSCAN_KEY"))
	fmt.Println("Abi Parse from", uri)
	resp := abi.GetAbiEthBase(uri, addr)
	return resp
}
func (abi AbiParser) GetAbiEthBase(uri string, addr common.Address) *models.AbiResponse {
	emptyAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	if addr == emptyAddress {
		return nil
	}

	response, err := http.Get(uri)
	if err != nil {
		time.Sleep(4 * time.Second)
		return abi.GetAbiEthBase(uri, addr)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		time.Sleep(4 * time.Second)
		return abi.GetAbiEthBase(uri, addr)
	}
	var jsonResp *models.EthScanResponse
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if jsonResp.Status == "1" {

		abiResp := &models.AbiResponse{
			Address: addr.String(),
			Abi:     jsonResp.Result,
		}
		return abiResp
	} else {
		if jsonResp.Result == "Max rate limit reached" {
			time.Sleep(4 * time.Second)
			return abi.GetAbiEthBase(uri, addr)
		}
		if jsonResp.Result == "Contract source code not verified" {
			return nil
		}
	}
	return nil
}

func (abi AbiParser) GetAbiBsc(addr common.Address) *models.AbiResponse {
	uri := fmt.Sprintf("https://api.bscscan.com/api?module=contract&action=getabi&address=%s&apikey=%s", addr.Hex(), os.Getenv("BSSSCAN_KEY"))
	return abi.GetAbiEthBase(uri, addr)
}

func (abi AbiParser) GetAbiPolygon(addr common.Address) *models.AbiResponse {
	uri := fmt.Sprintf("https://api.polygonscan.com/api?module=contract&action=getabi&address=%s&apikey=%s", addr.Hex(), os.Getenv("POLYGONSCAN_KEY"))
	return abi.GetAbiEthBase(uri, addr)
}

func (abi AbiParser) parse_from_api(chain int, address common.Address) (string, error) {
	var resp *models.AbiResponse
	switch chain {
	case 1:
		resp = abi.GetAbiEth(address)
		break
	case 137:
		resp = abi.GetAbiPolygon(address)
		break
	case 56:
		resp = abi.GetAbiBsc(address)
		break
	}

	if resp != nil {
		_, _ = abi.client.Collection("abi").InsertOne(context.TODO(), bson.M{"address": address.Hex(), "abi": resp.Abi, "chain": chain})
		return resp.Abi, nil
	} else {
		_, _ = abi.client.Collection("abi").InsertOne(context.TODO(), bson.M{"address": address.Hex(), "abi": nil, "chain": chain})
		return "", errors.New("not found")
	}

}

func (abi AbiParser) parse_from_db(chain int, address common.Address) (string, error) {

	var abi_data *models.AbiResponse
	resp := abi.client.Collection("abi").FindOne(context.TODO(), bson.M{"address": address.Hex(), "chain": chain})
	err := resp.Decode(&abi_data)
	if err == nil {
		return abi_data.Abi, nil
	}
	return "", errors.New("not found")
}

func (abi AbiParser) GetAbi(chain int, address common.Address) (string, error) {
	data, err := abi.parse_from_db(chain, address)
	if err != nil {
		fmt.Println("Abi parse from api")
		data, err = abi.parse_from_api(chain, address)
		if err != nil {
			return "", errors.New("not found")
		}
	}
	if data == "Contract source code not verified" {
		return "", errors.New("not found")
	}
	if data == "" {
		return "", errors.New("not found")
	}
	return data, nil
}
