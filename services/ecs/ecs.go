package ecs

import (
	"context"
	"encoding/json/v2"
	"errors"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"net/http"
	"net/url"
	"strconv"
)

// Options aliases shared per-call runtime options.
type Options = alicloud.CallOptions

// Client is a concurrency-safe ECS reference client. Construct with New.
type Client struct{ runtime *alicloud.Client }

// New validates and copies runtime configuration.
func New(config alicloud.Config) (*Client, error) {
	c, err := alicloud.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{runtime: c}, nil
}

// DescribeRegionsInput selects optional region-list filters. Nil means no filters.
type DescribeRegionsInput struct {
	// AcceptLanguage selects localized names; empty uses the service default.
	AcceptLanguage string
	// InstanceChargeType selects billing mode; empty uses the service default.
	InstanceChargeType string
	// ResourceType selects the resource category; empty uses the service default.
	ResourceType string
}

// Region is a selected region-list entry.
type Region struct {
	// RegionID is the region identifier.
	RegionID string `json:"RegionId"`
	// LocalName is the localized display name.
	LocalName string `json:"LocalName"`
	// RegionEndpoint is the service-advertised host, not automatically trusted as a resolver rule.
	RegionEndpoint string `json:"RegionEndpoint"`
	// Status is the service availability status.
	Status string `json:"Status"`
}

// DescribeRegionsOutput contains selected region information and transport metadata.
type DescribeRegionsOutput struct {
	// Regions contains the returned entries.
	Regions []Region
	// Metadata contains request ID, HTTP status and attempts.
	Metadata alicloud.Metadata
}

// DescribeRegionsAPI is the minimal interface for region-list consumers and fakes.
type DescribeRegionsAPI interface {
	// DescribeRegions lists region information.
	DescribeRegions(context.Context, *DescribeRegionsInput, ...func(*Options)) (*DescribeRegionsOutput, error)
}

// DescribeRegions lists regions using API 2014-05-26. Nil input means no filters;
// context cancellation and service errors retain their identity through wrapping.
func (c *Client) DescribeRegions(ctx context.Context, input *DescribeRegionsInput, opts ...func(*Options)) (*DescribeRegionsOutput, error) {
	q := url.Values{}
	if input != nil {
		set(q, "AcceptLanguage", input.AcceptLanguage)
		set(q, "InstanceChargeType", input.InstanceChargeType)
		set(q, "ResourceType", input.ResourceType)
	}
	var wire struct {
		Regions struct {
			Region []Region `json:"Region"`
		} `json:"Regions"`
	}
	meta, err := c.invoke(ctx, "DescribeRegions", "", q, &wire, opts...)
	if err != nil {
		return nil, err
	}
	return &DescribeRegionsOutput{Regions: wire.Regions.Region, Metadata: meta}, nil
}

// DescribeInstancesInput selects a region and optional filters. Models cover only
// these fields. Token and page-number parameters cannot be combined.
type DescribeInstancesInput struct {
	// RegionID defaults to the client's region.
	RegionID string
	// InstanceIDs is encoded as a JSON string array; at most 100 IDs are supported.
	InstanceIDs []string
	// ZoneID optionally limits results to a zone.
	ZoneID string
	// Status optionally filters lifecycle status.
	Status string
	// NextToken selects the next token page.
	NextToken string
	// MaxResults selects token pagination size; zero is omitted, 1..100 are accepted.
	MaxResults int
	// PageNumber selects legacy page-number pagination; zero is omitted.
	PageNumber int
	// PageSize selects legacy page size; zero is omitted, 1..100 are accepted.
	PageSize int
}

// Instance is a selected subset of instance properties.
type Instance struct {
	// InstanceID is the ECS instance identifier.
	InstanceID string `json:"InstanceId"`
	// InstanceName is the display name.
	InstanceName string `json:"InstanceName"`
	// RegionID is the instance region.
	RegionID string `json:"RegionId"`
	// ZoneID is the instance zone.
	ZoneID string `json:"ZoneId"`
	// Status is the lifecycle state.
	Status string `json:"Status"`
}

