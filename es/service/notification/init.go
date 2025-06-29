// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"net/url"
	"time"

	"github.com/HPInc/krypton-es/es/service/config"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger

	// Global context for the package.
	gCtx context.Context

	// cancel func helps in propagating shutdown
	// to queue watch
	gCancelFunc context.CancelFunc

	// Connection to sqs
	gSQS *sqs.Client

	// config settings
	notificationSettings *config.Notification

	// enroll queue url
	enrollQueueUrl, enrollErrorQueueUrl, pendingEnrollQueueUrl string
)

const (
	awsOperationTimeout     = time.Second * 5
	awsSqsVisibilityTimeout = 60
)

func Init(logger *zap.Logger, settings *config.Notification) error {
	var err error
	notificationSettings = settings
	esLogger = logger

	gCtx, gCancelFunc = context.WithCancel(context.Background())

	// Create sqs client
	if gSQS, err = newSQS(settings); err != nil {
		esLogger.Error("Failed to create queue client",
			zap.Error(err))
		return err
	}

	if err = getQueueUrls(); err != nil {
		esLogger.Error("Failed to create queue urls",
			zap.Error(err))
		return err
	}

	// watch queue for enrolled events
	go watchEnrollQueue()
	go watchEnrollErrorQueue()

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

// get queue urls using sqs api
func getQueueUrls() error {
	var err error
	// set enroll queue url
	if enrollQueueUrl, err = getQueueUrl(
		notificationSettings.EnrollName); err != nil {
		return err
	}

	// pending enroll queue url
	if pendingEnrollQueueUrl, err = getQueueUrl(
		notificationSettings.PendingEnrollName); err != nil {
		return err
	}

	// enroll error queue url
	if enrollErrorQueueUrl, err = getQueueUrl(
		notificationSettings.EnrollErrorName); err != nil {
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
		esLogger.Error("Failed to create queue url",
			zap.String("name", name),
			zap.Error(err))
		return "", err
	}
	return *urlResult.QueueUrl, err
}

// NewSQS returns a new sns client
func newSQS(settings *config.Notification) (*sqs.Client, error) {
	// make aws config. expects AWS_xx env vars
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		esLogger.Error("Failed to initialize an AWS session.",
			zap.Error(err),
		)
	}

	// Create an instance of the SQS client using the session.
	cliSQS := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.EndpointResolverV2 = &resolverV2{endpoint: settings.Endpoint}
	})
	if cliSQS == nil {
		esLogger.Error("Failed to create an SQS client for the session",
			zap.Error(err),
		)
		return nil, err
	}

	return cliSQS, nil
}

func Shutdown() {
	esLogger.Info("HP Krypton ES: signalling shutdown to enroll queue subscribers")
	if gCancelFunc != nil {
		esLogger.Info("Cancelling enroll and enroll error queue watches")
		gCancelFunc()
	}
}
