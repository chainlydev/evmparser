package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Block struct {
	Number      uint64             `bson:"number,omitempty"`
	Hash        string             `bson:"hash,omitempty"`
	TxHash      string             `bson:"tx_hash,omitempty"`
	Difficulty  string             `bson:"difficulty,omitempty"`
	GasUsed     uint64             `bson:"gas_used,omitempty"`
	GasLimit    uint64             `bson:"gas_limit,omitempty"`
	Root        string             `bson:"root,omitempty"`
	ReceivedAt  primitive.DateTime `bson:"received_at,omitempty"`
	ReceiptHash string             `bson:"receipt_hash,omitempty"`
	ParentHash  string             `bson:"parent_hash,omitempty"`
	Bloom       string             `bson:"bloom,omitempty"`
	Time        primitive.DateTime `bson:"time,omitempty"`
	BaseFee     uint64             `bson:"base_fee,omitempty"`
	Coinbase    string             `bson:"coinbase,omitempty"`
	Nonce       string             `bson:"nonce,omitempty"`
}
type Topic struct {
	Hash     string `bson:"hash"`
	Function string `bson:"function"`
}
type AbiResponse struct {
	Address string `bson:"address" json:"address"`
	Abi     string `bson:"abi" json:"abi"`
}
type PriceHistory struct {
	Date   primitive.DateTime `bson:"date"`
	High   float64            `bson:"high"`
	Low    float64            `bson:"Low"`
	Open   float64            `bson:"open"`
	Close  float64            `bson:"close"`
	Symbol string             `bson:"symbol"`
	Source string             `bson:"source"`
}

type EthScanResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

type Values struct {
	Wei        string  `bson:"wei"`
	GWei       float64 `bson:"gwei"`
	Eth        float64 `bson:"eth"`
	TimedPrice float64 `bson:"time_price"`
	TotalValue float64 `bson:"total_value"`
}

type TransactionDetails struct {
	From              string               `bson:"from,omitempty"`
	To                string               `bson:"to,omitempty"`
	Block             Block                `bson:"block,omitempty"`
	Hash              string               `bson:"hash,omitempty"`
	ContractAddress   string               `bson:"contract_address,omitempty"`
	GasUsed           uint64               `bson:"gas_used,omitempty"`
	GasPrice          primitive.Decimal128 `bson:"gas_price,omitempty"`
	GasFeeCap         primitive.Decimal128 `bson:"gas_fee_cap,omitempty"`
	GasTipCap         primitive.Decimal128 `bson:"gas_tip_cap,omitempty"`
	CumulativeGasUsed uint64               `bson:"cumulative_gas_used,omitempty"`
	Status            uint64               `bson:"status,omitempty"`
	Cost              primitive.Decimal128 `bson:"cost,omitempty"`
	Type              uint8                `bson:"type,omitempty"`
	Value             primitive.Decimal128 `bson:"value,omitempty"`
	TransactionIndex  uint64               `bson:"transaction_index,omitempty"`
	Bloom             string               `bson:"bloom,omitempty"`
	Logs              []Logs               `bson:"logs,omitempty"`
	Values            *Values              `bson:"values,omitempty"`
	Token             *TokenData           `bson:"token,omitempty"`
	TransactionType   string               `bson:"transaction_type,omitempty"`
	ContractType      string               `bson:"contract_type,omitempty"`
	IsNft             bool                 `bson:"is_nft,omitempty"`
	IsSwap            bool                 `bson:"is_swap,omitempty"`
	Nft               map[string]any       `bson:"nft,omitempty"`
	Data              string               `bson:"data,omitempty"`
	Date              primitive.DateTime   `bson:"date,omitempty"`
	InteractedTokens  []TokenData          `bson:"interacted_tokens,omitempty"`
}
type TransactionInfo struct {
	From                   string               `bson:"from,omitempty"`
	Chain                  int                  `bson:"chain"`
	To                     string               `bson:"to,omitempty"`
	Block                  Block                `bson:"block,omitempty"`
	Hash                   string               `bson:"hash,omitempty"`
	Date                   primitive.DateTime   `bson:"date,omitempty"`
	Type                   uint8                `bson:"type,omitempty"`
	Value                  primitive.Decimal128 `bson:"value,omitempty"`
	BaseToken              *TokenData           `bson:"base_token,omitempty"`
	TransactionCost        primitive.Decimal128 `bson:"transaction_cost"`
	TransactionFee         primitive.Decimal128 `bson:"transaction_fee"`
	TransactionIndex       uint64               `bson:"transaction_index,omitempty"`
	Status                 uint64               `bson:"status,omitempty"`
	IsNft                  bool                 `bson:"is_nft,omitempty"`
	WaitData               bool                 `bson:"wait_data,omitempty"`
	Nft                    map[string]any       `bson:"nft,omitempty"`
	IsSwap                 bool                 `bson:"is_swap,omitempty"`
	Swap                   map[string]any       `bson:"swap,omitempty"`
	TotalCostUSD           uint64               `bson:"total_cost_usd,omitempty"`
	InteractedTokens       []TokenData          `bson:"interacted_tokens,omitempty"`
	TransactionAmountValue uint64               `bson:"transaction_amount_value"`
	Detail                 TransactionDetails   `bson:"detail"`
}
