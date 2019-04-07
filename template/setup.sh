#/bin/bash -ex
echo ""
echo ""
echo "                   Welcome to OSA Labs"

export RESOURCE_URL={{.Hostname}}
CREDENTIALS=$(curl -sSk ${RESOURCE_URL}/credentials)
WORKER=$(curl -sSk ${RESOURCE_URL}/worker)
T="$(mktemp -d)"

if command -v python3 &>/dev/null; then
    USERNAME=$(echo ${CREDENTIALS} | python3 -c "import sys, json; print(json.load(sys.stdin)['username'])")
    PASSWORD=$(echo ${CREDENTIALS} | python3 -c "import sys, json; print(json.load(sys.stdin)['password'])")
    echo ${WORKER} | python3 -c "import sys, json; print(json.load(sys.stdin)['sshKey'])" > ${T}/id_rsa
    IP=$(echo ${WORKER} | python3 -c "import sys, json; print(json.load(sys.stdin)['ip'])")
else
    USERNAME=$(echo ${CREDENTIALS} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["username"]')
    PASSWORD=$(echo ${CREDENTIALS} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["password"]')
    echo ${WORKER} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["sshKey"]' > ${T}/id_rsa
    IP=$(echo ${WORKER} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["ip"]')
fi

chmod 600 ${T}/id_rsa

echo ""
echo "            Your Azure Portal credentials are:"
echo ""
echo "Username: ${USERNAME}"
echo "Password: ${PASSWORD}"

echo "Portal URL: https://portal.azure.com"

echo "Configuring workstation with unique ssh key to the bastion host..."
sleep 2
echo ""
echo "ssh command:"
echo ""
echo "ssh ${IP} -p 2222 -i ${T}/id_rsa"
echo ""
