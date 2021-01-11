OUTDIR = build

all: web seed lambda lut

clean:
	rm -rf $(OUTDIR)

lut:
	mkdir -p $(OUTDIR)/lut
	cp -R ./lut $(OUTDIR)

cuda: lut
	nvcc -o $(OUTDIR)/zeta_machine cuda/main.cu cuda/zeta.cu

web: 
	go build -o $(OUTDIR)/zeta_web ./cmd/web/main.go

lambda:
	go build -o $(OUTDIR)/zeta_lambda ./cmd/lambda/main.go
	~/go/bin/build-lambda-zip -o ./build/zeta_lambda.zip ./build/zeta_lambda

seed: 
	go build -o $(OUTDIR)/zeta_seed ./cmd/seed/main.go