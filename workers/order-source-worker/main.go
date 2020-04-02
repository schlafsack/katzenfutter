/*
 * MIT License
 *
 * Copyright (c) 2020 Tom Greasley
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
	"order-source-worker/configuration"
	"time"
)

func main() {

	log.Println("Starting...")

	c := configuration.New()
	broker := c.GetBrokerEndpoint()
	log.Println("Using broker:", broker)

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         broker,
		UsePlaintextConnection: true})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for {
		orderId := uuid.New().String()
		log.Println("New order ", orderId)
		variables := make(map[string]interface{})
		variables["order_id"] = orderId
		request, err := zbClient.NewCreateInstanceCommand().BPMNProcessId("order_process").
			LatestVersion().VariablesFromMap(variables)
		if err != nil {
			panic(err)
		}
		_, err = request.Send(ctx)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Duration(c.GetFrequency()) * time.Second)
	}
}
