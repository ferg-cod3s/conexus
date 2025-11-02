#!/bin/bash

# Generate self-signed certificates for development
# Usage: ./scripts/generate-dev-certs.sh [domain] [output_dir]

set -e

DOMAIN=${1:-"localhost"}
OUTPUT_DIR=${2:-"./data/tls"}
CERT_FILE="${OUTPUT_DIR}/cert.pem"
KEY_FILE="${OUTPUT_DIR}/key.pem"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Generating self-signed certificate for development..."
echo "Domain: $DOMAIN"
echo "Certificate: $CERT_FILE"
echo "Private Key: $KEY_FILE"

# Generate private key
openssl genrsa -out "$KEY_FILE" 2048

# Generate certificate signing request
openssl req -new -key "$KEY_FILE" -out "${OUTPUT_DIR}/cert.csr" -subj "/C=US/ST=Dev/L=Dev/O=Conexus/CN=$DOMAIN"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 -in "${OUTPUT_DIR}/cert.csr" -signkey "$KEY_FILE" -out "$CERT_FILE" \
  -extfile <(printf "subjectAltName=DNS:$DOMAIN,DNS:localhost,IP:127.0.0.1,IP:::1")

# Clean up CSR
rm "${OUTPUT_DIR}/cert.csr"

echo "Certificate generated successfully!"
echo ""
echo "To use in development, set the following environment variables:"
echo "export CONEXUS_TLS_ENABLED=true"
echo "export CONEXUS_TLS_CERT_FILE=$CERT_FILE"
echo "export CONEXUS_TLS_KEY_FILE=$KEY_FILE"
echo ""
echo "Or add to config.yml:"
echo "tls:"
echo "  enabled: true"
echo "  cert_file: $CERT_FILE"
echo "  key_file: $KEY_FILE"
echo ""
echo "Note: This certificate is self-signed and should only be used for development!"