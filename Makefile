.PHONY: run-tests test
IMAGE_NAME =: "python-tg-script"
testpath := manage
containername := manage_tg_bot_script
# Path to the environment file
TEST_ENV_FILE  =./configs/.env.test
LOCAL_ENV_FILE =./configs/.env.dev

build:
	docker build -t $(IMAGE_NAME) -f ./src/Dockerfile.manage .

run-tests:
	-@ docker run -t --name "$(containername)" \
		--env-file $(TEST_ENV_FILE) $(IMAGE_NAME) \
		pytest ${testpath}

create-session:
	-@ mkdir -p ./sessions
	-@ docker run -it --network=host --name "$(containername)" \
		-v "./sessions:/usr/sessions" \
		--env-file $(LOCAL_ENV_FILE) $(IMAGE_NAME) \
		python ./manage/login.py
		
clean:
	docker rm -f "/$(containername)" 

# Default target: build the image, run the container, and clean up afterward
test: build run-tests clean

black:
	black ./src/manage
ruff:
	ruff ./src/manage --fix

format: black ruff

login: build create-session clean
