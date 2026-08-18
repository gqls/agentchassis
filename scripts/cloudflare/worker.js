const B2_S3_ENDPOINT = 'https://s3.us-east-005.backblazeb2.com';  
const B2_BUCKET_NAME = 'portfolio-sites';

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const hostname = url.hostname;
    
    let path = url.pathname;
    if (path === '/' || path === '') {
      path = '/index.html';
    } else if (path.endsWith('/')) {
      // Directory-form URLs (/guides/, /tools/x/) — an object store has no
      // directory index, so without this the key is literally "guides/" and
      // every such address is a live 404, while /guides/index.html serves 200.
      // The git-hosted route (sites with sites.github_repo set) has always
      // served both forms; this is what makes the two routes agree.
      //
      // It can only turn a 404 into a 200: verified 2026-08-18 that ZERO keys
      // in the bucket end in "/", so no object that serves today stops serving.
      //
      // Deliberately NOT extended to the slashless form (/guides -> /guides/index.html):
      // that would mask genuine 404s for missing .html files, which is a
      // different decision and was not the one taken. See the concept register
      // (delivery.md, DGH-0xx) and LANDMINES.md's "/section/ URL 404s" entry.
      path += 'index.html';
    }

    // Health check
    if (url.pathname === '/worker-health') {
      return new Response('Worker is running!', { status: 200 });
    }
    
    // Debug endpoint
    if (url.searchParams.get('debug') === '1') {
      return new Response(JSON.stringify({
        hostname: hostname,
        path: path,
        hasCredentials: !!(env.B2_KEY_ID && env.B2_APP_KEY),
        b2Endpoint: B2_S3_ENDPOINT,
        bucket: B2_BUCKET_NAME
      }, null, 2), {
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // Build the S3 object key
    const objectKey = `${hostname}${path}`;
    
    try {
      const b2Response = await fetchFromB2(env, objectKey);
      
      if (!b2Response.ok) {
        // bugs_open/132: a missing path used to return the raw B2 error JSON,
        // leaking the bucket objectKey to the visitor. Serve the site's own
        // 404 page instead. Status MUST stay 404 — a soft-404 at 200 hides
        // every broken link from crawlers and from any link checker.
        if (b2Response.status === 404) {
          const fallback = await fetchFromB2(env, `${hostname}/404.html`);
          if (fallback.ok) {
            return new Response(fallback.body, {
              status: 404,
              headers: {
                'Content-Type': 'text/html; charset=utf-8',
                'Cache-Control': 'public, max-age=300'
              }
            });
          }
        }
        // No 404.html in the bucket, or a non-404 origin error: plain text,
        // same 404 status as before, nothing internal in the body.
        return new Response('Not found', {
          status: 404,
          headers: { 'Content-Type': 'text/plain; charset=utf-8' }
        });
      }
      
      const response = new Response(b2Response.body, b2Response);
      response.headers.set('Cache-Control', 'public, max-age=3600');
      return response;
      
    } catch (err) {
      // This will show us what's actually failing
      return new Response(JSON.stringify({
        error: 'Worker exception',
        message: err.message,
        stack: err.stack,
        objectKey: objectKey
      }, null, 2), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }
};

async function fetchFromB2(env, objectKey) {
  const region = 'us-east-005';  
  const bucket = B2_BUCKET_NAME;
  
  // Define these FIRST before using them
  const host = `s3.${region}.backblazeb2.com`;
  const endpoint = `https://${host}`;
  
  // URL encode the object key (important for dots in domain names)
  const encodedKey = objectKey.split('/').map(encodeURIComponent).join('/');
  const url = `${endpoint}/${bucket}/${encodedKey}`;
  
  const method = 'GET';
  const datetime = new Date().toISOString().replace(/[:-]|\.\d{3}/g, '');
  const date = datetime.slice(0, 8);
  
  const canonicalUri = `/${bucket}/${encodedKey}`;
  const canonicalQueryString = '';
  const canonicalHeaders = `host:${host}\nx-amz-content-sha256:UNSIGNED-PAYLOAD\nx-amz-date:${datetime}\n`;
  const signedHeaders = 'host;x-amz-content-sha256;x-amz-date';
  const payloadHash = 'UNSIGNED-PAYLOAD';
  
  const canonicalRequest = [
    method,
    canonicalUri,
    canonicalQueryString,
    canonicalHeaders,
    signedHeaders,
    payloadHash
  ].join('\n');
  
  const algorithm = 'AWS4-HMAC-SHA256';
  const credentialScope = `${date}/${region}/s3/aws4_request`;
  const canonicalRequestHash = await sha256Hex(canonicalRequest);
  
  const stringToSign = [
    algorithm,
    datetime,
    credentialScope,
    canonicalRequestHash
  ].join('\n');
  
  const signingKey = await getSignatureKey(env.B2_APP_KEY, date, region, 's3');
  const signature = await hmacHex(signingKey, stringToSign);
  
  const authorization = `${algorithm} Credential=${env.B2_KEY_ID}/${credentialScope}, SignedHeaders=${signedHeaders}, Signature=${signature}`;
  
  return fetch(url, {
    method: method,
    headers: {
      'Host': host,
      'x-amz-date': datetime,
      'x-amz-content-sha256': 'UNSIGNED-PAYLOAD',
      'Authorization': authorization
    }
  });
}

async function sha256Hex(message) {
  const msgBuffer = new TextEncoder().encode(message);
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
  return bufferToHex(hashBuffer);
}

async function hmac(key, message) {
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    typeof key === 'string' ? new TextEncoder().encode(key) : key,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );
  return await crypto.subtle.sign('HMAC', cryptoKey, new TextEncoder().encode(message));
}

async function hmacHex(key, message) {
  const sig = await hmac(key, message);
  return bufferToHex(sig);
}

async function getSignatureKey(key, dateStamp, region, service) {
  const kDate = await hmac('AWS4' + key, dateStamp);
  const kRegion = await hmac(kDate, region);
  const kService = await hmac(kRegion, service);
  const kSigning = await hmac(kService, 'aws4_request');
  return kSigning;
}

function bufferToHex(buffer) {
  return [...new Uint8Array(buffer)]
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}