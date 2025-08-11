#!/bin/bash

# 为特定域名生成SSL证书
# 使用方法: ./generate-domain-cert.sh example.com

set -e

if [ $# -eq 0 ]; then
    echo "❌ 请提供域名参数"
    echo "使用方法: $0 <domain-name>"
    echo "例如: $0 blockchain.example.com"
    exit 1
fi

DOMAIN=$1
CERT_DIR="./certs"
CERT_NAME="$DOMAIN"

# 创建证书目录
mkdir -p "$CERT_DIR"

echo "🔐 正在为域名 $DOMAIN 生成SSL证书..."

# 生成私钥
openssl genrsa -out "$CERT_DIR/$CERT_NAME.key" 2048

# 生成证书签名请求配置
cat > "$CERT_DIR/$CERT_NAME.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = CN
ST = Beijing
L = Beijing
O = Blockchain Browser
OU = Production
CN = $DOMAIN

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = $DOMAIN
DNS.2 = *.$DOMAIN
EOF

# 生成证书签名请求
openssl req -new -key "$CERT_DIR/$CERT_NAME.key" -out "$CERT_DIR/$CERT_NAME.csr" -config "$CERT_DIR/$CERT_NAME.conf"

# 生成自签名证书
openssl x509 -req -in "$CERT_DIR/$CERT_NAME.csr" -signkey "$CERT_DIR/$CERT_NAME.key" -out "$CERT_DIR/$CERT_NAME.crt" -days 365 -extensions v3_req -extfile "$CERT_DIR/$CERT_NAME.conf"

# 创建PEM格式证书
cat "$CERT_DIR/$CERT_NAME.crt" "$CERT_DIR/$CERT_NAME.key" > "$CERT_DIR/$CERT_NAME.pem"

echo "✅ 域名 $DOMAIN 的SSL证书生成完成！"
echo ""
echo "📁 证书文件位置："
echo "   证书文件: $CERT_DIR/$CERT_NAME.crt"
echo "   私钥文件: $CERT_DIR/$CERT_NAME.key"
echo "   PEM文件:  $CERT_DIR/$CERT_NAME.pem"
echo "   CSR文件:  $CERT_DIR/$CERT_NAME.csr (可用于申请正式证书)"
echo ""
echo "🔧 配置说明："
echo "   在 server/config.yaml 中设置："
echo "   server:"
echo "     tls_enabled: true"
echo "     cert_file: \"$(pwd)/$CERT_DIR/$CERT_NAME.crt\""
echo "     key_file: \"$(pwd)/$CERT_DIR/$CERT_NAME.key\""
echo ""
echo "📋 生产环境建议："
echo "   1. 将 $CERT_DIR/$CERT_NAME.csr 文件提交给CA机构申请正式证书"
echo "   2. 或使用 Let's Encrypt 获取免费证书："
echo "      certbot certonly --standalone -d $DOMAIN"

# 不删除CSR文件，可能用于申请正式证书
rm -f "$CERT_DIR/$CERT_NAME.conf"

echo ""
echo "🚀 现在可以启动HTTPS服务器了！"
