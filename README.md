# tech-challenge-notification-service

Asynchronous service for handling system notifications.

## Seeding email templates

There's no `PUT /templates/{type}` route wired up in this deploy (see the
comment on `aws_s3_bucket.templates` in `terraform/main.tf`), so the
templates in `templates/*.html` have to be uploaded to S3 directly —
including the `subject` object metadata, which the send flow requires
(Mailtrap rejects a send with no `Subject`):

```bash
BUCKET=<the S3_BUCKET_NAME org secret's value> ./scripts/seed-templates.sh
```

Needed again any time the bucket is recreated (e.g. an AWS Academy Lab
account reset) or a template's HTML/subject changes.
