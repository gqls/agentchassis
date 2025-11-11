#!/usr/bin/env python3
"""
Playwright Adapter for Website Capture
Listens on Kafka topic: system.adapter.playwright.requests
Performs website captures and returns results
"""

import asyncio
import base64
import json
import logging
import os
import sys
import traceback
from datetime import datetime
from typing import Dict, Any, List, Optional
from pathlib import Path

from aiokafka import AIOKafkaConsumer, AIOKafkaProducer
from playwright.async_api import async_playwright, Browser, Page
import boto3

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration from environment
KAFKA_BROKER = os.getenv('KAFKA_BROKER', 'kafka:9092')
KAFKA_GROUP_ID = os.getenv('KAFKA_GROUP_ID', 'playwright-adapter-group')
REQUEST_TOPIC = os.getenv('REQUEST_TOPIC', 'system.adapter.playwright.requests')
S3_ENDPOINT = os.getenv('S3_ENDPOINT', 'https://s3.us-west-002.backblazeb2.com')
S3_BUCKET = os.getenv('S3_BUCKET', 'website-captures')
AWS_ACCESS_KEY_ID = os.getenv('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY')


class PlaywrightAdapter:
    """Adapter for handling Playwright-based website captures"""

    def __init__(self):
        self.consumer: Optional[AIOKafkaConsumer] = None
        self.producer: Optional[AIOKafkaProducer] = None
        self.browser: Optional[Browser] = None
        self.s3_client = None
        self.playwright = None

    async def start(self):
        """Initialize Kafka connections and Playwright"""
        logger.info("Starting Playwright adapter...")

        # Initialize Kafka consumer
        self.consumer = AIOKafkaConsumer(
            REQUEST_TOPIC,
            bootstrap_servers=KAFKA_BROKER,
            group_id=KAFKA_GROUP_ID,
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            enable_auto_commit=True,
            auto_offset_reset='latest'
        )

        # Initialize Kafka producer
        self.producer = AIOKafkaProducer(
            bootstrap_servers=KAFKA_BROKER,
            value_serializer=lambda v: json.dumps(v).encode('utf-8')
        )

        await self.consumer.start()
        await self.producer.start()

        # Initialize Playwright
        self.playwright = await async_playwright().start()
        self.browser = await self.playwright.chromium.launch(
            headless=True,
            args=['--no-sandbox', '--disable-setuid-sandbox']
        )

        # Initialize S3 client if credentials are available
        if AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY:
            self.s3_client = boto3.client(
                's3',
                endpoint_url=S3_ENDPOINT,
                aws_access_key_id=AWS_ACCESS_KEY_ID,
                aws_secret_access_key=AWS_SECRET_ACCESS_KEY
            )

        logger.info("Playwright adapter started successfully")

    async def stop(self):
        """Cleanup resources"""
        logger.info("Stopping Playwright adapter...")

        if self.browser:
            await self.browser.close()

        if self.playwright:
            await self.playwright.stop()

        if self.consumer:
            await self.consumer.stop()

        if self.producer:
            await self.producer.stop()

        logger.info("Playwright adapter stopped")

    async def process_messages(self):
        """Main message processing loop"""
        logger.info(f"Listening for messages on topic: {REQUEST_TOPIC}")

        async for message in self.consumer:
            try:
                request = message.value
                logger.info(f"Received request: {request.get('request_id')}")

                # Process based on action type
                action = request.get('action', 'capture')

                if action == 'capture':
                    result = await self.handle_capture(request)
                elif action == 'capture_interactions':
                    result = await self.handle_capture_interactions(request)
                elif action == 'capture_scroll':
                    result = await self.handle_capture_scroll(request)
                else:
                    result = {
                        'success': False,
                        'error': f'Unknown action: {action}'
                    }

                # Send response back
                await self.send_response(request, result)

            except Exception as e:
                logger.error(f"Error processing message: {str(e)}")
                logger.error(traceback.format_exc())

                # Send error response
                await self.send_response(request, {
                    'success': False,
                    'error': str(e),
                    'traceback': traceback.format_exc()
                })

    async def handle_capture(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle basic website capture"""
        url = request.get('url')
        config = request.get('capture_config', {})

        logger.info(f"Capturing website: {url}")

        # Create a new page
        page = await self.browser.new_page()

        try:
            # Set viewport if specified
            if 'viewport' in config:
                await page.set_viewport_size(
                    width=config['viewport'].get('width', 1920),
                    height=config['viewport'].get('height', 1080)
                )

            # Set user agent if specified
            if 'user_agent' in config:
                await page.set_extra_http_headers({
                    'User-Agent': config['user_agent']
                })

            # Navigate to the page
            wait_until = config.get('wait_until', 'networkidle')
            await page.goto(url, wait_until=wait_until)

            # Wait a bit for dynamic content
            await asyncio.sleep(2)

            # Take screenshot
            screenshot_options = {'full_page': config.get('full_page', True)}
            screenshot_bytes = await page.screenshot(**screenshot_options)
            screenshot_base64 = base64.b64encode(screenshot_bytes).decode('utf-8')

            # Capture DOM if requested
            html_content = None
            if config.get('capture_dom', False):
                html_content = await page.content()

            # Capture styles if requested
            css_content = None
            if config.get('capture_styles', False):
                css_content = await self.extract_styles(page)

            # Extract computed styles if requested
            computed_styles = None
            if config.get('extract_computed_styles', False):
                computed_styles = await self.extract_computed_styles(page)

            # Get page metadata
            title = await page.title()
            url_final = page.url

            # Upload to S3 if configured
            s3_paths = {}
            if self.s3_client and request.get('upload_to_s3', False):
                s3_paths = await self.upload_to_s3(
                    request_id=request['request_id'],
                    screenshot_bytes=screenshot_bytes,
                    html_content=html_content,
                    css_content=css_content
                )

            return {
                'success': True,
                'screenshot_base64': screenshot_base64,
                'html_content': html_content,
                'css_content': css_content,
                'computed_styles': computed_styles,
                'metadata': {
                    'title': title,
                    'url': url_final,
                    'captured_at': datetime.utcnow().isoformat(),
                    'viewport': config.get('viewport', {}),
                },
                's3_paths': s3_paths
            }

        finally:
            await page.close()

    async def handle_capture_interactions(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Capture hover and interaction states"""
        url = request.get('url')
        config = request.get('capture_config', {})
        selectors = config.get('selectors', ['a', 'button'])
        max_elements = config.get('max_elements', 50)

        logger.info(f"Capturing interaction states for: {url}")

        page = await self.browser.new_page()
        hover_states = []

        try:
            await page.goto(url, wait_until='networkidle')
            await asyncio.sleep(1)

            for selector in selectors:
                elements = await page.query_selector_all(selector)

                for i, element in enumerate(elements[:max_elements]):
                    if not await element.is_visible():
                        continue

                    try:
                        # Get element info
                        bbox = await element.bounding_box()
                        if not bbox:
                            continue

                        # Capture normal state
                        normal_screenshot = await element.screenshot()

                        # Capture hover state if requested
                        hover_screenshot = None
                        if config.get('capture_hover', True):
                            await element.hover()
                            await asyncio.sleep(0.3)  # Wait for hover effects
                            hover_screenshot = await element.screenshot()

                        # Get computed styles
                        styles = await page.evaluate('''
                            (element) => {
                                const computed = window.getComputedStyle(element);
                                return {
                                    color: computed.color,
                                    backgroundColor: computed.backgroundColor,
                                    borderRadius: computed.borderRadius,
                                    boxShadow: computed.boxShadow,
                                    transform: computed.transform,
                                    transition: computed.transition
                                };
                            }
                        ''', element)

                        hover_states.append({
                            'selector': selector,
                            'index': i,
                            'bbox': bbox,
                            'normal_screenshot': base64.b64encode(normal_screenshot).decode('utf-8'),
                            'hover_screenshot': base64.b64encode(hover_screenshot).decode('utf-8') if hover_screenshot else None,
                            'styles': styles
                        })

                    except Exception as e:
                        logger.warning(f"Failed to capture element {selector}[{i}]: {str(e)}")

            return {
                'success': True,
                'hover_states': hover_states,
                'total_captured': len(hover_states),
                'metadata': {
                    'url': url,
                    'captured_at': datetime.utcnow().isoformat()
                }
            }

        finally:
            await page.close()

    async def handle_capture_scroll(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Capture scroll animations and behavior"""
        url = request.get('url')
        config = request.get('capture_config', {})
        scroll_intervals = config.get('scroll_intervals', [0, 25, 50, 75, 100])

        logger.info(f"Capturing scroll behavior for: {url}")

        page = await self.browser.new_page()
        scroll_captures = []

        try:
            await page.goto(url, wait_until='networkidle')
            await asyncio.sleep(2)

            # Get page dimensions
            dimensions = await page.evaluate('''
                () => ({
                    scrollHeight: document.documentElement.scrollHeight,
                    clientHeight: document.documentElement.clientHeight
                })
            ''')

            max_scroll = dimensions['scrollHeight'] - dimensions['clientHeight']

            for percentage in scroll_intervals:
                scroll_position = int(max_scroll * (percentage / 100))

                # Scroll to position
                await page.evaluate(f'window.scrollTo(0, {scroll_position})')
                await asyncio.sleep(0.5)  # Wait for scroll effects

                # Take screenshot
                screenshot = await page.screenshot()

                # Detect sticky elements if requested
                sticky_elements = []
                if config.get('detect_sticky_elements', False):
                    sticky_elements = await page.evaluate('''
                        () => {
                            const elements = document.querySelectorAll('*');
                            const sticky = [];
                            
                            for (const el of elements) {
                                const style = window.getComputedStyle(el);
                                if (style.position === 'sticky' || style.position === 'fixed') {
                                    sticky.push({
                                        tagName: el.tagName,
                                        className: el.className,
                                        id: el.id,
                                        position: style.position
                                    });
                                }
                            }
                            
                            return sticky;
                        }
                    ''')

                scroll_captures.append({
                    'percentage': percentage,
                    'scroll_position': scroll_position,
                    'screenshot': base64.b64encode(screenshot).decode('utf-8'),
                    'sticky_elements': sticky_elements
                })

            # Detect parallax effects if requested
            parallax_detected = False
            if config.get('detect_parallax', False):
                parallax_detected = await self.detect_parallax(page)

            return {
                'success': True,
                'scroll_captures': scroll_captures,
                'parallax_detected': parallax_detected,
                'page_dimensions': dimensions,
                'metadata': {
                    'url': url,
                    'captured_at': datetime.utcnow().isoformat()
                }
            }

        finally:
            await page.close()

    async def extract_styles(self, page: Page) -> str:
        """Extract all CSS from the page"""
        styles = await page.evaluate('''
            () => {
                const styles = [];
                
                // Get all stylesheets
                for (const sheet of document.styleSheets) {
                    try {
                        const rules = sheet.cssRules || sheet.rules;
                        for (const rule of rules) {
                            styles.push(rule.cssText);
                        }
                    } catch (e) {
                        // Cross-origin stylesheets will throw
                        console.log('Could not access stylesheet:', sheet.href);
                    }
                }
                
                // Get inline styles
                const elements = document.querySelectorAll('[style]');
                for (const el of elements) {
                    if (el.style.cssText) {
                        styles.push(`/* Inline style */ ${el.tagName} { ${el.style.cssText} }`);
                    }
                }
                
                return styles.join('\\n');
            }
        ''')
        return styles

    async def extract_computed_styles(self, page: Page) -> Dict[str, Any]:
        """Extract computed styles for key elements"""
        return await page.evaluate('''
            () => {
                const getElementStyles = (selector) => {
                    const element = document.querySelector(selector);
                    if (!element) return null;
                    
                    const computed = window.getComputedStyle(element);
                    return {
                        // Colors
                        color: computed.color,
                        backgroundColor: computed.backgroundColor,
                        
                        // Typography
                        fontFamily: computed.fontFamily,
                        fontSize: computed.fontSize,
                        fontWeight: computed.fontWeight,
                        lineHeight: computed.lineHeight,
                        
                        // Spacing
                        padding: computed.padding,
                        margin: computed.margin,
                        
                        // Borders
                        border: computed.border,
                        borderRadius: computed.borderRadius,
                        
                        // Effects
                        boxShadow: computed.boxShadow,
                        transform: computed.transform,
                        transition: computed.transition
                    };
                };
                
                return {
                    body: getElementStyles('body'),
                    header: getElementStyles('header'),
                    nav: getElementStyles('nav'),
                    main: getElementStyles('main'),
                    footer: getElementStyles('footer'),
                    h1: getElementStyles('h1'),
                    button: getElementStyles('button'),
                    link: getElementStyles('a')
                };
            }
        ''')

    async def detect_parallax(self, page: Page) -> bool:
        """Detect if page uses parallax scrolling effects"""
        # Simplified parallax detection
        # In production, this would be more sophisticated
        result = await page.evaluate('''
            () => {
                const elements = document.querySelectorAll('*');
                
                for (const el of elements) {
                    const style = window.getComputedStyle(el);
                    const transform = style.transform;
                    const bgAttachment = style.backgroundAttachment;
                    
                    // Check for common parallax indicators
                    if (bgAttachment === 'fixed' || 
                        transform.includes('translate3d') ||
                        el.classList.toString().toLowerCase().includes('parallax')) {
                        return true;
                    }
                }
                
                return false;
            }
        ''')
        return result

    async def upload_to_s3(self, request_id: str, **files) -> Dict[str, str]:
        """Upload files to S3 and return paths"""
        if not self.s3_client:
            return {}

        paths = {}
        timestamp = datetime.utcnow().strftime('%Y%m%d_%H%M%S')

        for file_type, content in files.items():
            if content is None:
                continue

            # Determine file extension
            ext = 'txt'
            if file_type == 'screenshot_bytes':
                ext = 'png'
            elif file_type == 'html_content':
                ext = 'html'
            elif file_type == 'css_content':
                ext = 'css'

            # Generate S3 key
            key = f"captures/{request_id}/{timestamp}_{file_type}.{ext}"

            # Upload to S3
            try:
                if isinstance(content, bytes):
                    self.s3_client.put_object(
                        Bucket=S3_BUCKET,
                        Key=key,
                        Body=content
                    )
                else:
                    self.s3_client.put_object(
                        Bucket=S3_BUCKET,
                        Key=key,
                        Body=content.encode('utf-8')
                    )

                paths[file_type] = f"s3://{S3_BUCKET}/{key}"
                logger.info(f"Uploaded {file_type} to {paths[file_type]}")

            except Exception as e:
                logger.error(f"Failed to upload {file_type}: {str(e)}")

        return paths

    async def send_response(self, request: Dict[str, Any], result: Dict[str, Any]):
        """Send response back to the orchestrator"""
        reply_topic = request.get('reply_to_topic')
        request_id = request.get('request_id')

        if not reply_topic:
            logger.warning(f"No reply_to_topic for request {request_id}")
            return

        response = {
            'request_id': request_id,
            'correlation_id': request.get('correlation_id'),
            'orchestration_id': request.get('orchestration_id'),
            'timestamp': datetime.utcnow().isoformat(),
            'result': result
        }

        try:
            await self.producer.send(reply_topic, value=response, key=request_id.encode('utf-8'))
            logger.info(f"Sent response for {request_id} to {reply_topic}")
        except Exception as e:
            logger.error(f"Failed to send response: {str(e)}")


async def main():
    """Main entry point"""
    adapter = PlaywrightAdapter()

    try:
        await adapter.start()
        await adapter.process_messages()
    except KeyboardInterrupt:
        logger.info("Received shutdown signal")
    except Exception as e:
        logger.error(f"Adapter error: {str(e)}")
        logger.error(traceback.format_exc())
    finally:
        await adapter.stop()


if __name__ == '__main__':
    asyncio.run(main())