import os
import json
import logging
from kafka import KafkaConsumer, KafkaProducer
from playwright.sync_api import sync_playwright
import boto3
from botocore.client import Config
import io

# --- Setup ---
logging.basicConfig(level=logging.INFO)
log = logging.getLogger("playwright-adapter")

KAFKA_BROKERS = os.environ.get("KAFKA_BROKERS", "kafka:9092")
REQUEST_TOPIC = "system.adapter.playwright.requests"
KAFKA_GROUP_ID = "playwright-adapter-group"

# Backblaze S3-Compatible Config
S3_ENDPOINT_URL = os.environ.get("S3_ENDPOINT_URL") # e.g., 'https://s3.us-west-000.backblazeb2.com'
S3_ACCESS_KEY = os.environ.get("S3_ACCESS_KEY")
S3_SECRET_KEY = os.environ.get("S3_SECRET_KEY")

log.info(f"Connecting to Kafka at {KAFKA_BROKERS} on topic {REQUEST_TOPIC}")

# --- Kafka and S3 Clients ---
try:
    consumer = KafkaConsumer(
        REQUEST_TOPIC,
        bootstrap_servers=KAFKA_BROKERS,
        auto_offset_reset='earliest',
        group_id=KAFKA_GROUP_ID,
        value_deserializer=lambda v: json.loads(v.decode('utf-8', 'ignore'))
    )

    producer = KafkaProducer(
        bootstrap_servers=KAFKA_BROKERS,
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )

    s3_client = boto3.client(
        's3',
        endpoint_url=S3_ENDPOINT_URL,
        aws_access_key_id=S3_ACCESS_KEY,
        aws_secret_access_key=S3_SECRET_KEY,
        config=Config(signature_version='s3v4')
    )
except Exception as e:
    log.fatal(f"Failed to initialize clients: {e}")
    exit(1)


# --- Core Functions ---
def capture_and_upload(url, s3_bucket, s3_path_prefix, logger):
    """
    Runs Playwright to capture screenshot and DOM, then uploads to S3.
    """
    logger.info(f"Starting capture for: {url}")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto(url, wait_until='networkidle', timeout=30000)

        # 1. Capture Screenshot
        ss_bytes = page.screenshot(full_page=True, type='png')
        ss_path = f"{s3_path_prefix}/desktop.png"

        # 2. Capture DOM
        dom_content = page.content()
        dom_path = f"{s3_path_prefix}/page.html"

        browser.close()

    # 3. Upload to S3 (Backblaze)
    try:
        # Upload Screenshot
        s3_client.put_object(
            Bucket=s3_bucket,
            Key=ss_path,
            Body=io.BytesIO(ss_bytes),
            ContentType='image/png'
        )
        logger.info(f"Uploaded screenshot to: {ss_path}")

        # Upload DOM
        s3_client.put_object(
            Bucket=s3_bucket,
            Key=dom_path,
            Body=dom_content.encode('utf-8'),
            ContentType='text/html; charset=utf-8'
        )
        logger.info(f"Uploaded DOM to: {dom_path}")

    except Exception as e:
        logger.error(f"Failed to upload to S3: {e}")
        raise

    return {"screenshot_s3_path": ss_path, "dom_s3_path": dom_path}

def build_response_message(original_headers, success, data, error_message):
    """
    Builds the Kafka response message in the expected format.
    """
    response_headers = {
        "correlation_id": original_headers.get("correlation_id"),
        "in_response_to_request_id": original_headers.get("request_id"),
        "message_type": "response",
        "sender_agent_type": "playwright-adapter"
    }

    response_body = {
        "success": success,
        "data": data if success else None,
        "error": error_message if not success else None
    }

    return {"headers": response_headers, "body": response_body}


# --- Main Consumer Loop ---
def main():
    log.info("Playwright Adapter service started. Waiting for messages...")
    for message in consumer:
        msg_data = message.value
        headers = {k: v.decode('utf-8') for (k, v) in message.headers}

        reply_to = msg_data.get("body", {}).get("reply_to_topic")
        if not reply_to:
            log.error(f"Missing 'reply_to_topic' in message: {msg_data}")
            continue

        log.info(f"Received job {headers.get('request_id')} for {reply_to}")

        try:
            action = msg_data.get("body", {}).get("action")
            if action == "capture_site":
                params = msg_data.get("body", {}).get("data", {})

                # Do the work
                result_data = capture_and_upload(
                    url=params.get("url"),
                    s3_bucket=params.get("s3_bucket"),
                    s3_path_prefix=params.get("s3_path_prefix"),
                    logger=log
                )

                # Send success response
                response_msg = build_response_message(headers, True, result_data, None)
            else:
                raise ValueError(f"Unknown action: {action}")

        except Exception as e:
            log.error(f"Job {headers.get('request_id')} failed: {e}")
            # Send error response
            response_msg = build_response_message(headers, False, None, str(e))

        try:
            producer.send(reply_to, value=response_msg)
            log.info(f"Sent response for {headers.get('request_id')} to {reply_to}")
        except Exception as e:
            log.error(f"Failed to send response to {reply_to}: {e}")

if __name__ == "__main__":
    main()