package xmlmodel_test

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/xmlmodel"
)

// These native wire tags mirror complete OSS DSL models, not the helper's
// extra response wrapper. Root choices come from the pinned comparison.
type owner struct {
	ID *string `json:"ID,omitempty"`
}
type acl struct {
	Access *struct {
		Grant *string `json:"Grant,omitempty"`
	} `json:"AccessControlList,omitempty"`
	Owner *owner `json:"Owner,omitempty"`
}
type rule struct {
	Origins []*string `json:"AllowedOrigin,omitempty"`
	Methods []*string `json:"AllowedMethod,omitempty"`
}
type cors struct {
	Rules []*rule `json:"CORSRule,omitempty"`
	Vary  *bool   `json:"ResponseVary,omitempty"`
}
type location struct {
	Region *string `json:"LocationConstraint,omitempty"`
}

func ptr[T any](v T) *T { return &v }

func TestNativeRootsPreserveWireModels(t *testing.T) {
	var permissions acl
	if err := xmlmodel.Decode(context.Background(), []byte(`<AccessControlPolicy><Owner><ID>synthetic</ID></Owner><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`), xmlmodel.Root{Name: xml.Name{Local: "AccessControlPolicy"}}, &permissions); err != nil {
		t.Fatal(err)
	}
	if permissions.Access == nil || permissions.Owner == nil || permissions.Access.Grant == nil || *permissions.Access.Grant != "private" || permissions.Owner.ID == nil || *permissions.Owner.ID != "synthetic" {
		t.Fatal("structured root lost ACL fields")
	}
	var configuration cors
	if err := xmlmodel.Decode(context.Background(), []byte(`<CORSConfiguration><CORSRule><AllowedOrigin>https://example.invalid</AllowedOrigin><AllowedOrigin></AllowedOrigin><AllowedMethod>GET</AllowedMethod></CORSRule><ResponseVary>false</ResponseVary><Future><Item>ignored</Item></Future></CORSConfiguration>`), xmlmodel.Root{Name: xml.Name{Local: "CORSConfiguration"}}, &configuration); err != nil {
		t.Fatal(err)
	}
	if len(configuration.Rules) != 1 || len(configuration.Rules[0].Origins) != 2 || *configuration.Rules[0].Origins[0] != "https://example.invalid" || *configuration.Rules[0].Origins[1] != "" || configuration.Vary == nil || *configuration.Vary {
		t.Fatal("repeated fields or explicit absence changed")
	}
	var result location
	root := xmlmodel.Root{Name: xml.Name{Local: "LocationConstraint"}, ScalarField: "LocationConstraint"}
	if err := xmlmodel.Decode(context.Background(), []byte(`<LocationConstraint>oss-cn-beijing</LocationConstraint>`), root, &result); err != nil || result.Region == nil || *result.Region != "oss-cn-beijing" {
		t.Fatal("scalar root lost its DSL field", err)
	}
	if err := xmlmodel.Decode(context.Background(), []byte(`<LocationConstraint/>`), root, &result); err != nil || result.Region == nil || *result.Region != "" {
		t.Fatal("explicit empty scalar root lost presence", err)
	}
}

func TestDeterministicRequestEscapingAndOwnership(t *testing.T) {
	input := cors{Rules: []*rule{{Origins: []*string{ptr("https://example.invalid?a=1&b=<中文>")}, Methods: []*string{ptr("GET"), ptr("")}}}, Vary: ptr(false)}
	root := xmlmodel.Root{Name: xml.Name{Local: "CORSConfiguration"}}
	data, err := xmlmodel.Encode(context.Background(), root, &input)
	if err != nil {
		t.Fatal(err)
	}
	want := `<CORSConfiguration><CORSRule><AllowedOrigin>https://example.invalid?a=1&amp;b=&lt;中文&gt;</AllowedOrigin><AllowedMethod>GET</AllowedMethod><AllowedMethod></AllowedMethod></CORSRule><ResponseVary>false</ResponseVary></CORSConfiguration>`
	if string(data) != want || *input.Rules[0].Origins[0] != "https://example.invalid?a=1&b=<中文>" {
		t.Fatal("request encoding or input ownership changed")
	}
	var decoded cors
	if err := xmlmodel.Decode(context.Background(), data, root, &decoded); err != nil || !reflect.DeepEqual(input, decoded) {
		t.Fatal("native fields changed", err)
	}
	data[0] = '?'
	second, err := xmlmodel.Encode(context.Background(), root, input)
	if err != nil || string(second) != want {
		t.Fatal("returned bytes retained shared state", err)
	}
	text, err := xmlmodel.Encode(context.Background(), xmlmodel.Root{Name: xml.Name{Local: "LocationConstraint"}, ScalarField: "LocationConstraint"}, location{ptr("")})
	if err != nil || string(text) != `<LocationConstraint></LocationConstraint>` {
		t.Fatal("scalar root encoding changed", err)
	}
}

