// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

func TestExtAuthz(t *testing.T) {
	server := NewExtAuthzServer()
	go server.run("localhost:0")

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", <-server.grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer func() { _ = conn.Close() }()
	grpcV3Client := authv3.NewAuthorizationClient(conn)

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{
			name:   "GRPCv3-allow",
			header: "allow",
			want:   int(codes.OK),
		},
		{
			name:   "GRPCv3-deny",
			header: "deny",
			want:   int(codes.PermissionDenied),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got int
			resp, err := grpcV3Client.Check(context.Background(), &authv3.CheckRequest{
				Attributes: &authv3.AttributeContext{
					Request: &authv3.AttributeContext_Request{
						Http: &authv3.AttributeContext_HttpRequest{
							Host:    "localhost",
							Path:    "/check",
							Headers: map[string]string{checkHeader: tc.header},
						},
					},
				},
			})
			if err != nil {
				t.Errorf(err.Error())
			} else {
				got = int(resp.Status.Code)
			}
			if got != tc.want {
				t.Errorf("want %d but got %d", tc.want, got)
			}
		})
	}
}
