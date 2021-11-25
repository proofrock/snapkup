.PHONY: list
list:
	@LC_ALL=C $(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

build-prepare:
	make clean
	mkdir bin

clean:
	rm -rf bin

build:
	make build-prepare
	cd src; CGO_ENABLED=1 go build -a -tags netgo,osusergo -ldflags '-w -extldflags "-static"' -o snapkup
	mv src/snapkup bin/

zbuild:
	make build
	cd bin; 7zr a -mx9 -t7z snapkup-linux-`uname -m`.7z snapkup
