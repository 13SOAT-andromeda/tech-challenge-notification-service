#!/usr/bin/env bash
set -euo pipefail

# Seeds all templates/*.html into the S3 bucket the Lambda reads from at
# runtime, including the Subject S3 object metadata the send flow requires
# (Mailtrap rejects a send with no subject — this is what broke the first
# real signup-email test: the templates had been uploaded with plain
# `aws s3 cp`, no --metadata). Subjects are intentionally ASCII-only, since
# S3 object metadata values can't contain non-ASCII bytes.
#
# Usage: BUCKET=<bucket-name> ./scripts/seed-templates.sh
#   BUCKET must match TF_VAR_s3_bucket_name in .github/workflows/deploy.yml
#   (i.e. the S3_BUCKET_NAME org secret's value).

: "${BUCKET:?Set BUCKET to the target S3 bucket name}"

cd "$(dirname "$0")/.."

declare -A SUBJECTS=(
  [BACKORDER_CREATED]="Backorder Criado"
  [ORDER_APPROVAL_REQUEST]="Aprovacao de Ordem de Servico"
  [ORDER_AWAITING_PAYMENT]="Pagamento Aguardado"
  [PROCESSING_FAILED]="Falha no processamento do video"
  [STOCK_RESERVATION_FAILED]="Falha na Reserva de Estoque"
  [STOCK_RESERVED]="Estoque Reservado"
  [email-verification]="Confirme seu e-mail"
)

for name in "${!SUBJECTS[@]}"; do
  file="templates/${name}.html"
  if [ ! -f "$file" ]; then
    echo "warning: skipping $file — not found" >&2
    continue
  fi
  echo "Uploading $file (subject: ${SUBJECTS[$name]})"
  aws s3 cp "$file" "s3://${BUCKET}/templates/${name}.html" \
    --content-type "text/html" \
    --metadata "subject=${SUBJECTS[$name]}"
done
