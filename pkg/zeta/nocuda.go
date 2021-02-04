// +build !BUILD_CUDA

package zeta

import (
	"context"
	"log"
	"time"
)

// ComputeRequest takes a JSON serialized tile, unmarshals it, computes the
// iteration data, base64 encodes the data and marshals it all back to JSON
func (t *Tile) ComputeRequest(ctx context.Context) {

	start := time.Now()
	algo := &Algo{}
	t.Data = algo.Compute(ctx, t.Min(), t.Max(), t.Width)
	log.Println("[tile] compute complete in ", time.Since(start), t)
}