func TestReuseWithRenamedModelsAndNamespace(t *testing.T) {
	type entry struct {
		Flag  *bool  `json:"Enabled,omitempty"`
		Count *int32 `json:"Count,omitempty"`
	}
	type renamed struct {
		Items []*entry `json:"Entry,omitempty"`
		Note  *string  `json:"Description,omitempty"`
	}
	root := xmlmodel.Root{Name: xml.Name{Local: "SyntheticCollection", Space: "urn:example:synthetic"}}
	var output renamed
	data := []byte(`<p:SyntheticCollection xmlns:p="urn:example:synthetic"><p:Entry><p:Enabled>1</p:Enabled><p:Count>42</p:Count></p:Entry><p:Entry><p:Enabled>0</p:Enabled></p:Entry><p:Description><![CDATA[<value>&]]></p:Description></p:SyntheticCollection>`)
	if err := xmlmodel.Decode(context.Background(), data, root, &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Items) != 2 || *output.Items[0].Count != 42 || !*output.Items[0].Flag || *output.Items[1].Flag || output.Items[1].Count != nil || *output.Note != "<value>&" {
		t.Fatal("renamed model lost native fields")
	}
	encoded, err := xmlmodel.Encode(context.Background(), root, output)
	if err != nil {
		t.Fatal(err)
	}
	var another renamed
	if err := xmlmodel.Decode(context.Background(), encoded, root, &another); err != nil || !reflect.DeepEqual(output, another) {
		t.Fatal("namespace binding changed", err)
	}
}

func TestInvalidDocumentsNeverPublishPartialOutput(t *testing.T) {
	type value struct {
		Secret *string `json:"Secret,omitempty"`
		Number *int32  `json:"Number,omitempty"`
		Flag   *bool   `json:"Flag,omitempty"`
	}
	root := xmlmodel.Root{Name: xml.Name{Local: "Root"}}
	cases := []string{
		`<Root><Secret>secret-fixture</Secret><Number>not-a-number</Number></Root>`,
		`<Root><Secret>secret-fixture</Secret><Number>2147483648</Number></Root>`,
		`<Root><Secret>one</Secret><Secret>secret-fixture</Secret></Root>`,
		`<Root><Secret>secret-fixture</Secret><Flag>yes</Flag></Root>`,
		`<Root><Secret>secret-fixture</Secret></Wrong>`, `<Wrong/>`, `<Root/><Root/>`,
		`<Root/>secret-fixture`, "<Root/>\u00a0", `<Root>secret-fixture<Secret/></Root>`,
		`<Root><Secret><Nested/></Secret></Root>`, `<Root>&unknown;</Root>`,
		`<!DOCTYPE Root [<!ENTITY x "secret-fixture">]><Root>&x;</Root>`,
		`<?custom secret-fixture?><Root/>`, `<Root><?xml version="1.0"?></Root>`,
		`<Root xmlns="urn:wrong"/>`, `<Root><Secret xmlns="urn:wrong">secret-fixture</Secret></Root>`,
		`<Root><p:Unknown/></Root>`, `<Root xmlns:p="urn:x" xmlns:q="urn:x" p:a="1" q:a="2"/>`,
		`<Root xmlns:p="urn:x" xmlns:p="urn:y"/>`, `<Root xmlns:xml="urn:wrong"/>`,
		`<?xml version="1.1"?><Root/>`, `<?xml nonsense?><Root/>`, `<!--bad--comment--><Root/>`,
		`<Root><Secret>` + string([]byte{0xff}) + `</Secret></Root>`, "", `<Root>`,
	}
	for i, data := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			original := ptr("retained")
			output := value{Secret: original}
			err := xmlmodel.Decode(context.Background(), []byte(data), root, &output)
			if !errors.Is(err, xmlmodel.ErrInvalid) || strings.Contains(err.Error(), "secret-fixture") || output.Secret != original || output.Number != nil || output.Flag != nil {
				t.Fatal("invalid document changed output or exposed data", err)
			}
		})
	}
}

func TestUTF8DeclarationAndAtomicReset(t *testing.T) {
	type output struct {
		Value *string `json:"Value,omitempty"`
	}
	value := output{ptr("old")}
	root := xmlmodel.Root{Name: xml.Name{Local: "Root"}}
	data := []byte("\xef\xbb\xbf" + `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Root><!--safe--></Root>`)
	if err := xmlmodel.Decode(context.Background(), data, root, &value); err != nil || value.Value != nil {
		t.Fatal("valid document did not publish fresh missing fields", err)
	}
}

// pollingContext cancels deterministically during work, without clock races.
type pollingContext struct {
	context.Context
	cancel    context.CancelFunc
	remaining int
}

