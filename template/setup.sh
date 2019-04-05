#/bin/bash -ex
echo ""
echo ""
echo "                   Welcome to OSA Labs"

export RESOURCE_URL={{.Hostname}}
CREDENTIALS=$(curl -sSk ${RESOURCE_URL}/credentials)
USERNAME=$(echo ${CREDENTIALS} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["username"]')
PASSWORD=$(echo ${CREDENTIALS} | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["password"]')

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
echo "ssh ${USERNAME}@bastion.osadev.cloud"
echo ""
echo ""
