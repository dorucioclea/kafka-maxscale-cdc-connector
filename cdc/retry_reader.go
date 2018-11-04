// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"context"
	"time"

	"github.com/golang/glog"
)

// RetryReader store the gtid of the last message and resume there on failure
type RetryReader struct {
	Reader        Reader
	GTIDExtractor interface {
		Parse(line []byte) (*GTID, error)
	}
}

// Read from the sub reader and retry if needed
func (r *RetryReader) Read(ctx context.Context, gtid *GTID, outch chan<- []byte) error {
	ch := make(chan []byte)
	defer close(ch)
	go func() {
		for line := range ch {
			newGtid, err := r.GTIDExtractor.Parse(line)
			if err != nil {
				glog.V(2).Infof("%v", err)
			}
			outch <- line
			gtid = newGtid
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := r.Reader.Read(ctx, gtid, ch); err != nil {
				glog.Warningf("read failed: %v", err)
			}
			glog.V(3).Infof("reader closed => restart")
			time.Sleep(10 * time.Second)
		}
	}
}
