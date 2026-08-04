/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package decimal

import (
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func Test_Dec(t *testing.T) {
	fmt.Printf("%#v", DecimalModel{exponent: 6})
}

func Test_Float(t *testing.T) {
	//var a = 0.1000000000000000055511151231257827021181583404541015625
	a := 0.11
	b := 0.1
	f1 := big.NewFloat(a)
	f2 := big.NewFloat(b)
	f3 := f2.Mul(f1, f2)
	data, _ := f3.MarshalText()
	fmt.Println(a*b, f3, string(data))
}

func Test_DB(t *testing.T) {
	t.Skip("requires live database connection")
	type DbTest struct {
		Id   uint64
		Dec  decimal.Decimal `gorm:"type:decimal(10,2)"`
		Time time.Time
	}
	db := &gorm.DB{}
	tx := db.Begin()
	/*	tx.DropTable(&DecTest{})
		tx.CreateTable(&DecTest{})*/
	d, err := decimal.NewFromString("0.1")
	if err != nil {
		t.Error(err)
	}
	var dec = DbTest{Dec: d}
	tx.Save(&dec)
	log.Println(dec.Id)
	var dec1 DbTest
	tx.First(&dec1)
	tx.Commit()
	fmt.Println(dec1)
}
