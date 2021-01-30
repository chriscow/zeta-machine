OUTDIR = build

all: web seed lambda

clean:
	rm -rf $(OUTDIR)

cuda: lut
	nvcc -o $(OUTDIR)/zeta_machine cuda/main.cu cuda/zeta.cu

web: 
	go build -o $(OUTDIR)/zeta_web ./cmd/web/.

lambda:
	go build -o $(OUTDIR)/zeta_lambda ./cmd/lambda/.
	~/go/bin/build-lambda-zip -o ./build/zeta_lambda.zip ./build/zeta_lambda

seed: 
	go build -o $(OUTDIR)/zeta_seed ./cmd/seed/.