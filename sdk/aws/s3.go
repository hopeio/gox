/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package aws

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hopeio/gox/log"
)

func upload(cli *s3.Client) {
	fp, _ := os.Open("s3_test.go")

	defer fp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	cli.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("hatlonely"),
		Key:    aws.String("hoper/s3_test.go"),
		Body:   fp,
	})

}

func download(service *s3.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	out, _ := service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("hatlonely"),
		Key:    aws.String("hoper/s3_test.go"),
	})

	defer out.Body.Close()
	scanner := bufio.NewScanner(out.Body)
	for scanner.Scan() {
		log.Info(scanner.Text())
	}
}

// 目录遍历
func listObjects(service *s3.Client) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	service.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String("hatlonely"),
		Prefix: aws.String("hoper/"),
	})
}
