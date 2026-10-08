package profilecreds

import "os"

func profileEnvironment() string { return os.Getenv("ALIBABA_CLOUD_PROFILE") }
