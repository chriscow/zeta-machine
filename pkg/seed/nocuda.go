// +build !BUILD_CUDA

package seed

import (
	"context"
	"zetamachine/pkg/zeta"
)

// Generate patch data via call to cuda zeta machine library
func (p *Patch) Generate(ctx context.Context) {
	algo := zeta.Algo{}
	p.Data = algo.Compute(ctx, p.GetMin(), p.GetMax(), PatchWidth*PatchWidth)
}
