package ossxml_test

import (
	"encoding/xml"
	"fmt"
	"testing"

	helper "github.com/alibabacloud-go/alibabacloud-gateway-oss-util/client"
	oss "github.com/alibabacloud-go/oss-20190517/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

const aclXML = `<AccessControlPolicy><Owner><ID>synthetic</ID></Owner><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`
const corsXML = `<CORSConfiguration><CORSRule><AllowedOrigin>https://example.invalid</AllowedOrigin><AllowedMethod>GET</AllowedMethod></CORSRule></CORSConfiguration>`
const locationXML = `<LocationConstraint>oss-cn-beijing</LocationConstraint>`

func parse(t *testing.T, action, input string) map[string]interface{} {
	t.Helper()
	value, err := helper.ParseXml(tea.String(input), tea.String(action))
	if err != nil {
		t.Fatal(err)
	}
	object, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected helper representation %T", value)
	}
	return object
}

func TestStructuredRootRepresentation(t *testing.T) {
	t.Run("ACL", func(t *testing.T) {
		parsed := parse(t, "GetBucketAcl", aclXML)
		if len(parsed) != 1 || parsed["AccessControlPolicy"] == nil {
			t.Fatal("expected the root registered by the pinned native helper")
		}
		var unchanged oss.GetBucketAclResponse
		if err := tea.Convert(map[string]interface{}{"body": parsed}, &unchanged); err != nil {
			t.Fatal(err)
		}
		if unchanged.Body == nil || unchanged.Body.Owner != nil || unchanged.Body.AccessControlList != nil {
			t.Fatal("unexpected unchanged helper/model conversion")
		}
		var normalized oss.GetBucketAclResponse
		if err := tea.Convert(map[string]interface{}{"body": parsed["AccessControlPolicy"]}, &normalized); err != nil {
			t.Fatal(err)
		}
		if normalized.Body == nil || normalized.Body.Owner == nil || normalized.Body.AccessControlList == nil || tea.StringValue(normalized.Body.Owner.ID) != "synthetic" || tea.StringValue(normalized.Body.AccessControlList.Grant) != "private" {
			t.Fatal("fixture normalization lost ACL fields")
		}
		var direct oss.GetBucketAclResponseBody
		if err := xml.Unmarshal([]byte(aclXML), &direct); err != nil || direct.Owner == nil || direct.AccessControlList == nil {
			t.Fatal("direct structured-body fixture did not preserve XML fields", err)
		}
	})
	t.Run("CORS", func(t *testing.T) {
		parsed := parse(t, "GetBucketCors", corsXML)
		if len(parsed) != 1 || parsed["CORSConfiguration"] == nil {
			t.Fatal("missing pinned CORS root")
		}
		var unchanged, normalized oss.GetBucketCorsResponse
		if err := tea.Convert(map[string]interface{}{"body": parsed}, &unchanged); err != nil {
			t.Fatal(err)
		}
		if unchanged.Body == nil || len(unchanged.Body.CORSRule) != 0 {
			t.Fatal("unexpected unchanged CORS conversion")
		}
		if err := tea.Convert(map[string]interface{}{"body": parsed["CORSConfiguration"]}, &normalized); err != nil {
			t.Fatal(err)
		}
		if normalized.Body == nil || len(normalized.Body.CORSRule) != 1 || len(normalized.Body.CORSRule[0].AllowedOrigin) != 1 || tea.StringValue(normalized.Body.CORSRule[0].AllowedOrigin[0]) != "https://example.invalid" {
			t.Fatal("fixture normalization lost CORS fields")
		}
	})
}

func TestScalarRootMustRemainAField(t *testing.T) {
	parsed := parse(t, "GetBucketLocation", locationXML)
	if len(parsed) != 1 || parsed["LocationConstraint"] == nil {
		t.Fatal("missing pinned scalar root")
	}
	var output oss.GetBucketLocationResponse
	if err := tea.Convert(map[string]interface{}{"body": parsed}, &output); err != nil || output.Body == nil || tea.StringValue(output.Body.LocationConstraint) != "oss-cn-beijing" {
		t.Fatal("scalar root must retain its DSL field", err)
	}
	var direct oss.GetBucketLocationResponseBody
	if err := xml.Unmarshal([]byte(locationXML), &direct); err != nil || direct.LocationConstraint != nil {
		t.Fatal("unexpected direct scalar-body decoding", err)
	}
}

func TestMalformedXMLHelperSuppressesError(t *testing.T) {
	input := `<AccessControlPolicy><Owner></AccessControlPolicy>`
	parsed := parse(t, "GetBucketAcl", input)
	if len(parsed) != 0 {
		t.Fatal("pinned helper no longer returns an empty map for malformed XML")
	}
	var output oss.GetBucketAclResponseBody
	if err := xml.Unmarshal([]byte(input), &output); err == nil {
		t.Fatal("standard XML decoder must report malformed XML")
	}
}

func Example() {
	parsed, err := helper.ParseXml(tea.String(locationXML), tea.String("GetBucketLocation"))
	if err != nil {
		panic(err)
	}
	var output oss.GetBucketLocationResponse
	if err := tea.Convert(map[string]interface{}{"body": parsed}, &output); err != nil {
		panic(err)
	}
	fmt.Println(tea.StringValue(output.Body.LocationConstraint))
	// Output: oss-cn-beijing
}
