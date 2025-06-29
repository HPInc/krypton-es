// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
func receiveMessage(queueUrl string, waitTimeSeconds int) (
	[]types.Message, error) {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	msgResult, err := gSQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   awsSqsVisibilityTimeout,
		WaitTimeSeconds:     int32(waitTimeSeconds),
	})
	if err != nil {
		esLogger.Error("Error recieving message",
			zap.Error(err))
		return nil, err
	}
	return msgResult.Messages, nil
}

func SendMessage(msg string) (*sqs.SendMessageOutput, error) {
	sqsMessage := &sqs.SendMessageInput{
		QueueUrl:    &pendingEnrollQueueUrl,
		MessageBody: &msg,
	}
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	output, err := gSQS.SendMessage(ctx, sqsMessage)
	if err != nil {
		return nil, fmt.Errorf(
			"could not send message to queue %v: %v",
			enrollQueueUrl,
			err)
	}
	return output, nil
}

func deleteMessage(queueUrl, receiptHandle string) error {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	_, err := gSQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: &receiptHandle,
	})
	return err
}
