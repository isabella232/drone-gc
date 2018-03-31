// Copyright 2018 Drone.IO Inc
// Use of this software is governed by the Business Source License
// that can be found in the LICENSE file.

package gc

import (
	"context"
	"time"

	"docker.io/go-docker"
)

// Collector defines a Docker container garbage collector.
type Collector interface {
	Collect(context.Context) error
}

type collector struct {
	client docker.APIClient

	whitelist []string // reserved containers
	reserved  []string // reserved images
	threshold int64    // target threshold in bytes
}

// New returns a garbage collector.
func New(client docker.APIClient, opt ...Option) Collector {
	c := new(collector)
	c.client = client
	for _, o := range opt {
		o(c)
	}
	return c
}

func (c *collector) Collect(ctx context.Context) error {
	c.collectContainers(ctx)
	c.collectDanglingImages(ctx)
	c.collectImages(ctx)
	c.collectNetworks(ctx)
	c.collectVolumes(ctx)
	return nil
}

// Schedule schedules the garbage collector to execute at the
// specified interval duration.
func Schedule(ctx context.Context, collector Collector, interval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			collector.Collect(ctx)
		}
	}
}