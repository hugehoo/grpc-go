/*
 *
 * Copyright 2026 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package resolver_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func TestAddressIsImmutable(t *testing.T) {
	typ := reflect.TypeOf(resolver.Address{})
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i); field.IsExported() {
			t.Errorf("resolver.Address field %q is exported", field.Name)
		}
	}

	originalAttrs := attributes.New("original", true)
	originalBalancerAttrs := attributes.New("original-balancer", true)
	originalMetadata := &struct{ value string }{"original"}
	original := resolver.NewAddress("original-address").
		WithServerName("original-server-name").
		WithAttributes(originalAttrs).
		WithBalancerAttributes(originalBalancerAttrs).
		WithMetadata(originalMetadata)
	updated := original.
		WithAddr("updated-address").
		WithServerName("updated-server-name").
		WithAttributes(nil).
		WithBalancerAttributes(nil).
		WithMetadata(nil)

	if got, want := original.Addr(), "original-address"; got != want {
		t.Errorf("original.Addr() = %q, want %q", got, want)
	}
	if got, want := original.ServerName(), "original-server-name"; got != want {
		t.Errorf("original.ServerName() = %q, want %q", got, want)
	}
	if got := original.Attributes(); got != originalAttrs {
		t.Errorf("original.Attributes() = %v, want %v", got, originalAttrs)
	}
	if got := original.BalancerAttributes(); got != originalBalancerAttrs {
		t.Errorf("original.BalancerAttributes() = %v, want %v", got, originalBalancerAttrs)
	}
	if got := original.Metadata(); got != originalMetadata {
		t.Errorf("original.Metadata() = %v, want %v", got, originalMetadata)
	}

	if got, want := updated.Addr(), "updated-address"; got != want {
		t.Errorf("updated.Addr() = %q, want %q", got, want)
	}
	if got, want := updated.ServerName(), "updated-server-name"; got != want {
		t.Errorf("updated.ServerName() = %q, want %q", got, want)
	}
	if got := updated.Attributes(); got != nil {
		t.Errorf("updated.Attributes() = %v, want nil", got)
	}
	if got := updated.BalancerAttributes(); got != nil {
		t.Errorf("updated.BalancerAttributes() = %v, want nil", got)
	}
	if got := updated.Metadata(); got != nil {
		t.Errorf("updated.Metadata() = %v, want nil", got)
	}
}

func TestAddressZeroValue(t *testing.T) {
	var address resolver.Address
	if got := address.Addr(); got != "" {
		t.Errorf("zero Address.Addr() = %q, want empty", got)
	}
	if got := address.ServerName(); got != "" {
		t.Errorf("zero Address.ServerName() = %q, want empty", got)
	}
	if address.Attributes() != nil || address.BalancerAttributes() != nil || address.Metadata() != nil {
		t.Errorf("zero Address has non-nil data: %v", address)
	}
}

func TestAddressMarshalJSON(t *testing.T) {
	address := resolver.NewAddress("address").WithServerName("server-name").WithMetadata("metadata")
	got, err := json.Marshal(address)
	if err != nil {
		t.Fatalf("json.Marshal(%v) failed: %v", address, err)
	}
	want := `{"Addr":"address","ServerName":"server-name","Attributes":null,"BalancerAttributes":null,"Metadata":"metadata"}`
	if string(got) != want {
		t.Errorf("json.Marshal(%v) = %s, want %s", address, got, want)
	}
}
