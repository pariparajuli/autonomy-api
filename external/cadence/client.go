package cadence

import (
	"context"

	"github.com/spf13/viper"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
)

const (
	ClientName     = "autonomy-worker"
	CadenceService = "cadence-frontend"
)

type CadenceClient struct {
	client client.Client
}

func BuildCadenceServiceClient(hostPort string) workflowserviceclient.Interface {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(ClientName))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: ch.NewSingleOutbound(hostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(CadenceService))
}

func NewClient() *CadenceClient {
	service := BuildCadenceServiceClient(viper.GetString("cadence.conn"))

	return &CadenceClient{
		client: client.NewClient(
			service,
			viper.GetString("cadence.domain"),
			&client.Options{
				MetricsScope:  tally.NoopScope,
				DataConverter: NewMsgPackDataConverter(),
			},
		),
	}
}

func (c *CadenceClient) StartWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (*workflow.Execution, error) {
	return c.client.StartWorkflow(ctx, options, workflow, args...)
}

func (c *CadenceClient) SignalWithStartWorkflow(ctx context.Context,
	workflowID string, signalName string, signalArg interface{},
	options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (*workflow.Execution, error) {
	return c.client.SignalWithStartWorkflow(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
}
