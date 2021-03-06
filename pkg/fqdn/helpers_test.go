// Copyright 2019 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !privileged_tests

package fqdn

import (
	"net"
	"time"

	"github.com/cilium/cilium/pkg/checker"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func (ds *DNSCacheTestSuite) TestKeepUniqueNames(c *C) {
	testData := []struct {
		argument []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "a", "c"}, []string{"a", "b", "c"}},
		{[]string{""}, []string{""}},
		{[]string{}, []string{}},
	}

	for _, item := range testData {
		val := KeepUniqueNames(item.argument)
		c.Assert(val, checker.DeepEquals, item.expected)
	}
}

func (ds *DNSCacheTestSuite) TestMapIPsToSelectors(c *C) {

	var (
		ciliumIP1 = net.ParseIP("1.2.3.4")
		ciliumIP2 = net.ParseIP("1.2.3.5")
	)

	log.Level = logrus.DebugLevel

	// Create DNS cache
	now := time.Now()
	cache := NewDNSCache(60)

	selectors := map[api.FQDNSelector]struct{}{
		ciliumIOSel: {},
	}

	// Empty cache.
	selsMissingIPs, selIPMapping := mapSelectorsToIPs(selectors, cache)
	c.Assert(len(selsMissingIPs), Equals, 1)
	c.Assert(selsMissingIPs[0], Equals, ciliumIOSel)
	c.Assert(len(selIPMapping), Equals, 0)

	// Just one IP.
	changed := cache.Update(now, prepareMatchName(ciliumIOSel.MatchName), []net.IP{ciliumIP1}, 100)
	c.Assert(changed, Equals, true)
	selsMissingIPs, selIPMapping = mapSelectorsToIPs(selectors, cache)
	c.Assert(len(selsMissingIPs), Equals, 0)
	c.Assert(len(selIPMapping), Equals, 1)
	ciliumIPs, ok := selIPMapping[ciliumIOSel]
	c.Assert(ok, Equals, true)
	c.Assert(len(ciliumIPs), Equals, 1)
	c.Assert(ciliumIPs[0].Equal(ciliumIP1), Equals, true)

	// Two IPs now.
	changed = cache.Update(now, prepareMatchName(ciliumIOSel.MatchName), []net.IP{ciliumIP1, ciliumIP2}, 100)
	c.Assert(changed, Equals, true)
	selsMissingIPs, selIPMapping = mapSelectorsToIPs(selectors, cache)
	c.Assert(len(selsMissingIPs), Equals, 0)
	c.Assert(len(selIPMapping), Equals, 1)
	ciliumIPs, ok = selIPMapping[ciliumIOSel]
	c.Assert(ok, Equals, true)
	c.Assert(len(ciliumIPs), Equals, 2)
	c.Assert(ciliumIPs[0].Equal(ciliumIP1), Equals, true)
	c.Assert(ciliumIPs[1].Equal(ciliumIP2), Equals, true)

	// Test with a MatchPattern.
	selectors = map[api.FQDNSelector]struct{}{
		ciliumIOSelMatchPattern: {},
	}
	selsMissingIPs, selIPMapping = mapSelectorsToIPs(selectors, cache)
	c.Assert(len(selsMissingIPs), Equals, 0)
	c.Assert(len(selIPMapping), Equals, 1)
	ciliumIPs, ok = selIPMapping[ciliumIOSelMatchPattern]
	c.Assert(ok, Equals, true)
	c.Assert(len(ciliumIPs), Equals, 2)
	c.Assert(ciliumIPs[0].Equal(ciliumIP1), Equals, true)
	c.Assert(ciliumIPs[1].Equal(ciliumIP2), Equals, true)

	selectors = map[api.FQDNSelector]struct{}{
		ciliumIOSelMatchPattern: {},
		ciliumIOSel:             {},
	}
	selsMissingIPs, selIPMapping = mapSelectorsToIPs(selectors, cache)
	c.Assert(len(selsMissingIPs), Equals, 0)
	c.Assert(len(selIPMapping), Equals, 2)
	ciliumIPs, ok = selIPMapping[ciliumIOSelMatchPattern]
	c.Assert(ok, Equals, true)
	c.Assert(len(ciliumIPs), Equals, 2)
	c.Assert(ciliumIPs[0].Equal(ciliumIP1), Equals, true)
	c.Assert(ciliumIPs[1].Equal(ciliumIP2), Equals, true)
	ciliumIPs, ok = selIPMapping[ciliumIOSel]
	c.Assert(ok, Equals, true)
	c.Assert(len(ciliumIPs), Equals, 2)
	c.Assert(ciliumIPs[0].Equal(ciliumIP1), Equals, true)
	c.Assert(ciliumIPs[1].Equal(ciliumIP2), Equals, true)
}
