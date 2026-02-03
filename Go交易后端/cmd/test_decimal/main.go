package main

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
)

func main() {
	// 测试decimal转换为float64
	testCases := []string{
		"1110.00000000",
		"10000.00000000",
		"0.00000000",
		"1234.56789012",
	}

	for _, tc := range testCases {
		dec, err := decimal.NewFromString(tc)
		if err != nil {
			fmt.Printf("解析失败: %s, err=%v\n", tc, err)
			continue
		}

		// 测试Float64转换
		f, err := dec.Float64()
		if err != nil {
			fmt.Printf("Float64转换失败: %s -> err=%v\n", tc, err)
		} else {
			fmt.Printf("转换成功: %s -> %.4f\n", tc, f)
		}
	}
}
