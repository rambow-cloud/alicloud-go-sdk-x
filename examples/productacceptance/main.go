// Command productacceptance demonstrates public product consumers without a cloud account.
package main

import (
	"context"
	"fmt"

	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func imageIDs(ctx context.Context, api ecs.DescribeImagesAPI, input *ecs.DescribeImagesInput) ([]string, error) {
	p, err := ecs.NewDescribeImagesPaginator(api, input)
	if err != nil {
		return nil, err
	}
	var ids []string
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		if page.Images != nil {
			for _, image := range page.Images.Image {
				if image.ImageID != nil {
					ids = append(ids, *image.ImageID)
				}
			}
		}
	}
	return ids, nil
}

func instanceIDs(ctx context.Context, api ecs.DescribeInstancesAPI, input *ecs.DescribeInstancesInput) ([]string, error) {
	p, err := ecs.NewDescribeInstancesPaginator(api, input)
	if err != nil {
		return nil, err
	}
	var ids []string
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		if page.Instances != nil {
			for _, instance := range page.Instances.Instance {
				if instance.InstanceID != nil {
					ids = append(ids, *instance.InstanceID)
				}
			}
		}
	}
	return ids, nil
}

type imagesFunc func(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error)

func (f imagesFunc) DescribeImages(ctx context.Context, in *ecs.DescribeImagesInput, opts ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
	return f(ctx, in, opts...)
}

func ptr[T any](value T) *T { return &value }

func main() {
	api := imagesFunc(func(_ context.Context, in *ecs.DescribeImagesInput, _ ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
		return &ecs.DescribeImagesOutput{TotalCount: ptr(int32(2)), PageSize: ptr(int32(1)), PageNumber: in.PageNumber, Images: &ecs.DescribeImagesOutputImages{Image: []ecs.DescribeImagesOutputImagesImage{{ImageID: ptr(fmt.Sprintf("fixture-image-%d", *in.PageNumber))}}}}, nil
	})
	ids, err := imageIDs(context.Background(), api, &ecs.DescribeImagesInput{PageSize: ptr(int32(1))})
	if err != nil {
		panic(err)
	}
	fmt.Printf("ECS image traversal: %v\n", ids)
}
