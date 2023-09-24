package ksqldb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

// Client -
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
	Statement   string `json:"statement"`
	Timestamp   string `json:"timestamp"`
}

type Payload struct {
	Ksql       string                 `json:"ksql"`
	Properties map[string]interface{} `json:"streamsProperties"`
}

// newClient -
func newClient(url, username, password *string) (*Client, error) {
	c := Client{
		client:   &http.Client{Timeout: 10 * time.Second},
		url:      *url,
		username: *username,
		password: *password,
	}

	return &c, nil
}

func (c *Client) doRequest(payload *Payload) (*Response, error) {

	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

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

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

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

func (c *Client) describe(name string) (*Source, error) {

	payload := Payload{
		Ksql: "DESCRIBE " + name + ";",
	}

	response, err := c.doRequest(&payload)
	if err != nil {
		return nil, err
	}

	return &response.Source, nil
}

func (c *Client) createStream(d *schema.ResourceData, materialized bool, source bool) (*Source, error) {
	return c.doCreateStream(d, source, materialized, false)
}

func (c *Client) updateStream(d *schema.ResourceData, materialized bool) (*Source, error) {
	// updating a stream is the same as creating it but with using "CREATE OR REPLACE" in the statement.
	// Therefore, it can't be a source stream.
	// Also, it must exist.
	return c.doCreateStream(d, false, materialized, true)
}

func (c *Client) doCreateStream(d *schema.ResourceData, source bool, materialized bool, mustExist bool) (*Source, error) {

	name := d.Get("name").(string)

	if mustExist {
		err := c.validateDoesExist(name)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.validateDoesNotExist(name)
		if err != nil {
			return nil, err
		}
	}

	ksql, err := createStreamKsql(name, source, materialized, d)
	properties := d.Get("properties").(map[string]interface{})

	payload := Payload{
		Ksql:       *ksql,
		Properties: properties,
	}

	_, err = c.doRequest(&payload)
	if err != nil {
		return nil, err
	}

	created, err := c.describe(name)
	if err != nil {
		return nil, fmt.Errorf("this shouldn't happen") // TODO
	}

	return created, nil
}

func (c *Client) dropStream(name string) error {

	err := c.validateDoesExist(name)
	if err != nil {
		return err
	}

	payload := Payload{
		Ksql: fmt.Sprintf("DROP STREAM %s;", name),
	}

	response, err := c.doRequest(&payload)
	if err != nil {
		return err
	}

	if response.ErrorCode != 0 {
		return errors.New(response.Message)
	}

	return nil
}

func (c *Client) validateDoesExist(name string) error {

	_, err := c.describe(name)
	if err != nil {
		return fmt.Errorf("there is no stream or table named %s", name)
	}

	return nil
}

func (c *Client) validateDoesNotExist(name string) error {

	_, err := c.describe(name)
	if err == nil {
		return fmt.Errorf("there is already a stream or a table named %s", name)
	}

	return nil
}
