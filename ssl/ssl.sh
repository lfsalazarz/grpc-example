SERVER_CN=issuer
# Generate new CA certificate ca.pem file
openssl genrsa 4096 > ca-key.pem
openssl req -new -x509 -nodes -days 3600 -key ca-key.pem -out ca-cert.pem -subj "/CN=${SERVER_CN}"

SERVER_CN=localhost
# create the server-side certificates
openssl req -newkey rsa:4096 -days 3600 -nodes -keyout server-key.pem -out server-req.pem -subj "/CN=${SERVER_CN}"
openssl rsa -in server-key.pem -out server-key.pem
openssl x509 -req -in server-req.pem -days 3600 -CA ca-cert.pem -CAkey ca-key.pem -set_serial 01 -out server-cert.pem

openssl verify -CAfile ca-cert.pem server-cert.pem
