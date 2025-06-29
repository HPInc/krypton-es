// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	dstsclient "github.com/HPInc/krypton-es/es-worker/service/client/dsts"
	"github.com/HPInc/krypton-es/es-worker/service/config"
	"github.com/HPInc/krypton-es/es-worker/service/structs"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	eswLogger *zap.Logger

	// dsts client connection
	gDSTSClient *structs.DSTSClient

	// ca client connection
	gCAClient *structs.CAClient

	// Global context for the package.
	gCtx context.Context

	// cancel func helps in propagating shutdown
	// to queue watch
	gCancelFunc context.CancelFunc

	// Connection to the devices database.
	gSQS *sqs.Client

	// Error channel
	ErrorChannel chan error

	// Interrupt channel
	InterruptChannel chan os.Signal

	// Stop channel for enroll watch
	PendingEnrollWatchStopChannel chan bool

	// Stop channel for registration watch
	PendingRegistrationWatchStopChannel chan bool

	// config settings
	settings *config.Notification

	// queue urls
	pendingEnrollQueueUrl, pendingRegistrationQueueUrl string
	enrolledQueueUrl, enrollErrorQueueUrl              string
)

const (
	awsOperationTimeout     = time.Second * 5
	awsSqsVisibilityTimeout = 60
)

const (
	EnrollTypeEnroll   string = "enroll"
	EnrollTypeUnenroll string = "unenroll"
	EnrollTypeRenew    string = "renew_enroll"
)

func Init(logger *zap.Logger, notificationSettings *config.Notification) error {
	var err error
	settings = notificationSettings
	eswLogger = logger

	ErrorChannel = make(chan error)
	InterruptChannel = make(chan os.Signal, 1)
	PendingEnrollWatchStopChannel = make(chan bool)
	PendingRegistrationWatchStopChannel = make(chan bool)
	signal.Notify(InterruptChannel, syscall.SIGINT, syscall.SIGTERM)

	gDSTSClient, err = dstsclient.Start(eswLogger)
	if err != nil {
		eswLogger.Error("dsts server connection failed. cannot continue.",
			zap.Error(err))
		return err
	}

	gCAClient, err = caclient.Start(eswLogger)
	if err != nil {
		eswLogger.Error("ca server connection failed. cannot continue.",
			zap.Error(err))
		return err
	}

	gCtx, gCancelFunc = context.WithCancel(context.Background())
	if gSQS, err = newSQS(); err != nil {
		eswLogger.Error("Failed to create queue client",
			zap.Error(err))
		return err
	}

	if err = getQueueUrls(); err != nil {
		eswLogger.Error("Failed to create queue urls",
			zap.Error(err))
		return err
	}

	go ProcessPendingEnrollQueue()
	go ProcessPendingRegistrationQueue()

	return nil
}

type resolverV2 struct {
	// Custom SQS endpoint, if configured.
	endpoint string
}

// make endpoint connection for transparent runs in local as well as cloud.
// Specify endpoint explicitly for local runs; cloud runs will load default
// config automatically. settings.Endpoint will not be set for cloud runs
func (r *resolverV2) ResolveEndpoint(ctx context.Context, params sqs.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.endpoint != "" {
		uri, err := url.Parse(r.endpoint)
		return smithyendpoints.Endpoint{
			URI: *uri,
		}, err
	}

	// delegate back to the default v2 resolver otherwise
	return sqs.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// NewSQS returns a new sns client
func newSQS() (*sqs.Client, error) {
	// make aws config. expects AWS_xx env vars
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		eswLogger.Error("Failed to initialize an AWS session.",
			zap.Error(err),
		)
	}

	// Create an instance of the SQS client using the session.
	cliSQS := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.EndpointResolverV2 = &resolverV2{endpoint: settings.Endpoint}
	})
	if cliSQS == nil {
		eswLogger.Error("Failed to create an SQS client for the session",
			zap.Error(err),
		)
		return nil, err
	}

	return cliSQS, nil
}

// get queue urls using sqs api
func getQueueUrls() error {
	var err error
	// pending enroll queue url
	if pendingEnrollQueueUrl, err = getQueueUrl(
		settings.PendingEnrollName); err != nil {
		return err
	}

	// pending registration queue url
	if pendingRegistrationQueueUrl, err = getQueueUrl(
		settings.PendingRegistrationName); err != nil {
		return err
	}

	// enrolled queue url
	if enrolledQueueUrl, err = getQueueUrl(
		settings.EnrollName); err != nil {
		return err
	}

	// enroll error queue url
	if enrollErrorQueueUrl, err = getQueueUrl(
		settings.EnrollErrorName); err != nil {
		return err
	}
	return nil
}

// helper to get queue url using sqs api
func getQueueUrl(name string) (string, error) {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	urlResult, err := gSQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &name,
	})
	if err != nil {
		eswLogger.Error("Failed to create queue url",
			zap.String("name", name),
			zap.Error(err))
		return "", err
	}
	return *urlResult.QueueUrl, err
}

func Shutdown() {
	eswLogger.Info("HP CEM ES WORKER: signalling shutdown to enroll queues")
	if gCancelFunc != nil {
		eswLogger.Info("Cancelling queue watches")
		gCancelFunc()
	}
	gCtx.Done()
	if gDSTSClient != nil && gDSTSClient.Conn != nil {
		_ = gDSTSClient.Conn.Close()
	}
	if gCAClient != nil && gCAClient.Conn != nil {
		_ = gCAClient.Conn.Close()
	}
}
