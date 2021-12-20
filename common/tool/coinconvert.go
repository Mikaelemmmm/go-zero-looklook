package tool

import "github.com/shopspring/decimal"

var oneHundredDecimal decimal.Decimal = decimal.NewFromInt(100)

//分转元
func Fen2Yuan(fen int64) float64 {
	y, _ := decimal.NewFromInt(fen).Div(oneHundredDecimal).Truncate(2).Float64()
	return y
}

//元转分
func Yuan2Fen(yuan float64) int64 {

	f, _ := decimal.NewFromFloat(yuan).Mul(oneHundredDecimal).Truncate(0).Float64()
	return int64(f)

}
