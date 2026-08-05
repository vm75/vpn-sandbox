.PHONY: build run start stop logs test test-clean debug debug-start debug-stop
default: build

build:
	podman build -f Containerfile -t vm75/vpn-sandbox .

test:
	podman compose -f data/compose.yml down
	podman build -f Containerfile -t localhost/vm75/vpn-sandbox:latest .
	podman compose -f data/compose.yml up -d --force-recreate

test-clean:
	podman compose -f data/compose.yml down

test-logs:
	podman logs vpn

test-sh:
	podman exec -ti vpn sh
