#!/bin/bash

# 生成正确的TLS证书脚本
# 修复密钥用途配置问题

set -e

# 证书目录（相对于项目根目录）
CERT_DIR="./certs"
CERT_NAME="localhost"

# 创建证书目录
mkdir -p "$CERT_DIR"

echo "🔐 正在生成正确的TLS证书..."

# 生成私钥
openssl genrsa -out "$CERT_DIR/$CERT_NAME.key" 2048

# 生成证书配置文件
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
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
basicConstraints = CA:FALSE

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
echo "     cert_file: \"./certs/localhost.crt\""
echo "     key_file: \"./certs/localhost.key\""
echo ""
echo "📚 修复的问题："
echo "   - 添加了 digitalSignature 密钥用途"
echo "   - 添加了 clientAuth 扩展密钥用途"
echo "   - 设置了 basicConstraints = CA:FALSE"
echo "   - 修复了 ERR_SSL_KEY_USAGE_INCOMPATIBLE 错误"
echo ""
echo "📍 使用方法："
echo "   在项目根目录执行：./scripts/generate-tls-cert.sh"

# 清理临时文件
rm -f "$CERT_DIR/$CERT_NAME.csr" "$CERT_DIR/$CERT_NAME.conf"

echo ""
echo "🚀 现在可以启动HTTPS服务器了！"
