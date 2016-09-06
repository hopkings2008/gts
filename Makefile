.PHONY: all clean

GOFLAGS=-gcflags "-N -l"

all:
	export CGO_ENABLED=0 && go build $(GOFLAGS) -o gts

clean:
	-@rm -f gts
