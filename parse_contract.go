package evm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chainlydev/evmparser/lib"
	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/logger"
	"github.com/influxdata/influxdb/pkg/slices"
)

type Contract struct {
	address       common.Address
	client        lib.MongoConnect
	chain         int
	is_proxy      bool
	proxy_address *common.Address
	type_name     string
	evm_contract  abi.ABI
}

func NewContract(address common.Address, chain int) *Contract {

	mongo := lib.NewMongo()
	return &Contract{address: address, chain: chain, client: *mongo, is_proxy: false, proxy_address: nil, type_name: "fittt"}
}

func (cn *Contract) init_web3(abi string) (*eth.Contract, *rpc.Client) {
	var rpcProviderURL = os.Getenv("ETH_PROVIDER")
	if cn.chain == 137 {
		rpcProviderURL = os.Getenv("POLYGON_PROVIDER")
	}
	if cn.chain == 56 {
		rpcProviderURL = "https://bsc-dataseed.binance.org/"
	}

	web3Item, _ := web3.NewWeb3(rpcProviderURL)
	rpc, _ := rpc.NewClient(rpcProviderURL, "")
	contract, _ := web3Item.Eth.NewContract(abi, cn.address.Hex())
	return contract, rpc
}

func (cn *Contract) detect_type_proxy_swap(contract *eth.Contract, abi_string string) string {
	methods := contract.AllMethods()
	for _, method := range methods {
		if strings.ContainsAny(strings.ToLower(method), "domainseparator") {
			return "SWAP"
		}
		if strings.ContainsAny(strings.ToLower(method), "implementation") {
			return "Proxy"
		}
	}
	return ""
}
func (cn *Contract) detect_type(contract *eth.Contract, abi_string string) string {
	methods := contract.AllMethods()

	if slices.ExistsIgnoreCase(methods, "name") &&
		slices.ExistsIgnoreCase(methods, "information") {
		return "Multi Type"
	} else if slices.ExistsIgnoreCase(methods, "symbol") &&
		slices.ExistsIgnoreCase(methods, "name") &&
		slices.ExistsIgnoreCase(methods, "decimals") &&
		slices.ExistsIgnoreCase(methods, "balanceOf") &&
		slices.ExistsIgnoreCase(methods, "transfer") &&
		slices.ExistsIgnoreCase(methods, "transferFrom") &&
		slices.ExistsIgnoreCase(methods, "approve") &&
		slices.ExistsIgnoreCase(methods, "allowance") &&
		slices.ExistsIgnoreCase(methods, "totalSupply") &&
		!slices.ExistsIgnoreCase(methods, "asset") &&
		!slices.ExistsIgnoreCase(methods, "granularity") {
		return "ERC20"
	} else if slices.ExistsIgnoreCase(methods, "balanceOf") &&
		slices.ExistsIgnoreCase(methods, "ownerOf") {
		return "ERC721"
	} else if slices.ExistsIgnoreCase(methods, "symbol") &&
		slices.ExistsIgnoreCase(methods, "name") &&
		slices.ExistsIgnoreCase(methods, "granularity") &&
		slices.ExistsIgnoreCase(methods, "balanceOf") {
		return "ERC777"
	} else if slices.ExistsIgnoreCase(methods, "asset") &&
		slices.ExistsIgnoreCase(methods, "name") &&
		slices.ExistsIgnoreCase(methods, "decimals") &&
		slices.ExistsIgnoreCase(methods, "balanceOf") {
		return "ERC4626"
	} else if slices.ExistsIgnoreCase(methods, "uri") &&
		slices.ExistsIgnoreCase(methods, "supportsinterface") {
		return "ERC1155"
	} else if slices.ExistsIgnoreCase(methods, "token0") &&
		slices.ExistsIgnoreCase(methods, "token1") {
		return "Swap"
	} else if slices.ExistsIgnoreCase(methods, "totalSupply") &&
		slices.ExistsIgnoreCase(methods, "balanceOf") &&
		slices.ExistsIgnoreCase(methods, "transfer") &&
		slices.ExistsIgnoreCase(methods, "allowance") &&
		!slices.ExistsIgnoreCase(methods, "asset") &&
		!slices.ExistsIgnoreCase(methods, "granularity") {
		return "ERC20"
	} else if slices.ExistsIgnoreCase(methods, "DOMAIN_SEPARATOR") {
		return "Multi Type"
	} else if slices.ExistsIgnoreCase(methods, "MIN_REGISTRATION_DURATION") &&
		slices.ExistsIgnoreCase(methods, "rentPrice") &&
		slices.ExistsIgnoreCase(methods, "available") {
		return "ETH Name Service"
	} else if slices.ExistsIgnoreCase(methods, "DomainSeparator") {
		return "SWAP"
	} else if slices.ExistsIgnoreCase(methods, "implementation") || slices.ExistsIgnoreCase(methods, "__implementation__") {
		return "Proxy"
	} else {
		if strings.Contains(abi_string, "implementation") {
			return "Proxy"
		} else {
			return ""
		}
	}
}

