#!/usr/bin/env python3
"""
Test script for Playwright Adapter
Tests the adapter by sending a capture request directly to Kafka
"""

import asyncio
import json
import uuid
from datetime import datetime
from aiokafka import AIOKafkaProducer, AIOKafkaConsumer
import argparse


async def test_capture(url: str, kafka_broker: str = 'localhost:9092'):
    """Test website capture through Kafka messaging"""

    request_id = str(uuid.uuid4())
    response_topic = f"test.responses.{request_id[:8]}"

    print(f"Testing capture for: {url}")
    print(f"Request ID: {request_id}")
    print(f"Response topic: {response_topic}")

    # Create producer
    producer = AIOKafkaProducer(
        bootstrap_servers=kafka_broker,
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )

    # Create consumer for response
    consumer = AIOKafkaConsumer(
        response_topic,
        bootstrap_servers=kafka_broker,
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        auto_offset_reset='earliest'
    )

    try:
        # Start connections
        await producer.start()
        await consumer.start()

        # Build test request
        request = {
            'request_id': request_id,
            'action': 'capture',
            'url': url,
            'capture_config': {
                'viewport': {'width': 1920, 'height': 1080},
                'full_page': True,
                'wait_until': 'networkidle',
                'capture_dom': True,
                'capture_styles': True,
                'extract_computed_styles': True
            },
            'correlation_id': f'test-{request_id[:8]}',
            'orchestration_id': f'test-orch-{request_id[:8]}',
            'reply_to_topic': response_topic,
            'timestamp': datetime.utcnow().isoformat()
        }

        print("\nSending capture request...")
        print(json.dumps(request, indent=2))

        # Send request
        await producer.send(
            'system.adapter.playwright.requests',
            value=request,
            key=request_id.encode('utf-8')
        )

        print("\nWaiting for response (timeout: 30s)...")

        # Wait for response
        try:
            async for msg in consumer:
                response = msg.value

                if response.get('request_id') == request_id:
                    print("\nReceived response!")

                    if response.get('result', {}).get('success'):
                        print("✓ Capture successful!")

                        result = response['result']
                        metadata = result.get('metadata', {})

                        print(f"\nCapture Details:")
                        print(f"  Title: {metadata.get('title', 'N/A')}")
                        print(f"  URL: {metadata.get('url', 'N/A')}")
                        print(f"  Captured at: {metadata.get('captured_at', 'N/A')}")

                        if result.get('screenshot_base64'):
                            print(f"  Screenshot size: {len(result['screenshot_base64'])} bytes (base64)")

                        if result.get('html_content'):
                            print(f"  HTML size: {len(result['html_content'])} characters")

                        if result.get('css_content'):
                            print(f"  CSS size: {len(result['css_content'])} characters")

                        if result.get('computed_styles'):
                            print(f"  Computed styles extracted: Yes")

                        if result.get('s3_paths'):
                            print(f"\n  S3 Uploads:")
                            for key, path in result['s3_paths'].items():
                                print(f"    {key}: {path}")
                    else:
                        print("✗ Capture failed!")
                        print(f"  Error: {response.get('result', {}).get('error', 'Unknown error')}")

                    break

        except asyncio.TimeoutError:
            print("✗ Timeout waiting for response")

    finally:
        await producer.stop()
        await consumer.stop()


async def test_interactions(url: str, kafka_broker: str = 'localhost:9092'):
    """Test interaction capture"""

    request_id = str(uuid.uuid4())
    response_topic = f"test.responses.{request_id[:8]}"

    print(f"Testing interaction capture for: {url}")

    producer = AIOKafkaProducer(
        bootstrap_servers=kafka_broker,
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )

    consumer = AIOKafkaConsumer(
        response_topic,
        bootstrap_servers=kafka_broker,
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        auto_offset_reset='earliest'
    )

    try:
        await producer.start()
        await consumer.start()

        request = {
            'request_id': request_id,
            'action': 'capture_interactions',
            'url': url,
            'capture_config': {
                'selectors': ['a', 'button', '[role="button"]'],
                'max_elements': 10,
                'capture_hover': True,
                'capture_focus': True
            },
            'reply_to_topic': response_topic,
            'timestamp': datetime.utcnow().isoformat()
        }

        print("Sending interaction capture request...")

        await producer.send(
            'system.adapter.playwright.requests',
            value=request,
            key=request_id.encode('utf-8')
        )

        print("Waiting for response...")

        try:
            async for msg in consumer:
                response = msg.value

                if response.get('request_id') == request_id:
                    if response.get('result', {}).get('success'):
                        print("✓ Interaction capture successful!")

                        result = response['result']
                        hover_states = result.get('hover_states', [])

                        print(f"\nCaptured {len(hover_states)} interactive elements")

                        for state in hover_states[:3]:  # Show first 3
                            print(f"\n  Element: {state['selector']}[{state['index']}]")
                            print(f"    Position: ({state['bbox']['x']}, {state['bbox']['y']})")
                            print(f"    Size: {state['bbox']['width']}x{state['bbox']['height']}")
                            if state.get('styles'):
                                print(f"    Styles captured: Yes")
                    else:
                        print("✗ Interaction capture failed!")
                        print(f"  Error: {response.get('result', {}).get('error', 'Unknown')}")

                    break

        except asyncio.TimeoutError:
            print("✗ Timeout")

    finally:
        await producer.stop()
        await consumer.stop()


