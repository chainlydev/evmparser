package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MyNumber uint64
type TokenData struct {
	TimedPrice      float64              `bson:"time_price,omitempty"`
	Time            primitive.DateTime   `bson:"time,omitempty"`
	TotalVolume     primitive.Decimal128 `bson:"total_volume,omitempty"`
	Tags            []string             `bson:"tags,omitempty"`
	Deployer        string               `bson:"deployer,omitempty"`
	DeployerAddress string               `bson:"deployer_address,omitempty"`
	Social          map[string]string    `bson:"social,omitempty"`
	Decimal         any                  `bson:"decimal,omitempty"`
	Icon            string               `bson:"icon,omitempty"`
	Name            string               `bson:"name,omitempty"`
	Symbol          string               `bson:"short,omitempty"`
	Address         string               `bson:"address,omitempty"`
	TokenType       string               `bson:"token_type,omitempty"`
	ProxyContract   bool                 `bson:"proxy,omitempty"`
	ProxyAddress    string               `bson:"proxy_address,omitempty"`
	Chain           int                  `bson:"chain,omitempty"`
	ERCType         string               `bson:"erc,omitempty"`
}

type TokenInfo struct {
	Address     string               `bson:"address"`
	Name        string               `bson:"name"`
	Symbol      string               `bson:"symbol"`
	Decimal     any                  `bson:"decimal"`
	TotalSupply primitive.Decimal128 `bson:"total_supply"`
}
