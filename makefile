OUTDIR = build

all: web

clean:
	rm -rf $(OUTDIR)

lut:
	mkdir -p $(OUTDIR)/lut
	cp -R ./cuda/lut $(OUTDIR)

cuda: lut
	nvcc -o $(OUTDIR)/zeta_machine cuda/main.cu cuda/zeta.cu

web: cuda
	go build -o $(OUTDIR)/zeta_web ./apps/web/main.go