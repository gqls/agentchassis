const B2_S3_ENDPOINT = 'https://s3.us-east-005.backblazeb2.com';  
const B2_BUCKET_NAME = 'portfolio-sites';

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const hostname = url.hostname;
    
    let path = url.pathname;
    if (path === '/' || path === '') {
      path = '/index.html';
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
        const errorText = await b2Response.text();
        return new Response(JSON.stringify({
          error: 'B2 returned error',
          objectKey: objectKey,
          status: b2Response.status,
          statusText: b2Response.statusText,
          body: errorText.slice(0, 500)
        }, null, 2), { 
          status: 404,
          headers: { 'Content-Type': 'application/json' }
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