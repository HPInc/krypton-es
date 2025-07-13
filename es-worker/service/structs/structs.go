// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package structs

import (
	pbca "github.com/HPInc/krypton-es/es-worker/protos/ca"
	pbdsts "github.com/HPInc/krypton-es/es-worker/protos/dsts"
	"google.golang.org/grpc"
)

type CAClient struct {
	Conn   *grpc.ClientConn
	Client pbca.CertificateAuthorityClient
}

type DSTSClient struct {
	Conn   *grpc.ClientConn
	Client pbdsts.DeviceSTSClient
}
