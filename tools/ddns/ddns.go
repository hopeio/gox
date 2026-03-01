package main

import (
	"context"
	"flag"
	"github.com/cloudflare/cloudflare-go"
	"github.com/hopeio/gox/log"
	neti "github.com/hopeio/gox/net"
	"github.com/hopeio/gox/scheduler/retry"
	"slices"
	"time"
)

// docker run --restart=always --name=ddns --net=host -d jybl/ddns --token=<token> --domain=<domain> --name=<name> --name=<name>
func main() {
	var token, domain string
	var names []string
	flag.StringVar(&token, "token", "", "Cloudflare API Token")
	flag.StringVar(&domain, "domain", "", "domain")
	flag.Func("name", "name", func(s string) error {
		names = append(names, s)
		return nil
	})
	flag.Parse()
	if token == "" {
		log.Fatal("token is empty")
	}
	if domain == "" {
		log.Fatal("domain is empty")
	}
	log.Info("domain: ", domain)
	if len(names) == 0 {
		log.Fatal("names is empty")
	}
	log.Infof("names: %v", names)
	// 初始化 Cloudflare API 客户端
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Fatalf("failed to initialize Cloudflare API: %v", err)
	}
	// 获取 Zone ID
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatalf("failed to get Zone ID: %v", err)
	}
	ipv6s, err := neti.IPv6s()
	if err != nil || len(ipv6s) == 0 {
		log.Fatalf("failed to get IPv6 ipv6s: %v", err)
	}
	lastIP := ipv6s[0].String()
	log.Infof("ipv6: %v", lastIP)
	needUpdateRecord := make([]cloudflare.UpdateDNSRecordParams, 0, 2)
	// 获取 DNS 记录
	ctx := context.Background()
	resourceContainer := cloudflare.ResourceContainer{Type: cloudflare.ZoneType, Identifier: zoneID, Level: cloudflare.ZoneRouteLevel}
	records, _, err := api.ListDNSRecords(ctx, &resourceContainer, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		log.Fatalf("failed to get DNS records: %v", err)
	}
	for _, record := range records {
		if record.Type == "AAAA" {
			if slices.Contains(names, record.Name) {
				needUpdateRecord = append(needUpdateRecord, cloudflare.UpdateDNSRecordParams{
					Type:    record.Type,
					Name:    record.Name,
					Content: record.Content,
					ID:      record.ID,
					TTL:     1,
				})
			}
		}
	}
	for _, record := range needUpdateRecord {
		if record.Content != lastIP {
			record.Content = lastIP
			retry.Run(func(int) bool {
				_, err = api.UpdateDNSRecord(ctx, &resourceContainer, record)
				if err != nil {
					log.Error(err)
				} else {
					log.Infof("init update record %s to %s", record.Name, record.Content)
				}
				return err != nil
			})

		}
	}
	updateTime := time.Now()
	timer := time.NewTimer(time.Hour)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			ipv6s, err = neti.IPv6s()
			if err != nil || len(ipv6s) == 0 {
				log.Errorf("failed to get IPv6 ipv6s: %v", err)
				timer.Reset(time.Second)
				continue
			}
			if lastIP != ipv6s[0].String() {
				lastIP = ipv6s[0].String()
				for _, record := range needUpdateRecord {
					record.Content = lastIP
					retry.Run(func(int) bool {
						_, err = api.UpdateDNSRecord(ctx, &resourceContainer, record)
						if err != nil {
							log.Error(err)
						} else {
							log.Infof("update record %s to %s", record.Name, record.Content)
						}
						return err != nil
					})

				}
				updateTime = time.Now()
				timer.Reset(time.Hour)
				continue
			}
			t := time.Now().Sub(updateTime)
			if t > time.Hour*24 {
				timer.Reset(time.Second * 30)
			} else if t > time.Hour*12 {
				timer.Reset(time.Minute)
			} else if t > time.Hour*6 {
				timer.Reset(time.Hour / 2)
			} else {
				timer.Reset(time.Hour)
			}
		}
	}
}
