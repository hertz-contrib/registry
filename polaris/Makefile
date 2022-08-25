polaris_version ?= v1.10.0

prepare:
	wget "https://github.com/polarismesh/polaris/releases/download/$(polaris_version)/polaris-server-release_$(polaris_version).linux.amd64.zip"
	unzip polaris-server-release_$(polaris_version).linux.amd64.zip
	bash polaris-server-release_$(polaris_version).linux.amd64/tool/start.sh

