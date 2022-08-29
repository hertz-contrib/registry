prepare:
	docker pull consul:latest
	docker run -d --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul:latest
