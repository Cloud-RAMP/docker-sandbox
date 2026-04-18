NUM_CONTAINERS=5

build_container:
	docker build -t container-sandbox .

start_containers:
	for i in $$(seq 0 $$(($(NUM_CONTAINERS)-1))); do \
        docker run -d \
            --name "container-sandbox-$$i" \
            -v "/tmp/cloud_ramp/sockets/$$i:/tmp/cloud_ramp/sockets/$$i" \
            container-sandbox $$i; \
    done

stop_containers:
	for i in $$(seq 0 $$(($(NUM_CONTAINERS)-1))); do \
        docker stop "container-sandbox-$$i"; \
    done

remove_containers:
	for i in $$(seq 0 $$(($(NUM_CONTAINERS)-1))); do \
        docker rm "container-sandbox-$$i"; \
    done

clean_containers:
	docker ps -q --filter "name=container-sandbox" | xargs -r docker stop
	docker ps -a -q --filter "name=container-sandbox" | xargs -r docker rm