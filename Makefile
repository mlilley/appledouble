TARGET	:=	appledouble
VERSION	:=	$(shell cat ./VERSION)
INSTALL_DIR	:= /usr/local/bin

.PHONY: all build install clean test

all: build

build:
	go build -v

install: 
	-mv appledouble $(INSTALL_DIR)

clean:
	-rm -f $(TARGET)

test:
	go test ./... -v
