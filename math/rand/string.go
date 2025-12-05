/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package rand

import "math/rand/v2"

func Chinese() string {
	runes := make([]rune, 5)
	for i := range runes {
		runes[i] = ChineseChar()
	}
	return string(runes)
}

func ChineseChar() rune {
	r := rand.N(500)
	return rune(r + 19968)
}

const codes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func English() string {
	bytes := make([]byte, 5)
	for i := range bytes {
		bytes[i] = codes[rand.N[int](52)]
	}
	return string(bytes)
}
