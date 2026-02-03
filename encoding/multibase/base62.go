package multibase

import (
	"fmt"
	"math/big"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// 计算前导零的数量
	leadingZeros := 0
	for leadingZeros < len(data) && data[leadingZeros] == 0 {
		leadingZeros++
	}

	// 将字节转换为大整数
	bigNum := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	var result []byte

	for bigNum.Sign() > 0 {
		remainder := new(big.Int)
		bigNum.DivMod(bigNum, base, remainder)
		result = append(result, base62Alphabet[remainder.Int64()])
	}

	// 添加前导零对应的字符
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '0') // 前导零在Base62中对应'0'
	}

	// 反转结果，因为我们是从低位开始计算的
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

var base62DecodeMap [256]byte

func init() {
	for i := range base62DecodeMap {
		base62DecodeMap[i] = 255
	}
	for i, char := range base62Alphabet {
		base62DecodeMap[byte(char)] = byte(i)
	}
}

// DecodeBase62 将Base62编码的字符串解码为字节数组
func DecodeBase62(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}

	// 计算前导零的数量
	leadingZeros := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '0' {
			leadingZeros++
		} else {
			break
		}
	}

	// 跳过前导零，处理剩余字符
	remaining := s[leadingZeros:]
	if remaining == "" {
		// 如果全是零
		result := make([]byte, leadingZeros)
		return result, nil
	}

	// 将Base62字符串转换为大整数
	bigNum := new(big.Int)
	base := big.NewInt(62)

	for i := 0; i < len(remaining); i++ {
		char := remaining[i]
		var index byte
		switch {
		case char >= '0' && char <= '9':
			index = char - '0'
			bigNum.Mul(bigNum, base)
			bigNum.Add(bigNum, big.NewInt(int64(index)))
		case char >= 'A' && char <= 'Z':
			index = char - 'A' + 10
			bigNum.Mul(bigNum, base)
			bigNum.Add(bigNum, big.NewInt(int64(index)))
		case char >= 'a' && char <= 'z':
			index = char - 'a' + 10
			bigNum.Mul(bigNum, base)
			bigNum.Add(bigNum, big.NewInt(int64(index)))
		default:
			return nil, fmt.Errorf("invalid Base62 character: %c", char)
		}

		// bigNum = bigNum * 62 + index
		bigNum.Mul(bigNum, base)
		bigNum.Add(bigNum, big.NewInt(int64(index)))
	}

	// 转换为字节数组
	byteArray := bigNum.Bytes()

	// 添加前导零
	if leadingZeros > 0 {
		result := make([]byte, leadingZeros+len(byteArray))
		copy(result[leadingZeros:], byteArray)
		return result, nil
	}

	return byteArray, nil
}
