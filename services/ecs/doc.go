// Package ecs provides handwritten reference clients for three read operations
// in API version 2014-05-26: DescribeRegions, DescribeInstances and
// DescribeInstanceStatus. Models deliberately cover selected fields only;
// unknown response fields are ignored. This is not full ECS coverage.
// Clients are concurrency safe. Input slices are copied; individual operation
// interfaces permit small fakes. Errors preserve alicloud.OperationError causes.
package ecs
