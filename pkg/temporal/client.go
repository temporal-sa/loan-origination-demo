package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	TaskQueue = "loan-origination-task-queue"
	Namespace = "default"
)

func NewClient() (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  "localhost:7233",
		Namespace: Namespace,
	})
}

func NewWorker(c client.Client) worker.Worker {
	return worker.New(c, TaskQueue, worker.Options{})
}