func (cn *Contract) get_rpc_call(chain int, msg interface{}) (error, interface{}) {

	jsonValue, _ := json.Marshal(msg)
	var rpcProviderURL = os.Getenv("ETH_PROVIDER")
	if chain == 137 {
		rpcProviderURL = os.Getenv("POLYGON_PROVIDER")
	}
	if chain == 56 {
		rpcProviderURL = "https://bsc-dataseed.binance.org/"
	}

	resp, err := http.Post(rpcProviderURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return err, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]any
	_ = json.Unmarshal(body, &response)
	return nil, response

}
func (cn *Contract) parse_proxy(chain int, contract *eth.Contract) (error, common.Address) {
	respImp, errImp := contract.Call("implementation")
	if errImp == nil {
		return nil, respImp.(common.Address)

	}
	know_hashes := []string{
		"0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103",
		"0x7050c9e0f4ca769c69bd3a8ef740bc37934f8e2c036e5a723fd8ee048ed3f8c3",
		"0xc5f16f0fcc639fa48a6947836d9850f504798523bf8c9a3a87d5876cf622bcf7",
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0xa3f0ad74e5423aebfd80d3ef4346578335a9a72aeaee59ff6cb3582b35133d50",
		"0x13f464f5b1a0affee1715000e48aa76e15a8bafca7bfca161fb73dde5768e2b2",
		"0x7050c9e0f4ca769c69bd3a8ef740bc37934f8e2c036e5a723fd8ee048ed3f8c3",
		"0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"}

	for _, hash := range know_hashes {
		time.Sleep(100 * time.Millisecond)
		params := []string{cn.address.Hex(), hash, "latest"}
		values := map[string]any{"jsonrpc": "2.0", "method": "eth_getStorageAt", "params": params, "id": 1}
		err, resp := cn.get_rpc_call(chain, values)
		if err == nil {
			if resp.(map[string]any)["result"] != nil {
				if resp.(map[string]any)["result"].(string) != "0x0000000000000000000000000000000000000000000000000000000000000000" {
					return nil, common.HexToAddress(resp.(map[string]any)["result"].(string))
				}
			}

		}
	}
	return errors.New("can't determinate proxied address"), common.HexToAddress("0x0")
}

func (cn *Contract) IsAbi() bool {

	return false
}
func (cn *Contract) InitContract() (*eth.Contract, *rpc.Client) {
	var contract *eth.Contract
	var client *rpc.Client
	cn.IsAbi()
	fmt.Println("Abi parsing", cn.address)
	abi_parser := NewAbiParser(cn.client)
	abi_string, err := abi_parser.GetAbi(cn.chain, cn.address)
	fmt.Println("Abi string", cn.address)
	if err != nil {
		contract, client = cn.detect_abi()
	}
	fmt.Println("Detect Abi", cn.address)
	if contract == nil {
		contract, client = cn.init_web3(abi_string)
	}
	fmt.Println("WEB3 Abi", cn.address)
	contract_type := cn.detect_type(contract, abi_string)
	fmt.Println("Detect Type", cn.address)
	if contract_type == "" {
		fmt.Println("Detect Proxy Swap", cn.address)
		contract_type = cn.detect_type_proxy_swap(contract, abi_string)
	}
	if contract_type == "Proxy" {
		fmt.Println("Parse Proxy", cn.address)
		err, proxy_addr := cn.parse_proxy(cn.chain, contract)

		if err != nil {
			logger.Error("proxy address error ", contract.Address().Hex(), cn.address.Hex(), contract_type)

		}
		cn.is_proxy = true
		cn.proxy_address = &proxy_addr
		fmt.Println("Proxy Addr", cn.address)
		abi_string, _ = abi_parser.GetAbi(cn.chain, proxy_addr)
		contract, client = cn.init_web3(abi_string)
		if contract != nil {

			contract_type = cn.detect_type(contract, abi_string)
		}

	}

	if contract_type == "" {
		logger.Error("can't determinate contract type ", cn.address.Hex())
	}
	cn.type_name = contract_type
	fmt.Println("Evm", cn.address)
	cn.evm_contract, _ = abi.JSON(bytes.NewReader([]byte(abi_string)))
	return contract, client
}

