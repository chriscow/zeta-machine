OUTDIR = build

all: web seed lambda

clean:
	rm -rf $(OUTDIR)

lut:
	mkdir -p $(OUTDIR)/lut
	cp -R ./cuda/lut $(OUTDIR)

cuda: lut
	nvcc -o $(OUTDIR)/zeta_machine cuda/main.cu cuda/zeta.cu

web: 
	go build -o $(OUTDIR)/webz ./cmd/web/main.go

lambda:
	go build -o $(OUTDIR)/lambdaz ./cmd/lambda/main.go

seed: 
	go build -o $(OUTDIR)/seedz ./cmd/seed/main.go