Absolutely! Using Firecrawl as an MVP makes a lot of sense - it's a managed service that handles the complexity of web scraping. Let me adapt the solution to use your half-built Firecrawl adapter instead of Playwright. This will significantly reduce the development load.

## Firecrawl Adapter MVP Implementation

### 1. Updated Firecrawl Adapter (Python)

```python
#!/usr/bin/env python3
"""
Firecrawl Adapter for Website Capture
Listens on Kafka topic: system.adapter.firecrawl.requests
Uses Firecrawl API for website scraping and analysis
"""

import asyncio
import json
import logging
import os
import traceback
from datetime import datetime
from typing import Dict, Any, Optional
import httpx
from aiokafka import AIOKafkaConsumer, AIOKafkaProducer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
KAFKA_BROKER = os.getenv('KAFKA_BROKER', 'kafka:9092')
KAFKA_GROUP_ID = os.getenv('KAFKA_GROUP_ID', 'firecrawl-adapter-group')
REQUEST_TOPIC = os.getenv('REQUEST_TOPIC', 'system.adapter.firecrawl.requests')
FIRECRAWL_API_KEY = os.getenv('FIRECRAWL_API_KEY')
FIRECRAWL_API_URL = os.getenv('FIRECRAWL_API_URL', 'https://api.firecrawl.dev/v1')


class FirecrawlAdapter:
    """Adapter for Firecrawl API integration"""
    
    def __init__(self):
        self.consumer: Optional[AIOKafkaConsumer] = None
        self.producer: Optional[AIOKafkaProducer] = None
        self.http_client: Optional[httpx.AsyncClient] = None
        
    async def start(self):
        """Initialize connections"""
        logger.info("Starting Firecrawl adapter...")
        
        # Initialize Kafka
        self.consumer = AIOKafkaConsumer(
            REQUEST_TOPIC,
            bootstrap_servers=KAFKA_BROKER,
            group_id=KAFKA_GROUP_ID,
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            enable_auto_commit=True,
            auto_offset_reset='latest'
        )
        
        self.producer = AIOKafkaProducer(
            bootstrap_servers=KAFKA_BROKER,
            value_serializer=lambda v: json.dumps(v).encode('utf-8')
        )
        
        await self.consumer.start()
        await self.producer.start()
        
        # Initialize HTTP client for Firecrawl API
        self.http_client = httpx.AsyncClient(
            timeout=120.0,
            headers={
                'Authorization': f'Bearer {FIRECRAWL_API_KEY}',
                'Content-Type': 'application/json'
            }
        )
        
        logger.info("Firecrawl adapter started successfully")
        
    async def stop(self):
        """Cleanup resources"""
        logger.info("Stopping Firecrawl adapter...")
        
        if self.http_client:
            await self.http_client.aclose()
            
        if self.consumer:
            await self.consumer.stop()
            
        if self.producer:
            await self.producer.stop()
            
        logger.info("Firecrawl adapter stopped")
        
    async def process_messages(self):
        """Main message processing loop"""
        logger.info(f"Listening for messages on topic: {REQUEST_TOPIC}")
        
        async for message in self.consumer:
            try:
                request = message.value
                logger.info(f"Received request: {request.get('request_id')} - Action: {request.get('action')}")
                
                # Route to appropriate handler
                action = request.get('action', 'scrape')
                
                if action == 'scrape':
                    result = await self.handle_scrape(request)
                elif action == 'crawl':
                    result = await self.handle_crawl(request)
                elif action == 'extract':
                    result = await self.handle_extract_structured(request)
                else:
                    result = {
                        'success': False,
                        'error': f'Unknown action: {action}'
                    }
                
                # Send response
                await self.send_response(request, result)
                
            except Exception as e:
                logger.error(f"Error processing message: {str(e)}")
                logger.error(traceback.format_exc())
                
                await self.send_response(request, {
                    'success': False,
                    'error': str(e),
                    'traceback': traceback.format_exc()
                })
    
    async def handle_scrape(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle single page scraping using Firecrawl"""
        url = request.get('url')
        config = request.get('capture_config', {})
        
        logger.info(f"Scraping website: {url}")
        
        # Build Firecrawl scrape request
        firecrawl_payload = {
            'url': url,
            'formats': config.get('formats', ['markdown', 'html', 'screenshot']),
            'includeTags': config.get('include_tags', []),
            'excludeTags': config.get('exclude_tags', []),
            'onlyMainContent': config.get('only_main_content', False),
            'waitFor': config.get('wait_for', 0),
            'screenshot': config.get('capture_screenshot', True),
            'screenshotConfig': {
                'fullPage': config.get('full_page', True),
                'width': config.get('viewport', {}).get('width', 1920),
                'height': config.get('viewport', {}).get('height', 1080)
            }
        }
        
        try:
            # Call Firecrawl API
            response = await self.http_client.post(
                f"{FIRECRAWL_API_URL}/scrape",
                json=firecrawl_payload
            )
            
            if response.status_code == 200:
                data = response.json()
                
                return {
                    'success': True,
                    'url': url,
                    'title': data.get('metadata', {}).get('title'),
                    'description': data.get('metadata', {}).get('description'),
                    'markdown_content': data.get('markdown'),
                    'html_content': data.get('html'),
                    'screenshot_url': data.get('screenshot'),
                    'clean_content': data.get('content'),
                    'metadata': data.get('metadata', {}),
                    'links': data.get('links', []),
                    'captured_at': datetime.utcnow().isoformat()
                }
            else:
                return {
                    'success': False,
                    'error': f'Firecrawl API error: {response.status_code}',
                    'details': response.text
                }
                
        except Exception as e:
            logger.error(f"Firecrawl scrape error: {str(e)}")
            return {
                'success': False,
                'error': str(e)
            }
    
    async def handle_crawl(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle multi-page crawling using Firecrawl"""
        url = request.get('url')
        config = request.get('crawl_config', {})
        
        logger.info(f"Crawling website: {url}")
        
        # Build Firecrawl crawl request
        firecrawl_payload = {
            'url': url,
            'limit': config.get('limit', 10),
            'maxDepth': config.get('max_depth', 2),
            'excludePaths': config.get('exclude_paths', []),
            'includePaths': config.get('include_paths', []),
            'allowBackwardLinks': config.get('allow_backward_links', False),
            'formats': config.get('formats', ['markdown', 'html'])
        }
        
        try:
            # Start crawl job
            response = await self.http_client.post(
                f"{FIRECRAWL_API_URL}/crawl",
                json=firecrawl_payload
            )
            
            if response.status_code == 200:
                crawl_data = response.json()
                job_id = crawl_data.get('id')
                
                # Poll for completion (simplified - in production use webhooks)
                max_attempts = 60
                for _ in range(max_attempts):
                    await asyncio.sleep(5)
                    
                    status_response = await self.http_client.get(
                        f"{FIRECRAWL_API_URL}/crawl/{job_id}"
                    )
                    
                    if status_response.status_code == 200:
                        status_data = status_response.json()
                        
                        if status_data.get('status') == 'completed':
                            return {
                                'success': True,
                                'job_id': job_id,
                                'pages': status_data.get('data', []),
                                'total_pages': len(status_data.get('data', [])),
                                'completed_at': datetime.utcnow().isoformat()
                            }
                        elif status_data.get('status') == 'failed':
                            return {
                                'success': False,
                                'error': 'Crawl job failed',
                                'job_id': job_id
                            }
                
                return {
                    'success': False,
                    'error': 'Crawl job timeout',
                    'job_id': job_id
                }
            else:
                return {
                    'success': False,
                    'error': f'Firecrawl API error: {response.status_code}',
                    'details': response.text
                }
                
        except Exception as e:
            logger.error(f"Firecrawl crawl error: {str(e)}")
            return {
                'success': False,
                'error': str(e)
            }
    
    async def handle_extract_structured(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Extract structured data using Firecrawl's LLM extraction"""
        url = request.get('url')
        schema = request.get('schema', {})
        
        logger.info(f"Extracting structured data from: {url}")
        
        # Build extraction request
        firecrawl_payload = {
            'url': url,
            'formats': ['extract'],
            'extract': {
                'schema': schema,
                'systemPrompt': request.get('system_prompt', 
                    "Extract the requested information from the webpage content"),
                'prompt': request.get('prompt', "")
            }
        }
        
        try:
            response = await self.http_client.post(
                f"{FIRECRAWL_API_URL}/scrape",
                json=firecrawl_payload
            )
            
            if response.status_code == 200:
                data = response.json()
                
                return {
                    'success': True,
                    'extracted_data': data.get('extract', {}),
                    'url': url,
                    'metadata': data.get('metadata', {}),
                    'extracted_at': datetime.utcnow().isoformat()
                }
            else:
                return {
                    'success': False,
                    'error': f'Firecrawl API error: {response.status_code}',
                    'details': response.text
                }
                
        except Exception as e:
            logger.error(f"Firecrawl extract error: {str(e)}")
            return {
                'success': False,
                'error': str(e)
            }
    
    async def send_response(self, request: Dict[str, Any], result: Dict[str, Any]):
        """Send response back to orchestrator"""
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
    adapter = FirecrawlAdapter()
    
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
```

