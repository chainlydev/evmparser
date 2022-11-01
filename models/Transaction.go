package models

import (
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	Index            int                  `json:"index" bson:"index"`
	From             string               `json:"from" bson:"From"`
	To               string               `json:"to" bson:"To"`
	Hash             string               `json:"hash" bson:"Hash"`
	Block            int64                `json:"block" bson:"Block"`
	Chain            int                  `json:"chain" bson:"Chain"`
	TotalCost        float64              `json:"total_cost" bson:"total_cost"`
	TransactionCost  float64              `json:"transaction_cost" bson:"transaction_cost"`
	Logs             []Logs               `json:"logs" bson:"logs"`
	IsNFT            bool                 `json:"is_nft" bson:"is_nft"`
	RawTransaction   *types.Transaction   `json:"raw_transaction" bson:"raw_transaction"`
	RawReciept       *types.Receipt       `json:"raw_reciept" bson:"raw_reciept"`
	IsSwap           bool                 `json:"is_swap" bson:"is_swap"`
	IsSwapNft        bool                 `json:"is_swap_nft" bson:"is_swap_nft"`
	ERCType          string               `json:"erc_type" bson:"erc_type"`
	Value            primitive.Decimal128 `json:"value" bson:"value"`
	InteractedTokens []string             `json:"interacted_tokens" bson:"interacted_tokens"`
	Swap             *Swap                `json:"swap" bson:"swap"`
	NFT              []*NFT               `json:"nft" bson:"nft"`
	Date             primitive.DateTime   `json:"date" bson:"date"`
}
