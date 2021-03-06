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
	"fmt"
	"fraud-check-worker/configuration"
	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
	"math/rand"
	"time"
)

const (
	OkColor    = "\033[1;36m%s\033[0m"
	HoldColor  = "\033[1;33m%s\033[0m"
	FraudColor = "\033[1;31m%s\033[0m"
)

func main() {

	log.Println("Starting...")

	rand.Seed(time.Now().UTC().UnixNano())

	c := configuration.New()
	broker := c.GetBrokerEndpoint()
	log.Println("Using broker:", broker)

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         broker,
		UsePlaintextConnection: true})
	if err != nil {
		panic(err)
	}

	jobWorker := zbClient.NewJobWorker().JobType("fraud_check_task").Handler(handleJob).Open()
	defer jobWorker.Close()

	jobWorker.AwaitClose()
}

func handleJob(client worker.JobClient, job entities.Job) {

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		_ = failJob(client, job)
		return
	}

	fraudy := rand.Intn(100) >= 80
	manual := rand.Intn(100) >= 50
	variables["fraud"] = fraudy
	variables["fraud_manual"] = fraudy && manual

	err = completeJob(client, job, variables)
	if err != nil {
		log.Println("order", variables["order_id"], "| MESSAGE DROPPED")
	} else {
		color := OkColor
		msg := "false"
		if fraudy {
			color = FraudColor
			msg = "true"
			if manual {
				color = HoldColor
				msg = "HOLD"
			}
		}
		log.Println("order", variables["order_id"], "| fraudy:", fmt.Sprintf(color, msg))
	}
}

func failJob(client worker.JobClient, job entities.Job) error {
	ctx := context.Background()
	_, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
	return err
}

func completeJob(client worker.JobClient, job entities.Job, variables map[string]interface{}) error {
	ctx := context.Background()
	request, _ := client.NewCompleteJobCommand().JobKey(job.GetKey()).VariablesFromMap(variables)
	_, err := request.Send(ctx)
	if err != nil {
		// failed to set the updated variables
		_ = failJob(client, job)
	}
	return err
}
