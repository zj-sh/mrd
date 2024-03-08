package util

import "github.com/shopspring/decimal"

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func Add[T Number](args ...T) float64 {
	var des []decimal.Decimal
	for _, item := range args {
		des = append(des, decimal.NewFromFloat(float64(item)))
	}
	result, _ := decimal.Sum(des[0], des[1:]...).Float64()
	return result
}
func Sub[T Number](args ...T) float64 {
	des := decimal.NewFromFloat(float64(args[0]))
	for _, item := range args[1:] {
		des = des.Sub(decimal.NewFromFloat(float64(item)))
	}
	result, _ := des.Float64()
	return result
}
func Mul[T Number](args ...T) float64 {
	des := decimal.NewFromFloat(float64(args[0]))
	for _, item := range args[1:] {
		des = des.Mul(decimal.NewFromFloat(float64(item)))
	}
	result, _ := des.Float64()
	return result
}
func Div[T Number](args ...T) float64 {
	des := decimal.NewFromFloat(float64(args[0]))
	for _, item := range args[1:] {
		des = des.Div(decimal.NewFromFloat(float64(item)))
	}
	result, _ := des.Float64()
	return result
}
func Round(num float64, round int32) float64 {
	d, _ := decimal.NewFromFloat(num).Round(round).Float64()
	return d
}