func (c *pollingContext) Err() error {
	c.remaining--
	if c.remaining == 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestCancellationDuringWork(t *testing.T) {
	type output struct {
		Value *string `json:"Value,omitempty"`
	}
	root := xmlmodel.Root{Name: xml.Name{Local: "Root"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	original := ptr("retained")
	value := output{original}
	err := xmlmodel.Decode(&pollingContext{ctx, cancel, 12}, []byte(`<Root><Value>new</Value>`+strings.Repeat(`<Future/>`, 50)+`</Root>`), root, &value)
	if !errors.Is(err, context.Canceled) || value.Value != original {
		t.Fatal("parse cancellation published partial output", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	type repeated struct {
		Items []string `json:"Item,omitempty"`
	}
	data, err := xmlmodel.Encode(&pollingContext{ctx, cancel, 12}, root, repeated{make([]string, 50)})
	if !errors.Is(err, context.Canceled) || data != nil {
		t.Fatal("encode cancellation published bytes", err)
	}
	if data, err := xmlmodel.Encode(context.Background(), root, repeated{make([]string, 65536)}); !errors.Is(err, xmlmodel.ErrLimit) || data != nil {
		t.Fatal("encode element limit missing", err)
	}
}

func TestLimitsCancellationAndUnsupportedBindings(t *testing.T) {
	root := xmlmodel.Root{Name: xml.Name{Local: "Root"}}
	type simple struct {
		Secret *string `json:"Secret,omitempty"`
	}
	for _, data := range []string{strings.Repeat("<Root>", 65) + strings.Repeat("</Root>", 65), "<Root>" + strings.Repeat("<Future/>", 65536) + "</Root>", "<Root>" + strings.Repeat(" ", 8<<20) + "</Root>"} {
		var output simple
		if err := xmlmodel.Decode(context.Background(), []byte(data), root, &output); !errors.Is(err, xmlmodel.ErrLimit) {
			t.Fatal("limit was not enforced", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var output simple
	if err := xmlmodel.Decode(ctx, []byte(`<Root/>`), root, &output); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if data, err := xmlmodel.Encode(ctx, root, simple{}); !errors.Is(err, context.Canceled) || data != nil {
		t.Fatal(err)
	}
	type cyclic struct {
		Child *cyclic `json:"Child,omitempty"`
	}
	duplicate := reflect.New(reflect.StructOf([]reflect.StructField{
		{Name: "First", Type: reflect.TypeFor[*string](), Tag: `json:"Value,omitempty"`},
		{Name: "Second", Type: reflect.TypeFor[*string](), Tag: `json:"Value,omitempty"`},
	})).Elem().Interface()
	type noTag struct{ Value string }
	type embedded struct{ simple }
	bindings := []any{nil, (*simple)(nil), 42, struct {
		Map map[string]string `json:"Map,omitempty"`
	}{}, struct {
		Bytes []byte `json:"Bytes,omitempty"`
	}{}, cyclic{}, duplicate, noTag{}, embedded{}}
	for _, input := range bindings {
		if data, err := xmlmodel.Encode(context.Background(), root, input); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
			t.Fatalf("unsupported binding %T accepted: %v", input, err)
		}
	}
	if data, err := xmlmodel.Encode(context.Background(), root, simple{ptr(strings.Repeat("<", 2<<20))}); !errors.Is(err, xmlmodel.ErrLimit) || data != nil {
		t.Fatal("escaped byte limit missing", err)
	}
	if data, err := xmlmodel.Encode(context.Background(), root, cors{Rules: make([]*rule, 65536)}); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
		t.Fatal("nil repeated item accepted", err)
	}
	for _, text := range []string{"bad\x00text", string([]byte{0xff})} {
		if data, err := xmlmodel.Encode(context.Background(), root, simple{ptr(text)}); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
			t.Fatal("invalid XML text accepted", err)
		}
	}
	if data, err := xmlmodel.Encode(context.Background(), xmlmodel.Root{Name: xml.Name{Local: "bad:name"}}, simple{}); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
		t.Fatal("qualified local name accepted", err)
	}
	if data, err := xmlmodel.Encode(context.Background(), xmlmodel.Root{Name: root.Name, ScalarField: "missing"}, simple{}); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
		t.Fatal("unknown scalar root field accepted", err)
	}
	if data, err := xmlmodel.Encode(context.Background(), root, struct {
		Value *float64 `json:"Value,omitempty"`
	}{ptr(math.Inf(1))}); !errors.Is(err, xmlmodel.ErrInvalid) || data != nil {
		t.Fatal("non-finite float accepted", err)
	}
}

func ExampleDecode() {
	type result struct {
		Region *string `json:"LocationConstraint,omitempty"`
	}
	root := xmlmodel.Root{Name: xml.Name{Local: "LocationConstraint"}, ScalarField: "LocationConstraint"}
	var output result
	if err := xmlmodel.Decode(context.Background(), []byte(`<LocationConstraint>oss-cn-beijing</LocationConstraint>`), root, &output); err != nil {
		panic(err)
	}
	fmt.Println(*output.Region)
	// Output: oss-cn-beijing
}