func (cn *Contract) GetType() string {
	return cn.type_name
}
func (cn *Contract) detect_abi() (*eth.Contract, *rpc.Client) {
	var contract *eth.Contract
	var client *rpc.Client
	var ERC20_String = `[
		{
		  "constant": true,
		  "inputs": [],
		  "name": "name",
		  "outputs": [
			{
			  "name": "",
			  "type": "string"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "constant": false,
		  "inputs": [
			{
			  "name": "_spender",
			  "type": "address"
			},
			{
			  "name": "_value",
			  "type": "uint256"
			}
		  ],
		  "name": "approve",
		  "outputs": [
			{
			  "name": "",
			  "type": "bool"
			}
		  ],
		  "payable": false,
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [],
		  "name": "totalSupply",
		  "outputs": [
			{
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "constant": false,
		  "inputs": [
			{
			  "name": "_from",
			  "type": "address"
			},
			{
			  "name": "_to",
			  "type": "address"
			},
			{
			  "name": "_value",
			  "type": "uint256"
			}
		  ],
		  "name": "transferFrom",
		  "outputs": [
			{
			  "name": "",
			  "type": "bool"
			}
		  ],
		  "payable": false,
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [],
		  "name": "decimals",
		  "outputs": [
			{
			  "name": "",
			  "type": "uint8"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [
			{
			  "name": "_owner",
			  "type": "address"
			}
		  ],
		  "name": "balanceOf",
		  "outputs": [
			{
			  "name": "balance",
			  "type": "uint256"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [],
		  "name": "symbol",
		  "outputs": [
			{
			  "name": "",
			  "type": "string"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "constant": false,
		  "inputs": [
			{
			  "name": "_to",
			  "type": "address"
			},
			{
			  "name": "_value",
			  "type": "uint256"
			}
		  ],
		  "name": "transfer",
		  "outputs": [
			{
			  "name": "",
			  "type": "bool"
			}
		  ],
		  "payable": false,
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [
			{
			  "name": "_owner",
			  "type": "address"
			},
			{
			  "name": "_spender",
			  "type": "address"
			}
		  ],
		  "name": "allowance",
		  "outputs": [
			{
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "payable": true,
		  "stateMutability": "payable",
		  "type": "fallback"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "name": "owner",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "name": "spender",
			  "type": "address"
			},
			{
			  "indexed": false,
			  "name": "value",
			  "type": "uint256"
			}
		  ],
		  "name": "Approval",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "name": "from",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "name": "to",
			  "type": "address"
			},
			{
			  "indexed": false,
			  "name": "value",
			  "type": "uint256"
			}
		  ],
		  "name": "Transfer",
		  "type": "event"
		}
	  ]`
	var ERC721_String = `[
		{
		  "anonymous": false,
		  "inputs": [{"indexed": true, "internalType": "address", "name": "owner", "type": "address"}, {
			"indexed": true,
			"internalType": "address",
			"name": "approved",
			"type": "address"
		  }, {"indexed": true, "internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "Approval",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [{"indexed": true, "internalType": "address", "name": "owner", "type": "address"}, {
			"indexed": true,
			"internalType": "address",
			"name": "operator",
			"type": "address"
		  }, {"indexed": false, "internalType": "bool", "name": "approved", "type": "bool"}],
		  "name": "ApprovalForAll",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [{"indexed": true, "internalType": "address", "name": "from", "type": "address"}, {
			"indexed": true,
			"internalType": "address",
			"name": "to",
			"type": "address"
		  }, {"indexed": true, "internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "Transfer",
		  "type": "event"
		},
		{
		  "inputs": [{"internalType": "address", "name": "to", "type": "address"}, {
			"internalType": "uint256",
			"name": "tokenId",
			"type": "uint256"
		  }], "name": "approve", "outputs": [], "stateMutability": "nonpayable", "type": "function"
		},
		{
		  "constant": true,
		  "inputs": [],
		  "name": "totalSupply",
		  "outputs": [
			{
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "payable": false,
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "owner", "type": "address"}],
		  "name": "balanceOf",
		  "outputs": [{"internalType": "uint256", "name": "balance", "type": "uint256"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "getApproved",
		  "outputs": [{"internalType": "address", "name": "operator", "type": "address"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "owner", "type": "address"}, {
			"internalType": "address",
			"name": "operator",
			"type": "address"
		  }],
		  "name": "isApprovedForAll",
		  "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "name",
		  "outputs": [{"internalType": "string", "name": "", "type": "string"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "ownerOf",
		  "outputs": [{"internalType": "address", "name": "owner", "type": "address"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "from", "type": "address"}, {
			"internalType": "address",
			"name": "to",
			"type": "address"
		  },
			{"internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "safeTransferFrom",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "from", "type": "address"}, {
			"internalType": "address",
			"name": "to",
			"type": "address"
		  },
			{"internalType": "uint256", "name": "tokenId", "type": "uint256"}, {
			  "internalType": "bytes",
			  "name": "data",
			  "type": "bytes"
			}], "name": "safeTransferFrom", "outputs": [], "stateMutability": "nonpayable", "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "operator", "type": "address"}, {
			"internalType": "bool",
			"name": "_approved",
			"type": "bool"
		  }], "name": "setApprovalForAll", "outputs": [], "stateMutability": "nonpayable", "type": "function"
		},
		{
		  "inputs": [{"internalType": "bytes4", "name": "interfaceId", "type": "bytes4"}],
		  "name": "supportsInterface",
		  "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "symbol",
		  "outputs": [{"internalType": "string", "name": "", "type": "string"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "tokenURI",
		  "outputs": [{"internalType": "string", "name": "", "type": "string"}],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [{"internalType": "address", "name": "from", "type": "address"}, {
			"internalType": "address",
			"name": "to",
			"type": "address"
		  }, {"internalType": "uint256", "name": "tokenId", "type": "uint256"}],
		  "name": "transferFrom",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		}
	  ]`
	var ERC1155_String = `[
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "operator",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "bool",
        "name": "approved",
        "type": "bool"
      }
    ],
    "name": "ApprovalForAll",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "operator",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256[]",
        "name": "ids",
        "type": "uint256[]"
      },
      {
        "indexed": false,
        "internalType": "uint256[]",
        "name": "values",
        "type": "uint256[]"
      }
    ],
    "name": "TransferBatch",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "operator",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "id",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "TransferSingle",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "string",
        "name": "value",
        "type": "string"
      },
      {
        "indexed": true,
        "internalType": "uint256",
        "name": "id",
        "type": "uint256"
      }
    ],
    "name": "URI",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "id",
        "type": "uint256"
      }
    ],
    "name": "balanceOf",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address[]",
        "name": "accounts",
        "type": "address[]"
      },
      {
        "internalType": "uint256[]",
        "name": "ids",
        "type": "uint256[]"
      }
    ],
    "name": "balanceOfBatch",
    "outputs": [
      {
        "internalType": "uint256[]",
        "name": "",
        "type": "uint256[]"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "operator",
        "type": "address"
      }
    ],
    "name": "isApprovedForAll",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "internalType": "uint256[]",
        "name": "ids",
        "type": "uint256[]"
      },
      {
        "internalType": "uint256[]",
        "name": "amounts",
        "type": "uint256[]"
      },
      {
        "internalType": "bytes",
        "name": "data",
        "type": "bytes"
      }
    ],
    "name": "safeBatchTransferFrom",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "id",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "internalType": "bytes",
        "name": "data",
        "type": "bytes"
      }
    ],
    "name": "safeTransferFrom",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "operator",
        "type": "address"
      },
      {
        "internalType": "bool",
        "name": "approved",
        "type": "bool"
      }
    ],
    "name": "setApprovalForAll",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes4",
        "name": "interfaceId",
        "type": "bytes4"
      }
    ],
    "name": "supportsInterface",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "id",
        "type": "uint256"
      }
    ],
    "name": "uri",
    "outputs": [
      {
        "internalType": "string",
        "name": "",
        "type": "string"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`
	contract, client = cn.init_web3(ERC20_String)
	_, err := contract.Call("totalSupply")
	if err != nil {
		contract, client = cn.init_web3(ERC721_String)
		_, err = contract.Call("ownerOf", big.NewInt(1))
		if err != nil {
			contract, client = cn.init_web3(ERC1155_String)
			_, err = contract.Call("uri", big.NewInt(1))
			if err != nil {
				// @TODO: digerleri de yapilacak
			}
			return contract, client
		}
		return contract, client

	}

	return contract, client
}
