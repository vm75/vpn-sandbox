.PHONY: build run clean start stop logs test debug debug-start debug-stop
default: build

build:
	podman build -f Containerfile -t vm75/vpn-sandbox .

run:
	podman compose -f compose.yml.example up -d

clean:
	podman compose -f compose.yml.example down

start:
	podman compose start

stop:
	podman compose stop

logs:
	podman logs vpn

sh:
	podman exec -ti vpn sh

test:
	podman build -f Containerfile -t localhost/vm75/vpn-sandbox:latest .
	podman compose -f data/compose.yml up -d --force-recreate

debug:
	podman run --rm --cap-add NET_ADMIN -p 8080:80 -p 1080:1080 -p 3128:3128 -v ./test:/data vm75/vpn-sandbox

debug-start:
	./debug/cmd.sh run

debug-stop:
	./debug/cmd.sh stop
