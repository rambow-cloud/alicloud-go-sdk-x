package ecs

import "errors"

// These reviewed constraints are prose-only and remain handwritten extensions.
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

func validateStatus(in DescribeInstanceStatusInput) error {
	if len(in.InstanceIDs) > 100 || in.PageNumber < 0 || in.PageSize < 0 || in.PageSize > 50 {
		return errors.New("ecs: invalid status-list parameters")
	}
	for _, id := range in.InstanceIDs {
		if id == "" {
			return errors.New("ecs: empty instance ID")
		}
	}
	return nil
}