### 2. Simplified Capture Actions for Firecrawl (Go)

```go
// internal/backend/agent-chassis/platform/orchestration/actions/firecrawl_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ============================================================================
// FIRECRAWL SCRAPE ACTION - Single page capture
// ============================================================================

// FirecrawlScrapeAction sends a scrape request to Firecrawl adapter
func FirecrawlScrapeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing FirecrawlScrapeAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	// Extract configuration
	config, ok := params.StepConfig.Config.(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
	}

	// Get URL from input data
	url := ""
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if targetURL, ok := inputData["target_url"].(string); ok {
			url = targetURL
		}
	}

	if url == "" {
		return nil, fmt.Errorf("target_url not found in input_data")
	}

	// Generate request ID
	requestID := uuid.New().String()
	adapterTopic := "system.adapter.firecrawl.requests"

	// Build scrape config
	captureConfig := map[string]interface{}{
		"formats":           []string{"markdown", "html", "screenshot"},
		"only_main_content": false,
		"capture_screenshot": true,
		"full_page":         true,
	}

	// Override with step config if provided
	if cc, ok := config["capture_config"].(map[string]interface{}); ok {
		for k, v := range cc {
			captureConfig[k] = v
		}
	}

	// Build request payload
	requestPayload := map[string]interface{}{
		"request_id":       requestID,
		"action":           "scrape",
		"url":              url,
		"capture_config":   captureConfig,
		"correlation_id":   params.ExecutionContext.CorrelationID,
		"orchestration_id": params.ExecutionContext.OrchestrationID,
		"reply_to_topic":   params.ExecutionContext.ResponsesTopic,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	params.Logger.Debug("Sending Firecrawl scrape request",
		zap.String("url", url),
		zap.String("request_id", requestID))

	// Send to Kafka
	err := params.Producer.SendMessage(ctx, adapterTopic, requestID, requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to send scrape request: %w", err)
	}

	// Return result indicating we're waiting for response
	return &CaptureSiteResult{
		Success:       true,
		RequestID:     requestID,
		TopicSentTo:   adapterTopic,
		AwaitResponse: true,
		CaptureMetadata: map[string]interface{}{
			"url":    url,
			"action": "scrape",
		},
	}, nil
}

// ============================================================================
// FIRECRAWL CRAWL ACTION - Multi-page crawl
// ============================================================================

// FirecrawlCrawlAction sends a crawl request for multiple pages
func FirecrawlCrawlAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing FirecrawlCrawlAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	config, ok := params.StepConfig.Config.(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
	}

	// Get URL
	url := ""
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if targetURL, ok := inputData["target_url"].(string); ok {
			url = targetURL
		}
	}

	if url == "" {
		return nil, fmt.Errorf("target_url not found")
	}

	requestID := uuid.New().String()
	adapterTopic := "system.adapter.firecrawl.requests"

	// Build crawl config
	crawlConfig := map[string]interface{}{
		"limit":     10,
		"max_depth": 2,
		"formats":   []string{"markdown", "html"},
	}

	// Override with step config
	if cc, ok := config["crawl_config"].(map[string]interface{}); ok {
		for k, v := range cc {
			crawlConfig[k] = v
		}
	}

	requestPayload := map[string]interface{}{
		"request_id":       requestID,
		"action":           "crawl",
		"url":              url,
		"crawl_config":     crawlConfig,
		"correlation_id":   params.ExecutionContext.CorrelationID,
		"orchestration_id": params.ExecutionContext.OrchestrationID,
		"reply_to_topic":   params.ExecutionContext.ResponsesTopic,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	params.Logger.Debug("Sending Firecrawl crawl request",
		zap.String("url", url),
		zap.Any("config", crawlConfig))

	err := params.Producer.SendMessage(ctx, adapterTopic, requestID, requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to send crawl request: %w", err)
	}

	return &CaptureSiteResult{
		Success:       true,
		RequestID:     requestID,
		TopicSentTo:   adapterTopic,
		AwaitResponse: true,
		CaptureMetadata: map[string]interface{}{
			"url":    url,
			"action": "crawl",
			"config": crawlConfig,
		},
	}, nil
}

// ============================================================================
// FIRECRAWL EXTRACT ACTION - Structured data extraction
// ============================================================================

// FirecrawlExtractAction extracts structured data using LLM
func FirecrawlExtractAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing FirecrawlExtractAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	config, ok := params.StepConfig.Config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing config for extract action")
	}

	// Get URL
	url := ""
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if targetURL, ok := inputData["target_url"].(string); ok {
			url = targetURL
		}
	}

	// Get schema from config
	schema, ok := config["schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema not found in config")
	}

	requestID := uuid.New().String()
	adapterTopic := "system.adapter.firecrawl.requests"

	requestPayload := map[string]interface{}{
		"request_id":       requestID,
		"action":           "extract",
		"url":              url,
		"schema":           schema,
		"system_prompt":    config["system_prompt"],
		"prompt":           config["prompt"],
		"correlation_id":   params.ExecutionContext.CorrelationID,
		"orchestration_id": params.ExecutionContext.OrchestrationID,
		"reply_to_topic":   params.ExecutionContext.ResponsesTopic,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	params.Logger.Debug("Sending Firecrawl extract request",
		zap.String("url", url),
		zap.Any("schema", schema))

	err := params.Producer.SendMessage(ctx, adapterTopic, requestID, requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to send extract request: %w", err)
	}

	return &CaptureSiteResult{
		Success:       true,
		RequestID:     requestID,
		TopicSentTo:   adapterTopic,
		AwaitResponse: true,
		CaptureMetadata: map[string]interface{}{
			"url":    url,
			"action": "extract",
		},
	}, nil
}
```

