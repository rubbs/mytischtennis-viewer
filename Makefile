

AIR=docker run -it --rm -w "$(PWD)" -v "$(PWD)":"$(PWD)" -p "$(AIR_PORT)":"$(AIR_PORT)" docker.io/cosmtrek/air

develop: PORT=8088
develop: AIR_PORT=$(PORT)
develop:
	$(AIR) --build.cmd "go build -o bin/api cmd/run.go" --build.bin "PORT=$(PORT) ./bin/api" --build.exclude_dir "bin" --build.include_dir "assets,cmd,parser"