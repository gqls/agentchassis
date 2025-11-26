# 🎨 VISUAL FLOW DIAGRAM

## Before: Firecrawl v1 + No Image Storage

```
┌─────────────────┐
│   Agent Sends   │
│   Scrape Req    │
└────────┬────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Webscrape Adapter (v1)                  │
│  ┌───────────────────────────────────────────┐  │
│  │  Send to Firecrawl v1:                    │  │
│  │  {                                        │  │
│  │    "url": "example.com",                  │  │
│  │    "includeRawHtml": true,  ❌            │  │
│  │    "screenshot": true,      ❌            │  │
│  │    "screenshotConfig": {...}❌            │  │
│  │  }                                        │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
    ❌ ERROR ❌
  "Bad Request"
  "Unrecognized keys"

─────────────────────────────────────────────────────

IF we used correct v1 format:

┌─────────────────┐
│   Agent Sends   │
│   Scrape Req    │
└────────┬────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Webscrape Adapter (v1)                  │
│  ┌───────────────────────────────────────────┐  │
│  │  Send to Firecrawl v1                     │  │
│  │  ✅ Request succeeds                       │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│            Firecrawl v1 API                     │
│  Scrapes page → Stores in Google Cloud         │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│              Response                           │
│  {                                              │
│    "markdown": "...",                           │
│    "html": "...",                               │
│    "screenshot": "https://storage.googleapis.   │
│                   com/firecrawl.../xyz.png"     │
│  }                                              │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Webscrape Adapter                       │
│  ┌───────────────────────────────────────────┐  │
│  │  ❌ Screenshot stays in Google Cloud      │  │
│  │  ❌ No image extraction                   │  │
│  │  ❌ Only HTML/MD stored in S3             │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│              S3 Bucket                          │
│  webscrape/client/date/id/                     │
│    ├── content.html                             │
│    ├── content.md                               │
│    └── metadata.json                            │
│                                                  │
│  ❌ No screenshot.png                           │
│  ❌ No images/                                  │
└─────────────────────────────────────────────────┘

Problems:
  ❌ Screenshot in Google Cloud (expires in 30 days)
  ❌ No control over screenshot storage
  ❌ Images not extracted or stored
  ❌ Incomplete data ownership
```

---

## After: Firecrawl v2 + Full S3 Storage

