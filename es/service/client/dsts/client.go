// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package dstsclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	pb "github.com/HPInc/krypton-es/es/protos/dsts"
	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DstsProtocolVersion = "v1"
	RetryCount          = 3
	RetryWaitDuration   = time.Second * 3
	operationTimeout    = time.Second * 3

	DefaultEnrollmentTokenLifetimeDays = 30
)

var (
	ErrFailedToValidateEnrollmentToken = errors.New("Failed to validate enrollment token")
	ErrInvalidEnrollmentToken          = errors.New("Invalid enrollment token")
	ErrFailedToCreateEnrollmentToken   = errors.New("Failed to create enrollment token")
	ErrFailedToGetEnrollmentToken      = errors.New("Failed to get enrollment token")
	ErrFailedToDeleteEnrollmentToken   = errors.New("Failed to delete enrollment token")
)

// dsts client struct
type DSTSClient struct {
	Conn   *grpc.ClientConn
	Client pb.DeviceSTSClient
}

// If there is an error, provide detail about
// error code and http status code suggestion
type ClientError struct {
	Error    error
	HttpCode int
}

type EnrollToken struct {
	TenantId  string `json:"tenant_id"`
	Token     string `json:"enroll_token"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger
	// client instance
	gClient DSTSClient
)

func Init(logger *zap.Logger) error {
	var err error
	esLogger = logger
	gClient = DSTSClient{}
	addr := fmt.Sprintf("%s:%d", config.Settings.DSTS.Host,
		config.Settings.DSTS.RpcPort)
	gClient.Conn, err = connectWithRetry(addr)
	if err != nil {
		esLogger.Error("Could not connect to DSTS.",
			zap.Error(err))
		return err
	}
	esLogger.Info("Connected to DSTS.", zap.String("Address:", addr))
	gClient.Client = pb.NewDeviceSTSClient(gClient.Conn)
	return Ping()
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
		esLogger.Error(
			"Could not connect to DSTS. Waiting to reconnect.",
			zap.Error(err))
		time.Sleep(RetryWaitDuration)
	}
	return c, err
}

func Ping() error {
	var err error
	for i := 0; i < RetryCount; i++ {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(),
			operationTimeout)
		defer cancelFunc()
		_, err = gClient.Client.Ping(
			ctx,
			&pb.PingRequest{Message: "ping"})
		if err == nil {
			esLogger.Info("Successful ping to DSTS")
			break
		}
		esLogger.Error("Could not ping dsts server",
			zap.Error(err))
		time.Sleep(RetryWaitDuration)
	}
	return err
}

func ValidateEnrollmentToken(token string) (string, *ClientError) {
	req := &pb.ValidateEnrollmentTokenRequest{
		Header:  newDstsProtocolHeader(),
		Version: DstsProtocolVersion,
		Token:   token,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	resp, err := gClient.Client.ValidateEnrollmentToken(ctx, req)
	if err != nil {
		return "", &ClientError{err, getHttpCode(codes.Internal)}
	}
	if resp.Header.Status != uint32(codes.OK) {
		esLogger.Error("ValidateEnrollmentToken: RPC failed",
			zap.Uint32("Status", resp.Header.Status))
		return "", &ClientError{
			ErrFailedToValidateEnrollmentToken,
			getHttpCode(codes.Code(resp.Header.Status)),
		}
	}
	if !resp.GetIsValid() {
		return "", &ClientError{
			ErrInvalidEnrollmentToken,
			getHttpCode(codes.InvalidArgument)}
	}
	return resp.GetTid(), nil
}

// create enrollment token
func CreateEnrollmentToken(tenantId string, lifetimeDays int32) (*EnrollToken, *ClientError) {
	req := &pb.CreateEnrollmentTokenRequest{
		Header:            newDstsProtocolHeader(),
		Version:           DstsProtocolVersion,
		Tid:               tenantId,
		TokenLifetimeDays: lifetimeDays,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	resp, err := gClient.Client.CreateEnrollmentToken(ctx, req)
	if err != nil {
		esLogger.Error("CreateEnrollmentToken: RPC failed",
			zap.Error(err))
		return nil, &ClientError{err, getHttpCode(codes.Internal)}
	}
	if resp.Header.Status != uint32(codes.OK) {
		esLogger.Error("CreateEnrollmentToken: RPC failed",
			zap.Uint32("Status", resp.Header.Status))
		return nil, &ClientError{
			ErrFailedToCreateEnrollmentToken,
			getHttpCode(codes.Code(resp.Header.Status)),
		}
	}
	return &EnrollToken{
		TenantId:  tenantId,
		Token:     resp.Token.Token,
		IssuedAt:  resp.Token.IssuedTime.Seconds,
		ExpiresAt: resp.Token.ExpiryTime.Seconds,
	}, nil
}

// get enrollment token for tenant
// must be created before
func GetEnrollmentToken(tenantId string) (*EnrollToken, *ClientError) {
	req := &pb.GetEnrollmentTokenRequest{
		Header:  newDstsProtocolHeader(),
		Version: DstsProtocolVersion,
		Tid:     tenantId,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	resp, err := gClient.Client.GetEnrollmentToken(ctx, req)
	if err != nil {
		esLogger.Error("GetEnrollmentToken: RPC failed",
			zap.Error(err))
		return nil, &ClientError{err, getHttpCode(codes.Internal)}
	}

	if resp.Header.Status != uint32(codes.OK) {
		esLogger.Error("GetEnrollmentToken: RPC failed",
			zap.Uint32("status", resp.Header.Status))
		return nil, &ClientError{
			ErrFailedToGetEnrollmentToken,
			getHttpCode(codes.Code(resp.Header.Status)),
		}
	}

	return &EnrollToken{
		TenantId:  tenantId,
		Token:     resp.Token.Token,
		IssuedAt:  resp.Token.IssuedTime.Seconds,
		ExpiresAt: resp.Token.ExpiryTime.Seconds,
	}, nil
}

// delete enrollment token
func DeleteEnrollmentToken(tenantId string) *ClientError {
	req := &pb.DeleteEnrollmentTokenRequest{
		Header:  newDstsProtocolHeader(),
		Version: DstsProtocolVersion,
		Tid:     tenantId,
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		operationTimeout)
	defer cancelFunc()
	resp, err := gClient.Client.DeleteEnrollmentToken(ctx, req)
	if err != nil {
		esLogger.Error("DeleteEnrollmentToken: RPC failed",
			zap.Error(err))
		return &ClientError{err, getHttpCode(codes.Internal)}
	}
	if resp.Header.Status != uint32(codes.OK) {
		esLogger.Error("DeleteEnrollmentToken: RPC failed",
			zap.Uint32("Status", resp.Header.Status))
		return &ClientError{
			ErrFailedToDeleteEnrollmentToken,
			getHttpCode(codes.Code(resp.Header.Status)),
		}
	}
	return nil
}

func Close() {
	if gClient.Conn != nil {
		_ = gClient.Conn.Close()
	}
}

func newDstsProtocolHeader() *pb.DstsRequestHeader {
	return &pb.DstsRequestHeader{
		ProtocolVersion: "v1",
		RequestId:       uuid.New().String(),
		RequestTime:     timestamppb.Now(),
	}
}

// translate grpc code to http code
// loose translation as a hint so grpc code stays encapsulated here.
// consumers for this api are ultimately http so this is justified.
func getHttpCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
