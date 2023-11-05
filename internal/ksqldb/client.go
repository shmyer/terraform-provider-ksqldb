package ksqldb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

// Client the ksqldb client object.
type Client struct {
	client   *http.Client
	url      string
	username string
	password string
}

type Response struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Source    Source `json:"sourceDescription"`
}

type Source struct {
	Name        string `json:"name"`
	KeyFormat   string `json:"keyFormat"`
	ValueFormat string `json:"valueFormat"`
	Topic       string `json:"topic"`
	Partitions  int64  `json:"partitions"`
	Replication int64  `json:"replication"`
	Statement   string `json:"statement"`
	Timestamp   string `json:"timestamp"`
}

type Payload struct {
	Ksql       string            `json:"ksql"`
	Properties map[string]string `json:"streamsProperties"`
}

// NewClient creates a new ksqldb Client.
func NewClient(url, username, password *string) *Client {
	return &Client{
		client:   &http.Client{Timeout: 10 * time.Second},
		url:      *url,
		username: *username,
		password: *password,
	}
}

func (c *Client) doRequest(ctx context.Context, payload *Payload) (*Response, error) {

	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Executing KSQL request: %s", rb))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ksql", c.url), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	// set headers according to ksqlDB HTTP API Reference: https://docs.ksqldb.io/en/latest/developer-guide/api/#content-types
	req.Header.Set("Accept", "application/vnd.ksql.v1+json")

	// limit access to the client to one resource at a time. Otherwise, a ProducerFencedException may occur in ksqlDB.
	mutex.Lock()

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// release
	mutex.Unlock()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Received KSQL response: %s", string(body)))

	if res.StatusCode != http.StatusOK {
		var obj Response
		if err := json.Unmarshal(body, &obj); err != nil {
			return nil, err
		}

		return nil, errors.New(obj.Message)
	}

	for _, b := range body {
		switch b {
		// These are the only valid whitespace in a JSON object.
		case ' ', '\n', '\r', '\t':
		case '[':
			var obj []Response
			if err := json.Unmarshal(body, &obj); err != nil {
				return nil, err
			}
			return &obj[0], nil
		case '{':
			var obj Response
			if err := json.Unmarshal(body, &obj); err != nil {
				return nil, err
			}
			return &obj, nil
		default:
			return nil, errors.New("response must be object or list")
		}
	}
	return nil, errors.New("response must be object or list")
}

func (c *Client) describe(ctx context.Context, name string) (*Source, error) {

	payload := Payload{
		Ksql: "DESCRIBE " + name + ";",
	}

	response, err := c.doRequest(ctx, &payload)
	if err != nil {
		return nil, err
	}

	return &response.Source, nil
}

func (c *Client) createStream(ctx context.Context, data StreamResourceModel, materialized bool, source bool) (*Source, error) {
	return c.doCreateStream(ctx, data, source, materialized, false)
}

func (c *Client) updateStream(ctx context.Context, data StreamResourceModel, materialized bool) (*Source, error) {
	// updating a stream is the same as creating it but with using "CREATE OR REPLACE" in the statement.
	// Therefore, it can't be a source stream.
	// Also, it must exist.
	return c.doCreateStream(ctx, data, false, materialized, true)
}

func (c *Client) doCreateStream(ctx context.Context, data StreamResourceModel, source bool, materialized bool, mustExist bool) (*Source, error) {

	name := data.Name.ValueString()

	if mustExist {
		err := c.validateDoesExist(ctx, name)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.validateDoesNotExist(ctx, name)
		if err != nil {
			return nil, err
		}
	}

	ksql := createStreamKsql(ctx, name, source, materialized, data)

	tflog.Info(ctx, fmt.Sprintf("Properties Raw: %v", data.Properties))

	properties := make(map[string]string, len(data.Properties.Elements()))
	data.Properties.ElementsAs(ctx, &properties, false)

	tflog.Info(ctx, fmt.Sprintf("Properties made: %v", properties))

	payload := Payload{
		Ksql:       *ksql,
		Properties: properties,
	}

	_, err := c.doRequest(ctx, &payload)
	if err != nil {
		return nil, err
	}

	created, err := c.describe(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("this shouldn't happen") // TODO
	}

	return created, nil
}

func (c *Client) dropStream(ctx context.Context, name string) error {

	err := c.validateDoesExist(ctx, name)
	if err != nil {
		return err
	}

	payload := Payload{
		Ksql: fmt.Sprintf("DROP STREAM %s;", name),
	}

	response, err := c.doRequest(ctx, &payload)
	if err != nil {
		return err
	}

	if response.ErrorCode != 0 {
		return errors.New(response.Message)
	}

	return nil
}

func (c *Client) validateDoesExist(ctx context.Context, name string) error {

	_, err := c.describe(ctx, name)
	if err != nil {
		return fmt.Errorf("there is no stream or table named %s", name)
	}

	return nil
}

func (c *Client) validateDoesNotExist(ctx context.Context, name string) error {

	_, err := c.describe(ctx, name)
	if err == nil {
		return fmt.Errorf("there is already a stream or a table named %s", name)
	}

	return nil
}
