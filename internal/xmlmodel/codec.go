// Package xmlmodel binds reviewed XML roots to concrete DSL wire models.
// It is an internal protocol helper, not an OSS client or root discovery policy.
package xmlmodel

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxBytes = 8 << 20
const maxDepth = 64
const maxElements = 65536

const xmlNamespace = "http://www.w3.org/XML/1998/namespace"
const xmlnsNamespace = "http://www.w3.org/2000/xmlns/"

var declarationPattern = regexp.MustCompile(`^version\s*=\s*(?:"1\.0"|'1\.0')(?:\s+encoding\s*=\s*(?:"[Uu][Tt][Ff]-8"|'[Uu][Tt][Ff]-8'))?(?:\s+standalone\s*=\s*(?:"(?:yes|no)"|'(?:yes|no)'))?\s*$`)

// ErrInvalid indicates malformed XML or an unsupported/ambiguous model binding.
// Its text never includes response bytes or field values.
var ErrInvalid = errors.New("xmlmodel: invalid document or model binding")

// ErrLimit indicates the codec's byte, nesting or element limit was exceeded.
var ErrLimit = errors.New("xmlmodel: document exceeds codec limit")

// Root supplies source-reviewed XML root semantics; no action-name inference occurs.
type Root struct {
	// Name is the exact expanded XML root name, including its namespace URI.
	Name xml.Name
	// ScalarField is the native wire field containing root text. Empty means
	// the root contains model children. Scalar bindings require one scalar field.
	ScalarField string
}

type field struct {
	name  string
	index int
}

func validName(name string) bool {
	if name == "" || !utf8.ValidString(name) {
		return false
	}
	for i, r := range name {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && (unicode.IsDigit(r) || unicode.IsMark(r) || r == '-' || r == '.')) {
			continue
		}
		return false
	}
	return true
}

func fields(t reflect.Type) ([]field, error) {
	var result []field
	seen := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		definition := t.Field(i)
		tag, present := definition.Tag.Lookup("json")
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}
		if !present || definition.PkgPath != "" || definition.Anonymous || !validName(name) || seen[name] {
			return nil, ErrInvalid
		}
		seen[name] = true
		result = append(result, field{name, i})
	}
	return result, nil
}

func validateType(t reflect.Type, active map[reflect.Type]bool, depth int) error {
	if depth > maxDepth {
		return ErrLimit
	}
	if active[t] {
		return ErrInvalid
	}
	active[t] = true
	defer delete(active, t)
	switch t.Kind() {
	case reflect.Pointer:
		if t.Elem().Kind() == reflect.Pointer || t.Elem().Kind() == reflect.Slice {
			return ErrInvalid
		}
		return validateType(t.Elem(), active, depth+1)
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Slice || t.Elem().Kind() == reflect.Uint8 {
			return ErrInvalid
		}
		return validateType(t.Elem(), active, depth+1)
	case reflect.Struct:
		modelFields, err := fields(t)
		if err != nil {
			return err
		}
		for _, f := range modelFields {
			if err := validateType(t.Field(f.index).Type, active, depth+1); err != nil {
				return err
			}
		}
		return nil
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return nil
	default:
		return ErrInvalid
	}
}

func binding(root Root, t reflect.Type) ([]field, error) {
	if !validName(root.Name.Local) || !validNamespace(root.Name.Space) || t.Kind() != reflect.Struct {
		return nil, ErrInvalid
	}
	if err := validateType(t, map[reflect.Type]bool{}, 0); err != nil {
		return nil, err
	}
	f, err := fields(t)
	if err != nil {
		return nil, err
	}
	if root.ScalarField != "" {
		if len(f) != 1 || f[0].name != root.ScalarField {
			return nil, ErrInvalid
		}
		typeOf := t.Field(f[0].index).Type
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.Kind() == reflect.Struct || typeOf.Kind() == reflect.Slice {
			return nil, ErrInvalid
		}
	}
	return f, nil
}

type element struct {
	name       xml.Name
	rawName    xml.Name
	namespaces map[string]string
	text       strings.Builder
	children   []*element
}

func validNamespace(namespace string) bool {
	if namespace == "" {
		return true
	}
	if !validText(namespace) || strings.IndexFunc(namespace, unicode.IsSpace) >= 0 || namespace == xmlnsNamespace {
		return false
	}
	parsed, err := url.Parse(namespace)
	return err == nil && parsed.IsAbs()
}

func resolveName(raw xml.Name, namespaces map[string]string, attribute bool) (xml.Name, error) {
	if !validName(raw.Local) || (raw.Space != "" && !validName(raw.Space)) {
		return xml.Name{}, ErrInvalid
	}
	if raw.Space == "" && attribute {
		return raw, nil
	}
	space, known := namespaces[raw.Space]
	if raw.Space != "" && !known {
		return xml.Name{}, ErrInvalid
	}
	return xml.Name{Space: space, Local: raw.Local}, nil
}

