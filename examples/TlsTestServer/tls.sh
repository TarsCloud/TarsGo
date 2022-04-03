#!/bin/bash
# 参考文档
# https://www.mingfer.cn/2020/06/13/altern-name/
# https://www.jianshu.com/p/ea5bc56211ee
# https://github.com/square/certstrap
# https://github.com/cloudflare/cfssl
# https://stackoverflow.com/questions/54622879/cannot-validate-certificate-for-ip-address-because-it-doesnt-contain-any-ip-s
# https://security.stackexchange.com/questions/74345/provide-subjectaltname-to-openssl-directly-on-the-command-line/183973#183973
# https://blog.csdn.net/u012094456/article/details/101352543
set -ex
mkdir -p ssl
cd ssl
# ca
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 1024 -out ca.crt \
  -subj "/C=CN/ST=BeiJing/L=BJ/O=Tars/OU=TarsGo/CN=TarsCa/emailAddress=tarsgo@tarscloud.com"

# server
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/C=CN/ST=BeiJing/L=BJ/O=Tars/OU=TarsGo/CN=server/emailAddress=tarsgo@tarscloud.com"\
  -addext "subjectAltName=IP:127.0.0.1"
#openssl req -text -in server.csr -noout -verify
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 500 \
  -extfile <(printf "subjectAltName=IP:127.0.0.1")
#openssl x509 -in server.crt -noout -text

# client
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/C=CN/ST=BeiJing/L=BJ/O=Tars/OU=TarsGo/CN=client/emailAddress=tarsgo@tarscloud.com"
#openssl req -text -in client.csr -noout -verify
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 500
cd -
