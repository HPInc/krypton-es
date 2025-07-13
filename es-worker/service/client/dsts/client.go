// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package dstsclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	pb "github.com/HPInc/krypton-es/es-worker/protos/dsts"
	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	"github.com/HPInc/krypton-es/es-worker/service/config"
	"github.com/HPInc/krypton-es/es-worker/service/structs"
	"go.uber.org/zap"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dstsProtocolVersion = "v1"
	retryCount          = 3
	retryWaitDuration   = time.Second * 3
	operationTimeout    = time.Second * 3
)

var (
	eswLogger *zap.Logger
)

type DeviceDetails struct {
	// this id is allocated by es and sent over
	// as part of unenroll request
	UnenrollId string `json:"unenroll_id"`
	// request id from es for trace
	RequestId string `json:"request_id"`
	TenantId  string `json:"tenant_id"`
	DeviceId  string `json:"device_id"`
	// this type helps in queue multiplexing
	// as we overload enroll queue for unenroll
	Type string `json:"type"`
}

func Start(logger *zap.Logger) (*structs.DSTSClient, error) {
	var err error
	eswLogger = logger
	c := structs.DSTSClient{}
	addr := fmt.Sprintf("%s:%d", config.Settings.DSTS.Host, config.Settings.DSTS.Port)
	c.Conn, err = connectWithRetry(addr)
	if err != nil {
		eswLogger.Error(
			"could not connect to dsts service",
			zap.Error(err))
		return nil, err
	}
	eswLogger.Info("connected to DSTS", zap.String("address: ", addr))
	c.Client = pb.NewDeviceSTSClient(c.Conn)
	return &c, Ping(&c)
}

func connectWithRetry(addr string) (*grpc.ClientConn, error) {
	var c *grpc.ClientConn
	var err error
	for i := 0; i < retryCount; i++ {
		c, err = grpc.Dial(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		eswLogger.Error("could not connect to ca service",
			zap.Error(err))
		time.Sleep(retryWaitDuration)
	}
	return c, err
}

func Ping(c *structs.DSTSClient) error {
	var err error
	for i := 0; i < retryCount; i++ {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(),
			operationTimeout)
		defer cancelFunc()
		_, err = c.Client.Ping(ctx, &pb.PingRequest{Message: "ping"})
		if err == nil {
			eswLogger.Info("successful ping to DSTS")
			break
		}
		eswLogger.Error("could not ping dsts server", zap.Error(err))
		time.Sleep(retryWaitDuration)
	}
	return err
}

func CreateDevice(c *structs.DSTSClient, dc *caclient.DeviceCertificate) error {
	certBytes, err := base64.StdEncoding.DecodeString(dc.Certificate)
	if err != nil {
		eswLogger.Error("invalid device cert", zap.Error(err))
		return err
	}

	createRequest := &pb.CreateDeviceRequest{
		Header:            newDstsProtocolHeader(dc.RequestId),
		Version:           dstsProtocolVersion,
		Tid:               dc.TenantId,
		DeviceId:          dc.DeviceId,
		DeviceCertificate: certBytes,
		ManagementService: dc.ManagementService,
		HardwareHash:      dc.HardwareHash,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	response, err := c.Client.CreateDevice(ctx, createRequest)
	if err != nil {
		eswLogger.Error("error creating device", zap.Error(err))
		return err
	}
	if response.Header.Status != uint32(codes.OK) {
		return fmt.Errorf(
			"%s: %d", "dsts create device error: ",
			response.Header.Status)
	}
	return nil
}

func UpdateDevice(c *structs.DSTSClient, dc *caclient.DeviceCertificate) error {
	certBytes, err := base64.StdEncoding.DecodeString(dc.Certificate)
	if err != nil {
		eswLogger.Error("invalid device cert", zap.Error(err))
		return err
	}

	updateRequest := &pb.UpdateDeviceRequest{
		Header:   newDstsProtocolHeader(dc.RequestId),
		Version:  dstsProtocolVersion,
		Tid:      dc.TenantId,
		DeviceId: dc.DeviceId,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"certificate"},
		},
		Update: &pb.DeviceUpdates{
			DeviceCertificate: certBytes},
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	response, err := c.Client.UpdateDevice(ctx, updateRequest)
	if err != nil {
		eswLogger.Error("error updating device", zap.Error(err))
		return err
	}
	if response.Header.Status != uint32(codes.OK) {
		return fmt.Errorf(
			"%s: %d", "dsts update device error: ",
			response.Header.Status)
	}
	return nil
}

func DeleteDevice(c *structs.DSTSClient, dd *DeviceDetails) error {
	deleteRequest := &pb.DeleteDeviceRequest{
		Header:   newDstsProtocolHeader(dd.RequestId),
		Version:  dstsProtocolVersion,
		Tid:      dd.TenantId,
		DeviceId: dd.DeviceId,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	response, err := c.Client.DeleteDevice(ctx, deleteRequest)
	if err != nil {
		eswLogger.Error("error deleting device", zap.Error(err))
		return err
	}
	if response.Header.Status != uint32(codes.OK) {
		return fmt.Errorf(
			"%s: %d", "dsts delete device error: ",
			response.Header.Status)
	}
	return nil
}

func newDstsProtocolHeader(requestId string) *pb.DstsRequestHeader {
	return &pb.DstsRequestHeader{
		ProtocolVersion: "v1",
		RequestId:       requestId,
		RequestTime:     timestamppb.Now(),
	}
}
