prepare:
	docker pull consul:latest
	#docker run -d --name=dev-consul --net=host  consul:latest
	docker run -d -p 8500:8500 --restart=always --name=consul consul:latest agent -server -bootstrap -ui -node=1 -client='0.0.0.0'
