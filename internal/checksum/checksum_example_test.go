package checksum_test

import (
	"context"
	"fmt"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/checksum"
)

func ExampleCRC64() {
	crc, err := checksum.CRC64(context.Background(), 0, []byte("123456789"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x\n", crc)
	fmt.Println(checksum.VerifyCRC64(crc, "11051210869376104954") == nil)
	md5, err := checksum.ContentMD5(context.Background(), []byte("abc"))
	if err != nil {
		panic(err)
	}
	fmt.Println(md5)
	// Output:
	// 995dc9bbdf1939fa
	// true
	// kAFQmDzST7DWlj99KOF/cg==
}