### 3. Simplified Website Capture Agent for Firecrawl

```sql
-- Simplified Website Capture Agent using Firecrawl
INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag
) VALUES (
    gen_random_uuid(),
    'website-capture-firecrawl',
    'Website Capture Agent (Firecrawl)',
    'Captures website content using Firecrawl API',
    'data-extraction',
    '{
        "workflow": {
            "start_step": "validate_input",
            "steps": {
                "validate_input": {
                    "action": "validate_url",
                    "description": "Validate and normalize URL",
                    "config": {
                        "url_field": "target_url",
                        "add_protocol_if_missing": true
                    },
                    "next_step": "scrape_main_page"
                },
                "scrape_main_page": {
                    "action": "firecrawl_scrape",
                    "description": "Scrape main page content",
                    "config": {
                        "capture_config": {
                            "formats": ["markdown", "html", "screenshot"],
                            "capture_screenshot": true,
                            "full_page": true,
                            "only_main_content": false
                        }
                    },
                    "next_step": "extract_structure"
                },
                "extract_structure": {
                    "action": "firecrawl_extract",
                    "description": "Extract page structure and components",
                    "config": {
                        "schema": {
                            "navigation": {"type": "array", "items": {"type": "string"}},
                            "hero_section": {"type": "object"},
                            "main_sections": {"type": "array"},
                            "footer_links": {"type": "array"},
                            "color_scheme": {"type": "object"},
                            "fonts_used": {"type": "array"}
                        },
                        "system_prompt": "Analyze the webpage and extract its structural components, design elements, and content organization"
                    },
                    "next_step": "crawl_subpages"
                },
                "crawl_subpages": {
                    "action": "firecrawl_crawl",
                    "description": "Crawl additional pages for context",
                    "config": {
                        "crawl_config": {
                            "limit": 5,
                            "max_depth": 1,
                            "formats": ["markdown", "html"]
                        }
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return captured data"
                }
            }
        },
        "processing_mode": "task",
        "adapter_topic": "system.adapter.firecrawl.requests",
        "timeout_seconds": 180
    }'::jsonb,
    true,
    ARRAY['capture', 'firecrawl', 'scraping'],
    'docker.io/aqls/agent-chassis',
    'v1.0.407'
);
```

