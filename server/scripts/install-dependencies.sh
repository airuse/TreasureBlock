#!/bin/bash

# TLS证书生成依赖安装脚本
# 支持主流Linux发行版

set -e

echo "🔧 检查并安装TLS证书生成所需依赖..."

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VERSION=$VERSION_ID
else
    echo "❌ 无法检测操作系统版本"
    exit 1
fi

echo "📊 检测到操作系统: $OS $VERSION"

# 检查OpenSSL是否已安装
if command -v openssl &> /dev/null; then
    echo "✅ OpenSSL 已安装: $(openssl version)"
else
    echo "📦 正在安装 OpenSSL..."
    
    case "$OS" in
        "Ubuntu"*|"Debian"*)
            sudo apt-get update
            sudo apt-get install -y openssl
            ;;
        "CentOS Linux"*|"Red Hat"*|"Rocky Linux"*|"AlmaLinux"*)
            sudo yum install -y openssl
            ;;
        "Fedora"*)
            sudo dnf install -y openssl
            ;;
        "SUSE"*|"openSUSE"*)
            sudo zypper install -y openssl
            ;;
        "Alpine Linux"*)
            sudo apk add openssl
            ;;
        "Arch Linux"*)
            sudo pacman -S openssl
            ;;
        *)
            echo "❌ 不支持的操作系统: $OS"
            echo "请手动安装 OpenSSL"
            exit 1
            ;;
    esac
    
    echo "✅ OpenSSL 安装完成"
fi

# 验证安装
if command -v openssl &> /dev/null; then
    echo "🎉 所有依赖已准备就绪！"
    echo "OpenSSL 版本: $(openssl version)"
    echo ""
    echo "📚 技术说明："
    echo "   - OpenSSL 是生成TLS证书的工具"
    echo "   - 虽然叫'OpenSSL'，但生成的是TLS证书"
    echo "   - TLS是SSL的现代继任者，更安全"
    echo ""
    echo "现在可以运行证书生成脚本："
    echo "  ./scripts/generate-tls-cert.sh"
else
    echo "❌ OpenSSL 安装失败，请手动安装"
    exit 1
fi
