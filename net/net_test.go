/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package net

import (
	"log"
	"testing"
)

func TestIP(t *testing.T) {
	log.Println(ExternalIP())
}

func TestIPV6(t *testing.T) {
	log.Println(LocalIPv6s())
	log.Println(IPv6s())
}

func TestCommonIP(t *testing.T) {
	log.Println(CommonIPV4())
	log.Println(CommonIPv6())
}