```
┌─────────────────┐
│   Agent Sends   │
│   Scrape Req    │
└────────┬────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Webscrape Adapter (v2)                  │
│  ┌───────────────────────────────────────────┐  │
│  │  Send to Firecrawl v2:                    │  │
│  │  {                                        │  │
│  │    "url": "example.com",                  │  │
│  │    "formats": [                           │  │
│  │      "markdown",                          │  │
│  │      "html",                              │  │
│  │      "rawHtml",                           │  │
│  │      "links",  ✅ NEW                     │  │
│  │      {                                    │  │
│  │        "type": "screenshot",  ✅ NEW      │  │
│  │        "fullPage": true                   │  │
│  │      }                                    │  │
│  │    ]                                      │  │
│  │  }                                        │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│            Firecrawl v2 API                     │
│  Scrapes page → Stores in Google Cloud         │
│  Extracts images → Returns all URLs            │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│              Response                           │
│  {                                              │
│    "success": true,                             │
│    "data": {                                    │
│      "markdown": "...",                         │
│      "html": "...",                             │
│      "rawHtml": "...",                          │
│      "screenshot": "https://storage.            │
│                     googleapis.com/...",        │
│      "links": [                                 │
│        "https://example.com/image1.jpg",        │
│        "https://example.com/image2.png"         │
│      ],                                         │
│      "metadata": {...}                          │
│    }                                            │
│  }                                              │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Webscrape Adapter                       │
│  ┌───────────────────────────────────────────┐  │
│  │  Process Response:                        │  │
│  │  1. Extract screenshot URL                │  │
│  │  2. Extract image URLs from links         │  │
│  │  3. Filter images (skip data URLs, etc)   │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Download Phase                          │
│  ┌───────────────────────────────────────────┐  │
│  │  Download Screenshot:                     │  │
│  │  GET https://storage.googleapis.com/...   │  │
│  │  ✅ 234 KB downloaded                      │  │
│  │                                           │  │
│  │  Download Images (up to 50):              │  │
│  │  GET https://example.com/image1.jpg       │  │
│  │  ✅ 45 KB downloaded                       │  │
│  │  GET https://example.com/image2.png       │  │
│  │  ✅ 78 KB downloaded                       │  │
│  │  ... (repeat for all images)              │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Upload Phase                            │
│  ┌───────────────────────────────────────────┐  │
│  │  Upload to S3:                            │  │
│  │  ✅ content.html                           │  │
│  │  ✅ content.md                             │  │
│  │  ✅ raw.html                               │  │
│  │  ✅ screenshot.png       ← NEW!           │  │
│  │  ✅ images/image_000.jpg ← NEW!           │  │
│  │  ✅ images/image_001.png ← NEW!           │  │
│  │  ✅ images/image_002.jpg ← NEW!           │  │
│  │  ✅ metadata.json                          │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│              S3 Bucket                          │
│  webscrape/client/date/id/                     │
│    ├── content.html                             │
│    ├── content.md                               │
│    ├── raw.html                                 │
│    ├── screenshot.png        ✅ NEW!            │
│    ├── metadata.json                            │
│    └── images/               ✅ NEW!            │
│        ├── image_000.jpg                        │
│        ├── image_001.png                        │
│        ├── image_002.jpg                        │
│        └── ...                                  │
└─────────────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Generate Presigned URLs                 │
│  ┌───────────────────────────────────────────┐  │
│  │  Create 7-day URLs for all files:         │  │
│  │  ✅ https://s3...backblazeb2.com/...html  │  │
│  │  ✅ https://s3...backblazeb2.com/...md    │  │
│  │  ✅ https://s3...backblazeb2.com/...png   │  │
│  │  ✅ https://s3...backblazeb2.com/...jpg   │  │
│  └───────────────────────────────────────────┘  │
└────────┬────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────────────┐
│         Send Response to Agent                  │
│  {                                              │
│    "url": "https://example.com",                │
│    "captured_at": "2025-11-12T19:45:23Z",      │
│    "markdown_content": "...",                   │
│    "html_content": "...",                       │
│    "screenshot_url": "https://storage.google    │
│                       apis.com/...",  (original)│
│    "images": [                                  │
│      {"url": "https://example.com/image1.jpg"}, │
│      {"url": "https://example.com/image2.png"}  │
│    ],                                           │
│    "storage": {             ✅ NEW!             │
│      "screenshot_uri": "s3://bucket/.../        │
│                        screenshot.png",         │
│      "screenshot_url": "https://s3...back       │
│                        blazeb2.com/...",        │
│      "screenshot_source": "url",                │
│      "images": [                                │
│        {                                        │
│          "index": 0,                            │
│          "original_url": "https://example.      │
│                          com/image1.jpg",       │
│          "s3_uri": "s3://bucket/.../images/     │
│                    image_000.jpg",              │
│          "s3_url": "https://s3...back           │
│                    blazeb2.com/...",            │
│          "size_bytes": 45678                    │
│        }                                        │
│      ],                                         │
│      "images_uploaded_count": 10,               │
│      "images_failed_count": 2                   │
│    }                                            │
│  }                                              │
└─────────────────────────────────────────────────┘

Benefits:
  ✅ Screenshot in YOUR S3 (permanent)
  ✅ All images in YOUR S3 (permanent)
  ✅ Complete data ownership
  ✅ Consistent storage structure
  ✅ Cost-effective (~$0.004/scrape)
  ✅ Presigned URLs for easy access
```

---

## Data Flow Comparison

