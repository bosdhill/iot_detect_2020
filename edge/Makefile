EDGE_RUN_FLAGS=--server-addr $(EDGE_SERVER) --app-server-addr $(APP_SERVER) --with-cuda=$(W_CUDA) --matprofile=$(EDGE_MAT_PROFILE)
edge:
	go build $(EDGE_BUILD_FLAGS)
run:
	make edge
	./edge $(EDGE_RUN_FLAGS) 
clean:
	rm edge