### 4. Docker Setup for Firecrawl Adapter

```dockerfile
# Dockerfile.firecrawl
FROM python:3.10-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy adapter code
COPY firecrawl_adapter.py .

# Run adapter
CMD ["python", "firecrawl_adapter.py"]
```

### 5. Requirements for Firecrawl Adapter

```txt
aiokafka==0.10.0
httpx==0.25.2
```

### 6. Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: firecrawl-adapter
  namespace: agent-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: firecrawl-adapter
  template:
    metadata:
      labels:
        app: firecrawl-adapter
    spec:
      containers:
      - name: firecrawl-adapter
        image: your-registry/firecrawl-adapter:latest
        env:
        - name: KAFKA_BROKER
          value: "kafka:9092"
        - name: REQUEST_TOPIC
          value: "system.adapter.firecrawl.requests"
        - name: FIRECRAWL_API_KEY
          valueFrom:
            secretKeyRef:
              name: firecrawl-credentials
              key: api-key
        - name: FIRECRAWL_API_URL
          value: "https://api.firecrawl.dev/v1"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 7. Quick Test Script

```python
#!/usr/bin/env python3
"""Test Firecrawl Adapter"""

import asyncio
import json
import uuid
from datetime import datetime
from aiokafka import AIOKafkaProducer, AIOKafkaConsumer


async def test_firecrawl_scrape(url: str):
    """Test Firecrawl scraping"""
    
    request_id = str(uuid.uuid4())
    response_topic = f"test.responses.{request_id[:8]}"
    
    producer = AIOKafkaProducer(
        bootstrap_servers='localhost:9092',
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )
    
    consumer = AIOKafkaConsumer(
        response_topic,
        bootstrap_servers='localhost:9092',
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        auto_offset_reset='earliest'
    )
    
    await producer.start()
    await consumer.start()
    
    try:
        # Send scrape request
        request = {
            'request_id': request_id,
            'action': 'scrape',
            'url': url,
            'capture_config': {
                'formats': ['markdown', 'html', 'screenshot'],
                'capture_screenshot': True,
                'full_page': True
            },
            'reply_to_topic': response_topic,
            'timestamp': datetime.utcnow().isoformat()
        }
        
        print(f"Sending Firecrawl scrape request for: {url}")
        await producer.send('system.adapter.firecrawl.requests', value=request)
        
        # Wait for response
        print("Waiting for response...")
        async for msg in consumer:
            response = msg.value
            if response.get('request_id') == request_id:
                result = response.get('result', {})
                if result.get('success'):
                    print("✓ Scrape successful!")
                    print(f"  Title: {result.get('title')}")
                    print(f"  Has markdown: {'markdown_content' in result}")
                    print(f"  Has HTML: {'html_content' in result}")
                    print(f"  Has screenshot: {'screenshot_url' in result}")
                    print(f"  Links found: {len(result.get('links', []))}")
                else:
                    print(f"✗ Scrape failed: {result.get('error')}")
                break
                
    finally:
        await producer.stop()
        await consumer.stop()


if __name__ == '__main__':
    url = input("Enter URL to test: ").strip()
    if not url.startswith(('http://', 'https://')):
        url = 'https://' + url
    
    asyncio.run(test_firecrawl_scrape(url))
```

