package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const ODOH_CONTENT_TYPE = "application/oblivious-dns-message"

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

func Handler(ctx context.Context, request Request) (Response, error) {

	targetHost := request.QueryStringParameters["targethost"]
	targetPath := request.QueryStringParameters["targetpath"]
	queryBody := request.Body

	matched, err := regexp.Match(`^/`, []byte(targetPath))

	if err != nil{
		fmt.Println(err)
		return Response{StatusCode: 500}, err
	}

	if !matched {
		targetPath = "/" + targetPath
	}

	queryUrl := fmt.Sprintf("https://%s%s", targetHost, targetPath)

	serializedQuery, err := base64.StdEncoding.DecodeString(queryBody)

	if err != nil{
		fmt.Println(err)
		return Response{StatusCode: 500}, err
	}

	req, err := http.NewRequest(http.MethodPost, queryUrl, bytes.NewBuffer(serializedQuery))
	if err != nil {
		fmt.Println(err)
		return Response{StatusCode: 500}, err
	}

	req.Header.Set("content-type", ODOH_CONTENT_TYPE)
	req.Header.Set("cache-control", "no-cache, no-store")
	req.Header.Set("accept", ODOH_CONTENT_TYPE)

	client := http.Client{}
	resolvedResp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return Response{StatusCode: 500}, err
	}

	resolvedBody, err := ioutil.ReadAll(resolvedResp.Body)

	if err != nil {
		fmt.Println(err)
		return Response{StatusCode: 500}, err
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: true,
		Body:            base64.StdEncoding.EncodeToString(resolvedBody),
		Headers: map[string]string{
			"Content-Type": ODOH_CONTENT_TYPE,
			"Content-Encoding": "deflate",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