### v1 (Before)
```
Agent → Adapter → Firecrawl v1 API
                      ↓
                Google Cloud Storage (screenshot)
                      ↓
          Adapter ← Response (with Google URL)
                      ↓
                  Your S3 (HTML/MD only)
                      ↓
          Agent ← Response (incomplete)
```

**Problems:**
- Screenshot expires in 30 days
- No image extraction
- Split storage (Google + S3)

### v2 (After)
```
Agent → Adapter → Firecrawl v2 API
                      ↓
                Google Cloud Storage (temp)
                      ↓
          Adapter ← Response (with Google URLs)
                      ↓
          Downloads screenshot & images
                      ↓
          Uploads ALL to Your S3
                      ↓
          Agent ← Complete response
```

**Benefits:**
- Everything in YOUR S3
- Permanent storage
- Complete control
- Organized structure

---

## Cost Flow

### Before
```
Firecrawl API: $X per scrape
              ↓
Google Storage: FREE (Firecrawl's cost)
              ↓
Your S3: ~$0.001 per scrape (HTML/MD only)
              ↓
Total: $X + $0.001
```

### After
```
Firecrawl API: $X per scrape (same)
              ↓
Google Storage: FREE (Firecrawl's cost, temporary)
              ↓
Download: FREE (from Google)
              ↓
Upload: FREE (to S3)
              ↓
Your S3: ~$0.004 per scrape (everything)
              ↓
Total: $X + $0.004

Additional cost: +$0.003 per scrape
For 1000 scrapes/month: +$3.00/month
```

**Worth it?** YES! You get:
- Permanent storage
- Complete ownership
- No expiration
- Organized structure
- All images included

---

## Timeline Comparison

### v1 Flow
```
0ms    → Request sent to Firecrawl
2000ms → Firecrawl scrapes page
        → Screenshots stored in Google Cloud
        → Response returned
2100ms → HTML/MD uploaded to S3
2200ms → Response sent to agent

Total: ~2.2 seconds
```

### v2 Flow
```
0ms    → Request sent to Firecrawl
2000ms → Firecrawl scrapes page
        → Screenshots stored in Google Cloud
        → Response returned
2100ms → Download screenshot (234 KB)
2600ms → Upload screenshot to S3
2700ms → Download 10 images (500 KB total)
3500ms → Upload 10 images to S3
3800ms → Upload HTML/MD to S3
3900ms → Response sent to agent

Total: ~3.9 seconds

Additional time: +1.7 seconds
```

**Impact:** Minimal - adds ~1-2 seconds per scrape

---

## Error Handling Flow

### v1
```
Firecrawl Error → Adapter catches
                       ↓
                Create error response
                       ↓
                ❌ Missing headers
                       ↓
                Validation fails
                       ↓
                ❌ Error response rejected
                       ↓
                Agent times out
```

### v2
```
Firecrawl Error → Adapter catches
                       ↓
                Create error response
                       ↓
                ✅ All headers included:
                   - client_id
                   - sender_agent_type
                   - in_response_to_step_name
                       ↓
                Validation passes
                       ↓
                ✅ Error response delivered
                       ↓
                Agent receives error properly
                       ↓
                Can retry or handle gracefully
```

---

## Summary

| Aspect | Before (v1) | After (v2) | Change |
|--------|------------|-----------|--------|
| **API** | v1 (deprecated) | v2 (current) | ✅ Upgraded |
| **Screenshot** | Google Cloud | Your S3 | ✅ Migrated |
| **Images** | Not extracted | Extracted + stored | ✅ Added |
| **Storage** | Split (Google+S3) | All in S3 | ✅ Unified |
| **Cost/scrape** | $X + $0.001 | $X + $0.004 | +$0.003 |
| **Time/scrape** | ~2.2s | ~3.9s | +1.7s |
| **Ownership** | Partial | Complete | ✅ Full |
| **Expiration** | 30 days | Never | ✅ Permanent |
| **Errors** | Fail validation | Pass validation | ✅ Fixed |

**Result:** Better architecture, complete ownership, minimal cost increase