func startElement(start xml.StartElement, parent *element) (*element, error) {
	namespaces := map[string]string{"xml": xmlNamespace}
	if parent != nil {
		for prefix, namespace := range parent.namespaces {
			namespaces[prefix] = namespace
		}
	}
	declared := map[string]bool{}
	for _, attr := range start.Attr {
		prefix := ""
		if attr.Name.Space == "xmlns" {
			prefix = attr.Name.Local
		} else if attr.Name.Space != "" || attr.Name.Local != "xmlns" {
			continue
		}
		if declared[prefix] || (prefix != "" && (!validName(prefix) || attr.Value == "")) || prefix == "xmlns" || !validNamespace(attr.Value) || (prefix == "xml") != (attr.Value == xmlNamespace) {
			return nil, ErrInvalid
		}
		declared[prefix] = true
		namespaces[prefix] = attr.Value
	}
	name, err := resolveName(start.Name, namespaces, false)
	if err != nil {
		return nil, err
	}
	seen := map[xml.Name]bool{}
	for _, attr := range start.Attr {
		if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
			continue
		}
		name, err := resolveName(attr.Name, namespaces, true)
		if err != nil || seen[name] {
			return nil, ErrInvalid
		}
		seen[name] = true
	}
	return &element{name: name, rawName: start.Name, namespaces: namespaces}, nil
}

func document(ctx context.Context, data []byte) (*element, error) {
	if len(data) > maxBytes {
		return nil, ErrLimit
	}
	if !utf8.Valid(data) {
		return nil, ErrInvalid
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var root *element
	var stack []*element
	count, declaration, firstToken := 0, false, true
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		token, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrInvalid
		}
		switch token := token.(type) {
		case xml.StartElement:
			count++
			if count > maxElements || len(stack) >= maxDepth {
				return nil, ErrLimit
			}
			var parent *element
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			child, err := startElement(token, parent)
			if err != nil {
				return nil, err
			}
			if len(stack) == 0 {
				if root != nil {
					return nil, ErrInvalid
				}
				root = child
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, child)
			}
			stack = append(stack, child)
		case xml.EndElement:
			if len(stack) == 0 || stack[len(stack)-1].rawName != token.Name {
				return nil, ErrInvalid
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 {
				if strings.Trim(string(token), " \t\r\n") != "" {
					return nil, ErrInvalid
				}
			} else {
				stack[len(stack)-1].text.Write(token)
			}
		case xml.Directive:
			return nil, ErrInvalid
		case xml.ProcInst:
			if !firstToken || root != nil || declaration || token.Target != "xml" || !validText(string(token.Inst)) || !declarationPattern.Match(token.Inst) {
				return nil, ErrInvalid
			}
			declaration = true
		case xml.Comment:
			if bytes.Contains(token, []byte("--")) || bytes.HasSuffix(token, []byte("-")) {
				return nil, ErrInvalid
			}
		}
		firstToken = false
	}
	if root == nil || len(stack) != 0 {
		return nil, ErrInvalid
	}
	return root, nil
}

