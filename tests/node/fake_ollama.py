import contextlib
import json
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


class FakeOllama:
    """Records request bodies; `delay` slows /api/chat; `busy_until` delays probes."""

    def __init__(self):
        self.bodies = []
        self.delay = 0.0
        self.busy_until = 0.0
        self.loaded = ["model-a"]
        self.status = 200
        self.raw_body = None
        self.ps_raw = None
        fake = self

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                if fake.ps_raw is not None:
                    self._reply_raw(200, fake.ps_raw)
                    return
                self._reply({"models": [{"name": m} for m in fake.loaded]})

            def do_POST(self):
                body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                fake.bodies.append(body)
                if body.get("keep_alive") == 0:
                    fake.loaded = []
                    fake.busy_until = 0.0
                    self._reply({"done": True})
                    return
                probe = body.get("options", {}).get("num_predict") == 1
                wait = max(0.0, fake.busy_until - time.time()) if probe else fake.delay
                time.sleep(wait)
                if fake.raw_body is not None:
                    self._reply_raw(fake.status, fake.raw_body)
                    return
                self._reply(
                    {
                        "message": {"content": '{"label": "x"}'},
                        "prompt_eval_count": 7,
                        "eval_count": 3,
                    }
                )

            def _reply(self, payload):
                self._reply_raw(200, json.dumps(payload).encode())

            def _reply_raw(self, status, data):
                self.send_response(status)
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                with contextlib.suppress(BrokenPipeError, ConnectionResetError):
                    self.wfile.write(data)

            def log_message(self, *args):
                pass

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.server.daemon_threads = True
        self._thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self._thread.start()
        self.url = f"http://127.0.0.1:{self.server.server_address[1]}"

    def close(self):
        self.server.shutdown()
        self.server.server_close()
        self._thread.join(timeout=2)