### 8. Integration with Coordinator

Add these cases to your coordinator's executeStep:

```go
// In coordinator.go executeStep function
case "firecrawl_scrape":
    result, err = actions.FirecrawlScrapeAction(ctx, params)
    
case "firecrawl_crawl":
    result, err = actions.FirecrawlCrawlAction(ctx, params)
    
case "firecrawl_extract":
    result, err = actions.FirecrawlExtractAction(ctx, params)
```

## Benefits of Using Firecrawl MVP

1. **Simpler Implementation**:
    - No need to manage Playwright browsers
    - Handles JavaScript rendering automatically
    - Built-in screenshot capabilities
    - Clean markdown extraction

2. **Reduced Infrastructure**:
    - Lightweight adapter (just HTTP calls)
    - Less memory/CPU usage
    - No browser management

3. **Built-in Features**:
    - Automatic content cleaning
    - Structured data extraction with LLM
    - Multi-page crawling
    - Format conversion (HTML → Markdown)

4. **Quick to Deploy**:
    - Single API key setup
    - Minimal dependencies
    - Works immediately

## Migration Path

When you're ready to move beyond the MVP:

1. **Keep Both Adapters**: Run Firecrawl and Playwright in parallel
2. **Route by Complexity**: Use Firecrawl for simple captures, Playwright for complex interactions
3. **Gradual Migration**: Move specific features to Playwright as needed
4. **Fallback Strategy**: Use Firecrawl as fallback when Playwright fails

This MVP approach lets you start capturing websites immediately while building the more complex components (vision analysis, synthesis, content generation) in parallel. The Kafka-based architecture means you can swap adapters without changing the rest of your system.

