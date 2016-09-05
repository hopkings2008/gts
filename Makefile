.PHONY: all clean

GOFLAGS=-gcflags "-N -l"

all:
	go build $(GOFLAGS) -o gts

clean:
	-@rm -f gts
