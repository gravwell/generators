/*************************************************************************
 * Copyright 2019 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package main

import (
	"fmt"
	"math/rand"
	"time"

	rd "github.com/Pallinder/go-randomdata"
	"github.com/gravwell/ingest"
	"github.com/gravwell/ingest/entry"
)

const (
	streamBlock = 10
)

func throw(igst *ingest.IngestMuxer, tag entry.EntryTag, cnt uint64, dur time.Duration) (err error) {
	sp := dur / time.Duration(cnt)
	ts := time.Now()
	for i := uint64(0); i < cnt; i++ {
		dt := genData(ts)
		if err = igst.WriteEntry(&entry.Entry{
			TS:   entry.FromStandard(ts),
			Tag:  tag,
			SRC:  src,
			Data: dt,
		}); err != nil {
			return
		}
		ts = ts.Add(-1 * sp)
		totalBytes += uint64(len(dt))
	}
	return
}

func stream(igst *ingest.IngestMuxer, tag entry.EntryTag, cnt uint64) (err error) {
	var blksize uint64
	if cnt < streamBlock {
		blksize = 1
	} else {
		blksize = streamBlock
	}
	sp := time.Second / time.Duration((cnt / blksize))

loop:
	for {
		for i := uint64(0); i < blksize; i++ {
			ts := time.Now()
			dt := genData(ts)
			if err = igst.WriteEntry(&entry.Entry{
				TS:   entry.FromStandard(ts),
				Tag:  tag,
				SRC:  src,
				Data: dt,
			}); err != nil {
				break loop
			}
			totalBytes += uint64(len(dt))
		}
		time.Sleep(sp)
	}
	return
}

func genConnData(ts time.Time) []byte {
	uid := randomBase62(17)

	orig, resp := ips()

	orig_port, resp_port := ports()

	local_orig := "T"
	local_resp := "T"
	if rand.Int()%2 == 0 {
		local_orig = "F"
	}
	if rand.Int()%2 == 0 {
		local_resp = "F"
	}

	proto := protos[rand.Intn(len(protos))]
	service := "-"
	if svcs, ok := services[proto]; ok {
		service = svcs[rand.Intn(len(svcs))]
	}

	duration := float64(rand.Intn(10)) + rand.Float64()

	orig_bytes := rand.Intn(1000000000)
	resp_bytes := rand.Intn(1000000000)

	orig_pkts := (orig_bytes / (40 + rand.Intn(65000)))
	resp_pkts := (orig_bytes / (40 + rand.Intn(65000)))

	orig_ip_bytes := orig_bytes + rand.Intn(500)
	resp_ip_bytes := resp_bytes + rand.Intn(500)

	conn_state := states[rand.Intn(len(states))]

	missed_bytes := 0

	history := histories[rand.Intn(len(histories))]

	tunnel_parents := "-"

	return []byte(fmt.Sprintf("%d.%6d\t%s\t%s\t%d\t%s\t%d\t%s\t%s\t%.6f\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
		ts.Unix(), ts.UnixNano()%ts.Unix(), uid,
		orig, orig_port,
		resp, resp_port,
		proto, service,
		duration,
		orig_bytes, resp_bytes,
		conn_state,
		local_orig, local_resp,
		missed_bytes,
		history,
		orig_pkts, orig_ip_bytes,
		resp_pkts, resp_ip_bytes,
		tunnel_parents,
	))
}

func ips() (string, string) {
	if *enableIPv6 && (rand.Int()&3) == 0 {
		//more IPv4 than 6
		return rd.IpV6Address(), rd.IpV6Address()
	}
	return rd.IpV4Address(), rd.IpV4Address()
}

func ports() (int, int) {
	var orig_port, resp_port int
	if rand.Int()%2 == 0 {
		orig_port = 1 + rand.Intn(2048)
		resp_port = 2048 + rand.Intn(0xffff-2048)
	} else {
		orig_port = 2048 + rand.Intn(0xffff-2048)
		resp_port = 1 + rand.Intn(2048)
	}
	return orig_port, resp_port
}

func randomBase62(l int) string {
	r := make([]byte, l)
	for i := 0; i < l; i++ {
		r[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(r)
}