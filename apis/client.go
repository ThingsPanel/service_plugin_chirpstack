package apis

import (
	"context"
	"fmt"
	"github.com/chirpstack/chirpstack/api/go/v4/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"plugin_chirpstack/model"
)

var (
	// This must point to the API interface
	server = "104.156.140.42:8080"

	// The API token (retrieved using the web-interface)
	apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJjaGlycHN0YWNrIiwiaXNzIjoiY2hpcnBzdGFjayIsInN1YiI6ImFkMmMxODRhLTAxNjAtNGUyNi1iYTBkLWQzMzM4OWZhYTFiMCIsInR5cCI6ImtleSJ9.IiODJrVcbeVTtoRTW_jaqmaOP0VHY7rrXEOEOJfvPOU"
)

type APIToken string

func (a APIToken) GetRequestMetadata(ctx context.Context, url ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", a),
	}, nil
}

func (a APIToken) RequireTransportSecurity() bool {
	return false
}

type ChirpStackClient struct {
	client api.DeviceServiceClient
}

func NewClient(server, apiToken string) *ChirpStackClient {
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(APIToken(apiToken)),
		grpc.WithInsecure(), // remove this when using TLS
	}

	// connect to the gRPC server
	conn, err := grpc.Dial(server, dialOpts...)
	if err != nil {
		panic(err)
	}
	// define the DeviceService client
	return &ChirpStackClient{
		client: api.NewDeviceServiceClient(conn),
	}

}

func (c *ChirpStackClient) GetDeviceList(ctx context.Context, applicationId string, limit, offset uint32) (int, []model.DeviceItem, error) {
	var (
		total int
		list  []model.DeviceItem
	)
	logrus.Info("applicationId", applicationId, limit, offset)
	resp, err := c.client.List(context.Background(), &api.ListDevicesRequest{
		Limit:         limit,
		Offset:        offset,
		ApplicationId: applicationId,
	})
	logrus.Error(resp)
	if err != nil {
		return total, list, err
	}
	total = int(resp.TotalCount)
	for _, v := range resp.Result {
		list = append(list, model.DeviceItem{
			DeviceNumber: v.DevEui,
			DeviceName:   v.Name,
			Description:  v.DeviceProfileName,
		})
	}
	return total, list, nil
}
