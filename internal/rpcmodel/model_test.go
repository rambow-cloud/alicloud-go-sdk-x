package rpcmodel

import (
	"context"
	"encoding/hex"
	"errors"
	"math"
	"net/url"
	"reflect"
	"testing"
)

func TestPolicyHelpersTypedNilTokensAndCancellation(t *testing.T) {
	var value *struct{}
	if !IsNil(value) || !IsNil(nil) || IsNil(struct{}{}) {
		t.Fatal("typed nil detection")
	}
	first, err := NewClientToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewClientToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(*first) != 32 || *first == *second {
		t.Fatal("invalid token entropy or length")
	}
	if _, err := hex.DecodeString(*first); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewClientToken(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func ptr[T any](v T) *T { return &v }

func TestQueryPresenceWidthsCaseAndRepeatedContainers(t *testing.T) {
	type member struct {
		Key   *string `json:"Key"`
		Lower *string `json:"key"`
		Value *string `json:"Value"`
	}
	input := struct {
		Missing *string  `json:"Missing"`
		Empty   *string  `json:"Empty"`
		Zero    *int32   `json:"Zero"`
		False   *bool    `json:"False"`
		OwnerID *int64   `json:"OwnerId"`
		IDs     *string  `json:"InstanceIds"`
		Tags    []member `json:"Tag"`
		Nested  *struct {
			Values []int32 `json:"Values"`
		} `json:"Nested"`
		Map map[string]string `json:"Map"`
	}{Empty: ptr(""), Zero: ptr(int32(0)), False: ptr(false), OwnerID: ptr(int64(9007199254740993)), IDs: ptr(`["i-1","i-2"]`), Tags: []member{{Key: ptr("a & b"), Lower: ptr("lower"), Value: ptr("")}, {Key: ptr("second")}}, Map: map[string]string{"z": "last", "a": "first"}}
	input.Nested = &struct {
		Values []int32 `json:"Values"`
	}{[]int32{3, 4}}
	q, err := Query(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"Empty": {""}, "Zero": {"0"}, "False": {"false"}, "OwnerId": {"9007199254740993"}, "InstanceIds": {`["i-1","i-2"]`}, "Tag.1.Key": {"a & b"}, "Tag.1.key": {"lower"}, "Tag.1.Value": {""}, "Tag.2.Key": {"second"}, "Nested.Values.1": {"3"}, "Nested.Values.2": {"4"}, "Map.a": {"first"}, "Map.z": {"last"}}
	if !reflect.DeepEqual(q, want) {
		t.Fatalf("wire query: %v", q)
	}
}

func TestSnapshotOwnsPointersMapsSlicesAndPreservesPresence(t *testing.T) {
	type nested struct {
		Value   *string
		Numbers []int32
	}
	input := struct {
		Data    []nested
		Map     map[string][]*string
		Empty   []string
		Missing []string
	}{Data: []nested{{ptr("original"), []int32{1}}}, Map: map[string][]*string{"a": {ptr("original")}}, Empty: []string{}}
	out, err := Snapshot(context.Background(), &input)
	if err != nil {
		t.Fatal(err)
	}
	*out.Data[0].Value = "changed"
	out.Data[0].Numbers[0] = 2
	*out.Map["a"][0] = "changed"
	out.Map["b"] = nil
	if *input.Data[0].Value != "original" || input.Data[0].Numbers[0] != 1 || *input.Map["a"][0] != "original" || len(input.Map) != 1 || out.Empty == nil || out.Missing != nil {
		t.Fatal("snapshot aliases or lost presence")
	}
	nilCopy, err := Snapshot[nested](context.Background(), nil)
	if err != nil || nilCopy == nil {
		t.Fatal(err)
	}
}

func TestCyclesCollisionsUnsupportedAndCancellation(t *testing.T) {
	type recursive struct {
		Next *recursive `json:"Next"`
	}
	in := &recursive{}
	in.Next = in
	if _, err := Snapshot(context.Background(), in); err == nil {
		t.Fatal("cycle accepted")
	}
	if _, err := Query(context.Background(), in); err == nil {
		t.Fatal("invalid/cyclic query accepted")
	}
	conflict := struct {
		A string            `json:"a.b"`
		B map[string]string `json:"a"`
	}{A: "one", B: map[string]string{"b": "two"}}
	if _, err := Query(context.Background(), conflict); err == nil {
		t.Fatal("wire collision accepted")
	}
	for _, value := range []any{math.NaN(), math.Inf(1), make(chan int)} {
		if _, err := Query(context.Background(), struct {
			Value any `json:"Value"`
		}{value}); err == nil {
			t.Fatal("unsupported query accepted")
		}
	}
	if _, err := Query(context.Background(), struct {
		Value *float64 `json:"Value"`
	}{ptr(math.NaN())}); err == nil {
		t.Fatal("non-finite scalar accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Snapshot(ctx, &conflict); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Query(ctx, conflict); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
