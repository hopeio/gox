/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package chromedp

import (
	"context"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	"github.com/hopeio/gox/log"
)

// Rewrite returns the result.
func Rewrite(ctx context.Context, url string, localPath string) chromedp.Action[chromedp.Void] {
	data, err := os.ReadFile(localPath)
	if err != nil {
		log.Errorf("failed read file: %v", err)
	}
	// chromedp 0.23 起 ListenTarget 换成 Listen（顺序保证不变），事件回调仍收 any。
	if err := chromedp.Listen(ctx, func(ev any) {
		if ev, ok := ev.(*fetch.EventRequestPaused); ok {
			go func() {
				// 0.23 的 fetch.FulfillRequest 从函数式 builder 改成 cdp.Command 值，
				// 参数走类型化 struct；Body 收原始字节（旧版自己 base64）。
				_, err := cdp.Call(ctx, chromedp.FromContext(ctx).Target, fetch.FulfillRequest, fetch.FulfillRequestParams{
					RequestID:       ev.RequestID,
					ResponseCode:    200,
					ResponseHeaders: []*fetch.HeaderEntry{{Name: "Content-Type", Value: "application/javascript"}},
					Body:            data,
				})
				if err != nil {
					log.Errorf("failed to run: %v", err)
				}
			}()
		}
	}); err != nil {
		log.Errorf("failed to listen: %v", err)
	}
	return func(ctx context.Context, t *chromedp.Target) (chromedp.Void, error) {
		_, err := cdp.Call(ctx, t, fetch.Enable, fetch.EnableParams{
			Patterns: []*fetch.RequestPattern{{URLPattern: url, RequestStage: fetch.RequestStageRequest}},
		})
		return chromedp.Void{}, err
	}
}
