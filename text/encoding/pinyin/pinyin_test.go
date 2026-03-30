/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package py

import (
	"testing"
)

func TestPinyin(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "N",
			want: `n`,
		},
		{
			name: "【",
			want: `【`,
		},
		{
			name: "[",
			want: `[`,
		},
		{
			name: ",",
			want: `,`,
		},
		{
			name: "。",
			want: `。`,
		},
		{
			name: "中",
			want: `z`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FistLetter(test.name); got != test.want {
				t.Fatalf(" got: %s\nwant: %s\n", got, test.want)
			}
		})
	}

}
