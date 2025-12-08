/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"testing"
	"time"

	jsonx "github.com/hopeio/gox/encoding/json"
)

type Foo struct {
	T1 Time[milliTime]
	T2 Time[nanoTime]
}

func TestType(t *testing.T) {
	foo := Foo{T1: Time[milliTime](time.Now()), T2: Time[nanoTime](time.Now())}
	data, _ := jsonx.Marshal(&foo)
	t.Log(string(data))
}

func TestTimestamp(t *testing.T) {
	t.Log(time.Unix(-62135596800, 0)) // 0001-01-01 08:00:00 +0800 CST
	t.Log(time.Unix(-62135596899, 0)) // 0001-01-01 07:58:21 +0800 CST
}

type Foo1 struct {
	T1 DateTime
	T2 Date
	T3 SecondTime
	T4 NanoTime
}

func TestUnionTime(t *testing.T) {

	foo := Foo1{T1: DateTime(time.Now().Unix()),
		T2: Date(time.Now().Unix()),
		T3: NewTime[secondTime](time.Now()),
		T4: NewTime[nanoTime](time.Now()),
	}
	data, _ := jsonx.Marshal(&foo)
	t.Log(string(data)) // {"T1":"2023-02-09 15:00:49","T2":"2023-02-09","T3":1675926049,"T4":1675926049057035300}
	data = []byte(`{"T1":"2023-02-09 15:00:49","T2":"2023-02-09","T3":1675926049,"T4":1675926049057035300}`)
	foo1 := Foo1{
		T3: SecondTime{},
		T4: NanoTime{},
	}
	jsonx.Unmarshal(data, &foo1)
	t.Log(foo1)
}

func TestTimeScan(t *testing.T) {
	var d Date
	date := time.Now()
	err := d.Scan(date)
	if err != nil {
		t.Error(err)
	}
	t.Log(d)
}
