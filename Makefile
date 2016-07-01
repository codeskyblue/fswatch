build-all:
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="linux darwin" \
	-arch="amd64 386" \
	-output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" .