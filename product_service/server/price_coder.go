package main

import (
	"math"
	"strconv"
)

// PriceCoder - code price to int
type PriceCoder struct {
	base uint8
}

// NewPriceCoder - make new price coder, base - digits after point
func NewPriceCoder(base uint8) *PriceCoder {
	return &PriceCoder{base: base}
}

// Encode - convert string number as int64
func (pc *PriceCoder) Encode(price string) (int64, error) {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, err
	}

	return int64(math.RoundToEven(p * math.Pow10(int(pc.base)))), nil
}
