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
	server        string
	apiToken      APIToken
	applicationId string
}

func NewClient(server, apiToken, applicationId string) *ChirpStackClient {
	return &ChirpStackClient{
		server:        server,
		apiToken:      APIToken(apiToken),
		applicationId: applicationId,
	}
}

func (c *ChirpStackClient) GetDeviceList(ctx context.Context, limit, offset uint32) (int, []model.DeviceItem, error) {
	var (
		total int
		list  []model.DeviceItem
	)
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(c.apiToken),
		grpc.WithInsecure(), // remove this when using TLS
	}
	// connect to the gRPC server
	conn, err := grpc.Dial(c.server, dialOpts...)
	if err != nil {
		panic(err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	client := api.NewDeviceServiceClient(conn)
	logrus.Info("applicationId", c.applicationId, limit, offset)
	resp, err := client.List(context.Background(), &api.ListDevicesRequest{
		Limit:         limit,
		Offset:        offset,
		ApplicationId: c.applicationId,
	})

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
	logrus.Error("数据:", list)
	return total, list, nil
}
