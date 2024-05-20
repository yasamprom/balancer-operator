package slicer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log"

	model "github.com/yasamprom/balancer-operator/internal/model"
)

const (
	connectType = "application/json"
)

type Config struct {
	Host string
	Port string
}

type Client struct {
	host string
	port string
}

func New(cfg Config) *Client {
	return &Client{host: cfg.Host, port: cfg.Port}
}

func (c *Client) NotifyEvents(ctx context.Context, nodes model.UpdateNodes) error {
	bytesMsg, err := json.Marshal(nodes)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bytesMsg)
	_, err = http.Post(fmt.Sprintf("http://%s:%s/api/v1/notify_nodes", c.host, c.port), connectType, reader)
	log.Printf("request: %v", string(bytesMsg))
	if err != nil {
		log.Printf("http client error: %v", err)
		return err
	}
	return nil
}

func (c *Client) NotifyState(ctx context.Context, state model.UpdateNodes) error {
	jsonBody, err := json.Marshal(state)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	requestURL := fmt.Sprintf("http://%s:%s/update_state", c.host, c.port)
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		log.Printf("http client error: %v", err)
		return err
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return err
	}
	return nil
}