func decode(ctx context.Context, node *element, value reflect.Value, scalar bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if value.Kind() == reflect.Pointer {
		value.Set(reflect.New(value.Type().Elem()))
		return decode(ctx, node, value.Elem(), scalar)
	}
	if value.Kind() == reflect.Struct && !scalar {
		if strings.Trim(node.text.String(), " \t\r\n") != "" {
			return ErrInvalid
		}
		modelFields, _ := fields(value.Type())
		byName := map[string]int{}
		for _, f := range modelFields {
			byName[f.name] = f.index
		}
		seen := map[string]bool{}
		for _, child := range node.children {
			index, known := byName[child.name.Local]
			if !known {
				continue
			}
			if child.name.Space != node.name.Space {
				return ErrInvalid
			}
			field := value.Field(index)
			if field.Kind() == reflect.Slice {
				item := reflect.New(field.Type().Elem()).Elem()
				if err := decode(ctx, child, item, false); err != nil {
					return err
				}
				field.Set(reflect.Append(field, item))
			} else {
				if seen[child.name.Local] {
					return ErrInvalid
				}
				seen[child.name.Local] = true
				if err := decode(ctx, child, field, false); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if len(node.children) != 0 {
		return ErrInvalid
	}
	text := node.text.String()
	var err error
	switch value.Kind() {
	case reflect.String:
		value.SetString(text)
	case reflect.Bool:
		var parsed bool
		if text != "true" && text != "false" && text != "1" && text != "0" {
			return ErrInvalid
		}
		parsed, err = strconv.ParseBool(text)
		value.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var parsed int64
		parsed, err = strconv.ParseInt(text, 10, value.Type().Bits())
		value.SetInt(parsed)
	case reflect.Float32, reflect.Float64:
		var parsed float64
		parsed, err = strconv.ParseFloat(text, value.Type().Bits())
		if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return ErrInvalid
		}
		value.SetFloat(parsed)
	default:
		return ErrInvalid
	}
	if err != nil {
		return ErrInvalid
	}
	return nil
}

// Decode parses one bounded UTF-8 XML document and publishes a fresh model only
// on success. Limits are 8 MiB, 64 element levels and 65,536 elements. Unknown
// children are ignored; duplicate singular fields and unsupported types fail.
// Context cancellation is preserved. Callers own output and must not share it
// concurrently. Input bytes are never mutated or retained.
func Decode(ctx context.Context, data []byte, root Root, output any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	value := reflect.ValueOf(output)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return ErrInvalid
	}
	modelFields, err := binding(root, value.Elem().Type())
	if err != nil {
		return err
	}
	node, err := document(ctx, data)
	if err != nil {
		return err
	}
	if node.name != root.Name {
		return ErrInvalid
	}
	temporary := reflect.New(value.Elem().Type()).Elem()
	if root.ScalarField != "" {
		err = decode(ctx, node, temporary.Field(modelFields[0].index), true)
	} else {
		err = decode(ctx, node, temporary, false)
	}
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	value.Elem().Set(temporary)
	return nil
}

type boundedBuffer struct{ bytes.Buffer }

func (b *boundedBuffer) Write(data []byte) (int, error) {
	if len(data) > maxBytes-b.Len() {
		return 0, ErrLimit
	}
	return b.Buffer.Write(data)
}

func validText(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	for _, r := range text {
		if r != '\t' && r != '\n' && r != '\r' && (r < 0x20 || r == 0xfffe || r == 0xffff) {
			return false
		}
	}
	return true
}

func encode(ctx context.Context, encoder *xml.Encoder, value reflect.Value, name xml.Name, count *int, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return encode(ctx, encoder, value.Elem(), name, count, depth)
	}
	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if value.Index(i).Kind() == reflect.Pointer && value.Index(i).IsNil() {
				return ErrInvalid
			}
			if err := encode(ctx, encoder, value.Index(i), name, count, depth); err != nil {
				return err
			}
		}
		return nil
	}
	*count++
	if *count > maxElements || depth > maxDepth {
		return ErrLimit
	}
	start := xml.StartElement{Name: name}
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	if value.Kind() == reflect.Struct {
		modelFields, _ := fields(value.Type())
		for _, f := range modelFields {
			if err := encode(ctx, encoder, value.Field(f.index), xml.Name{Space: name.Space, Local: f.name}, count, depth+1); err != nil {
				return err
			}
		}
	} else {
		var text string
		switch value.Kind() {
		case reflect.String:
			text = value.String()
		case reflect.Bool:
			text = strconv.FormatBool(value.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			text = strconv.FormatInt(value.Int(), 10)
		case reflect.Float32, reflect.Float64:
			if math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
				return ErrInvalid
			}
			text = strconv.FormatFloat(value.Float(), 'g', -1, value.Type().Bits())
		default:
			return ErrInvalid
		}
		if !validText(text) {
			return ErrInvalid
		}
		if err := encoder.EncodeToken(xml.CharData(text)); err != nil {
			return err
		}
	}
	return encoder.EncodeToken(start.End())
}

// Encode returns a new bounded deterministic XML document in model field order.
// It omits nil optional pointers/slices, preserves explicit empty scalar values,
// escapes text and rejects nil array items. Empty arrays have no XML elements.
// Limits match Decode. It never mutates/retains input; concurrent use requires
// independent input ownership. Context errors are preserved; failures return no bytes.
func Encode(ctx context.Context, root Root, input any) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	value := reflect.ValueOf(input)
	if !value.IsValid() {
		return nil, ErrInvalid
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, ErrInvalid
		}
		value = value.Elem()
	}
	modelFields, err := binding(root, value.Type())
	if err != nil {
		return nil, err
	}
	if root.ScalarField != "" {
		value = value.Field(modelFields[0].index)
		if value.Kind() == reflect.Pointer && value.IsNil() {
			return nil, ErrInvalid
		}
	}
	buffer := &boundedBuffer{}
	encoder := xml.NewEncoder(buffer)
	count := 0
	if err := encode(ctx, encoder, value, root.Name, &count, 1); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
