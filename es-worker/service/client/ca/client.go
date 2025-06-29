// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package caclient

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	pb "github.com/HPInc/krypton-es/es-worker/protos/ca"
	"github.com/HPInc/krypton-es/es-worker/service/config"
	"github.com/HPInc/krypton-es/es-worker/service/structs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CaProtocolVersion = "v1"
	RetryCount        = 3
	RetryWaitDuration = time.Second * 3
	operationTimeout  = time.Second * 3

	CertificateTypeEnroll = "enroll"
	CertificateTypeRenew  = "renew_enroll"
)

// struct to hold certificate and device data
// main struct used in communication between queues
// es will see this struct via "enrolled" queue message
// at the end of a successful enroll
type DeviceCertificate struct {
	// request id for tracking
	RequestId string `json:"request_id"`
	// tenant id from original request
	TenantId string `json:"tenant_id"`
	// enroll id created and passed in from es
	EnrollId string `json:"enroll_id"`
	// device id is created by dsts at a separate step
	DeviceId string `json:"device_id"`
	// certificate returned by ca
	Certificate string `json:"certificate"`
	// receipt handle for queue management
	ReceiptHandle string `json:"receipt_handle"`
	// enroll type (renew or normal enroll)
	Type string `json:"type"`
	// target service like HP Connect
	ManagementService string `json:"mgmt_service"`
	// ca root and signing certificates
	ParentCertificates string `json:"parent_certificates"`
	// hardware hash if specified for device add to dsts
	HardwareHash string `json:"hardware_hash"`
}

var (
	eswLogger *zap.Logger
)

func Start(logger *zap.Logger) (*structs.CAClient, error) {
	var err error
	eswLogger = logger
	c := structs.CAClient{}
	addr := fmt.Sprintf("%s:%d", config.Settings.CA.Host, config.Settings.CA.Port)
	c.Conn, err = connectWithRetry(addr)
	if err != nil {
		return nil, err
	}
	eswLogger.Info("connected to CA", zap.String("Address:", addr))
	c.Client = pb.NewCertificateAuthorityClient(c.Conn)
	return &c, Ping(&c)
}

func connectWithRetry(addr string) (*grpc.ClientConn, error) {
	var c *grpc.ClientConn
	var err error
	for i := 0; i < RetryCount; i++ {
		c, err = grpc.Dial(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		eswLogger.Error("could not connect to ca service",
			zap.Error(err))
		time.Sleep(RetryWaitDuration)
	}
	return c, err
}

func Ping(c *structs.CAClient) error {
	var err error
	for i := 0; i < RetryCount; i++ {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(),
			operationTimeout)
		defer cancelFunc()
		_, err = c.Client.Ping(ctx, &pb.PingRequest{Message: "ping"})
		if err == nil {
			eswLogger.Info("successful ping to CA")
			break
		}
		eswLogger.Error("could not ping ca server", zap.Error(err))
		time.Sleep(RetryWaitDuration)
	}
	return err
}

func CreateDeviceCertificate(
	c *structs.CAClient,
	requestId, tenantID string,
	csr []byte) (*DeviceCertificate, error) {
	createRequest := &pb.CreateDeviceCertificateRequest{
		Header:  newCaProtocolHeader(requestId),
		Version: CaProtocolVersion,
		Tid:     tenantID,
		Csr:     csr,
	}
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	response, err := c.Client.CreateDeviceCertificate(ctx, createRequest)
	if err != nil {
		eswLogger.Error("create device certificate failed",
			zap.Error(err))
		return nil, err
	}
	if response.Header.Status != uint32(codes.OK) {
		eswLogger.Error("Response from certificate authority",
			zap.Uint32("code:", response.Header.Status))
		return nil, errors.New("failed to create device certifiate!")
	}
	return &DeviceCertificate{
		RequestId: response.Header.RequestId,
		DeviceId:  response.DeviceId,
		Certificate: base64.StdEncoding.EncodeToString(
			response.DeviceCertificate),
		ParentCertificates: base64.StdEncoding.EncodeToString(
			response.ParentCertificates),
		Type: CertificateTypeEnroll,
	}, nil
}

func RenewDeviceCertificate(
	c *structs.CAClient,
	requestId, tenantId, deviceId string,
	csr []byte) (*DeviceCertificate, error) {
	renewRequest := &pb.RenewDeviceCertificateRequest{
		Header:   newCaProtocolHeader(requestId),
		Version:  CaProtocolVersion,
		Tid:      tenantId,
		DeviceId: deviceId,
		Csr:      csr,
	}
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	response, err := c.Client.RenewDeviceCertificate(ctx, renewRequest)
	if err != nil {
		eswLogger.Error("renew device certificate failed",
			zap.Error(err))
		return nil, err
	}
	if response.Header.Status != uint32(codes.OK) {
		eswLogger.Error("Response from certificate authority",
			zap.Uint32("code:", response.Header.Status))
		return nil, errors.New("failed to renew device certifiate!")
	}
	return &DeviceCertificate{
		RequestId: response.Header.RequestId,
		DeviceId:  response.DeviceId,
		Certificate: base64.StdEncoding.EncodeToString(
			response.DeviceCertificate),
		ParentCertificates: base64.StdEncoding.EncodeToString(
			response.ParentCertificates),
		Type: CertificateTypeRenew,
	}, nil
}

func newCaProtocolHeader(requestId string) *pb.CaRequestHeader {
	return &pb.CaRequestHeader{
		ProtocolVersion: "v1",
		RequestId:       requestId,
		RequestTime:     timestamppb.Now(),
	}
}