async def test_scroll(url: str, kafka_broker: str = 'localhost:9092'):
    """Test scroll behavior capture"""

    request_id = str(uuid.uuid4())
    response_topic = f"test.responses.{request_id[:8]}"

    print(f"Testing scroll capture for: {url}")

    producer = AIOKafkaProducer(
        bootstrap_servers=kafka_broker,
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )

    consumer = AIOKafkaConsumer(
        response_topic,
        bootstrap_servers=kafka_broker,
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        auto_offset_reset='earliest'
    )

    try:
        await producer.start()
        await consumer.start()

        request = {
            'request_id': request_id,
            'action': 'capture_scroll',
            'url': url,
            'capture_config': {
                'scroll_intervals': [0, 50, 100],
                'capture_at_each': True,
                'detect_parallax': True,
                'detect_sticky_elements': True
            },
            'reply_to_topic': response_topic,
            'timestamp': datetime.utcnow().isoformat()
        }

        print("Sending scroll capture request...")

        await producer.send(
            'system.adapter.playwright.requests',
            value=request,
            key=request_id.encode('utf-8')
        )

        print("Waiting for response...")

        try:
            async for msg in consumer:
                response = msg.value

                if response.get('request_id') == request_id:
                    if response.get('result', {}).get('success'):
                        print("✓ Scroll capture successful!")

                        result = response['result']

                        print(f"\nScroll Analysis:")
                        print(f"  Captures: {len(result.get('scroll_captures', []))}")
                        print(f"  Parallax detected: {result.get('parallax_detected', False)}")

                        dimensions = result.get('page_dimensions', {})
                        print(f"  Page height: {dimensions.get('scrollHeight', 0)}px")

                        for capture in result.get('scroll_captures', []):
                            print(f"    {capture['percentage']}%: position {capture['scroll_position']}px")
                            if capture.get('sticky_elements'):
                                print(f"      Sticky elements: {len(capture['sticky_elements'])}")
                    else:
                        print("✗ Scroll capture failed!")
                        print(f"  Error: {response.get('result', {}).get('error', 'Unknown')}")

                    break

        except asyncio.TimeoutError:
            print("✗ Timeout")

    finally:
        await producer.stop()
        await consumer.stop()


async def main():
    parser = argparse.ArgumentParser(description='Test Playwright Adapter')
    parser.add_argument('url', help='URL to test capture')
    parser.add_argument('--broker', default='localhost:9092', help='Kafka broker address')
    parser.add_argument('--test', choices=['capture', 'interactions', 'scroll', 'all'],
                        default='capture', help='Test type')

    args = parser.parse_args()

    # Ensure URL has protocol
    if not args.url.startswith(('http://', 'https://')):
        args.url = 'https://' + args.url

    print(f"Playwright Adapter Test")
    print(f"=" * 50)
    print(f"URL: {args.url}")
    print(f"Broker: {args.broker}")
    print(f"Test: {args.test}")
    print(f"=" * 50)

    if args.test == 'capture' or args.test == 'all':
        await test_capture(args.url, args.broker)

    if args.test == 'interactions' or args.test == 'all':
        print("\n" + "=" * 50)
        await test_interactions(args.url, args.broker)

    if args.test == 'scroll' or args.test == 'all':
        print("\n" + "=" * 50)
        await test_scroll(args.url, args.broker)

    print("\n" + "=" * 50)
    print("Test complete!")


if __name__ == '__main__':
    asyncio.run(main())