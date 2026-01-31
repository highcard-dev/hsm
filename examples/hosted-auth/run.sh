
#check if cleanup env var is et

if [ -n "$CLEANUP" ]; then
    rm *.json
    docker compose down --remove-orphans
fi



docker run --rm -it -v $PWD:/home/step smallstep/step-cli step crypto jwk create jwk.pub.json jwk.json --no-password --insecure
#generate keyset
cat jwk.pub.json | docker run --rm -i -v $PWD:/home/step smallstep/step-cli step crypto jwk keyset add jwks.json

export JWT_TOKEN=$(docker run --rm -v $PWD:/home/step smallstep/step-cli step crypto jwt sign \
  --key jwk.json \
  --iss "auth.example.com" \
  --aud "api.example.com" \
  --sub gsp-user --subtle)

#explicitly initialize session.json, for better console experience
docker compose run --rm hsm login

docker compose up -d

echo "Waiting for server to start..."
sleep 5

export HSM_URL=http://localhost:8080

../../scripts/gsp-update-hytale-server.sh
../../scripts/gsp-start-hytale-server.sh