.PHONY: run-tests test
testpath := manage

# Path to the environment file
TEST_ENV_FILE  =./src/configs/.env.test
LOCAL_ENV_FILE =./src/configs/.env.dev
IMAGE_NAME = "manage_tg_script"

build:
	docker build -t $(IMAGE_NAME) -f ./src/Dockerfile.manage .

run-tests:
	-@ docker run -t --name test_manage_tg_script --env-file $(TEST_ENV_FILE) $(IMAGE_NAME) \
		pytest ${testpath}

clean-tests:
	docker rm -f /test_manage_tg_script   

# Default target: build the image, run the container, and clean up afterward
test: build run-tests clean-tests

black:
	black ./src/manage
ruff:
	ruff ./src/manage --fix

format: black ruff

