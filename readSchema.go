package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/imroc/req/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func readSchemaFromFile(inputFileName string) string {
	b, err := os.ReadFile(inputFileName) // just pass the file name
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	return string(b)
}

func readSchemaFromUrl(url string, key string) string {
	if !strings.HasSuffix("/v1/schema/read", url) {
		url = url + "/v1/schema/read"
	}

	var request = req.R()
	if key != "" {
		request.SetBearerAuthToken(key)
	}

	resp, err := request.Post(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		fmt.Println(resp.String())
		os.Exit(1)
	}

	var data SchemaResponse
	err = json.Unmarshal(resp.Bytes(), &data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return data.SchemaText
}

func readSchemaFromGrpc(host string, key string, insecureGrpc bool) string {
	var options []grpc.DialOption
	if insecureGrpc {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if key != "" {
			options = append(options, grpcutil.WithInsecureBearerToken(key))
		}
	} else {
		transport, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		options = append(options, transport)
		if key != "" {
			options = append(options, grpcutil.WithBearerToken(key))
		}
	}

	client, err := authzed.NewClient(host, options...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	response, err := client.ReadSchema(context.Background(), &v1.ReadSchemaRequest{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return response.SchemaText
}

type SchemaResponse struct {
	SchemaText string `json:"schemaText"`
}
