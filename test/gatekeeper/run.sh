#!/bin/bash
trap 'kill $(jobs -p)' EXIT

go build -o test-server
./test-server -logfile=test-server.log &
echo "start test backend app"

cd ../../cmd/hekate
go build
./hekate &
echo "start hekate"

cd ../../test/gatekeeper

# wait server up
sleep 1

SERVER_ADDR="http://localhost:18443"
URL="$SERVER_ADDR/adminapi/v1"

# Get Master Token
token_info=`curl --insecure -s -X POST $SERVER_ADDR/authapi/v1/project/master/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=password" \
  -d "client_id=portal" \
  -d 'grant_type=password'`
master_access_token=`echo $token_info | jq -r .access_token`
# echo $master_access_token

# register client
curl --insecure -s -X POST $URL/project/master/client \
  -H "Authorization: Bearer $master_access_token" \
  -H "Content-Type: application/json" \
  -d "@client.json"

ls keycloak-proxy-linux-amd64 > /dev/null 2>&1
if [ $? != 0 ]; then
  wget https://github.com/keycloak/keycloak-gatekeeper/releases/download/v2.3.0/keycloak-proxy-linux-amd64
  chmod +x keycloak-proxy-linux-amd64
fi
./keycloak-proxy-linux-amd64 --config=config.yml --verbose > gatekeeper.log 2>&1 &

# wait gatekeeper up
sleep 1

# echo "access without gatekeeper"
# curl http://localhost:10000/hello
# echo ""

token_info=`curl --insecure -s -X POST $SERVER_ADDR/authapi/v1/project/master/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=password" \
  -d "client_id=gatekeeper" \
  -d 'grant_type=password'`
access_token=`echo $token_info | jq -r .access_token`

echo "access with gatekeeper"
curl http://localhost:3000/hello \
  -H "Authorization: Bearer $access_token"
echo ""

# wait