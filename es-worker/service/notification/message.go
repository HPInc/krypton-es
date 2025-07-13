// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"encoding/json"
	"fmt"

	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// receieve next message for queueName
func receiveMessage(queueUrl string) ([]types.Message, error) {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	msgResult, err := gSQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   awsSqsVisibilityTimeout,
	})
	if err != nil {
		eswLogger.Error("Error recieving message",
			zap.Error(err))
		return nil, err
	}
	return msgResult.Messages, nil
}

// Outgoing enrolled message to be consumed by enroll server
// Contains device certificate. Enroll will match the enroll id
// from this message and mark device as enrolled.
func SendEnrolledMessage(dc *caclient.DeviceCertificate) error {
	jsonstring, err := json.Marshal(dc)
	if err != nil {
		eswLogger.Error("Error sending enroll message",
			zap.Error(err))
		return err
	}
	return sendMessage(enrolledQueueUrl, string(jsonstring))
}

// generic send message
func sendMessage(queueUrl string, msg string) error {
	sqsMessage := &sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: &msg,
	}
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := gSQS.SendMessage(ctx, sqsMessage)
	if err != nil {
		fmt.Printf("could not send message to queue %v: %v\n", queueUrl, err)
		return err
	}
	return nil
}

// generic delete message from queue by name and handle
func deleteMessage(queueUrl string, receiptHandle string) error {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := gSQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: &receiptHandle,
	})
	return err
}
