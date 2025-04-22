#!/usr/bin/env python3
import sys
import os
import datetime
import hashlib
import hmac
import http.client
import time

class ProgressFileReader:
    def __init__(self, file_path, chunk_size=8192):
        self.file = open(file_path, "rb")
        self.file_path = file_path
        self.total_size = os.path.getsize(file_path)
        self.read_bytes = 0
        self.chunk_size = chunk_size
        self.last_print = 0

    def read(self, size=-1):
        data = self.file.read(size if size > 0 else self.chunk_size)
        if data:
            self.read_bytes += len(data)
            self._print_progress()
        return data

    def _print_progress(self):
        percent = self.read_bytes / self.total_size * 100
        read_mb = self.read_bytes / (1024 * 1024)
        total_mb = self.total_size / (1024 * 1024)
        now = time.time()
        if now - self.last_print > 0.1 or self.read_bytes == self.total_size:
            print(f"\rUploading {self.file_path}... {percent:.1f}% " +
                f"({read_mb:.2f}/{total_mb:.2f} MB)", end="", flush=True)
            self.last_print = now

    def __len__(self):
        return self.total_size

    def close(self):
        self.file.close()

def sign(key, msg):
    return hmac.new(key, msg.encode("utf-8"), hashlib.sha256).digest()

def get_signature_key(key, date_stamp, region, service):
    k_date = sign(("AWS4" + key).encode("utf-8"), date_stamp)
    k_region = sign(k_date, region)
    k_service = sign(k_region, service)
    k_signing = sign(k_service, "aws4_request")
    return k_signing

def sha256_hexdigest_file(file_path):
    h = hashlib.sha256()
    with open(file_path, "rb") as f:
        while chunk := f.read(8192):
            h.update(chunk)
    return h.hexdigest()

def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <source_file> <bucket/path>",
            file=sys.stderr)
        sys.exit(1)

    source_file_path = sys.argv[1]
    dest_path = sys.argv[2]

    aws_access_key = os.environ.get("AWS_ACCESS_KEY_ID")
    aws_secret_key = os.environ.get("AWS_SECRET_ACCESS_KEY")
    cloudflare_account_id = os.environ.get("CLOUDFLARE_ACCOUNT_ID")
    region = os.environ.get("AWS_DEFAULT_REGION", "us-east-1")

    if not aws_access_key or not aws_secret_key or not cloudflare_account_id:
        print("Missing required environment variables", file=sys.stderr)
        sys.exit(1)

    if "/" not in dest_path:
        print("Destination path must include bucket and " +
            "object key (e.g. bucket/file.txt)", file=sys.stderr)
        sys.exit(1)

    bucket, s3_key = dest_path.split("/", 1)
    host = f"{cloudflare_account_id}.r2.cloudflarestorage.com"
    uri = f"/{bucket}/{s3_key}"
    method = "PUT"

    ext = os.path.splitext(source_file_path)[1].lower()
    if ext == ".html":
        content_type = "text/html"
    elif ext == ".json":
        content_type = "application/json"
    elif ext == ".sig":
        content_type = "application/pgp-signature"
    else:
        content_type = "application/octet-stream"

    content_length = str(os.path.getsize(source_file_path))
    payload_hash = sha256_hexdigest_file(source_file_path)

    now = datetime.datetime.utcnow()
    amz_date = now.strftime("%Y%m%dT%H%M%SZ")
    date_stamp = now.strftime("%Y%m%d")

    canonical_headers = (
        f"content-length:{content_length}\n"
        f"content-type:{content_type}\n"
        f"host:{host}\n"
        f"x-amz-content-sha256:{payload_hash}\n"
        f"x-amz-date:{amz_date}\n"
    )
    signed_headers = "content-length;content-type;host;" + \
        "x-amz-content-sha256;x-amz-date"
    canonical_request = (
        f"{method}\n"
        f"{uri}\n"
        f"\n"
        f"{canonical_headers}\n"
        f"{signed_headers}\n"
        f"{payload_hash}"
    )

    credential_scope = f"{date_stamp}/{region}/s3/aws4_request"
    string_to_sign = (
        f"AWS4-HMAC-SHA256\n"
        f"{amz_date}\n"
        f"{credential_scope}\n"
        f"{hashlib.sha256(canonical_request.encode()).hexdigest()}"
    )

    signing_key = get_signature_key(aws_secret_key, date_stamp, region, "s3")
    signature = hmac.new(signing_key, string_to_sign.encode("utf-8"),
        hashlib.sha256).hexdigest()

    authorization_header = (
        f"AWS4-HMAC-SHA256 Credential={aws_access_key}/{credential_scope}, "
        f"SignedHeaders={signed_headers}, Signature={signature}"
    )

    headers = {
        "Host": host,
        "Content-Type": content_type,
        "Content-Length": content_length,
        "X-Amz-Date": amz_date,
        "X-Amz-Content-SHA256": payload_hash,
        "Authorization": authorization_header
    }

    conn = http.client.HTTPSConnection(host)
    reader = ProgressFileReader(source_file_path)

    try:
        conn.request(method, uri, body=reader, headers=headers)
        response = conn.getresponse()
        print(f"\nStatus: {response.status} {response.reason}")
        body = response.read().decode()
        if response.status >= 300:
            print(body, file=sys.stderr)
            sys.exit(1)
        else:
            print(f"Upload successful {dest_path}")
    finally:
        reader.close()

if __name__ == "__main__":
    main()
