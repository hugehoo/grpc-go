/*
 *
 * Copyright 2021 gRPC authors.
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

package resolver

import (
	"slices"
	"testing"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/internal/grpctest"
)

type s struct {
	grpctest.Tester
}

func Test(t *testing.T) {
	grpctest.RunSubTests(t, s{})
}

func (s) TestAddressWithers(t *testing.T) {
	originalAttrs := attributes.New("original", true)
	newAttrs := attributes.New("new", true)
	originalBalancerAttrs := attributes.New("original-balancer", true)
	newBalancerAttrs := attributes.New("new-balancer", true)
	originalMetadata := &struct{ value string }{"original"}
	newMetadata := &struct{ value string }{"new"}

	original := NewAddress("original-address").
		WithServerName("original-server-name").
		WithAttributes(originalAttrs).
		WithBalancerAttributes(originalBalancerAttrs).
		WithMetadata(originalMetadata)
	updated := original.
		WithAddr("new-address").
		WithServerName("new-server-name").
		WithAttributes(newAttrs).
		WithBalancerAttributes(newBalancerAttrs).
		WithMetadata(newMetadata)

	if got, want := original, (NewAddress("original-address").WithServerName("original-server-name").WithAttributes(originalAttrs).WithBalancerAttributes(originalBalancerAttrs).WithMetadata(originalMetadata)); !got.Equal(want) {
		t.Fatalf("original Address changed: got %v, want %v", got, want)
	}
	if got, want := updated, (NewAddress("new-address").WithServerName("new-server-name").WithAttributes(newAttrs).WithBalancerAttributes(newBalancerAttrs).WithMetadata(newMetadata)); !got.Equal(want) {
		t.Fatalf("updated Address = %v, want %v", got, want)
	}
}

func (s) TestEndpointWithersOwnAddressSlice(t *testing.T) {
	addresses := []Address{NewAddress("one"), NewAddress("two")}
	originalAttrs := attributes.New("original", true)
	newAttrs := attributes.New("new", true)
	original := NewEndpoint(addresses...).WithAttributes(originalAttrs)

	addresses[0] = NewAddress("changed")
	if got, want := slices.Collect(original.Addresses()), []Address{NewAddress("one"), NewAddress("two")}; !slices.EqualFunc(got, want, Address.Equal) {
		t.Fatalf("NewEndpoint() retained the caller's slice: got %v, want %v", got, want)
	}

	updatedAddresses := []Address{NewAddress("three"), NewAddress("four")}
	updated := original.WithAddresses(updatedAddresses...).WithAttributes(newAttrs)
	updatedAddresses[0] = NewAddress("changed-again")
	updated = updated.WithAddress(1, NewAddress("five"))

	if got, want := slices.Collect(original.Addresses()), []Address{NewAddress("one"), NewAddress("two")}; !slices.EqualFunc(got, want, Address.Equal) {
		t.Fatalf("original Endpoint changed: got %v, want %v", got, want)
	}
	if original.Attributes() != originalAttrs {
		t.Fatalf("original Endpoint attributes changed: got %v, want %v", original.Attributes(), originalAttrs)
	}
	if got, want := slices.Collect(updated.Addresses()), []Address{NewAddress("three"), NewAddress("five")}; !slices.EqualFunc(got, want, Address.Equal) {
		t.Fatalf("updated Endpoint addresses = %v, want %v", got, want)
	}
	if updated.Attributes() != newAttrs {
		t.Fatalf("updated Endpoint attributes = %v, want %v", updated.Attributes(), newAttrs)
	}
	if got, want := updated.AddressCount(), 2; got != want {
		t.Fatalf("updated.AddressCount() = %d, want %d", got, want)
	}
	if got, want := updated.Address(1), NewAddress("five"); !got.Equal(want) {
		t.Fatalf("updated.Address(1) = %v, want %v", got, want)
	}
	if !updated.Equal(NewEndpoint(NewAddress("three"), NewAddress("five")).WithAttributes(newAttrs)) {
		t.Fatalf("updated.Equal(equivalent endpoint) = false, want true")
	}
}

// TestValidateEndpoints tests different scenarios of resolver addresses being
// validated by the ValidateEndpoint helper.
func (s) TestValidateEndpoints(t *testing.T) {
	addr1 := NewAddress("addr1")
	addr2 := NewAddress("addr2")
	addr3 := NewAddress("addr3")
	addr4 := NewAddress("addr4")
	tests := []struct {
		name      string
		endpoints []Endpoint
		wantErr   bool
	}{
		{
			name: "duplicate-address-across-endpoints",
			endpoints: []Endpoint{
				NewEndpoint([]Address{addr1}...),
				NewEndpoint([]Address{addr1}...),
			},
			wantErr: false,
		},
		{
			name: "duplicate-address-same-endpoint",
			endpoints: []Endpoint{
				NewEndpoint([]Address{addr1, addr1}...),
			},
			wantErr: false,
		},
		{
			name: "duplicate-address-across-endpoints-plural-addresses",
			endpoints: []Endpoint{
				NewEndpoint([]Address{addr1, addr2, addr3}...),
				NewEndpoint([]Address{addr3, addr4}...),
			},
			wantErr: false,
		},
		{
			name: "no-shared-addresses",
			endpoints: []Endpoint{
				NewEndpoint([]Address{addr1, addr2}...),
				NewEndpoint([]Address{addr3, addr4}...),
			},
			wantErr: false,
		},
		{
			name: "endpoint-with-no-addresses",
			endpoints: []Endpoint{
				NewEndpoint([]Address{addr1, addr2}...),
				NewEndpoint([]Address{}...),
			},
			wantErr: false,
		},
		{
			name:      "empty-endpoints-list",
			endpoints: []Endpoint{},
			wantErr:   true,
		},
		{
			name:      "endpoint-list-with-no-addresses",
			endpoints: []Endpoint{{}, {}},
			wantErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateEndpoints(test.endpoints)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateEndpoints() wantErr: %v, got: %v", test.wantErr, err)
			}
		})
	}
}
