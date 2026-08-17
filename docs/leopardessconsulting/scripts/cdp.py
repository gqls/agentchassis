"""Minimal CDP driver over a hand-rolled WebSocket client (stdlib only).
No node, no puppeteer, no websocket-client on this box."""
import base64, json, os, socket, struct, subprocess, sys, time, urllib.request

def http_json(port, path):
    with urllib.request.urlopen(f"http://127.0.0.1:{port}{path}", timeout=10) as r:
        return json.load(r)

class WS:
    def __init__(self, url):
        # ws://127.0.0.1:PORT/devtools/page/ID
        rest = url[len("ws://"):]
        hostport, _, path = rest.partition("/")
        host, _, port = hostport.partition(":")
        self.s = socket.create_connection((host, int(port)), timeout=30)
        key = base64.b64encode(os.urandom(16)).decode()
        req = (f"GET /{path} HTTP/1.1\r\nHost: {hostport}\r\nUpgrade: websocket\r\n"
               f"Connection: Upgrade\r\nSec-WebSocket-Key: {key}\r\n"
               f"Sec-WebSocket-Version: 13\r\n\r\n")
        self.s.sendall(req.encode())
        buf = b""
        while b"\r\n\r\n" not in buf:
            buf += self.s.recv(4096)
        assert b"101" in buf.split(b"\r\n")[0], buf[:200]
        self.buf = buf.split(b"\r\n\r\n", 1)[1]
        self.id = 0

    def _recv_exact(self, n):
        while len(self.buf) < n:
            chunk = self.s.recv(65536)
            if not chunk:
                raise EOFError("socket closed")
            self.buf += chunk
        out, self.buf = self.buf[:n], self.buf[n:]
        return out

    def send(self, method, params=None, session=None):
        self.id += 1
        msg = {"id": self.id, "method": method, "params": params or {}}
        if session:
            msg["sessionId"] = session
        payload = json.dumps(msg).encode()
        mask = os.urandom(4)
        n = len(payload)
        hdr = b"\x81"
        if n < 126:
            hdr += struct.pack("!B", 0x80 | n)
        elif n < 65536:
            hdr += struct.pack("!BH", 0x80 | 126, n)
        else:
            hdr += struct.pack("!BQ", 0x80 | 127, n)
        masked = bytes(b ^ mask[i % 4] for i, b in enumerate(payload))
        self.s.sendall(hdr + mask + masked)
        return self.id

    def recv(self):
        b0, b1 = self._recv_exact(2)
        n = b1 & 0x7F
        if n == 126:
            n = struct.unpack("!H", self._recv_exact(2))[0]
        elif n == 127:
            n = struct.unpack("!Q", self._recv_exact(8))[0]
        return json.loads(self._recv_exact(n).decode())

    def call(self, method, params=None, session=None):
        want = self.send(method, params, session)
        deadline = time.time() + 60
        while time.time() < deadline:
            m = self.recv()
            if m.get("id") == want:
                if "error" in m:
                    raise RuntimeError(f"{method}: {m['error']}")
                return m.get("result", {})
        raise TimeoutError(method)

def evaluate(ws, expr):
    r = ws.call("Runtime.evaluate", {"expression": expr, "returnByValue": True,
                                     "awaitPromise": True})
    if r.get("exceptionDetails"):
        raise RuntimeError(r["exceptionDetails"].get("text", "js error") + " :: " +
                           json.dumps(r["exceptionDetails"].get("exception", {}))[:300])
    return r["result"].get("value")
