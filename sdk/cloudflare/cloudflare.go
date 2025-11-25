package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hopeio/gox/log"
	netx "github.com/hopeio/gox/net"
)

func DDNSV6(ctx context.Context, api *cloudflare.API, zoneID, recordID string) error {
	resourceContainer := cloudflare.ResourceContainer{Type: cloudflare.ZoneType, Identifier: zoneID, Level: cloudflare.ZoneRouteLevel}
	record, err := api.GetDNSRecord(ctx, &resourceContainer, recordID)
	if err != nil {
		return fmt.Errorf("failed to get DNS record: %v", err)
	}
	addresses, err := netx.IPv6Addresses()
	if err != nil {
		return err
	}
	if len(addresses) == 0 || strings.HasPrefix(addresses[0], "fe80") {
		return fmt.Errorf("no IPv6 addresses found")
	}
	currentIP := addresses[0]
	if record.Content == currentIP {
		log.Printf("DNS record for %s (%s) is already up-to-date with IP %s\n", record.Name, record.Type, currentIP)
		return nil
	}
	proxied := true
	updatedRecord := cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		Content: currentIP,
		Proxied: &proxied,
		TTL:     1,
	}

	_, err = api.UpdateDNSRecord(ctx, &resourceContainer, updatedRecord)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}
	return nil
}
