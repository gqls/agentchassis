const B2_BUCKET_NAME = 'portfolio-sites';

export default {
    async fetch(request, env, ctx) {
        const url = new URL(request.url);
        const hostname = url.hostname;

        let path = url.pathname;
        if (path === '/' || path === '') {
            path = '/index.html';
        }

        if (url.pathname === '/worker-health') {
            return new Response('Worker is running!', { status: 200 });
        }

        if (url.searchParams.get('debug') === '1') {
            return new Response(JSON.stringify({
                hostname: hostname,
                path: path,
                hasCredentials: !!(env.B2_KEY_ID && env.B2_APP_KEY)
            }, null, 2), {
                headers: { 'Content-Type': 'application/json' }
            });
        }

        const objectKey = `${hostname}${path}`;

        try {
            const b2Response = await fetchFromB2(env, objectKey);

            if (!b2Response.ok) {
                return new Response('Not found', { status: 404 });
            }

            const response = new Response(b2Response.body, b2Response);
            response.headers.set('Cache-Control', 'public, max-age=3600');
            return response;

        } catch (err) {
            return new Response('Server error', { status: 500 });
        }
    }
};

async function fetchFromB2(env, objectKey) {
    const region = 'us-east-005';
    const bucket = B2_BUCKET_NAME;
    const host = `s3.${region}.backblazeb2.com`;
    const endpoint = `https://${host}`;

    const encodedKey = objectKey.split('/').map(encodeURIComponent).join('/');
    const url = `${endpoint}/${bucket}/${encodedKey}`;

    const method = 'GET';
    const datetime = new Date().toISOString().replace(/[:-]|\.\d{3}/g, '');
    const date = datetime.slice(0, 8);

    const canonicalUri = `/${bucket}/${encodedKey}`;
    const canonicalHeaders = `host:${host}\nx-amz-content-sha256:UNSIGNED-PAYLOAD\nx-amz-date:${datetime}\n`;
    const signedHeaders = 'host;x-amz-content-sha256;x-amz-date';

    const canonicalRequest = [method, canonicalUri, '', canonicalHeaders, signedHeaders, 'UNSIGNED-PAYLOAD'].join('\n');

    const algorithm = 'AWS4-HMAC-SHA256';
    const credentialScope = `${date}/${region}/s3/aws4_request`;
    const canonicalRequestHash = await sha256Hex(canonicalRequest);
    const stringToSign = [algorithm, datetime, credentialScope, canonicalRequestHash].join('\n');

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
    return await hmac(kService, 'aws4_request');
}

function bufferToHex(buffer) {
    return [...new Uint8Array(buffer)].map(b => b.toString(16).padStart(2, '0')).join('');
}