// DescribeInstancesOutput contains selected instance properties and pagination metadata.
type DescribeInstancesOutput struct {
	// Instances contains the returned entries.
	Instances []Instance
	// NextToken is empty when token pagination finishes.
	NextToken string
	// PageNumber is meaningful only in legacy page-number mode.
	PageNumber int
	// PageSize is meaningful only in legacy page-number mode.
	PageSize int
	// TotalCount is meaningful only in legacy page-number mode.
	TotalCount int
	// Metadata describes the HTTP operation.
	Metadata alicloud.Metadata
}

// DescribeInstancesAPI is the minimal interface for paginators and fakes.
type DescribeInstancesAPI interface {
	// DescribeInstances retrieves a selected instance-list page.
	DescribeInstances(context.Context, *DescribeInstancesInput, ...func(*Options)) (*DescribeInstancesOutput, error)
}

// DescribeInstances retrieves one page; nil input uses the client's region.
// It copies InstanceIDs and validates pagination modes before sending.
func (c *Client) DescribeInstances(ctx context.Context, input *DescribeInstancesInput, opts ...func(*Options)) (*DescribeInstancesOutput, error) {
	var in DescribeInstancesInput
	if input != nil {
		in = *input
		in.InstanceIDs = append([]string(nil), input.InstanceIDs...)
	}
	if err := validateInstances(in); err != nil {
		return nil, validation(ctx, "DescribeInstances", err)
	}
	q := url.Values{}
	set(q, "ZoneId", in.ZoneID)
	set(q, "Status", in.Status)
	set(q, "NextToken", in.NextToken)
	setInt(q, "MaxResults", in.MaxResults)
	setInt(q, "PageNumber", in.PageNumber)
	setInt(q, "PageSize", in.PageSize)
	if len(in.InstanceIDs) > 0 {
		data, err := json.Marshal(in.InstanceIDs)
		if err != nil {
			return nil, validation(ctx, "DescribeInstances", err)
		}
		q.Set("InstanceIds", string(data))
	}
	region := in.RegionID
	if region == "" {
		region = c.runtime.Region()
	}
	q.Set("RegionId", region)
	var wire struct {
		Instances struct {
			Instance []Instance `json:"Instance"`
		} `json:"Instances"`
		NextToken  string `json:"NextToken"`
		PageNumber int    `json:"PageNumber"`
		PageSize   int    `json:"PageSize"`
		TotalCount int    `json:"TotalCount"`
	}
	meta, err := c.invoke(ctx, "DescribeInstances", region, q, &wire, opts...)
	if err != nil {
		return nil, err
	}
	return &DescribeInstancesOutput{Instances: wire.Instances.Instance, NextToken: wire.NextToken, PageNumber: wire.PageNumber, PageSize: wire.PageSize, TotalCount: wire.TotalCount, Metadata: meta}, nil
}
func validateInstances(in DescribeInstancesInput) error {
	if len(in.InstanceIDs) > 100 || in.MaxResults < 0 || in.MaxResults > 100 || in.PageNumber < 0 || in.PageSize < 0 || in.PageSize > 100 {
		return errors.New("ecs: invalid instance-list parameters")
	}
	for _, id := range in.InstanceIDs {
		if id == "" {
			return errors.New("ecs: empty instance ID")
		}
	}
	if (in.NextToken != "" || in.MaxResults != 0) && (in.PageNumber != 0 || in.PageSize != 0) {
		return errors.New("ecs: token and page-number parameters cannot be combined")
	}
	return nil
}

