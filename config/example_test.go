package config_test

import (
	"context"
	"fmt"
	"os"

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
)

func ExampleLoadDefaultConfig() {
	file, err := os.CreateTemp("", "alicloud-profile-example-*.json")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(`{"profiles":[{"name":"example","mode":"StsToken","region_id":"cn-hangzhou","access_key_id":"fictional-key","access_key_secret":"fictional-secret","sts_token":"fictional-token","sts_expiration":4070908800}]}`)
	if err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(file.Name()), config.WithSharedConfigProfile("example"))
	if err != nil {
		panic(err)
	}
	value, err := cfg.CredentialsProvider.Retrieve(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.Region, value.Source, value.ExpiresAt.Year())
	// Output: cn-hangzhou Profile.StsToken 2099
}
