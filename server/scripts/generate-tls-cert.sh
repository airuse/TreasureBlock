#!/bin/bash

# 生成TLS证书脚本
# 用于开发环境的自签名证书

set -e

CERT_DIR="./certs"
CERT_NAME="localhost"

# 创建证书目录
mkdir -p "$CERT_DIR"

echo "🔐 正在生成TLS证书..."

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
OU = Development
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# 生成证书签名请求
openssl req -new -key "$CERT_DIR/$CERT_NAME.key" -out "$CERT_DIR/$CERT_NAME.csr" -config "$CERT_DIR/$CERT_NAME.conf"

# 生成自签名证书
openssl x509 -req -in "$CERT_DIR/$CERT_NAME.csr" -signkey "$CERT_DIR/$CERT_NAME.key" -out "$CERT_DIR/$CERT_NAME.crt" -days 365 -extensions v3_req -extfile "$CERT_DIR/$CERT_NAME.conf"

# 创建PEM格式证书
cat "$CERT_DIR/$CERT_NAME.crt" "$CERT_DIR/$CERT_NAME.key" > "$CERT_DIR/$CERT_NAME.pem"

echo "✅ TLS证书生成完成！"
echo ""
echo "📁 证书文件位置："
echo "   证书文件: $CERT_DIR/$CERT_NAME.crt"
echo "   私钥文件: $CERT_DIR/$CERT_NAME.key"
echo "   PEM文件:  $CERT_DIR/$CERT_NAME.pem"
echo ""
echo "🔧 配置说明："
echo "   在 server/config.yaml 中设置："
echo "   server:"
echo "     tls_enabled: true"
echo "     cert_file: \"$(pwd)/$CERT_DIR/$CERT_NAME.crt\""
echo "     key_file: \"$(pwd)/$CERT_DIR/$CERT_NAME.key\""
echo ""
echo "⚠️  注意：这是开发用的自签名证书，浏览器会显示安全警告。"
echo "   生产环境请使用由可信CA签发的证书。"
echo ""
echo "📚 技术说明："
echo "   - 生成的是TLS证书，不是SSL证书"
echo "   - TLS是SSL的现代继任者，更安全"
echo "   - 支持TLS 1.2和TLS 1.3协议"

# 清理临时文件
rm -f "$CERT_DIR/$CERT_NAME.csr" "$CERT_DIR/$CERT_NAME.conf"

echo ""
echo "🚀 现在可以启动HTTPS服务器了！"
