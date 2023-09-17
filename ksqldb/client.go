package ksqldb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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
	Ksql string `json:"ksql"`
}

// NewClient -
func NewClient(url, username, password *string) (*Client, error) {
	c := Client{
		client:   &http.Client{Timeout: 10 * time.Second},
		url:      *url,
		username: *username,
		password: *password,
	}

	return &c, nil
}

func (c *Client) DoRequest(payload *Payload) (*Response, error) {

	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ksql", c.url), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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

func (c *Client) Describe(name string) (*Source, error) {

	payload := Payload{
		Ksql: "DESCRIBE " + name + ";",
	}

	response, err := c.DoRequest(&payload)
	if err != nil {
		return nil, err
	}

	return &response.Source, nil
}

func (c *Client) CreateStream(d *schema.ResourceData, source bool) (*Source, error) {

	name := d.Get("name").(string)

	err := c.ValidateDoesNotExist(name)
	if err != nil {
		return nil, err
	}

	ksql, err := createStreamKsql(name, source, d)

	payload := Payload{
		Ksql: *ksql,
	}

	_, err = c.DoRequest(&payload)
	if err != nil {
		return nil, err
	}

	created, err := c.Describe(name)
	if err != nil {
		return nil, fmt.Errorf("this shouldn't happen") // TODO
	}

	return created, nil
}

func (c *Client) UpdateStream(d *schema.ResourceData) (*Source, error) {

	name := d.Get("name").(string)

	err := c.ValidateDoesExist(name)
	if err != nil {
		return nil, err
	}

	ksql, err := createStreamKsql(name, false, d)

	payload := Payload{
		Ksql: *ksql,
	}

	_, err = c.DoRequest(&payload)
	if err != nil {
		return nil, err
	}

	created, err := c.Describe(name)
	if err != nil {
		return nil, fmt.Errorf("this shouldn't happen") // TODO
	}

	return created, nil
}

func createStreamKsql(name string, source bool, d *schema.ResourceData) (*string, error) {

	kafkaTopic := d.Get("kafka_topic").(string)
	keyFormat := d.Get("key_format").(string)
	valueFormat := d.Get("value_format").(string)
	keySchemaId := d.Get("key_schema_id").(int)
	valueSchemaId := d.Get("value_schema_id").(int)

	var sb strings.Builder

	sb.WriteString("CREATE")

	if source {
		sb.WriteString(" SOURCE")
	} else {
		sb.WriteString(" OR REPLACE")
	}

	sb.WriteString(" STREAM ")
	sb.WriteString(name)

	// TODO columns?

	comma := false
	sb.WriteString(" WITH (")
	if &kafkaTopic != nil {
		comma = checkCommaForString(&sb, &kafkaTopic, comma)
		sb.WriteString(fmt.Sprintf("KAFKA_TOPIC = '%s'", kafkaTopic))
	}
	if &keyFormat != nil {
		comma = checkCommaForString(&sb, &keyFormat, comma)
		sb.WriteString(fmt.Sprintf("KEY_FORMAT = '%s'", keyFormat))
	}
	if &valueFormat != nil {
		comma = checkCommaForString(&sb, &valueFormat, comma)
		sb.WriteString(fmt.Sprintf("VALUE_FORMAT = '%s'", valueFormat))
	}
	if &keySchemaId != nil && keySchemaId != -1 {
		comma = checkCommaForInt(&sb, &keySchemaId, comma)
		sb.WriteString(fmt.Sprintf("KEY_SCHEMA_ID = %d", keySchemaId))
	}
	if &valueSchemaId != nil && keySchemaId != -1 {
		comma = checkCommaForInt(&sb, &valueSchemaId, comma)
		sb.WriteString(fmt.Sprintf("VALUE_SCHEMA_ID = %d", valueSchemaId))
	}

	sb.WriteString(");")

	ksql := sb.String()

	return &ksql, nil
}

func checkCommaForString(sb *strings.Builder, str *string, comma bool) bool {
	if str != nil {
		if comma {
			sb.WriteString(", ")
		}
		return true
	}
	return comma
}

func checkCommaForInt(sb *strings.Builder, i *int, comma bool) bool {
	if i != nil && *i != -1 {
		if comma {
			sb.WriteString(", ")
		}
		return true
	}
	return comma
}

func (c *Client) DropStream(name string) error {

	err := c.ValidateDoesExist(name)
	if err != nil {
		return err
	}

	payload := Payload{
		Ksql: fmt.Sprintf("DROP STREAM %s;", name),
	}

	response, err := c.DoRequest(&payload)
	if err != nil {
		return err
	}

	if response.ErrorCode != 0 {
		return errors.New(response.Message)
	}

	return nil
}

func (c *Client) ValidateDoesExist(name string) error {

	_, err := c.Describe(name)
	if err != nil {
		return fmt.Errorf("there is no stream or table named %s", name)
	}

	return nil
}

func (c *Client) ValidateDoesNotExist(name string) error {

	_, err := c.Describe(name)
	if err == nil {
		return fmt.Errorf("there is already a stream or a table named %s", name)
	}

	return nil
}
