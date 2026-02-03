package multibase

import (
	"errors"
	"math/big"
)

const base58Alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte
var ErrInvalidBase58 = errors.New("invalid base58")

func init() {
	for i := 0; i < len(base58Alphabet); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(base58Alphabet); i++ {
		decodeBase58Map[base58Alphabet[i]] = byte(i)
	}
}

// EncodeBase58 将字节数组编码为Base58字符串
func EncodeBase58(data []byte) string {
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
	base := big.NewInt(58)
	var result []byte

	// 进行进制转换
	for bigNum.Sign() > 0 {
		remainder := new(big.Int)
		bigNum.DivMod(bigNum, base, remainder)
		result = append(result, base58Alphabet[remainder.Int64()])
	}

	// 添加前导零对应的字符（Base58中用'1'表示前导零）
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	// 反转结果，因为我们是从低位开始计算的
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// DecodeBase58 将 Base58 编码的字符串解码为字节数组
func DecodeBase58(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}

	// 计算前导 '1' 的数量（对应原始数据的前导零字节）
	leadingOnes := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	// 跳过前导 '1'，处理剩余字符
	remaining := s[leadingOnes:]
	if remaining == "" {
		// 如果全是 '1'，返回相应数量的零字节
		return make([]byte, leadingOnes), nil
	}

	// 将 Base58 字符串转换为大整数
	bigNum := new(big.Int)
	base := big.NewInt(58)

	for i := 0; i < len(remaining); i++ {
		char := remaining[i]

		// 查找字符在 Base58 字符集中的索引
		index := int(decodeBase58Map[char])
		if index == 0xFF {
			return nil, ErrInvalidBase58
		}

		// bigNum = bigNum * 58 + index
		bigNum.Mul(bigNum, base)
		bigNum.Add(bigNum, big.NewInt(int64(index)))
	}

	// 转换为字节数组
	byteArray := bigNum.Bytes()

	// 添加前导零字节
	if leadingOnes > 0 {
		result := make([]byte, leadingOnes+len(byteArray))
		copy(result[leadingOnes:], byteArray)
		return result, nil
	}

	return byteArray, nil
}
