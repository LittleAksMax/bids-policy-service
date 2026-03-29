package repository

import "go.mongodb.org/mongo-driver/bson/primitive"

type Policy struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Marketplace string             `bson:"marketplace" json:"marketplace"`
	Name        string             `bson:"name" json:"name"`
	Script      string             `bson:"script" json:"script"`
}

const (
	MpUK = "UK"
	MpDE = "DE"
	MpFR = "FR"
	MpIT = "IT"
	MpES = "ES"
	MpUS = "US"
	MpCA = "CA"
	MpMX = "MX"
	MpBR = "BR"
	MpAE = "AE"
	MpBE = "BE"
	MpEG = "EG"
	MpIE = "IE"
	MpIN = "IN"
	MpNL = "NL"
	MpPL = "PL"
	MpSA = "SA"
	MpSE = "SE"
	MpTR = "TR"
	MpZA = "ZA"
	MpAU = "AU"
	MpJP = "JP"
	MpSG = "SG"
)

func IsValidMarketplace(marketplace string) bool {
	switch marketplace {
	case MpUK, MpDE, MpFR, MpIT, MpES, MpUS, MpCA, MpMX,
		MpBR, MpAE, MpBE, MpEG, MpIE, MpIN, MpNL, MpPL, MpSA, MpSE, MpTR, MpZA, MpAU, MpJP, MpSG:
		return true
	default:
		return false
	}
}