// DescribeInstanceStatusInput selects status entries. Nil uses the client's region.
type DescribeInstanceStatusInput struct {
	// RegionID defaults to the client's region.
	RegionID string
	// InstanceIDs are encoded as InstanceId.N; at most 100 IDs are supported.
	InstanceIDs []string
	// ZoneID optionally selects a zone.
	ZoneID string
	// PageNumber is one-based; zero uses the service default of one.
	PageNumber int
	// PageSize accepts 1..50; zero uses the service default of ten.
	PageSize int
}

// InstanceStatus contains an instance identifier and lifecycle state.
type InstanceStatus struct {
	// InstanceID is the instance identifier.
	InstanceID string `json:"InstanceId"`
	// Status is the service lifecycle state.
	Status string `json:"Status"`
}

// DescribeInstanceStatusOutput contains returned status entries.
type DescribeInstanceStatusOutput struct {
	// InstanceStatuses contains the selected entries.
	InstanceStatuses []InstanceStatus
	// TotalCount is the matching status count.
	TotalCount int
	// PageNumber is the returned page number.
	PageNumber int
	// PageSize is the returned page size.
	PageSize int
	// Metadata contains transport information.
	Metadata alicloud.Metadata
}

// DescribeInstanceStatusAPI is the minimal interface for waiters and fakes.
type DescribeInstanceStatusAPI interface {
	// DescribeInstanceStatus retrieves one status page.
	DescribeInstanceStatus(context.Context, *DescribeInstanceStatusInput, ...func(*Options)) (*DescribeInstanceStatusOutput, error)
}

// DescribeInstanceStatus retrieves one status page and copies all input slices.
func (c *Client) DescribeInstanceStatus(ctx context.Context, input *DescribeInstanceStatusInput, opts ...func(*Options)) (*DescribeInstanceStatusOutput, error) {
	var in DescribeInstanceStatusInput
	if input != nil {
		in = *input
		in.InstanceIDs = append([]string(nil), input.InstanceIDs...)
	}
	if len(in.InstanceIDs) > 100 || in.PageNumber < 0 || in.PageSize < 0 || in.PageSize > 50 {
		return nil, validation(ctx, "DescribeInstanceStatus", errors.New("ecs: invalid status-list parameters"))
	}
	q := url.Values{}
	for i, id := range in.InstanceIDs {
		if id == "" {
			return nil, validation(ctx, "DescribeInstanceStatus", errors.New("ecs: empty instance ID"))
		}
		q.Set("InstanceId."+strconv.Itoa(i+1), id)
	}
	set(q, "ZoneId", in.ZoneID)
	setInt(q, "PageNumber", in.PageNumber)
	setInt(q, "PageSize", in.PageSize)
	region := in.RegionID
	if region == "" {
		region = c.runtime.Region()
	}
	q.Set("RegionId", region)
	var wire struct {
		InstanceStatuses struct {
			InstanceStatus []InstanceStatus `json:"InstanceStatus"`
		} `json:"InstanceStatuses"`
		TotalCount int `json:"TotalCount"`
		PageNumber int `json:"PageNumber"`
		PageSize   int `json:"PageSize"`
	}
	meta, err := c.invoke(ctx, "DescribeInstanceStatus", region, q, &wire, opts...)
	if err != nil {
		return nil, err
	}
	return &DescribeInstanceStatusOutput{InstanceStatuses: wire.InstanceStatuses.InstanceStatus, TotalCount: wire.TotalCount, PageNumber: wire.PageNumber, PageSize: wire.PageSize, Metadata: meta}, nil
}
func (c *Client) invoke(ctx context.Context, name, region string, q url.Values, output any, opts ...func(*Options)) (alicloud.Metadata, error) {
	return c.runtime.Invoke(ctx, alicloud.Operation{Service: "ecs", Name: name, Version: "2014-05-26", Idempotent: true}, alicloud.Request{Method: http.MethodPost, Region: region, Query: q, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}}, output, opts...)
}
func validation(ctx context.Context, name string, err error) error {
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return &alicloud.OperationError{Service: "ecs", Operation: name, Err: err}
}
func set(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}
func setInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}
