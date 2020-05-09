#!/bin/bash

SERVER_ADDR="http://localhost:18443"
CLI_DIR="../cmd/hctl"
TEST_DIR=$PWD

function test_command() {
  ./hctl $@
  if [ $? != 0 ]; then
    echo "hctl $@ command expect success, but got error"
    exit 1
  fi
}

curl $SERVER_ADDR/healthz -s -o /dev/null
if [ $? != 0 ]; then
  echo "Before test, please run a server"
  exit 1
fi

cd $CLI_DIR
go build

# login
test_command login --project master --name admin --password password

# config
## TODO(get)
## TODO(set)

# project
## create
test_command project add --file $TEST_DIR/inputs/project_create.json

## get all
test_command project get

## get
test_command project get --name new-project

## TODO(update)

## delete
test_command project delete --name new-project

# user
## create
test_command user add --project master --file $TEST_DIR/inputs/user_create.json

## get all
test_command user get --project master

## get
test_command user get --project master --name user1

## TODO(update)

## add user role
test_command user role add --project master --user user1 --role read-project

## TODO(delete user role)

## TODO(password change)

## delete
test_command user delete --project master --name user1

# client
## create
test_command client add --project master --file $TEST_DIR/inputs/client_create.json

## get all
test_command client get --project master

## get
test_command client get --project master --id oidc-client

## TODO(update)

## delete
test_command client delete --project master --id oidc-client

# role
## create
test_command role add --project master --name viewer

## TODO(get all)

## TODO(get)

## TODO(update)

## TODO(delete)

# logout
test_command logout

echo "successfully finished test"