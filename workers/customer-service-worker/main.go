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
	"customer-service-worker/configuration"
	"encoding/json"
	"fmt"
	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
)

const (
	SuccessColor = "\033[1;36m%s\033[0m"
	FailColor    = "\033[1;31m%s\033[0m"
)

func main() {

	log.Println("Starting order-service-worker...")

	c := configuration.New()
	broker := c.GetBrokerEndpoint()
	log.Println("Using broker:", broker)

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         broker,
		UsePlaintextConnection: true})
	if err != nil {
		panic(err)
	}

	dispatchSuccessWorker := zbClient.NewJobWorker().JobType("notify_dispatch_success_task").Handler(handleAllocationSuccess).Open()
	defer dispatchSuccessWorker.Close()

	dispatchFailWorker := zbClient.NewJobWorker().JobType("notify_dispatch_fail_task").Handler(handleAllocationFail).Open()
	defer dispatchFailWorker.Close()

	dispatchSuccessWorker.AwaitClose()
	dispatchFailWorker.AwaitClose()

}

func handleAllocationSuccess(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	orderId := variables["order_id"]
	lines, ok := variables["dispatch_list"].(map[string]interface{})
	if !ok {
		// failed to set the updated variables
		failJob(client, job)
		return
	}
	lineIds := toJson(lines)
	msg := fmt.Sprint("Joy! Your cat will be happy and fat and will not starve today."+
		" Dispatch success for lines ", lineIds, " in order ", orderId)
	log.Println("order", orderId, "| message:\n", fmt.Sprintf(SuccessColor, msg))
}

func handleAllocationFail(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	orderId := variables["order_id"]
	lines, ok := variables["fail_list"].(map[string]interface{})
	if !ok {
		// failed to set the updated variables
		failJob(client, job)
		return
	}
	lineIds := toJson(lines)

	msg := fmt.Sprint("Commiserations! I'm sorry, your cat is going to get thin today."+
		" Dispatch fail for lines ", lineIds, " in order ", orderId)
	log.Println("order", orderId, "| message:\n", fmt.Sprintf(FailColor, msg))

}

func failJob(client worker.JobClient, job entities.Job) {
	ctx := context.Background()
	_, _ = client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
}

func toJson(payload interface{}) string {
	str, err := json.Marshal(payload)
	if err != nil {
		return "<unknown>"
	}
	return string(str)
}
