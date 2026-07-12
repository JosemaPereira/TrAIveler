# CloudFront & S3 Frontend Delivery Module
#
# Creates the static-hosting origin for the React frontend build (a private
# S3 bucket, never publicly readable) and a CloudFront distribution that
# fronts it over HTTPS. Access from CloudFront to the bucket is granted
# exclusively through an Origin Access Identity (OAI) — there is no public
# bucket policy and no public ACL anywhere in this module.
#
# SPA routing: React Router resolves routes client-side, so a direct request
# for a deep link (e.g. /trips/42) has no matching object in the S3 bucket
# and S3 returns 404. custom_error_response rewrites that 404 to
# /index.html and returns it as HTTP 200, letting the SPA's router take over
# instead of showing a broken CloudFront/S3 error page.

# ---------------------------------------------------------------------------
# S3 origin bucket
#
# Private by default: no bucket ACL is set (new buckets default to the
# BucketOwnerEnforced ownership setting, which disables ACLs entirely), and
# aws_s3_bucket_public_access_block explicitly blocks any public ACL or
# public bucket policy from ever being applied. The only read access is the
# OAI-scoped bucket policy below.
# ---------------------------------------------------------------------------

resource "aws_s3_bucket" "frontend" {
  bucket = "traveler-${var.environment}-frontend"

  tags = {
    Name        = "${var.environment}-frontend"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ---------------------------------------------------------------------------
# CloudFront Origin Access Identity + bucket policy
#
# The OAI is the only principal allowed to read from the bucket. The bucket
# policy below grants it s3:GetObject on bucket objects only — no other
# principal (and no anonymous/public access) is ever granted access.
# ---------------------------------------------------------------------------

resource "aws_cloudfront_origin_access_identity" "frontend" {
  comment = "OAI for ${var.environment} frontend S3 bucket"
}

data "aws_iam_policy_document" "frontend_oai_read" {
  statement {
    sid       = "AllowCloudFrontOAIRead"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.frontend.arn}/*"]

    principals {
      type        = "AWS"
      identifiers = [aws_cloudfront_origin_access_identity.frontend.iam_arn]
    }
  }
}

resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  policy = data.aws_iam_policy_document.frontend_oai_read.json
}

# ---------------------------------------------------------------------------
# CloudFront distribution
#
# viewer_protocol_policy = redirect-to-https ensures the CDN never serves
# plaintext HTTP. custom_error_response is the SPA-routing rewrite described
# above — required for React Router's client-side routes to work on a
# direct load or page refresh.
# ---------------------------------------------------------------------------

resource "aws_cloudfront_distribution" "frontend" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  aliases             = [var.domain_name]
  price_class         = var.price_class

  origin {
    domain_name = aws_s3_bucket.frontend.bucket_regional_domain_name
    origin_id   = "s3-frontend"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.frontend.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "s3-frontend"
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  # SPA routing: rewrite S3's 404 (no object for a client-side route) to
  # index.html, returned as HTTP 200, so React Router owns the navigation.
  custom_error_response {
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 0
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = var.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  tags = {
    Name        = "${var.environment}-frontend-cdn"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}
