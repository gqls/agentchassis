# Create the JSON structure
cat > dockerconfig.json <<EOF
{
"auths": {
"docker.io": {
"username": "aqls",
"password": "AaD02432123!",
"email": "aaa@designconsultancy.co.uk",
"auth": "YXFsczpBYUQwMjQzMjEyMyE="
}
}
}
EOF

# Base64 encode it
cat dockerconfig.json | base64 -w 0