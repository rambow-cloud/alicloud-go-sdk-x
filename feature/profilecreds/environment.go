package profilecreds

import (
	"os"
	"strings"
)

func profileEnvironment() string { return os.Getenv("ALIBABA_CLOUD_PROFILE") }
func externalDisabled() bool {
	v := os.Getenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS")
	return v == "1" || strings.EqualFold(v, "true")
}
func metadataDisabled() bool {
	v := os.Getenv("ALIBABA_CLOUD_ECS_METADATA_DISABLED")
	return v == "1" || strings.EqualFold(v, "true")
}
