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
	"encoding/json"
	"errors"
	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
	"math/rand"
	"time"
	"warehouse-worker/configuration"
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

	stockAllocationWorker := zbClient.NewJobWorker().JobType("allocate_stock_task").Handler(handleAllocateStockJob).Open()
	defer stockAllocationWorker.Close()

	createPickListAllocationWorker := zbClient.NewJobWorker().JobType("create_picklist_task").Handler(handleCreatePickListJob).Open()
	defer createPickListAllocationWorker.Close()

	createPickPackWorker := zbClient.NewJobWorker().JobType("pick_pack_task").Handler(handlePickPackJob).Open()
	defer createPickPackWorker.Close()

	stockAllocationWorker.AwaitClose()

}

func handleAllocateStockJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	allocationlist, faillist, err := allocateStock(variables["consignment"])
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	variables["allocation_list"] = allocationlist
	variables["fail_list"] = faillist
	variables["complete_allocation"] = len(faillist) <= 0

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	log.Println("order", variables["order_id"], "| allocation: success", toJson(allocationlist), " fail", toJson(faillist))
}

func handleCreatePickListJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	picklist, err := createPickList(variables["allocation_list"])
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	variables["pick_list"] = picklist

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	log.Println("order", variables["order_id"], "| created pick list:", toJson(picklist))
}

func handlePickPackJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	dispatchlist, err := createDispatchList(variables["pick_list"])
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	variables["dispatch_list"] = dispatchlist

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	log.Println("order", variables["order_id"], "| dispatched:", toJson(dispatchlist))
}

func allocateStock(consignment interface{}) (map[string]interface{}, map[string]interface{}, error) {
	consignmentMap, ok := consignment.(map[string]interface{})
	if !ok {
		return nil, nil, errors.New("unable to get the pick list")
	}
	allocationlist := make(map[string]interface{})
	faillist := make(map[string]interface{})
	for key, element := range consignmentMap {
		if rand.Intn(100) >= 3 {
			allocationlist[key] = element
		} else {
			faillist[key] = element
		}
	}
	return allocationlist, faillist, nil
}

func createPickList(allocationlist interface{}) (map[string]interface{}, error) {
	allocationMap, ok := allocationlist.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to get the allocation list")
	}
	return allocationMap, nil
}

func createDispatchList(picklist interface{}) (map[string]interface{}, error) {
	pickMap, ok := picklist.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to get the pick list")
	}
	return pickMap, nil
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
